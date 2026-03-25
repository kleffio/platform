package domain

import "time"

// Status values for a GameServer.
type Status string

const (
	StatusProvisioning Status = "provisioning"
	StatusRunning      Status = "running"
	StatusStopped      Status = "stopped"
	StatusDeleted      Status = "deleted"
	StatusError        Status = "error"
)

// GameServer is the platform's record of a provisioned server instance.
type GameServer struct {
	ID             string
	OrganizationID string
	OwnerID        string
	BlueprintID    string
	Name           string
	Status         Status
	NodeID         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
