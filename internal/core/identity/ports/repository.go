package ports

import (
	"context"

	"github.com/kleff/platform/internal/core/identity/domain"
)

// UserRepository defines the persistence contract for User aggregates.
type UserRepository interface {
	// FindByID returns the user with the given ID, or ErrNotFound.
	FindByID(ctx context.Context, id string) (*domain.User, error)
	// FindByExternalID returns the user matching the OIDC subject claim.
	FindByExternalID(ctx context.Context, externalID string) (*domain.User, error)
	// FindByEmail returns the user with the given email, or ErrNotFound.
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	// Save creates or updates a user.
	Save(ctx context.Context, user *domain.User) error
}

// OrganizationRepository defines the persistence contract for Organization aggregates.
type OrganizationRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Organization, error)
	FindBySlug(ctx context.Context, slug string) (*domain.Organization, error)
	ListByUserID(ctx context.Context, userID string) ([]*domain.Organization, error)
	Save(ctx context.Context, org *domain.Organization) error
}

// MembershipRepository manages org membership records.
type MembershipRepository interface {
	FindMembership(ctx context.Context, userID, orgID string) (*domain.Membership, error)
	ListMembers(ctx context.Context, orgID string) ([]*domain.Membership, error)
	Save(ctx context.Context, m *domain.Membership) error
	Delete(ctx context.Context, userID, orgID string) error
}
