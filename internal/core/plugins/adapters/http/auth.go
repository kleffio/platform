package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	commonhttp "github.com/kleff/go-common/adapters/http"
	"github.com/kleff/go-common/domain"
	pluginsv1 "github.com/kleffio/plugin-sdk-go/v1"
	"github.com/kleffio/platform/internal/core/plugins/ports"
	"github.com/kleffio/platform/internal/shared/middleware"
)

// AuthHandler owns all platform-level auth endpoints.
// Plugins act as IDP adapters — the platform standardizes the HTTP surface.
type AuthHandler struct {
	manager ports.PluginManager
	logger  *slog.Logger
}

func NewAuthHandler(manager ports.PluginManager, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{manager: manager, logger: logger}
}

// RegisterPublicRoutes attaches unauthenticated auth endpoints.
func (h *AuthHandler) RegisterPublicRoutes(r chi.Router) {
	r.Post("/api/v1/auth/login", h.handleLogin)
	r.Post("/api/v1/auth/register", h.handleRegister)
	r.Get("/api/v1/auth/config", h.handleConfig)
	r.Post("/api/v1/auth/refresh", h.handleRefresh)
	r.Post("/api/v1/auth/token-exchange", h.handleTokenExchange)
}

// RegisterRoutes attaches authenticated auth endpoints.
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/v1/auth/me", h.handleMe)
	r.Get("/api/v1/plugins/ui-manifests", h.handleUIManifests)
	r.Post("/api/v1/auth/change-password", h.handleChangePassword)
	r.Get("/api/v1/auth/sessions", h.handleListSessions)
	r.Delete("/api/v1/auth/sessions", h.handleRevokeAllSessions)
	r.Delete("/api/v1/auth/sessions/{sessionID}", h.handleRevokeSession)
}

// POST /api/v1/auth/login
func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" || req.Password == "" {
		commonhttp.Error(w, domain.NewBadRequest("username and password are required"))
		return
	}
	tok, err := h.manager.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if isPluginError(err, pluginsv1.ErrorCodeUnauthorized) {
			commonhttp.Error(w, domain.NewUnauthorized("invalid username or password"))
			return
		}
		h.logger.Warn("login failed", "error", err)
		commonhttp.Error(w, domain.NewInternal(err))
		return
	}
	commonhttp.Success(w, tok)
}

