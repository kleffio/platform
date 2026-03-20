package domain

import "time"

// AuditEvent records a security-relevant action in the system.
// All state-mutating API calls should produce an audit event.
type AuditEvent struct {
	ID             string
	OrganizationID string
	ActorID        string // user ID or system identifier
	ActorType      string // "user" | "system" | "api_key"
	Action         string // e.g. "deployment.created", "user.invited"
	ResourceType   string // e.g. "deployment", "organization"
	ResourceID     string
	IPAddress      string
	UserAgent      string
	Outcome        AuditOutcome
	Metadata       map[string]any
	OccurredAt     time.Time
}

// AuditOutcome indicates whether the action succeeded or was denied.
type AuditOutcome string

const (
	AuditOutcomeSuccess AuditOutcome = "success"
	AuditOutcomeDenied  AuditOutcome = "denied"
	AuditOutcomeFailure AuditOutcome = "failure"
)
