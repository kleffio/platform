package bootstrap

import (
	"net/http"

	commonhttp "github.com/kleff/go-common/adapters/http"
	"github.com/kleff/platform/internal/shared/middleware"
)

// buildRouter assembles the main HTTP router for the platform API.
// All routes are versioned under /api/v1.
func buildRouter(c *Container) http.Handler {
	// ── Unauthenticated mux (health probes only) ────────────────────────────
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// ── Authenticated API mux ───────────────────────────────────────────────
	// All domain routes run behind RequireAuth. The mux is registered under
	// /api/ so the auth middleware only wraps those paths.
	apiMux := http.NewServeMux()
	c.IdentityHandler.RegisterRoutes(apiMux)
	c.OrganizationsHandler.RegisterRoutes(apiMux)
	c.DeploymentsHandler.RegisterRoutes(apiMux)
	c.NodesHandler.RegisterRoutes(apiMux)
	c.BillingHandler.RegisterRoutes(apiMux)
	c.UsageHandler.RegisterRoutes(apiMux)
	c.AuditHandler.RegisterRoutes(apiMux)
	c.AdminHandler.RegisterRoutes(apiMux)

	mux.Handle("/api/", middleware.RequireAuth(apiMux))

	// ── Global middleware stack ─────────────────────────────────────────────
	var handler http.Handler = mux
	handler = commonhttp.RequestID(handler)
	handler = commonhttp.Logger(c.Logger)(handler)
	handler = commonhttp.Recover(c.Logger)(handler)
	handler = commonhttp.CORS(c.Config.CORSAllowedOrigins...)(handler)

	return handler
}
