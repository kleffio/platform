package domain

import "time"

// Role constants for project membership.
const (
	RoleOwner      = "owner"
	RoleMaintainer = "maintainer"
	RoleDeveloper  = "developer"
	RoleViewer     = "viewer"
)

// ValidRoles is the ordered set of project roles from least to most privileged.
var ValidRoles = []string{RoleViewer, RoleDeveloper, RoleMaintainer, RoleOwner}

// RoleRank returns a numeric rank for a role (higher = more privileged).
func RoleRank(role string) int {
	for i, r := range ValidRoles {
		if r == role {
			return i
		}
	}
	return -1
}

// ProjectMember is a user with a role in a project.
type ProjectMember struct {
	ProjectID   string    `json:"project_id"`
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	InvitedBy   string    `json:"invited_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// ProjectInvite is a pending email invitation to a project.
type ProjectInvite struct {
	ID           string     `json:"id"`
	ProjectID    string     `json:"project_id"`
	InvitedEmail string     `json:"invited_email"`
	Role         string     `json:"role"`
	Token        string     `json:"token,omitempty"`
	TokenHash    string     `json:"-"`
	InvitedBy    string     `json:"invited_by"`
	ExpiresAt    time.Time  `json:"expires_at"`
	AcceptedAt   *time.Time `json:"accepted_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type Project struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Slug           string    `json:"slug"`
	Name           string    `json:"name"`
	IsDefault      bool      `json:"is_default"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Connection describes a logical link between two workloads.
// Kind: "network" | "dependency" | "traffic"
type Connection struct {
	ID               string    `json:"id"`
	ProjectID        string    `json:"project_id"`
	SourceWorkloadID string    `json:"source_workload_id"`
	TargetWorkloadID string    `json:"target_workload_id"`
	Kind             string    `json:"kind"`
	Label            string    `json:"label"`
	CreatedAt        time.Time `json:"created_at"`
}

// GraphNode persists the canvas (x, y) position of a workload node.
type GraphNode struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"project_id"`
	WorkloadID string    `json:"workload_id"`
	PositionX  float64   `json:"position_x"`
	PositionY  float64   `json:"position_y"`
	UpdatedAt  time.Time `json:"updated_at"`
}
