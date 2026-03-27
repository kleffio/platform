package middleware

import (
	"context"
	"io"
	"net/http"

	pluginsv1 "github.com/kleffio/plugin-sdk/v1"
)

// PluginRouter is the subset of PluginManager used by PluginRouteInterceptor.
type PluginRouter interface {
	MatchPluginRoute(method, path string) (pluginID string, public bool, ok bool)
	HandlePluginRoute(ctx context.Context, pluginID string, req *pluginsv1.HTTPRequest) (*pluginsv1.HTTPResponse, error)
}

// PluginRouteInterceptor wraps the entire handler stack. For routes declared by
// a plugin (via CapabilityAPIRoutes), it intercepts the request, optionally
// validates the bearer token for non-public routes, forwards via gRPC Handle,
// and writes the plugin's raw HTTP response. All other requests pass through.
func PluginRouteInterceptor(router PluginRouter, verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pluginID, public, ok := router.MatchPluginRoute(r.Method, r.URL.Path)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			req := &pluginsv1.HTTPRequest{
				Method:   r.Method,
				Path:     r.URL.Path,
				RawQuery: r.URL.RawQuery,
				Headers:  extractHeaders(r),
			}
			if body, err := io.ReadAll(r.Body); err == nil {
				req.Body = body
			}

			if !public {
				token := extractBearer(r)
				if token == "" {
					http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
					return
				}
				result, err := verifier.Verify(r.Context(), token)
				if err != nil {
					http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
					return
				}
				req.UserID = result.Subject
				req.Roles = result.Roles
			}

			resp, err := router.HandlePluginRoute(r.Context(), pluginID, req)
			if err != nil {
				http.Error(w, `{"error":"plugin unavailable"}`, http.StatusServiceUnavailable)
				return
			}

			for k, v := range resp.Headers {
				w.Header().Set(k, v)
			}
			if w.Header().Get("Content-Type") == "" {
				w.Header().Set("Content-Type", "application/json")
			}
			w.WriteHeader(resp.StatusCode)
			_, _ = w.Write(resp.Body)
		})
	}
}

func extractHeaders(r *http.Request) map[string]string {
	out := make(map[string]string, len(r.Header))
	for k := range r.Header {
		out[k] = r.Header.Get(k)
	}
	return out
}
