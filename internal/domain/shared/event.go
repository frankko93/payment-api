package shared

import "time"

// Event represents the base interface for all domain events
type Event interface {
	EventType() string
	OccurredAt() time.Time
	Metadata() Metadata
}

// Metadata contains contextual information about the event
type Metadata struct {
	ClientID  string            `json:"clientId"`
	RequestID string            `json:"requestId"`
	Source    string            `json:"source"`
	Extra     map[string]string `json:"extra,omitempty"`
}

// BaseEvent provides common functionality for all events
type BaseEvent struct {
	eventType  string
	occurredAt time.Time
	metadata   Metadata
}

func (e BaseEvent) EventType() string {
	return e.eventType
}

func (e BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e BaseEvent) Metadata() Metadata {
	return e.metadata
}

// NewBaseEvent creates a new base event (exported for use in other packages)
func NewBaseEvent(eventType string, metadata Metadata) BaseEvent {
	return BaseEvent{
		eventType:  eventType,
		occurredAt: time.Now().UTC(),
		metadata:   metadata,
	}
}
