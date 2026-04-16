package middleware

import (
	"context"
	"net/http"
	"strings"
)

const nodeClaimsKey contextKey = "node_claims"

type NodeClaims struct {
	NodeID string
}

type NodeTokenVerifier interface {
	VerifyNodeToken(ctx context.Context, rawToken string) (string, error)
}

// NodeClaimsFromContext returns node auth claims injected by RequireNodeAuth.
func NodeClaimsFromContext(ctx context.Context) (*NodeClaims, bool) {
	claims, ok := ctx.Value(nodeClaimsKey).(*NodeClaims)
	return claims, ok
}

// RequireNodeBootstrap protects node registration with a shared bootstrap secret.
func RequireNodeBootstrap(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.TrimSpace(secret) == "" {
				http.Error(w, `{"error":"node bootstrap not configured"}`, http.StatusServiceUnavailable)
				return
			}
			token := extractBearer(r)
			if token == "" {
				token = r.Header.Get("X-Node-Secret")
			}
			if token == "" || token != secret {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireNodeAuth validates a node token and injects the node identity in context.
func RequireNodeAuth(verifier NodeTokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if verifier == nil {
				http.Error(w, `{"error":"node auth not configured"}`, http.StatusServiceUnavailable)
				return
			}
			token := extractBearer(r)
			if token == "" {
				token = r.Header.Get("X-Node-Token")
			}
			if token == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			nodeID, err := verifier.VerifyNodeToken(r.Context(), token)
			if err != nil || nodeID == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), nodeClaimsKey, &NodeClaims{NodeID: nodeID})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
