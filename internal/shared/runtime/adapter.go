// Package runtime defines the RuntimeAdapter interface that abstracts container
// lifecycle operations across Docker, Kubernetes, and manual deployments.
// The same interface drives plugin container management (platform) and game
// server container management (kleff-agent).
package runtime

import (
	"context"
	"time"
)

// RuntimeAdapter abstracts container lifecycle operations.
// Implementations must be safe for concurrent use.
type RuntimeAdapter interface {
	// Deploy pulls the image (if not present) and starts a container according
	// to spec. If a container with spec.ID already exists, it is replaced.
	// Deploy is idempotent.
	Deploy(ctx context.Context, spec ContainerSpec) error

	// Remove stops and removes the container with the given ID.
	// Returns nil if the container does not exist.
	Remove(ctx context.Context, id string) error

	// Start starts a stopped container. No-op if already running.
	Start(ctx context.Context, id string) error

	// Stop stops a running container gracefully (SIGTERM, then SIGKILL after
	// timeout). No-op if already stopped.
	Stop(ctx context.Context, id string) error

	// Status returns the current state of the container.
	Status(ctx context.Context, id string) (ContainerStatus, error)

	// Endpoint returns the host:port at which the container's primary port can
	// be reached from within the platform API process.
	// Docker: uses container-name DNS ("kleff-idp-auth0:50051").
	// Kubernetes: uses Service DNS.
	Endpoint(ctx context.Context, id string, port int) (string, error)

	// Logs returns the last n lines of container stdout+stderr.
	Logs(ctx context.Context, id string, lines int) ([]string, error)
}

// ContainerSpec describes a container to be deployed.
type ContainerSpec struct {
	// ID is the unique name/identifier for this container.
	// Convention: "kleff-{plugin-id}", e.g. "kleff-idp-keycloak"
	ID string

	// Image is the fully qualified Docker image reference.
	Image string

	// Command overrides the container's default CMD (optional).
	Command []string

	// Env is the set of environment variables injected into the container.
	// Secret values are resolved before being passed here.
	Env map[string]string

	// Ports maps container port → host port.
	// Leave HostPort as 0 to let the runtime assign a random port.
	Ports []PortMapping

	// Labels are key/value metadata attached to the container.
	// The platform always sets "kleff.io/managed=true" and
	// "kleff.io/plugin-id={id}".
	Labels map[string]string

	// Volumes declares named Docker volumes to mount into the container.
	Volumes []VolumeMount

	// User overrides the container's default user (e.g. "root", "1000", "1000:1000").
	// Empty means use the image default.
	User string

	// Resources constrains CPU and memory (optional, 0 = unlimited).
	Resources ResourceLimits

	// RestartPolicy controls container restart behaviour.
	RestartPolicy RestartPolicy
}

// VolumeMount maps a named Docker volume to a path inside the container.
type VolumeMount struct {
	Name   string // Docker volume name (created automatically if absent)
	Target string // Absolute path inside the container
}

// PortMapping maps a container port to an optional host port.
type PortMapping struct {
	ContainerPort int
	HostPort      int    // 0 = auto-assign
	Protocol      string // "tcp" or "udp"
}

// ResourceLimits constrains CPU and memory usage.
type ResourceLimits struct {
	CPUMillicores int64 // 0 = unlimited. e.g. 500 = 0.5 CPU
	MemoryMB      int64 // 0 = unlimited
}

// RestartPolicy controls container restart behaviour.
type RestartPolicy string

const (
	RestartAlways    RestartPolicy = "always"
	RestartOnFailure RestartPolicy = "on-failure"
	RestartNever     RestartPolicy = "never"
)

// ContainerStatus describes the current state of a deployed container.
type ContainerStatus struct {
	ID      string
	State   ContainerState
	Image   string
	Since   time.Time // when the container entered the current state
	Message string    // optional human-readable detail (e.g. exit reason)
}

// ContainerState enumerates the possible container states.
type ContainerState string

const (
	StateRunning  ContainerState = "running"
	StateStopped  ContainerState = "stopped"
	StateStarting ContainerState = "starting"
	StateFailed   ContainerState = "failed"
	StateUnknown  ContainerState = "unknown"
	StateNotFound ContainerState = "not_found"
)
