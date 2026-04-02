package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	commonhttp "github.com/kleff/go-common/adapters/http"
	"github.com/kleff/go-common/domain"
	"github.com/kleffio/platform/internal/core/plugins/ports"
)

// SetupHandler exposes unauthenticated plugin installation endpoints
// for the initial platform setup (before any IDP is configured).
// All endpoints return 403 once an IDP is active.
type SetupHandler struct {
	manager  ports.PluginManager
	registry ports.PluginRegistry
	logger   *slog.Logger
}

func NewSetupHandler(
	manager ports.PluginManager,
	registry ports.PluginRegistry,
	logger *slog.Logger,
) *SetupHandler {
	return &SetupHandler{manager: manager, registry: registry, logger: logger}
}

// RegisterPublicRoutes attaches setup routes — no auth required.
func (h *SetupHandler) RegisterPublicRoutes(r chi.Router) {
	r.Get("/api/v1/setup/catalog", h.requireSetupMode(h.handleCatalog))
	r.Post("/api/v1/setup/install", h.requireSetupMode(h.handleInstall))
}

// requireSetupMode wraps a handler and rejects requests if an IDP is already active.
func (h *SetupHandler) requireSetupMode(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.manager.HasIdentityProvider() {
			commonhttp.Error(w, domain.NewForbidden("setup already complete"))
			return
		}
		next(w, r)
	}
}

// GET /api/v1/setup/catalog
func (h *SetupHandler) handleCatalog(w http.ResponseWriter, r *http.Request) {
	catalog, err := h.registry.ListCatalog(r.Context())
	if err != nil {
		h.logger.Error("setup: list catalog", "error", err)
		commonhttp.Error(w, err)
		return
	}
	commonhttp.Success(w, map[string]any{"plugins": catalog})
}

// POST /api/v1/setup/install — installs the plugin and sets it as the active IDP.
func (h *SetupHandler) handleInstall(w http.ResponseWriter, r *http.Request) {
	var req installRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		commonhttp.Error(w, domain.NewBadRequest("invalid request body"))
		return
	}
	if req.ID == "" {
		commonhttp.Error(w, domain.NewBadRequest("id is required"))
		return
	}

	manifest, err := h.registry.GetManifest(r.Context(), req.ID)
	if err != nil {
		h.logger.Error("setup: get manifest", "id", req.ID, "error", err)
		commonhttp.Error(w, err)
		return
	}
	if manifest == nil {
		commonhttp.Error(w, domain.NewBadRequest("plugin "+req.ID+" not found in catalog"))
		return
	}

	// Validate required fields.
	for _, field := range manifest.Config {
		if field.Required {
			if _, ok := req.Config[field.Key]; !ok {
				commonhttp.Error(w, domain.NewBadRequest("required config field missing: "+field.Key))
				return
			}
		}
	}

	// Use a detached context with a generous timeout for install.
	// Companion images (e.g. Keycloak ~400MB) can take minutes to pull
	// and must not be cancelled when the HTTP request completes.
	installCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	p, err := h.manager.Install(installCtx, manifest, req.Config)
	if err != nil {
		h.logger.Warn("setup: install plugin failed", "id", req.ID, "error", err)
		commonhttp.Error(w, err)
		return
	}

	// Automatically set as active IDP if it's an IDP plugin.
	if manifest.Type == "idp" {
		if err := h.manager.SetActiveIDP(installCtx, p.ID); err != nil {
			h.logger.Warn("setup: set active IDP failed", "id", p.ID, "error", err)
			// Plugin is installed but activation failed — don't fail the whole request.
		}
	}

	commonhttp.Created(w, toResponse(p))
}

