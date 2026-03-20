// Package middleware provides shared HTTP middleware for the platform API.
package middleware

import (
	"context"
	"net/http"
	"strings"
)

// contextKey is unexported to prevent collision with other packages.
type contextKey string

const claimsKey contextKey = "jwt_claims"

// Claims holds the parsed JWT claims injected by the auth middleware.
type Claims struct {
	Subject string // OIDC sub
	Email   string
	Roles   []string
	OrgID   string // organization derived from token claims
}

// ClaimsFromContext retrieves parsed JWT claims from the request context.
// Returns (nil, false) if no claims are present (unauthenticated request).
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claimsKey).(*Claims)
	return c, ok
}

// RequireAuth is a middleware that validates the Bearer token in the
// Authorization header. The token is verified against the OIDC authority
// configured at startup.
//
// TODO: implement JWT verification using jwks endpoint from OIDC discovery.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearer(r)
		if token == "" {
			http.Error(w, `{"error":"unauthorized","code":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// TODO: validate JWT signature and expiry, extract claims.
		_ = token

		// Stub: inject placeholder claims so downstream handlers can compile.
		claims := &Claims{Subject: "stub"}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole ensures the caller has at least one of the specified roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				http.Error(w, `{"error":"unauthorized","code":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			for _, role := range claims.Roles {
				if allowed[role] {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, `{"error":"forbidden","code":"forbidden"}`, http.StatusForbidden)
		})
	}
}

func extractBearer(r *http.Request) string {
	v := r.Header.Get("Authorization")
	if after, ok := strings.CutPrefix(v, "Bearer "); ok {
		return after
	}
	return ""
}
