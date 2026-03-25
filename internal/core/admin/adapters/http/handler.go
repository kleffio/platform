package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
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

// RegisterRoutes attaches all admin routes to the provided mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+basePath+"/organizations", h.listOrgs)
	mux.HandleFunc("GET "+basePath+"/users", h.listUsers)
	mux.HandleFunc("POST "+basePath+"/users/{id}/suspend", h.suspendUser)
	mux.HandleFunc("POST "+basePath+"/organizations/{id}/suspend", h.suspendOrg)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) listOrgs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"data": []map[string]any{
			{"id": "org-001", "name": "Acme Gaming", "plan": "pro", "status": "active", "createdAt": "2024-11-01T00:00:00Z"},
			{"id": "org-002", "name": "Night Owl Servers", "plan": "starter", "status": "active", "createdAt": "2025-01-15T00:00:00Z"},
		},
	})
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"data": []map[string]any{
			{"id": "usr-001", "username": "admin", "email": "admin@kleff.dev", "roles": []string{"admin"}, "status": "active"},
			{"id": "usr-002", "username": "alice", "email": "alice@example.com", "roles": []string{}, "status": "active"},
			{"id": "usr-003", "username": "bob", "email": "bob@example.com", "roles": []string{}, "status": "suspended"},
		},
	})
}

func (h *Handler) suspendUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	writeJSON(w, map[string]any{"message": "user suspended", "id": id})
}

func (h *Handler) suspendOrg(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	writeJSON(w, map[string]any{"message": "organization suspended", "id": id})
}
