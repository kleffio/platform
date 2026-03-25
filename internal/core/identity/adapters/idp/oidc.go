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

// OIDCConfig holds configuration for the generic OIDC adapter.
// This adapter supports any OIDC provider that implements the Resource Owner
// Password Credentials (ROPC) grant. Registration is not supported — users
// must be created directly in the identity provider.
type OIDCConfig struct {
	// Issuer is the OIDC issuer URL, e.g. "https://accounts.example.com".
	// Used to derive the token endpoint if TokenEndpoint is not set.
	Issuer string
	// TokenEndpoint is the full URL to the OAuth2 token endpoint.
	// If empty, derived from Issuer as "{issuer}/oauth2/token".
	TokenEndpoint string
	// ClientID is the OAuth2 client ID.
	ClientID string
	// ClientSecret is the OAuth2 client secret. May be empty for public clients.
	ClientSecret string
	// Scope is the requested OAuth2 scope.
	// Defaults to "openid profile email offline_access".
	Scope string
}

// OIDCAdapter implements ports.IdentityProvider for any generic OIDC provider.
// Only Login is supported; Register returns ErrNotSupported.
type OIDCAdapter struct {
	cfg    OIDCConfig
	client *http.Client
}

// NewOIDCAdapter creates a generic OIDC identity provider adapter.
func NewOIDCAdapter(cfg OIDCConfig) *OIDCAdapter {
	if cfg.Scope == "" {
		cfg.Scope = "openid profile email offline_access"
	}
	if cfg.TokenEndpoint == "" && cfg.Issuer != "" {
		cfg.TokenEndpoint = strings.TrimRight(cfg.Issuer, "/") + "/oauth2/token"
	}
	return &OIDCAdapter{cfg: cfg, client: &http.Client{}}
}

// Login authenticates via the Resource Owner Password Credentials grant.
func (a *OIDCAdapter) Login(ctx context.Context, username, password string) (*ports.Token, error) {
	form := "grant_type=password" +
		"&client_id=" + a.cfg.ClientID +
		"&username=" + username +
		"&password=" + password +
		"&scope=" + strings.ReplaceAll(a.cfg.Scope, " ", "+")

	if a.cfg.ClientSecret != "" {
		form += "&client_secret=" + a.cfg.ClientSecret
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.cfg.TokenEndpoint, strings.NewReader(form))
	if err != nil {
		return nil, fmt.Errorf("oidc login: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("oidc login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, domain.NewUnauthorized("invalid username or password")
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("oidc login: unexpected status %d: %s", resp.StatusCode, b)
	}

	var tok ports.Token
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, fmt.Errorf("oidc login: decode response: %w", err)
	}
	return &tok, nil
}

// Register is not supported by the generic OIDC adapter.
// Users must be provisioned directly in the identity provider.
func (a *OIDCAdapter) Register(_ context.Context, _ ports.RegisterRequest) (string, error) {
	return "", domain.NewBadRequest("user registration is not supported for this identity provider; create users directly in your IDP")
}
