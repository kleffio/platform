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
	ListByMember(ctx context.Context, userID string) ([]*domain.Project, error)
	Save(ctx context.Context, project *domain.Project) error

	// Connections (workload links)
	ListConnections(ctx context.Context, projectID string) ([]*domain.Connection, error)
	FindConnection(ctx context.Context, connectionID string) (*domain.Connection, error)
	CreateConnection(ctx context.Context, conn *domain.Connection) error
	DeleteConnection(ctx context.Context, connectionID string) error

	// Graph node positions (canvas layout)
	ListGraphNodes(ctx context.Context, projectID string) ([]*domain.GraphNode, error)
	UpsertGraphNode(ctx context.Context, node *domain.GraphNode) error

	// Project members
	ListMembers(ctx context.Context, projectID string) ([]*domain.ProjectMember, error)
	GetMember(ctx context.Context, projectID, userID string) (*domain.ProjectMember, error)
	AddMember(ctx context.Context, member *domain.ProjectMember) error
	UpdateMemberRole(ctx context.Context, projectID, userID, role string) error
	RemoveMember(ctx context.Context, projectID, userID string) error

	// Project invites
	ListInvites(ctx context.Context, projectID string) ([]*domain.ProjectInvite, error)
	FindInviteByToken(ctx context.Context, tokenHash string) (*domain.ProjectInvite, error)
	FindActiveInviteByEmail(ctx context.Context, projectID, email string) (*domain.ProjectInvite, error)
	CreateInvite(ctx context.Context, invite *domain.ProjectInvite) error
	AcceptInvite(ctx context.Context, tokenHash, userID, email, displayName string) (*domain.ProjectInvite, error)
	RevokeInvite(ctx context.Context, projectID, inviteID string) error
}
