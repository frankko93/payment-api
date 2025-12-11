package fakes

import (
	"context"
	"sync"

	"github.com/franco/payment-api/internal/domain/shared"
)

// EventPublisherFake is a fake implementation of EventPublisher for testing
type EventPublisherFake struct {
	mu              sync.RWMutex
	PublishedEvents []shared.Event
}

// NewEventPublisherFake creates a new EventPublisherFake
func NewEventPublisherFake() *EventPublisherFake {
	return &EventPublisherFake{
		PublishedEvents: make([]shared.Event, 0),
	}
}

// Publish captures an event
func (f *EventPublisherFake) Publish(ctx context.Context, event shared.Event, topicArn string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.PublishedEvents = append(f.PublishedEvents, event)
	return nil
}

// GetPublishedEvents returns all published events
func (f *EventPublisherFake) GetPublishedEvents() []shared.Event {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.PublishedEvents
}

// GetEventsByType returns events of a specific type
func (f *EventPublisherFake) GetEventsByType(eventType string) []shared.Event {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var events []shared.Event
	for _, event := range f.PublishedEvents {
		if event.EventType() == eventType {
			events = append(events, event)
		}
	}
	return events
}

// Reset clears all published events
func (f *EventPublisherFake) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.PublishedEvents = make([]shared.Event, 0)
}
