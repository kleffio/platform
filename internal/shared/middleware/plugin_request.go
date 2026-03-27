package middleware

import (
	"context"
	"net/http"

	commonhttp "github.com/kleff/go-common/adapters/http"
	"github.com/kleff/go-common/domain"
)

// PluginRequestRunner is the subset of PluginManager used by PluginRequest.
type PluginRequestRunner interface {
	RunMiddleware(ctx context.Context, userID string, roles []string, method, path string) error
}

// PluginRequest fans out to all plugins that declared CapabilityAPIMiddleware.
// Must sit inside RequireAuth so claims are already in context.
// If any plugin denies the request it short-circuits with 403.
func PluginRequest(runner PluginRequestRunner) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			if err := runner.RunMiddleware(r.Context(), claims.Subject, claims.Roles, r.Method, r.URL.Path); err != nil {
				commonhttp.Error(w, domain.NewForbidden(err.Error()))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
