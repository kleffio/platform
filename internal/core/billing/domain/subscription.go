package domain

import "time"

// SubscriptionStatus reflects the lifecycle of a billing subscription.
type SubscriptionStatus string

const (
	SubscriptionTrialing  SubscriptionStatus = "trialing"
	SubscriptionActive    SubscriptionStatus = "active"
	SubscriptionPastDue   SubscriptionStatus = "past_due"
	SubscriptionCanceled  SubscriptionStatus = "canceled"
	SubscriptionUnpaid    SubscriptionStatus = "unpaid"
	SubscriptionPaused    SubscriptionStatus = "paused"
)

// BillingInterval controls how frequently a subscription renews.
type BillingInterval string

const (
	BillingMonthly BillingInterval = "monthly"
	BillingYearly  BillingInterval = "yearly"
)

// Plan describes what a subscription tier includes.
type Plan struct {
	ID             string
	Tier           string
	Name           string
	Description    string
	PricePerMonth  int // in cents
	PricePerYear   int // in cents
	MaxGameServers int
	MaxTeamMembers int
	SupportLevel   string
}

// Subscription is the billing aggregate root.
type Subscription struct {
	ID                 string
	OrganizationID     string
	Plan               Plan
	Status             SubscriptionStatus
	Interval           BillingInterval
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CancelAtPeriodEnd  bool
	TrialEnd           *time.Time
	ExternalID         string // e.g. Stripe subscription ID
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// IsActive returns true if the subscription is in a state that grants access.
func (s *Subscription) IsActive() bool {
	return s.Status == SubscriptionActive || s.Status == SubscriptionTrialing
}
