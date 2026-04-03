// Package http exposes the plugin marketplace REST API.
//
// Routes (all under /api/v1/admin/plugins, require admin role):
//
//	GET    /api/v1/admin/plugins/catalog          – list remote plugin catalog
//	POST   /api/v1/admin/plugins/catalog/refresh  – force catalog refresh
//	GET    /api/v1/admin/plugins                  – list installed plugins
//	POST   /api/v1/admin/plugins                  – install a plugin
//	GET    /api/v1/admin/plugins/{id}             – get a single plugin
//	PATCH  /api/v1/admin/plugins/{id}/config      – update plugin config
//	POST   /api/v1/admin/plugins/{id}/enable      – enable plugin
//	POST   /api/v1/admin/plugins/{id}/disable     – disable plugin
//	DELETE /api/v1/admin/plugins/{id}             – remove plugin
//	GET    /api/v1/admin/plugins/{id}/status      – live container/gRPC status
//	POST   /api/v1/admin/plugins/{id}/set-active  – set as active IDP
//
// Authenticated (any role):
//
//	GET    /api/v1/plugins/catalog  – marketplace catalog; admins see all, others see type=ui only
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
	plugindomain "github.com/kleffio/platform/internal/core/plugins/domain"
	"github.com/kleffio/platform/internal/core/plugins/ports"
	"github.com/kleffio/platform/internal/shared/middleware"
)

// Handler is the plugin marketplace HTTP handler.
type Handler struct {
	manager  ports.PluginManager
	registry ports.PluginRegistry
	logger   *slog.Logger
}

// NewHandler wires the marketplace handler.
func NewHandler(
	manager ports.PluginManager,
	registry ports.PluginRegistry,
	logger *slog.Logger,
) *Handler {
	return &Handler{manager: manager, registry: registry, logger: logger}
}

// RegisterPublicRoutes attaches authenticated-but-not-admin plugin routes.
func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	r.Get("/api/v1/plugins/catalog", h.handleMarketplaceCatalog)
}

// RegisterRoutes attaches all plugin routes to the provided router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/api/v1/admin/plugins/catalog", h.handleListCatalog)
	r.Post("/api/v1/admin/plugins/catalog/refresh", h.handleRefreshCatalog)
	r.Get("/api/v1/admin/plugins", h.handleListInstalled)
	r.Post("/api/v1/admin/plugins", h.handleInstall)
	r.Get("/api/v1/admin/plugins/{id}", h.handleGetPlugin)
	r.Patch("/api/v1/admin/plugins/{id}/config", h.handleConfigure)
	r.Post("/api/v1/admin/plugins/{id}/enable", h.handleEnable)
	r.Post("/api/v1/admin/plugins/{id}/disable", h.handleDisable)
	r.Delete("/api/v1/admin/plugins/{id}", h.handleRemove)
	r.Get("/api/v1/admin/plugins/{id}/status", h.handleStatus)
	r.Post("/api/v1/admin/plugins/{id}/set-active", h.handleSetActive)
}

// ── Catalog ───────────────────────────────────────────────────────────────────

func (h *Handler) handleListCatalog(w http.ResponseWriter, r *http.Request) {
	catalog, err := h.registry.ListCatalog(r.Context())
	if err != nil {
		h.logger.Error("list catalog", "error", err)
		commonhttp.Error(w, err)
		return
	}
	commonhttp.Success(w, map[string]any{
		"plugins":   catalog,
		"cached_at": h.registry.CachedAt(),
	})
}

// handleMarketplaceCatalog serves the plugin marketplace to all authenticated users.
// Admins receive the full catalog; everyone else receives only type=ui plugins.
func (h *Handler) handleMarketplaceCatalog(w http.ResponseWriter, r *http.Request) {
	catalog, err := h.registry.ListCatalog(r.Context())
	if err != nil {
		h.logger.Error("marketplace catalog", "error", err)
		commonhttp.Error(w, err)
		return
	}

	claims, _ := middleware.ClaimsFromContext(r.Context())
	isAdmin := false
	for _, role := range claims.Roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		filtered := catalog[:0]
		for _, p := range catalog {
			if p.Type == "ui" {
				filtered = append(filtered, p)
			}
		}
		catalog = filtered
	}

	commonhttp.Success(w, map[string]any{
		"plugins":   catalog,
		"cached_at": h.registry.CachedAt(),
	})
}

func (h *Handler) handleRefreshCatalog(w http.ResponseWriter, r *http.Request) {
	if err := h.registry.Refresh(r.Context()); err != nil {
		h.logger.Error("refresh catalog", "error", err)
		commonhttp.Error(w, err)
		return
	}
	commonhttp.Success(w, map[string]string{"cached_at": h.registry.CachedAt()})
}

// ── Installed plugins ─────────────────────────────────────────────────────────

