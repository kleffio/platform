package ports

import (
	"context"

	"github.com/kleffio/platform/internal/core/billing/domain"
)

// SubscriptionRepository is the persistence port for Subscription aggregates.
type SubscriptionRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Subscription, error)
	FindByOrganizationID(ctx context.Context, orgID string) (*domain.Subscription, error)
	Save(ctx context.Context, s *domain.Subscription) error
}

// InvoiceRepository manages billing invoices.
type InvoiceRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Invoice, error)
	ListByOrganizationID(ctx context.Context, orgID string, page, limit int) ([]*domain.Invoice, int, error)
	Save(ctx context.Context, inv *domain.Invoice) error
}
