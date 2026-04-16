package queue

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisPublisher struct {
	client *redis.Client
}

func NewRedisPublisher(url, password string, useTLS bool) (*RedisPublisher, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	if password != "" {
		opts.Password = password
	}
	if useTLS {
		opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}
	return &RedisPublisher{client: client}, nil
}

func (p *RedisPublisher) Enqueue(ctx context.Context, job *Job) error {
	if job == nil {
		return fmt.Errorf("job is required")
	}
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal queue job: %w", err)
	}
	if err := p.client.LPush(ctx, PendingListKey(), payload).Err(); err != nil {
		return fmt.Errorf("enqueue queue job: %w", err)
	}
	return nil
}

type RedisEnqueuer struct {
	publisher Publisher
}

func NewRedisEnqueuer(url, password string, useTLS bool) (*RedisEnqueuer, error) {
	publisher, err := NewRedisPublisher(url, password, useTLS)
	if err != nil {
		return nil, err
	}
	return &RedisEnqueuer{publisher: publisher}, nil
}

func (e *RedisEnqueuer) Enqueue(ctx context.Context, jobID string, spec WorkloadSpec) error {
	return e.EnqueueAction(ctx, jobID, JobTypeServerProvision, spec)
}

func (e *RedisEnqueuer) EnqueueAction(ctx context.Context, jobID string, jobType JobType, spec WorkloadSpec) error {
	job, err := newJobWithID(jobID, jobType, spec.ServerID, spec, 5)
	if err != nil {
		return fmt.Errorf("create job: %w", err)
	}
	if err := e.publisher.Enqueue(ctx, job); err != nil {
		return fmt.Errorf("enqueue job: %w", err)
	}
	return nil
}
