package middleware

import "context"

// TokenVerifier validates a bearer token and returns its claims.
// Implemented by both the OIDC JWT validator and the Hydra introspector.
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (*VerifyResult, error)
}

// VerifyResult holds the verified claims from a token.
type VerifyResult struct {
	Subject string
	Roles   []string
}
