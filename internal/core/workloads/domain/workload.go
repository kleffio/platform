package domain

import "time"

type WorkloadState string

const (
	WorkloadPending WorkloadState = "pending"
	WorkloadRunning WorkloadState = "running"
	WorkloadStopped WorkloadState = "stopped"
	WorkloadDeleted WorkloadState = "deleted"
	WorkloadFailed  WorkloadState = "failed"
)

type Workload struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	OrganizationID string        `json:"organization_id"`
	ProjectID      string        `json:"project_id"`
	OwnerID        string        `json:"owner_id"`
	BlueprintID    string        `json:"blueprint_id"`
	Image          string        `json:"image"`
	RuntimeRef     string        `json:"runtime_ref"`
	Endpoint       string        `json:"endpoint"`
	NodeID         string        `json:"node_id"`
	State          WorkloadState `json:"state"`
	ErrorMessage   string        `json:"error_message"`
	CPUMillicores  int64         `json:"cpu_millicores"`
	MemoryBytes    int64         `json:"memory_bytes"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

type DaemonStatusUpdate struct {
	WorkloadID    string
	Status        WorkloadState
	RuntimeRef    string
	Endpoint      string
	NodeID        string
	ErrorMessage  string
	ObservedAt    time.Time
	CPUMillicores int64
	MemoryMB      int64
	NetworkRxMB   float64
	NetworkTxMB   float64
	DiskReadMB    float64
	DiskWriteMB   float64
}

// WorkloadStatusChanged is emitted after daemon callbacks are persisted.
type WorkloadStatusChanged struct {
	WorkloadID string
	Status     WorkloadState
	NodeID     string
	Endpoint   string
}

func (e WorkloadStatusChanged) EventName() string {
	return "workload.status_changed"
}
