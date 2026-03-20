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
