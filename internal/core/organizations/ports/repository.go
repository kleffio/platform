package ports

import (
	"context"

	"github.com/kleffio/platform/internal/core/organizations/domain"
)

// OrganizationRepository is the persistence port for Organization aggregates.
type OrganizationRepository interface {
	// Org CRUD
	FindByID(ctx context.Context, id string) (*domain.Organization, error)
	FindBySlug(ctx context.Context, slug string) (*domain.Organization, error)
	Save(ctx context.Context, org *domain.Organization) error
	Update(ctx context.Context, org *domain.Organization) error
	Delete(ctx context.Context, id string) error

	// Membership
	ListByUserID(ctx context.Context, userID string) ([]*domain.Organization, error)
	ListMembers(ctx context.Context, orgID string) ([]*domain.Member, error)
	GetMember(ctx context.Context, orgID, userID string) (*domain.Member, error)
	FindMemberByEmail(ctx context.Context, email string) (*domain.Member, error)
	AddMember(ctx context.Context, member *domain.Member) error
	UpdateMemberRole(ctx context.Context, orgID, userID, role string) error
	RemoveMember(ctx context.Context, orgID, userID string) error
	CountOwners(ctx context.Context, orgID string) (int, error)

	// Bootstrap: upsert org row + add caller as owner if not already a member.
	// Used to migrate existing "org-<slug>" orgs on first login.
	EnsureOrgWithOwner(ctx context.Context, orgID, orgName, userID, email, displayName string) error

	// Invites
	CreateInvite(ctx context.Context, invite *domain.Invite) error
	FindInviteByToken(ctx context.Context, tokenHash string) (*domain.Invite, error)
	ListInvites(ctx context.Context, orgID string) ([]*domain.Invite, error)
	AcceptInvite(ctx context.Context, inviteID, userID, email, displayName string) error
	RevokeInvite(ctx context.Context, inviteID string) error
}
