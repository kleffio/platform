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

// OryConfig holds configuration for the Ory (Kratos + Hydra) adapter.
//
// Login flow:
//  1. Kratos native API authenticates the user and returns a session token.
//  2. If HydraAdminURL is provided the session is exchanged for a proper
//     OAuth2 JWT access token via the Hydra admin token-for-session endpoint.
//     Otherwise the Kratos session token is returned directly as access_token.
//
// Registration flow:
//
//	Kratos native registration API creates the identity.
type OryConfig struct {
	// KratosPublicURL is the root URL of the Kratos public API,
	// e.g. "https://kratos.example.com" or "http://kratos:4433".
	// Required.
	KratosPublicURL string

	// HydraAdminURL is the Ory Hydra admin API root URL,
	// e.g. "http://hydra:4445".
	// Optional — when set, the adapter exchanges the Kratos session for a
	// Hydra JWT access token after login.
	HydraAdminURL string

	// HydraClientID is the OAuth2 client ID used when creating tokens via
	// the Hydra admin API. Required when HydraAdminURL is set.
	HydraClientID string
}

// OryAdapter implements ports.IdentityProvider for the Ory Kratos + Hydra stack.
type OryAdapter struct {
	cfg    OryConfig
	client *http.Client
}

// NewOryAdapter creates an Ory identity provider adapter.
func NewOryAdapter(cfg OryConfig) *OryAdapter {
	cfg.KratosPublicURL = strings.TrimRight(cfg.KratosPublicURL, "/")
	if cfg.HydraAdminURL != "" {
		cfg.HydraAdminURL = strings.TrimRight(cfg.HydraAdminURL, "/")
	}
	return &OryAdapter{cfg: cfg, client: &http.Client{}}
}

// Login authenticates the user via the Kratos native (API-first) login flow.
// No browser redirect is required.
func (a *OryAdapter) Login(ctx context.Context, username, password string) (*ports.Token, error) {
	// Step 1 — initialize a native login flow.
	flowID, err := a.initFlow(ctx, "login")
	if err != nil {
		return nil, fmt.Errorf("ory login: init flow: %w", err)
	}

	// Step 2 — submit credentials.
	payload, _ := json.Marshal(map[string]any{
		"method":     "password",
		"identifier": username,
		"password":   password,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.cfg.KratosPublicURL+"/self-service/login?flow="+flowID,
		strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("ory login: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ory login: %w", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusBadRequest {
		return nil, domain.NewUnauthorized("invalid username or password")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ory login: unexpected status %d: %s", resp.StatusCode, b)
	}

	var result struct {
		SessionToken string `json:"session_token"`
		Session      struct {
			Identity struct {
				ID string `json:"id"`
			} `json:"identity"`
		} `json:"session"`
	}
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, fmt.Errorf("ory login: decode response: %w", err)
	}

	// If Hydra is configured, exchange the Kratos session for a JWT access token.
	if a.cfg.HydraAdminURL != "" && result.Session.Identity.ID != "" {
		tok, err := a.hydraTokenForSubject(ctx, result.Session.Identity.ID)
		if err != nil {
			return nil, fmt.Errorf("ory login: hydra token exchange: %w", err)
		}
		return tok, nil
	}

	// No Hydra — return the Kratos session token directly.
	// The session token is an opaque bearer token; validate it with the Kratos
	// session introspection endpoint (/sessions/whoami) rather than JWKS.
	return &ports.Token{AccessToken: result.SessionToken}, nil
}

// Register creates a new identity via the Kratos native registration flow.
func (a *OryAdapter) Register(ctx context.Context, req ports.RegisterRequest) (string, error) {
	flowID, err := a.initFlow(ctx, "registration")
	if err != nil {
		return "", fmt.Errorf("ory register: init flow: %w", err)
	}

	payload, _ := json.Marshal(map[string]any{
		"method":   "password",
		"password": req.Password,
		"traits": map[string]any{
			"email":      req.Email,
			"username":   req.Username,
			"name": map[string]any{
				"first": req.FirstName,
				"last":  req.LastName,
			},
		},
	})

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.cfg.KratosPublicURL+"/self-service/registration?flow="+flowID,
		strings.NewReader(string(payload)))
	if err != nil {
		return "", fmt.Errorf("ory register: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ory register: %w", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusConflict {
		return "", domain.NewConflict("a user with that email or username already exists")
	}
	if resp.StatusCode == http.StatusBadRequest {
		// Kratos returns 400 for validation errors (duplicate email/username,
		// password policy violations, etc.). Try to extract a useful message.
		var errBody struct {
			UI struct {
				Messages []struct {
					Text string `json:"text"`
				} `json:"messages"`
			} `json:"ui"`
		}
		if json.Unmarshal(b, &errBody) == nil && len(errBody.UI.Messages) > 0 {
			return "", domain.NewBadRequest(errBody.UI.Messages[0].Text)
		}
		return "", domain.NewBadRequest("registration failed")
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("ory register: unexpected status %d: %s", resp.StatusCode, b)
	}

	var result struct {
		Identity struct {
			ID string `json:"id"`
		} `json:"identity"`
	}
	if err := json.Unmarshal(b, &result); err != nil {
		return "", fmt.Errorf("ory register: decode response: %w", err)
	}
	return result.Identity.ID, nil
}

// initFlow initialises a Kratos native (API-first) self-service flow and
// returns the flow ID.  kind is "login" or "registration".
func (a *OryAdapter) initFlow(ctx context.Context, kind string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		a.cfg.KratosPublicURL+"/self-service/"+kind+"/api", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("init %s flow: status %d: %s", kind, resp.StatusCode, b)
	}

	var flow struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&flow); err != nil {
		return "", err
	}
	if flow.ID == "" {
		return "", fmt.Errorf("init %s flow: empty flow id", kind)
	}
	return flow.ID, nil
}

// hydraTokenForSubject uses the Hydra admin API to create an access token for
// a known subject. This is only called when HydraAdminURL is configured.
func (a *OryAdapter) hydraTokenForSubject(ctx context.Context, subject string) (*ports.Token, error) {
	payload, _ := json.Marshal(map[string]any{
		"subject":    subject,
		"client_id":  a.cfg.HydraClientID,
		"extra":      map[string]any{},
		"grant_type": "client_credentials",
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.cfg.HydraAdminURL+"/admin/oauth2/token",
		strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("hydra token: status %d: %s", resp.StatusCode, b)
	}

	var tok ports.Token
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, fmt.Errorf("hydra token: decode: %w", err)
	}
	return &tok, nil
}
