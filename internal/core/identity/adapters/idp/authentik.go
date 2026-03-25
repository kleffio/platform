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

// AuthentikConfig holds configuration for the Authentik adapter.
type AuthentikConfig struct {
	// BaseURL is the Authentik server root URL, e.g. "https://authentik.example.com".
	BaseURL string
	// ClientID is the OAuth2 application client ID with the Password grant enabled.
	ClientID string
	// ClientSecret is the OAuth2 application client secret.
	ClientSecret string
	// APIToken is a service account API token used for user creation via the
	// Authentik REST API. The token must have the "can_access_admin_interface"
	// permission or be scoped to user management.
	APIToken string
	// FlowSlug is the authentication flow slug used for the ROPC token endpoint.
	// Defaults to "default-authentication-flow".
	FlowSlug string
}

// AuthentikAdapter implements ports.IdentityProvider for Authentik.
type AuthentikAdapter struct {
	cfg    AuthentikConfig
	client *http.Client
}

// NewAuthentikAdapter creates an Authentik identity provider adapter.
func NewAuthentikAdapter(cfg AuthentikConfig) *AuthentikAdapter {
	if cfg.FlowSlug == "" {
		cfg.FlowSlug = "default-authentication-flow"
	}
	return &AuthentikAdapter{cfg: cfg, client: &http.Client{}}
}

func (a *AuthentikAdapter) baseURL() string {
	return strings.TrimRight(a.cfg.BaseURL, "/")
}

// Login authenticates via the Resource Owner Password Credentials grant.
// Authentik exposes this at /application/o/token/ for applications with the
// "Resource owner password" grant type enabled.
func (a *AuthentikAdapter) Login(ctx context.Context, username, password string) (*ports.Token, error) {
	form := strings.NewReader(
		"grant_type=password" +
			"&client_id=" + a.cfg.ClientID +
			"&client_secret=" + a.cfg.ClientSecret +
			"&username=" + username +
			"&password=" + password +
			"&scope=openid+profile+email+offline_access",
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.baseURL()+"/application/o/token/", form)
	if err != nil {
		return nil, fmt.Errorf("authentik login: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("authentik login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, domain.NewUnauthorized("invalid username or password")
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("authentik login: unexpected status %d: %s", resp.StatusCode, b)
	}

	var tok ports.Token
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, fmt.Errorf("authentik login: decode response: %w", err)
	}
	return &tok, nil
}

// Register creates a user via the Authentik Core Users API.
// Requires a service account API token with user management permissions.
func (a *AuthentikAdapter) Register(ctx context.Context, req ports.RegisterRequest) (string, error) {
	payload, _ := json.Marshal(map[string]any{
		"username":   req.Username,
		"email":      req.Email,
		"name":       strings.TrimSpace(req.FirstName + " " + req.LastName),
		"password":   req.Password,
		"is_active":  true,
		"attributes": map[string]any{},
	})

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.baseURL()+"/api/v3/core/users/", strings.NewReader(string(payload)))
	if err != nil {
		return "", fmt.Errorf("authentik register: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.cfg.APIToken)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("authentik register: %w", err)
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusBadRequest {
		var e map[string]any
		_ = json.Unmarshal(b, &e)
		if _, dup := e["username"]; dup {
			return "", domain.NewConflict("a user with that username already exists")
		}
		if _, dup := e["email"]; dup {
			return "", domain.NewConflict("a user with that email already exists")
		}
		return "", domain.NewBadRequest(fmt.Sprintf("create user failed: %s", b))
	}
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("authentik register: unexpected status %d: %s", resp.StatusCode, b)
	}

	var created struct {
		PK int `json:"pk"`
	}
	if err := json.Unmarshal(b, &created); err != nil {
		return "", fmt.Errorf("authentik register: decode response: %w", err)
	}
	return fmt.Sprintf("%d", created.PK), nil
}
