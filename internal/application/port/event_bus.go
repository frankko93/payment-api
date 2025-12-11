package port

import (
	"context"

	"github.com/franco/payment-api/internal/domain/shared"
)

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	Publish(ctx context.Context, event shared.Event, topicArn string) error
}

// EventConsumer defines the interface for consuming events
type EventConsumer interface {
	StartConsuming(queueURL string, handler func(ctx context.Context, event shared.Event) error)
}
