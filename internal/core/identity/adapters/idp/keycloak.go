package idp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kleff/go-common/domain"
	"github.com/kleff/platform/internal/core/identity/ports"
)

// KeycloakConfig holds configuration for the Keycloak adapter.
type KeycloakConfig struct {
	// BaseURL is the Keycloak server root, e.g. "http://keycloak:8080".
	BaseURL string
	// Realm is the Keycloak realm name, e.g. "kleff".
	Realm string
	// ClientID is the OIDC client used for the Direct Access Grant (ROPC).
	// Must have "Direct access grants" enabled in Keycloak.
	ClientID string
	// AdminUser is the Keycloak admin username (master realm).
	AdminUser string
	// AdminPassword is the Keycloak admin password.
	AdminPassword string
}

// KeycloakAdapter implements ports.IdentityProvider for Keycloak 24+.
type KeycloakAdapter struct {
	cfg    KeycloakConfig
	client *http.Client
}

// NewKeycloakAdapter creates a Keycloak identity provider adapter.
func NewKeycloakAdapter(cfg KeycloakConfig) *KeycloakAdapter {
	return &KeycloakAdapter{cfg: cfg, client: &http.Client{}}
}

// Login authenticates via the Direct Access Grant (Resource Owner Password Credentials).
func (a *KeycloakAdapter) Login(ctx context.Context, username, password string) (*ports.Token, error) {
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		strings.TrimRight(a.cfg.BaseURL, "/"), a.cfg.Realm)

	form := url.Values{
		"grant_type": {"password"},
		"client_id":  {a.cfg.ClientID},
		"username":   {username},
		"password":   {password},
		"scope":      {"openid"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL,
		strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("keycloak login: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("keycloak login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, domain.NewUnauthorized("invalid username or password")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("keycloak login: unexpected status %d: %s", resp.StatusCode, body)
	}

	var tok ports.Token
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, fmt.Errorf("keycloak login: decode response: %w", err)
	}
	return &tok, nil
}

// Register creates a user via the Keycloak Admin REST API.
// It performs three steps: obtain admin token → create user → set password.
func (a *KeycloakAdapter) Register(ctx context.Context, req ports.RegisterRequest) (string, error) {
	adminToken, err := a.adminToken(ctx)
	if err != nil {
		return "", fmt.Errorf("keycloak register: admin token: %w", err)
	}

	userID, err := a.createUser(ctx, adminToken, req)
	if err != nil {
		return "", err // already wrapped with domain error
	}

	if err := a.setPassword(ctx, adminToken, userID, req.Password); err != nil {
		return "", fmt.Errorf("keycloak register: set password: %w", err)
	}

	return userID, nil
}

// adminToken fetches a short-lived admin-cli token from the master realm.
func (a *KeycloakAdapter) adminToken(ctx context.Context) (string, error) {
	tokenURL := fmt.Sprintf("%s/realms/master/protocol/openid-connect/token",
		strings.TrimRight(a.cfg.BaseURL, "/"))

	form := url.Values{
		"grant_type": {"password"},
		"client_id":  {"admin-cli"},
		"username":   {a.cfg.AdminUser},
		"password":   {a.cfg.AdminPassword},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL,
		strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, body)
	}

	var tok struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return "", err
	}
	return tok.AccessToken, nil
}

// createUser calls POST /admin/realms/{realm}/users and returns the new user's Keycloak ID.
func (a *KeycloakAdapter) createUser(ctx context.Context, adminToken string, req ports.RegisterRequest) (string, error) {
	usersURL := fmt.Sprintf("%s/admin/realms/%s/users",
		strings.TrimRight(a.cfg.BaseURL, "/"), a.cfg.Realm)

	firstName := req.FirstName
	if firstName == "" {
		firstName = req.Username // Keycloak 24+ requires firstName
	}

	payload, err := json.Marshal(map[string]any{
		"username":        req.Username,
		"email":           req.Email,
		"firstName":       firstName,
		"lastName":        req.LastName,
		"enabled":         true,
		"emailVerified":   true,
		"requiredActions": []string{},
	})
	if err != nil {
		return "", fmt.Errorf("keycloak register: marshal user: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, usersURL,
		strings.NewReader(string(payload)))
	if err != nil {
		return "", fmt.Errorf("keycloak register: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("keycloak register: create user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return "", domain.NewConflict("a user with that username or email already exists")
	}
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		var kcErr struct {
			ErrorMessage string `json:"errorMessage"`
		}
		_ = json.Unmarshal(body, &kcErr)
		msg := kcErr.ErrorMessage
		if msg == "" {
			msg = fmt.Sprintf("create user failed (status %d)", resp.StatusCode)
		}
		return "", domain.NewBadRequest(msg)
	}

	// Keycloak returns the new user URL in the Location header: .../users/{id}
	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("keycloak register: no Location header in response")
	}
	parts := strings.Split(strings.TrimRight(location, "/"), "/")
	return parts[len(parts)-1], nil
}

// setPassword calls PUT /admin/realms/{realm}/users/{id}/reset-password.
func (a *KeycloakAdapter) setPassword(ctx context.Context, adminToken, userID, password string) error {
	pwURL := fmt.Sprintf("%s/admin/realms/%s/users/%s/reset-password",
		strings.TrimRight(a.cfg.BaseURL, "/"), a.cfg.Realm, userID)

	payload, _ := json.Marshal(map[string]any{
		"type":      "password",
		"value":     password,
		"temporary": false,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, pwURL,
		strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		var kcErr struct {
			ErrorMessage string `json:"errorMessage"`
		}
		_ = json.Unmarshal(body, &kcErr)
		msg := kcErr.ErrorMessage
		if msg == "" {
			msg = fmt.Sprintf("set password failed (status %d)", resp.StatusCode)
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}
