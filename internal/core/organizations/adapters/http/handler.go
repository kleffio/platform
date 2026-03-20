package http

import (
	"log/slog"
	"net/http"

	commonhttp "github.com/kleff/go-common/adapters/http"
	"github.com/kleff/go-common/domain"
)

const basePath = "/api/v1/organizations"

// Handler groups all HTTP endpoints for the organizations module.
type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{logger: logger}
}

// RegisterRoutes attaches all organizations routes to the provided mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+basePath, h.list)
	mux.HandleFunc("POST "+basePath, h.create)
	mux.HandleFunc("GET "+basePath+"/{id}", h.get)
	mux.HandleFunc("PATCH "+basePath+"/{id}", h.update)
	mux.HandleFunc("DELETE "+basePath+"/{id}", h.delete)

	// Members sub-resource
	mux.HandleFunc("GET "+basePath+"/{id}/members", h.listMembers)
	mux.HandleFunc("POST "+basePath+"/{id}/members", h.addMember)
	mux.HandleFunc("DELETE "+basePath+"/{id}/members/{userId}", h.removeMember)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}

func (h *Handler) listMembers(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}

func (h *Handler) addMember(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}

func (h *Handler) removeMember(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}