func (h *Handler) handleListInstalled(w http.ResponseWriter, r *http.Request) {
	plugins, err := h.manager.ListPlugins(r.Context())
	if err != nil {
		commonhttp.Error(w, err)
		return
	}
	activeID := h.manager.GetActiveIDPID()
	responses := toResponses(plugins)
	for i := range responses {
		if responses[i].ID == activeID {
			responses[i].IsActiveIDP = true
		}
	}
	commonhttp.Success(w, map[string]any{"plugins": responses})
}

func (h *Handler) handleGetPlugin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := h.manager.GetPlugin(r.Context(), id)
	if err != nil {
		commonhttp.Error(w, err)
		return
	}
	commonhttp.Success(w, toResponse(p))
}

// ── Install ───────────────────────────────────────────────────────────────────

type installRequest struct {
	ID      string            `json:"id"`
	Version string            `json:"version"`
	Config  map[string]string `json:"config"`
}

func (h *Handler) handleInstall(w http.ResponseWriter, r *http.Request) {
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
		h.logger.Error("install: get manifest", "id", req.ID, "error", err)
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
		h.logger.Warn("install plugin failed", "id", req.ID, "error", err)
		commonhttp.Error(w, domain.NewBadRequest(err.Error()))
		return
	}
	commonhttp.Created(w, toResponse(p))
}

// ── Configure ─────────────────────────────────────────────────────────────────

type configureRequest struct {
	Config map[string]string `json:"config"`
}

func (h *Handler) handleConfigure(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req configureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		commonhttp.Error(w, domain.NewBadRequest("invalid request body"))
		return
	}

	if err := h.manager.Reconfigure(r.Context(), id, req.Config); err != nil {
		h.logger.Warn("configure plugin failed", "id", id, "error", err)
		commonhttp.Error(w, err)
		return
	}
	commonhttp.NoContent(w)
}

// ── Enable / Disable ──────────────────────────────────────────────────────────

func (h *Handler) handleEnable(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.manager.Enable(r.Context(), id); err != nil {
		h.logger.Warn("enable plugin failed", "id", id, "error", err)
		commonhttp.Error(w, err)
		return
	}
	commonhttp.NoContent(w)
}

func (h *Handler) handleDisable(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.manager.Disable(r.Context(), id); err != nil {
		h.logger.Warn("disable plugin failed", "id", id, "error", err)
		commonhttp.Error(w, err)
		return
	}
	commonhttp.NoContent(w)
}

// ── Remove ────────────────────────────────────────────────────────────────────

func (h *Handler) handleRemove(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.manager.Remove(r.Context(), id); err != nil {
		h.logger.Warn("remove plugin failed", "id", id, "error", err)
		commonhttp.Error(w, domain.NewBadRequest(err.Error()))
		return
	}
	commonhttp.NoContent(w)
}

// ── Status ────────────────────────────────────────────────────────────────────

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := h.manager.GetPlugin(r.Context(), id)
	if err != nil {
		commonhttp.Error(w, err)
		return
	}
	commonhttp.Success(w, map[string]string{
		"id":     p.ID,
		"status": string(p.Status),
	})
}

// ── Set active IDP ────────────────────────────────────────────────────────────

func (h *Handler) handleSetActive(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.manager.SetActiveIDP(r.Context(), id); err != nil {
		h.logger.Warn("set active IDP failed", "id", id, "error", err)
		commonhttp.Error(w, domain.NewBadRequest(err.Error()))
		return
	}
	commonhttp.NoContent(w)
}

// ── Response mapping ──────────────────────────────────────────────────────────

type pluginResponse struct {
	ID          string                    `json:"id"`
	Type        string                    `json:"type"`
	DisplayName string                    `json:"display_name"`
	Image       string                    `json:"image"`
	Version     string                    `json:"version"`
	GRPCAddr    string                    `json:"grpc_addr"`
	FrontendURL string                    `json:"frontend_url"`
	Config      json.RawMessage           `json:"config"`
	Enabled     bool                      `json:"enabled"`
	Status      plugindomain.PluginStatus `json:"status"`
	InstalledAt string                    `json:"installed_at"`
	UpdatedAt   string                    `json:"updated_at"`
	IsActiveIDP bool                      `json:"is_active_idp"`
}

func toResponse(p *plugindomain.Plugin) pluginResponse {
	return pluginResponse{
		ID:          p.ID,
		Type:        p.Type,
		DisplayName: p.DisplayName,
		Image:       p.Image,
		Version:     p.Version,
		GRPCAddr:    p.GRPCAddr,
		FrontendURL: p.FrontendURL,
		Config:      p.Config, // secrets already stripped from Config field
		Enabled:     p.Enabled,
		Status:      p.Status,
		InstalledAt: p.InstalledAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toResponses(plugins []*plugindomain.Plugin) []pluginResponse {
	out := make([]pluginResponse, len(plugins))
	for i, p := range plugins {
		out[i] = toResponse(p)
	}
	return out
}
