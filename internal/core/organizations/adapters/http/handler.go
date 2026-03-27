package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const basePath = "/api/v1/organizations"

// Handler groups all HTTP endpoints for the organizations module.
type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{logger: logger}
}

// RegisterRoutes attaches all organizations routes to the provided router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get(basePath, h.list)
	r.Post(basePath, h.create)
	r.Get(basePath+"/{id}", h.get)
	r.Patch(basePath+"/{id}", h.update)
	r.Delete(basePath+"/{id}", h.delete)

	// Members sub-resource
	r.Get(basePath+"/{id}/members", h.listMembers)
	r.Post(basePath+"/{id}/members", h.addMember)
	r.Delete(basePath+"/{id}/members/{userId}", h.removeMember)
}

func notImplemented(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(`{"error":"not implemented"}`))
}

func (h *Handler) list(w http.ResponseWriter, _ *http.Request)              { notImplemented(w) }
func (h *Handler) create(w http.ResponseWriter, _ *http.Request)            { notImplemented(w) }
func (h *Handler) get(w http.ResponseWriter, _ *http.Request)               { notImplemented(w) }
func (h *Handler) update(w http.ResponseWriter, _ *http.Request)            { notImplemented(w) }
func (h *Handler) delete(w http.ResponseWriter, _ *http.Request)            { notImplemented(w) }
func (h *Handler) listMembers(w http.ResponseWriter, _ *http.Request)       { notImplemented(w) }
func (h *Handler) addMember(w http.ResponseWriter, _ *http.Request)         { notImplemented(w) }
func (h *Handler) removeMember(w http.ResponseWriter, _ *http.Request)      { notImplemented(w) }
