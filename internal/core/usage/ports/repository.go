package ports

import (
	"context"

	"github.com/kleffio/platform/internal/core/usage/domain"
)

type UsageRepository interface {
	Save(ctx context.Context, record *domain.UsageRecord) error
	ListLatestByProject(ctx context.Context, projectID string) ([]*domain.WorkloadMetrics, error)
}
