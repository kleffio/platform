package ports

import "context"

// JobType mirrors the daemon's job type constants.
type JobType string

const (
	JobTypeProvision JobType = "server.provision"
	JobTypeStart     JobType = "server.start"
	JobTypeStop      JobType = "server.stop"
	JobTypeDelete    JobType = "server.delete"
	JobTypeRestart   JobType = "server.restart"
)

// PortRequirement describes a port a server container needs to expose.
type PortRequirement struct {
	TargetPort int
	Protocol   string
}

// ServerJob is the payload the platform publishes to the daemon's Redis queue.
// It must match the ServerOperationPayload the daemon expects exactly.
type ServerJob struct {
	JobType          JobType
	OwnerID          string
	ServerID         string
	BlueprintID      string
	Image            string
	EnvOverrides     map[string]string
	MemoryBytes      int64
	CPUMillicores    int64
	PortRequirements []PortRequirement
}

// GameServerQueue publishes jobs for the daemon to consume.
type GameServerQueue interface {
	Publish(ctx context.Context, job ServerJob) error
}
