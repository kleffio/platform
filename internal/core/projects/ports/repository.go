package ports

import (
	"context"

	"github.com/kleffio/platform/internal/core/projects/domain"
)

type ProjectRepository interface {
	// Core project CRUD
	EnsureOrganization(ctx context.Context, organizationID, name string) error
	FindByID(ctx context.Context, id string) (*domain.Project, error)
	FindBySlug(ctx context.Context, organizationID, slug string) (*domain.Project, error)
	ListByOrganization(ctx context.Context, organizationID string) ([]*domain.Project, error)
	Save(ctx context.Context, project *domain.Project) error

	// Connections (workload links)
	ListConnections(ctx context.Context, projectID string) ([]*domain.Connection, error)
	FindConnection(ctx context.Context, connectionID string) (*domain.Connection, error)
	CreateConnection(ctx context.Context, conn *domain.Connection) error
	DeleteConnection(ctx context.Context, connectionID string) error

	// Graph node positions (canvas layout)
	ListGraphNodes(ctx context.Context, projectID string) ([]*domain.GraphNode, error)
	UpsertGraphNode(ctx context.Context, node *domain.GraphNode) error
}
