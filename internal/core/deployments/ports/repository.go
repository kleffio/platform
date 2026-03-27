package ports

import (
	"context"

	"github.com/kleffio/platform/internal/core/deployments/domain"
)

// DeploymentRepository is the persistence port for Deployment records.
type DeploymentRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Deployment, error)
	ListByGameServer(ctx context.Context, gameServerID string, page int, limit int) ([]*domain.Deployment, int, error)
	ListByOrganization(ctx context.Context, orgID string, page int, limit int) ([]*domain.Deployment, int, error)
	Save(ctx context.Context, d *domain.Deployment) error
}
