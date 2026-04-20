package domain

import "time"

// Organization is an isolated tenant with its own members, projects, and
// billing. All user-owned resources are scoped to an organization.
type Organization struct {
	ID        string
	Name      string
	Slug      string
	LogoURL   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Role constants for organization membership.
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// Member represents a user's membership in an organization.
type Member struct {
	OrgID       string
	UserID      string
	Email       string
	DisplayName string
	Role        string
	CreatedAt   time.Time
}

// Invite is a pending email invitation to join an organization.
type Invite struct {
	ID           string
	OrgID        string
	InvitedEmail string
	Role         string
	Token        string // raw token (only available at creation time)
	TokenHash    string
	InvitedBy    string
	ExpiresAt    time.Time
	AcceptedAt   *time.Time
	CreatedAt    time.Time
}
