package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// JobType matches the daemon's JobType constants.
type JobType string

const (
	JobTypeServerProvision JobType = "server.provision"
	JobTypeServerStart     JobType = "server.start"
	JobTypeServerStop      JobType = "server.stop"
	JobTypeServerRestart   JobType = "server.restart"
	JobTypeServerDelete    JobType = "server.delete"
)

// JobStatus matches daemon constants.
type JobStatus string

const (
	JobStatusPending JobStatus = "pending"
)

// Job mirrors the daemon's jobs.Job struct for JSON compatibility.
// The daemon deserializes this directly, so field names must match exactly.
type Job struct {
	JobID       string          `json:"job_id"`
	JobType     JobType         `json:"job_type"`
	ResourceID  string          `json:"resource_id"`
	Payload     json.RawMessage `json:"payload"`
	Status      JobStatus       `json:"status"`
	Attempts    int             `json:"attempts"`
	MaxAttempts int             `json:"max_attempts"`
	CreatedAt   time.Time       `json:"created_at"`
}

// WorkloadSpec mirrors the daemon's ports.WorkloadSpec for JSON compatibility.
type WorkloadSpec struct {
	OwnerID          string            `json:"owner_id"`
	ServerID         string            `json:"server_id"`
	BlueprintID      string            `json:"blueprint_id"`
	Image            string            `json:"image"`
	BlueprintVersion string            `json:"blueprint_version,omitempty"`
	EnvOverrides     map[string]string `json:"env_overrides,omitempty"`
	MemoryBytes      int64             `json:"memory_bytes,omitempty"`
	CPUMillicores    int64             `json:"cpu_millicores,omitempty"`
	PortRequirements []PortRequirement `json:"port_requirements,omitempty"`
	RuntimeHints     RuntimeHints      `json:"runtime_hints,omitempty"`
}

// PortRequirement mirrors the daemon's ports.PortRequirement.
type PortRequirement struct {
	TargetPort int    `json:"target_port"`
	Protocol   string `json:"protocol"`
}

// RuntimeHints mirrors the daemon's ports.RuntimeHints.
type RuntimeHints struct {
	KubernetesStrategy string `json:"kubernetes_strategy,omitempty"`
	ExposeUDP          bool   `json:"expose_udp,omitempty"`
	PersistentStorage  bool   `json:"persistent_storage,omitempty"`
	StoragePath        string `json:"storage_path,omitempty"`
	StorageGB          int    `json:"storage_gb,omitempty"`
}

// Enqueuer can push jobs onto the daemon's work queue.
type Enqueuer interface {
	Enqueue(ctx context.Context, jobID string, spec WorkloadSpec) error
	EnqueueAction(ctx context.Context, jobID string, jobType JobType, spec WorkloadSpec) error
}

// newJob builds a Job ready to be serialized and pushed to Redis.
func newJob(jobID string, jobType JobType, resourceID string, payload any) (*Job, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	return &Job{
		JobID:       jobID,
		JobType:     jobType,
		ResourceID:  resourceID,
		Payload:     raw,
		Status:      JobStatusPending,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now().UTC(),
	}, nil
}
