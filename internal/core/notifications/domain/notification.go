// Package domain contains the pure business types for the notifications module.
package domain

import "time"

// Type classifies what kind of event a notification represents.
type Type string

const (
	TypeSystem            Type = "system"
	TypeBilling           Type = "billing"
	TypeOrgInvitation     Type = "org_invitation"
	TypeProjectInvitation Type = "project_invitation"
	TypeDeployment        Type = "deployment"
	TypeWorkload          Type = "workload"
	TypeSecurity          Type = "security"
)

// Notification is a single inbox item for a user.
type Notification struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	Type      Type           `json:"type"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Data      map[string]any `json:"data,omitempty"`
	ReadAt    *time.Time     `json:"read_at,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// IsRead returns true when the user has already read this notification.
func (n *Notification) IsRead() bool {
	return n.ReadAt != nil
}

// ListFilter controls which notifications are returned.
type ListFilter struct {
	UnreadOnly bool
	Limit      int
	Offset     int
}
