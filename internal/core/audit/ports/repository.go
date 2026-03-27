package ports

import (
	"context"

	"github.com/kleffio/platform/internal/core/audit/domain"
)

// AuditRepository is the write-only port for audit events.
// Audit events are append-only; updates and deletes are not permitted.
type AuditRepository interface {
	Append(ctx context.Context, event *domain.AuditEvent) error
	ListByOrganization(ctx context.Context, orgID string, page, limit int) ([]*domain.AuditEvent, int, error)
}
