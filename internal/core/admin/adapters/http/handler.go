package http

import (
	"log/slog"
	"net/http"

	commonhttp "github.com/kleff/go-common/adapters/http"
	"github.com/kleff/go-common/domain"
)

const basePath = "/api/v1/admin"

// Handler groups all HTTP endpoints for the admin module.
// These routes are restricted to platform operators (staff role).
type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{logger: logger}
}

// RegisterRoutes attaches all admin routes to the provided mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+basePath+"/organizations", h.listOrgs)
	mux.HandleFunc("GET "+basePath+"/users", h.listUsers)
	mux.HandleFunc("POST "+basePath+"/users/{id}/suspend", h.suspendUser)
	mux.HandleFunc("POST "+basePath+"/organizations/{id}/suspend", h.suspendOrg)
}

func (h *Handler) listOrgs(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewForbidden("admin access required"))
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewForbidden("admin access required"))
}

func (h *Handler) suspendUser(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewForbidden("admin access required"))
}

func (h *Handler) suspendOrg(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewForbidden("admin access required"))
}
