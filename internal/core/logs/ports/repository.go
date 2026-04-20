package ports

import (
	"context"

	"github.com/kleffio/platform/internal/core/logs/domain"
)

// LogRepository persists and retrieves workload log lines.
// Implementations of this interface are the extension point for log-sink
// plugins (Loki, Elasticsearch, etc.) — a plugin can wrap or replace the
// default Postgres store by implementing this interface.
type LogRepository interface {
	// SaveBatch persists a batch of log lines. Implementations should be
	// idempotent or at least tolerant of duplicate lines on retry.
	SaveBatch(ctx context.Context, lines []*domain.LogLine) error

	// ListByWorkload returns up to limit lines for workloadID ordered by ts DESC.
	ListByWorkload(ctx context.Context, workloadID string, limit int) ([]*domain.LogLine, error)
}
