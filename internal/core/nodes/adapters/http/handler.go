package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const basePath = "/api/v1/nodes"

// Handler groups all HTTP endpoints for the nodes module.
type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{logger: logger}
}

// RegisterRoutes attaches all node routes to the provided router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get(basePath, h.list)
	r.Get(basePath+"/{id}", h.get)
	r.Post(basePath+"/{id}/drain", h.drain)
}

func notImplemented(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(`{"error":"not implemented"}`))
}

func (h *Handler) list(w http.ResponseWriter, _ *http.Request)  { notImplemented(w) }
func (h *Handler) get(w http.ResponseWriter, _ *http.Request)   { notImplemented(w) }
func (h *Handler) drain(w http.ResponseWriter, _ *http.Request) { notImplemented(w) }
