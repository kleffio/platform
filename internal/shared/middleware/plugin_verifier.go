package middleware

import (
	"context"
	"fmt"

	"github.com/kleffio/platform/internal/core/plugins/ports"
)

// PluginTokenVerifier implements TokenVerifier by delegating to the active
// IDP plugin's ValidateToken gRPC method.
type PluginTokenVerifier struct {
	manager ports.PluginManager
}

func NewPluginTokenVerifier(manager ports.PluginManager) *PluginTokenVerifier {
	return &PluginTokenVerifier{manager: manager}
}

func (v *PluginTokenVerifier) Verify(ctx context.Context, rawToken string) (*VerifyResult, error) {
	claims, err := v.manager.ValidateToken(ctx, rawToken)
	if err != nil {
		return nil, fmt.Errorf("token validation: %w", err)
	}
	roles := claims.Roles
	if roles == nil {
		roles = []string{}
	}
	return &VerifyResult{Subject: claims.Subject, Email: claims.Email, Roles: roles}, nil
}
