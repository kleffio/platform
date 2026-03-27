package ports

import (
	"context"

	"github.com/kleffio/platform/internal/core/organizations/domain"
)

// OrganizationRepository is the persistence port for Organization aggregates.
type OrganizationRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Organization, error)
	FindBySlug(ctx context.Context, slug string) (*domain.Organization, error)
	ListByUserID(ctx context.Context, userID string) ([]*domain.Organization, error)
	Save(ctx context.Context, org *domain.Organization) error
	Delete(ctx context.Context, id string) error
}
