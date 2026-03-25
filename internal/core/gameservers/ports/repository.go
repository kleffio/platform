package ports

import (
	"context"

	"github.com/kleff/platform/internal/core/gameservers/domain"
)

type GameServerRepository interface {
	Save(ctx context.Context, gs *domain.GameServer) error
	FindByID(ctx context.Context, id string) (*domain.GameServer, error)
	ListByOrg(ctx context.Context, orgID string) ([]*domain.GameServer, error)
	UpdateStatus(ctx context.Context, id string, status domain.Status) error
}
