package application

import (
	"context"
	"fmt"

	"github.com/kleffio/platform/internal/core/nodes/ports"
)

type TokenVerifier struct {
	repo ports.NodeRepository
}

func NewTokenVerifier(repo ports.NodeRepository) *TokenVerifier {
	return &TokenVerifier{repo: repo}
}

func (v *TokenVerifier) VerifyNodeToken(ctx context.Context, rawToken string) (string, error) {
	hash := HashNodeToken(rawToken)
	node, err := v.repo.FindByTokenHash(ctx, hash)
	if err != nil {
		return "", fmt.Errorf("verify node token: %w", err)
	}
	return node.ID, nil
}
