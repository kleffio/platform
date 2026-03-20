package domain

import "time"

// InvoiceStatus reflects the lifecycle of an invoice.
type InvoiceStatus string

const (
	InvoiceDraft         InvoiceStatus = "draft"
	InvoiceOpen          InvoiceStatus = "open"
	InvoicePaid          InvoiceStatus = "paid"
	InvoiceVoid          InvoiceStatus = "void"
	InvoiceUncollectible InvoiceStatus = "uncollectible"
)

// Invoice is an immutable billing record.
type Invoice struct {
	ID             string
	OrganizationID string
	SubscriptionID string
	Status         InvoiceStatus
	Number         string
	Lines          []InvoiceLineItem
	Subtotal       int // cents
	Tax            int // cents
	Total          int // cents
	Currency       string
	DueDate        *time.Time
	PaidAt         *time.Time
	ExternalID     string // e.g. Stripe invoice ID
	CreatedAt      time.Time
}

// InvoiceLineItem is a single line on an invoice.
type InvoiceLineItem struct {
	ID          string
	Description string
	Quantity    int
	UnitAmount  int // cents
	TotalAmount int // cents
}