// POST /api/v1/auth/register
func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username  string `json:"username"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" || req.Email == "" || req.Password == "" {
		commonhttp.Error(w, domain.NewBadRequest("username, email, and password are required"))
		return
	}
	userID, err := h.manager.Register(r.Context(), &pluginsv1.RegisterRequest{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		if isPluginError(err, pluginsv1.ErrorCodeConflict) {
			commonhttp.Error(w, domain.NewConflict("user already exists"))
			return
		}
		h.logger.Warn("register failed", "error", err)
		commonhttp.Error(w, domain.NewInternal(err))
		return
	}
	commonhttp.Created(w, map[string]string{"user_id": userID})
}

// GET /api/v1/auth/config
func (h *AuthHandler) handleConfig(w http.ResponseWriter, r *http.Request) {
	// Always return a fresh response — the active IDP can change at runtime and
	// we don't want a browser-cached stale config redirecting to the wrong IDP.
	w.Header().Set("Cache-Control", "no-store")
	cfg, err := h.manager.GetOIDCConfig(r.Context())
	if err != nil {
		h.logger.Warn("get OIDC config failed", "error", err)
	}
	// An IDP plugin is active but its container hasn't finished starting yet —
	// the gRPC call failed (cfg == nil) or returned empty authority/client_id.
	// Return ready:false so the frontend keeps polling every 3 s.
	// Use HasIdentityProvider() for enabled so the login page shows the spinner
	// rather than the form (which would show even when no IDP is configured).
	if cfg == nil || cfg.Authority == "" || cfg.ClientID == "" {
		hasIDP := h.manager.HasIdentityProvider()
		commonhttp.Success(w, map[string]any{
			"enabled":        hasIDP,
			"setup_required": !hasIDP,
			"ready":          false,
		})
		return
	}
	resp := map[string]any{
		"enabled":   true,
		"authority": cfg.Authority,
		"client_id": cfg.ClientID,
		"jwks_uri":  cfg.JwksURI,
		"scopes":    cfg.Scopes,
		"auth_mode": cfg.AuthMode,
		"ready":     true,
	}
	if cfg.TokenEndpoint != "" {
		resp["token_endpoint"] = cfg.TokenEndpoint
	}
	if cfg.AuthorizationEndpoint != "" {
		resp["authorization_endpoint"] = cfg.AuthorizationEndpoint
	}
	if cfg.UserinfoEndpoint != "" {
		resp["userinfo_endpoint"] = cfg.UserinfoEndpoint
	}
	if cfg.EndSessionEndpoint != "" {
		resp["end_session_endpoint"] = cfg.EndSessionEndpoint
	}
	commonhttp.Success(w, resp)
}

// GET /api/v1/auth/me
func (h *AuthHandler) handleMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		commonhttp.Error(w, domain.NewUnauthorized("unauthorized"))
		return
	}
	roles := claims.Roles
	if roles == nil {
		roles = []string{}
	}
	commonhttp.Success(w, map[string]any{
		"user_id": claims.Subject,
		"email":   claims.Email,
		"roles":   roles,
	})
}

// GET /api/v1/plugins/ui-manifests
func (h *AuthHandler) handleUIManifests(w http.ResponseWriter, r *http.Request) {
	manifests, err := h.manager.GetUIManifests(r.Context())
	if err != nil {
		h.logger.Warn("ui-manifests: failed", "error", err)
		commonhttp.Error(w, domain.NewInternal(err))
		return
	}
	if manifests == nil {
		manifests = []*pluginsv1.UIManifest{}
	}
	commonhttp.Success(w, map[string]any{"plugins": manifests})
}

// POST /api/v1/auth/refresh
func (h *AuthHandler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		commonhttp.Error(w, domain.NewBadRequest("refresh_token is required"))
		return
	}
	tok, err := h.manager.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if isPluginError(err, pluginsv1.ErrorCodeUnauthorized) {
			commonhttp.Error(w, domain.NewUnauthorized("refresh token is invalid or expired"))
			return
		}
		h.logger.Warn("refresh token failed", "error", err)
		commonhttp.Error(w, domain.NewInternal(err))
		return
	}
	commonhttp.Success(w, tok)
}

// POST /api/v1/auth/token-exchange
// Proxies the PKCE authorization-code token exchange to the IDP's internal
// token endpoint, avoiding CORS issues when the browser cannot POST cross-origin.
func (h *AuthHandler) handleTokenExchange(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.manager.GetOIDCConfig(r.Context())
	if err != nil || cfg == nil || cfg.InternalTokenEndpoint == "" {
		h.logger.Warn("token-exchange: internal token endpoint unavailable", "error", err)
		commonhttp.Error(w, domain.NewInternal(fmt.Errorf("token endpoint not available")))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		commonhttp.Error(w, domain.NewInternal(err))
		return
	}
	proxyReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, cfg.InternalTokenEndpoint, bytes.NewReader(body))
	if err != nil {
		commonhttp.Error(w, domain.NewInternal(err))
		return
	}
	if ct := r.Header.Get("Content-Type"); ct != "" {
		proxyReq.Header.Set("Content-Type", ct)
	}
	// Forward the public host so the IDP uses the public URL in the iss claim.
	// Without this, Authentik issues tokens with iss=internal-hostname which
	// won't match the metadata.issuer the frontend sees.
	if cfg.TokenEndpoint != "" {
		if u, err := url.Parse(cfg.TokenEndpoint); err == nil {
			proxyReq.Header.Set("X-Forwarded-Host", u.Host)
			proxyReq.Header.Set("X-Forwarded-Proto", u.Scheme)
		}
	}
	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		h.logger.Warn("token-exchange: proxy request failed", "error", err)
		commonhttp.Error(w, domain.NewInternal(err))
		return
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		commonhttp.Error(w, domain.NewInternal(err))
		return
	}
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBody)
}

// POST /api/v1/auth/change-password
func (h *AuthHandler) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		commonhttp.Error(w, domain.NewUnauthorized("unauthorized"))
		return
	}
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CurrentPassword == "" || req.NewPassword == "" {
		commonhttp.Error(w, domain.NewBadRequest("current_password and new_password are required"))
		return
	}
	if err := h.manager.ChangePassword(r.Context(), claims.Subject, req.CurrentPassword, req.NewPassword); err != nil {
		if isPluginError(err, pluginsv1.ErrorCodeUnauthorized) {
			commonhttp.Error(w, domain.NewUnauthorized("current password is incorrect"))
			return
		}
		h.logger.Warn("change password failed", "error", err)
		commonhttp.Error(w, domain.NewInternal(err))
		return
	}
	commonhttp.Success(w, map[string]string{"status": "ok"})
}

// GET /api/v1/auth/sessions?sid=<session_state>
func (h *AuthHandler) handleListSessions(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		commonhttp.Error(w, domain.NewUnauthorized("unauthorized"))
		return
	}
	currentSID := r.URL.Query().Get("sid")
	sessions, err := h.manager.ListSessions(r.Context(), claims.Subject, currentSID)
	if err != nil {
		h.logger.Warn("list sessions failed", "error", err)
		commonhttp.Error(w, domain.NewInternal(err))
		return
	}
	if sessions == nil {
		sessions = []*pluginsv1.Session{}
	}
	commonhttp.Success(w, map[string]any{"sessions": sessions})
}

// DELETE /api/v1/auth/sessions  — revoke all sessions except the current one
func (h *AuthHandler) handleRevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		commonhttp.Error(w, domain.NewUnauthorized("unauthorized"))
		return
	}
	currentSID := r.URL.Query().Get("sid")
	if err := h.manager.RevokeAllSessions(r.Context(), claims.Subject, currentSID); err != nil {
		h.logger.Warn("revoke all sessions failed", "error", err)
		commonhttp.Error(w, domain.NewInternal(err))
		return
	}
	commonhttp.Success(w, map[string]string{"status": "ok"})
}

// DELETE /api/v1/auth/sessions/{sessionID}
func (h *AuthHandler) handleRevokeSession(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		commonhttp.Error(w, domain.NewUnauthorized("unauthorized"))
		return
	}
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		commonhttp.Error(w, domain.NewBadRequest("sessionID is required"))
		return
	}
	if err := h.manager.RevokeSession(r.Context(), claims.Subject, sessionID); err != nil {
		h.logger.Warn("revoke session failed", "error", err)
		commonhttp.Error(w, domain.NewInternal(err))
		return
	}
	commonhttp.Success(w, map[string]string{"status": "ok"})
}

// isPluginError checks whether err is a *pluginsv1.PluginError with the given code.
func isPluginError(err error, code pluginsv1.ErrorCode) bool {
	pe, ok := err.(*pluginsv1.PluginError)
	return ok && pe.Code == code
}
