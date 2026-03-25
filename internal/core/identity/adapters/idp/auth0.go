package idp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kleff/go-common/domain"
	"github.com/kleff/platform/internal/core/identity/ports"
)

// Auth0Config holds configuration for the Auth0 adapter.
type Auth0Config struct {
	// Domain is the Auth0 tenant domain, e.g. "dev-xxxxx.us.auth0.com".
	Domain string
	// ClientID is the application client ID used for the password grant (ROPC).
	// The application must have the "Password" grant type enabled in Auth0.
	ClientID string
	// ClientSecret is the application client secret.
	ClientSecret string
	// Audience is the API audience/identifier, e.g. "https://api.example.com".
	Audience string
	// Connection is the Auth0 database connection name.
	// Defaults to "Username-Password-Authentication".
	Connection string
	// MgmtClientID is the Machine-to-Machine application client ID used to
	// obtain a Management API token for user creation.
	MgmtClientID string
	// MgmtClientSecret is the Management API application client secret.
	MgmtClientSecret string
}

// Auth0Adapter implements ports.IdentityProvider for Auth0.
type Auth0Adapter struct {
	cfg    Auth0Config
	client *http.Client
}

// NewAuth0Adapter creates an Auth0 identity provider adapter.
func NewAuth0Adapter(cfg Auth0Config) *Auth0Adapter {
	if cfg.Connection == "" {
		cfg.Connection = "Username-Password-Authentication"
	}
	return &Auth0Adapter{cfg: cfg, client: &http.Client{}}
}

func (a *Auth0Adapter) baseURL() string {
	return "https://" + strings.TrimRight(a.cfg.Domain, "/")
}

// Login authenticates via the Resource Owner Password Credentials grant.
// The Auth0 application must have the "Password" grant type enabled.
func (a *Auth0Adapter) Login(ctx context.Context, username, password string) (*ports.Token, error) {
	body, _ := json.Marshal(map[string]string{
		"grant_type":    "password",
		"client_id":     a.cfg.ClientID,
		"client_secret": a.cfg.ClientSecret,
		"username":      username,
		"password":      password,
		"audience":      a.cfg.Audience,
		"scope":         "openid profile email offline_access",
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.baseURL()+"/oauth/token", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("auth0 login: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth0 login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, domain.NewUnauthorized("invalid username or password")
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("auth0 login: unexpected status %d: %s", resp.StatusCode, b)
	}

	var tok ports.Token
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, fmt.Errorf("auth0 login: decode response: %w", err)
	}
	return &tok, nil
}

// Register creates a user via the Auth0 Management API.
// Requires a Machine-to-Machine application with the "create:users" permission.
func (a *Auth0Adapter) Register(ctx context.Context, req ports.RegisterRequest) (string, error) {
	mgmtToken, err := a.managementToken(ctx)
	if err != nil {
		return "", fmt.Errorf("auth0 register: management token: %w", err)
	}

	name := strings.TrimSpace(req.FirstName + " " + req.LastName)
	if name == "" {
		name = req.Username
	}
	payload, _ := json.Marshal(map[string]any{
		"connection":  a.cfg.Connection,
		"email":       req.Email,
		"password":    req.Password,
		"username":    req.Username,
		"given_name":  req.FirstName,
		"family_name": req.LastName,
		"name":        name,
	})

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.baseURL()+"/api/v2/users", strings.NewReader(string(payload)))
	if err != nil {
		return "", fmt.Errorf("auth0 register: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+mgmtToken)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("auth0 register: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return "", domain.NewConflict("a user with that username or email already exists")
	}
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		var e struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(b, &e)
		msg := e.Message
		if msg == "" {
			msg = fmt.Sprintf("create user failed (status %d)", resp.StatusCode)
		}
		return "", domain.NewBadRequest(msg)
	}

	var created struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return "", fmt.Errorf("auth0 register: decode response: %w", err)
	}
	return created.UserID, nil
}

// managementToken obtains a short-lived Management API token via client credentials.
func (a *Auth0Adapter) managementToken(ctx context.Context) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     a.cfg.MgmtClientID,
		"client_secret": a.cfg.MgmtClientSecret,
		"audience":      a.baseURL() + "/api/v2/",
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.baseURL()+"/oauth/token", strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, b)
	}

	var tok struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return "", err
	}
	return tok.AccessToken, nil
}
