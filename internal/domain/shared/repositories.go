package shared

import "context"

// NOTE: Repository interfaces are now defined in application/orchestrator
// to avoid circular dependencies. This file can be removed.

// IdempotencyStore defines operations for idempotency tracking
type IdempotencyStore interface {
	GetPaymentIDByKey(ctx context.Context, idempotencyKey string) (string, error)
	SaveKey(ctx context.Context, idempotencyKey, paymentID string) error
}

// EventStore defines operations for event persistence
type EventStore interface {
	Append(ctx context.Context, event Event, paymentID string) error
	ListByPaymentID(ctx context.Context, paymentID string) ([]StoredEvent, error)
}

// StoredEvent represents a persisted event
type StoredEvent struct {
	EventID    string
	EventType  string
	PaymentID  string
	Payload    string
	OccurredAt string
	Metadata   string
}
