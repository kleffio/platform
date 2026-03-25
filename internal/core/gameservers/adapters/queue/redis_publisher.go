// Package queue provides the Redis queue adapter that publishes daemon jobs.
// The wire format must exactly match the daemon's jobs.Job + payloads.ServerOperationPayload.
package queue

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/kleff/platform/internal/core/gameservers/ports"
)

const keyPending = "repo:queue:pending"

// daemonJob mirrors gameserver-daemon/internal/workers/jobs.Job.
type daemonJob struct {
	JobID       string          `json:"job_id"`
	JobType     string          `json:"job_type"`
	ResourceID  string          `json:"resource_id"`
	Payload     json.RawMessage `json:"payload"`
	Status      string          `json:"status"`
	Attempts    int             `json:"attempts"`
	MaxAttempts int             `json:"max_attempts"`
	CreatedAt   time.Time       `json:"created_at"`
}

// daemonPayload mirrors gameserver-daemon/internal/workers/payloads.ServerOperationPayload.
type daemonPayload struct {
	OwnerID          string            `json:"owner_id"`
	ServerID         string            `json:"server_id"`
	BlueprintID      string            `json:"blueprint_id"`
	Image            string            `json:"image"`
	EnvOverrides     map[string]string `json:"env_overrides,omitempty"`
	MemoryBytes      int64             `json:"memory_bytes,omitempty"`
	CPUMillicores    int64             `json:"cpu_millicores,omitempty"`
	PortRequirements []portRequirement `json:"port_requirements,omitempty"`
}

type portRequirement struct {
	TargetPort int    `json:"target_port"`
	Protocol   string `json:"protocol"`
}

// RedisPublisher publishes daemon jobs to the Redis queue.
type RedisPublisher struct {
	client *redis.Client
}

func NewRedisPublisher(redisURL, password string, useTLS bool) (*RedisPublisher, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	opts.Password = password
	if useTLS {
		opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &RedisPublisher{client: client}, nil
}

func (p *RedisPublisher) Publish(ctx context.Context, job ports.ServerJob) error {
	var portReqs []portRequirement
	for _, pr := range job.PortRequirements {
		portReqs = append(portReqs, portRequirement{
			TargetPort: pr.TargetPort,
			Protocol:   pr.Protocol,
		})
	}

	payload := daemonPayload{
		OwnerID:          job.OwnerID,
		ServerID:         job.ServerID,
		BlueprintID:      job.BlueprintID,
		Image:            job.Image,
		EnvOverrides:     job.EnvOverrides,
		MemoryBytes:      job.MemoryBytes,
		CPUMillicores:    job.CPUMillicores,
		PortRequirements: portReqs,
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	djob := daemonJob{
		JobID:       uuid.NewString(),
		JobType:     string(job.JobType),
		ResourceID:  job.ServerID,
		Payload:     rawPayload,
		Status:      "pending",
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now().UTC(),
	}

	data, err := json.Marshal(djob)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}

	if err := p.client.LPush(ctx, keyPending, data).Err(); err != nil {
		return fmt.Errorf("lpush to %s: %w", keyPending, err)
	}

	return nil
}
