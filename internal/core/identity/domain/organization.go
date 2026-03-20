package domain

import "time"

// Organization is the top-level tenancy boundary. All resources belong to
// an organization.
type Organization struct {
	ID        string
	Name      string
	Slug      string
	LogoURL   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Membership links a User to an Organization with a specific Role.
type Membership struct {
	UserID         string
	OrganizationID string
	Role           UserRole
	JoinedAt       time.Time
}
