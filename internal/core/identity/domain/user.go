package domain

import (
	"time"
)

// UserRole defines what a user can do within an organization.
type UserRole string

const (
	RoleOwner   UserRole = "owner"
	RoleAdmin   UserRole = "admin"
	RoleMember  UserRole = "member"
	RoleBilling UserRole = "billing"
	RoleViewer  UserRole = "viewer"
)

// User is the identity aggregate root. It represents an authenticated person
// who can belong to one or more organizations.
type User struct {
	ID          string
	Email       string
	DisplayName string
	AvatarURL   string
	ExternalID  string // OIDC subject claim
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsValid returns true if the user has the minimum required fields.
func (u *User) IsValid() bool {
	return u.ID != "" && u.Email != ""
}
