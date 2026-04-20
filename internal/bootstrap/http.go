package bootstrap

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	commonhttp "github.com/kleff/go-common/adapters/http"
	"github.com/kleffio/platform/internal/shared/middleware"
)

func buildRouter(c *Container) http.Handler {
	r := chi.NewRouter()

	// Global middleware — outermost to innermost
	r.Use(commonhttp.CORS(c.Config.CORSAllowedOrigins...))
	r.Use(commonhttp.Recover(c.Logger))
	r.Use(commonhttp.Logger(c.Logger))
	r.Use(commonhttp.RequestID)
	r.Use(middleware.PluginRouteInterceptor(c.PluginManager, c.TokenVerifier))

	// Health probes
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	r.Get("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Internal daemon-facing routes — shared secret only, no user auth.
	c.DeploymentsHandler.RegisterInternalRoutes(r)

	// Public auth routes (login, register, config) — no bearer token required.
	c.AuthHandler.RegisterPublicRoutes(r)

	// Public setup routes — only active before the first IDP is installed.
	c.SetupHandler.RegisterPublicRoutes(r)

	// Catalog (crates + blueprints) is public — no login needed to browse.
	c.CatalogHandler.RegisterRoutes(r)

	// Daemon node bootstrap route. Uses a shared bootstrap secret and returns a
	// node token for subsequent machine-to-machine calls.
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireNodeBootstrap(c.Config.NodeBootstrapSecret))
		c.NodesHandler.RegisterPublicRoutes(r)
	})

	// Daemon internal callback routes. Require a per-node token issued during
	// registration, not end-user JWT auth.
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireNodeAuth(c.NodeVerifier))
		c.WorkloadsHandler.RegisterInternalRoutes(r)
		c.LogsHandler.RegisterInternalRoutes(r)
	})

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(c.TokenVerifier))
		r.Use(middleware.PluginRequest(c.PluginManager))

		c.AuthHandler.RegisterRoutes(r)
		c.PluginsHandler.RegisterPublicRoutes(r)
		c.OrganizationsHandler.RegisterRoutes(r)
		c.ProjectsHandler.RegisterRoutes(r)
		c.WorkloadsHandler.RegisterRoutes(r)
		c.DeploymentsHandler.RegisterRoutes(r)
		c.NodesHandler.RegisterRoutes(r)
		c.BillingHandler.RegisterRoutes(r)
		c.UsageHandler.RegisterRoutes(r)
		c.LogsHandler.RegisterRoutes(r)
		c.AuditHandler.RegisterRoutes(r)
		c.NotificationsHandler.RegisterRoutes(r)

		// Admin routes — additionally require the "admin" role.
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole("admin"))
			c.AdminHandler.RegisterRoutes(r)
			c.PluginsHandler.RegisterRoutes(r)
		})
	})

	return r
}
