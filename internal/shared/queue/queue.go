package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kleffio/platform/internal/shared/ids"
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

const (
	pendingListKey = "repo:queue:pending"
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
	ProjectID        string            `json:"project_id"`
	ProjectSlug      string            `json:"project_slug"`
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

// Publisher pushes jobs into the daemon queue backend.
type Publisher interface {
	Enqueue(ctx context.Context, job *Job) error
}

// Enqueuer can push jobs onto the daemon's work queue.
type Enqueuer interface {
	Enqueue(ctx context.Context, jobID string, spec WorkloadSpec) error
	EnqueueAction(ctx context.Context, jobID string, jobType JobType, spec WorkloadSpec) error
}

// NopPublisher is used when queue config is not set.
type NopPublisher struct{}

func (NopPublisher) Enqueue(_ context.Context, _ *Job) error {
	return fmt.Errorf("daemon queue publisher is not configured")
}

// NopEnqueuer is used when queue config is not set.
type NopEnqueuer struct{}

func (NopEnqueuer) Enqueue(_ context.Context, _ string, _ WorkloadSpec) error {
	return fmt.Errorf("daemon queue enqueuer is not configured")
}

func (NopEnqueuer) EnqueueAction(_ context.Context, _ string, _ JobType, _ WorkloadSpec) error {
	return fmt.Errorf("daemon queue enqueuer is not configured")
}

func NewJob(jobType JobType, resourceID string, payload any, maxAttempts int) (*Job, error) {
	return newJobWithID("", jobType, resourceID, payload, maxAttempts)
}

func PendingListKey() string {
	return pendingListKey
}

func newJobWithID(jobID string, jobType JobType, resourceID string, payload any, maxAttempts int) (*Job, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal job payload: %w", err)
	}
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	if jobID == "" {
		jobID = ids.New()
	}
	return &Job{
		JobID:       jobID,
		JobType:     jobType,
		ResourceID:  resourceID,
		Payload:     raw,
		Status:      JobStatusPending,
		Attempts:    0,
		MaxAttempts: maxAttempts,
		CreatedAt:   time.Now().UTC(),
	}, nil
}
