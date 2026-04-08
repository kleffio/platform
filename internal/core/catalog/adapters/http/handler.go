package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kleffio/platform/internal/core/catalog/ports"
)

// Handler exposes the catalog (crates, blueprints, constructs) over HTTP.
type Handler struct {
	repo   ports.CatalogRepository
	logger *slog.Logger
}

func NewHandler(repo ports.CatalogRepository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/api/v1/crates", h.listCrates)
	r.Get("/api/v1/crates/{id}", h.getCrate)
	r.Get("/api/v1/blueprints", h.listBlueprints)
	r.Get("/api/v1/blueprints/{id}", h.getBlueprint)
	r.Get("/api/v1/constructs", h.listConstructs)
	r.Get("/api/v1/constructs/{id}", h.getConstruct)
}

func (h *Handler) listCrates(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	crates, err := h.repo.ListCrates(r.Context(), category)
	if err != nil {
		h.internalError(w, "list crates", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"crates": crates})
}

func (h *Handler) getCrate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	crate, err := h.repo.GetCrate(r.Context(), id)
	if err != nil {
		h.notFound(w, err)
		return
	}

	writeJSON(w, http.StatusOK, crate)
}

func (h *Handler) listBlueprints(w http.ResponseWriter, r *http.Request) {
	crateID := r.URL.Query().Get("crate")

	blueprints, err := h.repo.ListBlueprints(r.Context(), crateID)
	if err != nil {
		h.internalError(w, "list blueprints", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"blueprints": blueprints})
}

func (h *Handler) getBlueprint(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	blueprint, err := h.repo.GetBlueprint(r.Context(), id)
	if err != nil {
		h.notFound(w, err)
		return
	}

	writeJSON(w, http.StatusOK, blueprint)
}

func (h *Handler) listConstructs(w http.ResponseWriter, r *http.Request) {
	crateID := r.URL.Query().Get("crate")
	blueprintID := r.URL.Query().Get("blueprint")

	constructs, err := h.repo.ListConstructs(r.Context(), crateID, blueprintID)
	if err != nil {
		h.internalError(w, "list constructs", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"constructs": constructs})
}

func (h *Handler) getConstruct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	construct, err := h.repo.GetConstruct(r.Context(), id)
	if err != nil {
		h.notFound(w, err)
		return
	}

	writeJSON(w, http.StatusOK, construct)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) internalError(w http.ResponseWriter, op string, err error) {
	h.logger.Error(op, "error", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}

func (h *Handler) notFound(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
}
