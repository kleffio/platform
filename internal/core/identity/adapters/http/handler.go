package http

import (
	"log/slog"
	"net/http"

	commonhttp "github.com/kleff/go-common/adapters/http"
	"github.com/kleff/go-common/domain"
)

const basePath = "/api/v1/identity"

// Handler groups all HTTP endpoints for the identity module.
type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{logger: logger}
}

// RegisterRoutes attaches all identity routes to the provided mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+basePath+"/me", h.getMe)
	mux.HandleFunc("PATCH "+basePath+"/me", h.updateMe)
}

// GET /api/v1/identity/me — returns the currently authenticated user.
func (h *Handler) getMe(w http.ResponseWriter, r *http.Request) {
	// TODO: extract user from JWT claims injected by auth middleware.
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}

// PATCH /api/v1/identity/me — updates the current user's profile.
func (h *Handler) updateMe(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}
