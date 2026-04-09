package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

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
}

// RegisterRoutes attaches authenticated auth endpoints.
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/v1/auth/me", h.handleMe)
	r.Get("/api/v1/plugins/ui-manifests", h.handleUIManifests)
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
	commonhttp.Success(w, map[string]any{
		"enabled":   true,
		"authority": cfg.Authority,
		"client_id": cfg.ClientID,
		"jwks_uri":  cfg.JwksURI,
		"scopes":    cfg.Scopes,
		"auth_mode": cfg.AuthMode,
		"ready":     true,
	})
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
	commonhttp.Success(w, manifests)
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

// isPluginError checks whether err is a *pluginsv1.PluginError with the given code.
func isPluginError(err error, code pluginsv1.ErrorCode) bool {
	pe, ok := err.(*pluginsv1.PluginError)
	return ok && pe.Code == code
}
