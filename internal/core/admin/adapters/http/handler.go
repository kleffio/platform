package http

import (
	"log/slog"
	"net/http"

	commonhttp "github.com/kleff/go-common/adapters/http"
	"github.com/kleff/go-common/domain"

	middleware "github.com/kleff/platform/packages/adapters/http"
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
//
//	func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
//		mux.HandleFunc("GET "+basePath+"/organizations", h.listOrgs)
//		mux.HandleFunc("GET "+basePath+"/users", h.listUsers)
//		mux.HandleFunc("POST "+basePath+"/users/{id}/suspend", h.suspendUser)
//		mux.HandleFunc("POST "+basePath+"/organizations/{id}/suspend", h.suspendOrg)
//	}
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Organizations
	mux.HandleFunc("GET "+basePath+"/orgs", middleware.RequireAdmin(h.listOrgs))
	mux.HandleFunc("GET "+basePath+"/orgs/{id}", middleware.RequireAdmin(h.getOrgDetail))

	// Game Servers
	mux.HandleFunc("GET "+basePath+"/gameservers", middleware.RequireAdmin(h.listGameServers))
	mux.HandleFunc("POST "+basePath+"/gameservers/{id}/stop", middleware.RequireAdmin(h.stopGameServer))

	// Users
	mux.HandleFunc("GET "+basePath+"/users", middleware.RequireAdmin(h.listUsers))
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

// GET /api/v1/admin/orgs/{id}
func (h *Handler) getOrgDetail(w http.ResponseWriter, r *http.Request) {

	commonhttp.Error(w, domain.NewForbidden("admin access required"))

}

// GET /api/v1/admin/gameservers
func (h *Handler) listGameServers(w http.ResponseWriter, r *http.Request) {

	commonhttp.Error(w, domain.NewForbidden("admin access required"))

}

// POST /api/v1/admin/gameservers/{id}/stop
func (h *Handler) stopGameServer(w http.ResponseWriter, r *http.Request) {

	commonhttp.Error(w, domain.NewForbidden("admin access required"))

}
