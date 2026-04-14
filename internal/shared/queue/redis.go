package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// keyPending is the Redis list key the daemon reads from.
// Must match the daemon's internal constant "repo:queue:pending".
const keyPending = "repo:queue:pending"

// RedisEnqueuer pushes provision jobs onto the daemon's Redis queue.
type RedisEnqueuer struct {
	client *redis.Client
}

func NewRedisEnqueuer(url string) (*RedisEnqueuer, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect to redis: %w", err)
	}

	return &RedisEnqueuer{client: client}, nil
}

func (e *RedisEnqueuer) Enqueue(ctx context.Context, jobID string, spec WorkloadSpec) error {
	return e.EnqueueAction(ctx, jobID, JobTypeServerProvision, spec)
}

func (e *RedisEnqueuer) EnqueueAction(ctx context.Context, jobID string, jobType JobType, spec WorkloadSpec) error {
	job, err := newJob(jobID, jobType, spec.ServerID, spec)
	if err != nil {
		return fmt.Errorf("create job: %w", err)
	}

	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("serialize job: %w", err)
	}

	if err := e.client.LPush(ctx, keyPending, data).Err(); err != nil {
		return fmt.Errorf("redis lpush: %w", err)
	}

	return nil
}
