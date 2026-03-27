package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	commonhttp "github.com/kleff/go-common/adapters/http"
)

const basePath = "/api/v1/admin"

// Handler groups all HTTP endpoints for the admin module.
// These routes are restricted to users with the "admin" realm role.
type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{logger: logger}
}

// RegisterRoutes attaches all admin routes to the provided router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get(basePath+"/organizations", h.listOrgs)
	r.Get(basePath+"/users", h.listUsers)
	r.Post(basePath+"/users/{id}/suspend", h.suspendUser)
	r.Post(basePath+"/organizations/{id}/suspend", h.suspendOrg)
}

func (h *Handler) listOrgs(w http.ResponseWriter, _ *http.Request) {
	commonhttp.Success(w, []map[string]any{
		{"id": "org-001", "name": "Acme Gaming", "plan": "pro", "status": "active", "createdAt": "2024-11-01T00:00:00Z"},
		{"id": "org-002", "name": "Night Owl Servers", "plan": "starter", "status": "active", "createdAt": "2025-01-15T00:00:00Z"},
	})
}

func (h *Handler) listUsers(w http.ResponseWriter, _ *http.Request) {
	commonhttp.Success(w, []map[string]any{
		{"id": "usr-001", "username": "admin", "email": "admin@kleff.dev", "roles": []string{"admin"}, "status": "active"},
		{"id": "usr-002", "username": "alice", "email": "alice@example.com", "roles": []string{}, "status": "active"},
		{"id": "usr-003", "username": "bob", "email": "bob@example.com", "roles": []string{}, "status": "suspended"},
	})
}

func (h *Handler) suspendUser(w http.ResponseWriter, r *http.Request) {
	commonhttp.Success(w, map[string]string{"message": "user suspended", "id": chi.URLParam(r, "id")})
}

func (h *Handler) suspendOrg(w http.ResponseWriter, r *http.Request) {
	commonhttp.Success(w, map[string]string{"message": "organization suspended", "id": chi.URLParam(r, "id")})
}
