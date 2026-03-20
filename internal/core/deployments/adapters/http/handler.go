package http

import (
	"log/slog"
	"net/http"

	commonhttp "github.com/kleff/go-common/adapters/http"
	"github.com/kleff/go-common/domain"
)

const basePath = "/api/v1/deployments"

// Handler groups all HTTP endpoints for the deployments module.
type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{logger: logger}
}

// RegisterRoutes attaches all deployment routes to the provided mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+basePath, h.list)
	mux.HandleFunc("POST "+basePath, h.create)
	mux.HandleFunc("GET "+basePath+"/{id}", h.get)
	mux.HandleFunc("POST "+basePath+"/{id}/cancel", h.cancel)
	mux.HandleFunc("POST "+basePath+"/{id}/rollback", h.rollback)
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

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}

func (h *Handler) rollback(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}
