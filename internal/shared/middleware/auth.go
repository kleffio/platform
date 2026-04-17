// Package middleware provides shared HTTP middleware for the platform API.
package middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const claimsKey contextKey = "jwt_claims"

// Claims holds verified identity injected into the request context by RequireAuth.
type Claims struct {
	Subject  string
	Username string
	Email    string
	Roles    []string
}

// VerifyResult is returned by TokenVerifier.Verify on success.
type VerifyResult struct {
	Subject  string
	Username string
	Email    string
	Roles    []string
}

// TokenVerifier validates a raw bearer token and returns its claims.
// Implemented by PluginTokenVerifier, which delegates to the active IDP plugin.
type TokenVerifier interface {
	Verify(ctx context.Context, rawToken string) (*VerifyResult, error)
}

// ClaimsFromContext retrieves verified claims injected by RequireAuth.
// Returns (nil, false) if the request was not authenticated.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claimsKey).(*Claims)
	return c, ok
}

// RequireAuth validates the Bearer token on every request using the provided
// TokenVerifier. On success, claims are injected into the request context.
func RequireAuth(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			ctx := context.WithValue(r.Context(), claimsKey, &Claims{
				Subject:  result.Subject,
				Username: result.Username,
				Email:    result.Email,
				Roles:    result.Roles,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole ensures the caller has at least one of the specified roles.
// Must be used downstream of RequireAuth.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			for _, role := range claims.Roles {
				if allowed[role] {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
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
