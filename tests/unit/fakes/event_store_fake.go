package fakes

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/franco/payment-api/internal/domain/shared"
	"github.com/google/uuid"
)

// EventStoreFake is a fake implementation of EventStore for testing
type EventStoreFake struct {
	mu     sync.RWMutex
	events map[string][]shared.StoredEvent
}

// NewEventStoreFake creates a new EventStoreFake
func NewEventStoreFake() *EventStoreFake {
	return &EventStoreFake{
		events: make(map[string][]shared.StoredEvent),
	}
}

// Append stores an event
func (f *EventStoreFake) Append(ctx context.Context, event shared.Event, paymentID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	payloadBytes, _ := json.Marshal(event)
	metadataBytes, _ := json.Marshal(event.Metadata())

	storedEvent := shared.StoredEvent{
		EventID:    uuid.New().String(),
		EventType:  event.EventType(),
		PaymentID:  paymentID,
		Payload:    string(payloadBytes),
		OccurredAt: event.OccurredAt().String(),
		Metadata:   string(metadataBytes),
	}

	f.events[paymentID] = append(f.events[paymentID], storedEvent)
	return nil
}

// ListByPaymentID retrieves all events for a payment
func (f *EventStoreFake) ListByPaymentID(ctx context.Context, paymentID string) ([]shared.StoredEvent, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	events, exists := f.events[paymentID]
	if !exists {
		return []shared.StoredEvent{}, nil
	}

	return events, nil
}
