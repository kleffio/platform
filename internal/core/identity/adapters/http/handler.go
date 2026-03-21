package http

import (
	"log/slog"
	"net/http"

	commonhttp "github.com/kleff/go-common/adapters/http"
	"github.com/kleff/go-common/domain"
	"github.com/kleff/platform/internal/shared/middleware"
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

type meResponse struct {
	UserID string `json:"user_id"`
}

// GET /api/v1/identity/me — returns the currently authenticated user.
func (h *Handler) getMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		commonhttp.Error(w, domain.NewUnauthorized("unauthorized"))
		return
	}
	commonhttp.Success(w, meResponse{UserID: claims.Subject})
}

// PATCH /api/v1/identity/me — updates the current user's profile.
func (h *Handler) updateMe(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}
