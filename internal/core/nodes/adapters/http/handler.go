package http

import (
	"log/slog"
	"net/http"

	commonhttp "github.com/kleff/go-common/adapters/http"
	"github.com/kleff/go-common/domain"
)

const basePath = "/api/v1/nodes"

// Handler groups all HTTP endpoints for the nodes module.
type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{logger: logger}
}

// RegisterRoutes attaches all node routes to the provided mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+basePath, h.list)
	mux.HandleFunc("GET "+basePath+"/{id}", h.get)
	mux.HandleFunc("POST "+basePath+"/{id}/drain", h.drain)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}

func (h *Handler) drain(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}
