package domain

import "time"

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
