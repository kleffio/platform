package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const basePath = "/api/v1/usage"

// Handler groups all HTTP endpoints for the usage module.
type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{logger: logger}
}

// RegisterRoutes attaches all usage routes to the provided router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get(basePath+"/summary", h.getSummary)
	r.Get(basePath+"/records", h.listRecords)
}

func notImplemented(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(`{"error":"not implemented"}`))
}

func (h *Handler) getSummary(w http.ResponseWriter, _ *http.Request)   { notImplemented(w) }
func (h *Handler) listRecords(w http.ResponseWriter, _ *http.Request)  { notImplemented(w) }
