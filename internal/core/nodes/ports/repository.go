package ports

import (
	"context"

	"github.com/kleffio/platform/internal/core/nodes/domain"
)

// NodeRepository is the persistence port for Node aggregates.
type NodeRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Node, error)
	FindByHostname(ctx context.Context, hostname string) (*domain.Node, error)
	ListByRegion(ctx context.Context, region string) ([]*domain.Node, error)
	ListAll(ctx context.Context) ([]*domain.Node, error)
	Save(ctx context.Context, node *domain.Node) error
}
