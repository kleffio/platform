package ports

import (
	"context"

	"github.com/kleffio/platform/internal/core/workloads/domain"
)

type Repository interface {
	CreateWorkload(ctx context.Context, workload *domain.Workload) error
	FindByID(ctx context.Context, workloadID string) (*domain.Workload, error)
	FindByProjectAndName(ctx context.Context, projectID, name string) (*domain.Workload, error)
	ListByProject(ctx context.Context, projectID string) ([]*domain.Workload, error)
	SaveDeployment(ctx context.Context, deployment *DeploymentRecord) error
	DeleteWorkload(ctx context.Context, workloadID string) error
	UpdateState(ctx context.Context, workloadID string, state domain.WorkloadState, errorMessage string) error
	UpdateFromDaemon(ctx context.Context, update domain.DaemonStatusUpdate) error
}

type DeploymentRecord struct {
	ID             string
	OrganizationID string
	ProjectID      string
	WorkloadID     string
	Action         string
	Status         string
	InitiatedBy    string
}
