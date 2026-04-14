package domain

import "time"

// DeploymentStatus tracks the lifecycle of a deployment operation.
type DeploymentStatus string

const (
	DeploymentPending    DeploymentStatus = "pending"
	DeploymentInProgress DeploymentStatus = "in_progress"
	DeploymentSucceeded  DeploymentStatus = "succeeded"
	DeploymentFailed     DeploymentStatus = "failed"
	DeploymentRolledBack DeploymentStatus = "rolled_back"
)

// Deployment represents a single intent to deploy a game server version.
// It is a control-plane record — it describes desired state and tracks
// progress. The actual execution is handled by daemon workers.
type Deployment struct {
	ID             string
	OrganizationID string
	GameServerID   string
	ServerName     string           // human-readable name; becomes container/pod name
	BlueprintID    string
	Version        string
	Status         DeploymentStatus
	InitiatedBy    string // user ID
	StartedAt      time.Time
	FinishedAt     *time.Time
	FailureReason  string
	Address        string // host:port reported by daemon after provisioning
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// IsTerminal returns true if the deployment has reached a final state.
func (d *Deployment) IsTerminal() bool {
	return d.Status == DeploymentSucceeded ||
		d.Status == DeploymentFailed ||
		d.Status == DeploymentRolledBack
}
