package orchestrator

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/franco/payment-api/internal/domain/payment"
	"github.com/franco/payment-api/internal/domain/shared"
	"github.com/google/uuid"
)

// EventPublisher publishes events
type EventPublisher interface {
	Publish(ctx context.Context, event shared.Event, topicArn string) error
}

// ExternalGatewayMock simulates an external payment gateway
type ExternalGatewayMock struct {
	eventStore     shared.EventStore
	eventPublisher EventPublisher
	topicArn       string
	alwaysSuccess  bool
}

// NewExternalGatewayMock creates a new ExternalGatewayMock
func NewExternalGatewayMock(
	eventStore shared.EventStore,
	eventPublisher EventPublisher,
	topicArn string,
	alwaysSuccess bool,
) *ExternalGatewayMock {
	return &ExternalGatewayMock{
		eventStore:     eventStore,
		eventPublisher: eventPublisher,
		topicArn:       topicArn,
		alwaysSuccess:  alwaysSuccess,
	}
}

// HandleExternalPaymentRequested processes external payment requests
func (g *ExternalGatewayMock) HandleExternalPaymentRequested(ctx context.Context, event shared.Event) error {
	externalEvent, ok := event.(*payment.ExternalPaymentRequestedEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", event)
	}

	// Simulate processing
	var resultEvent shared.Event

	if g.alwaysSuccess {
		// Always succeed
		resultEvent = payment.NewExternalPaymentSucceededEvent(
			externalEvent.PaymentID(),
			uuid.New().String(), // external transaction ID
			event.Metadata(),
		)
	} else {
		// Random success/failure
		if rand.Float32() < 0.8 { // 80% success rate
			resultEvent = payment.NewExternalPaymentSucceededEvent(
				externalEvent.PaymentID(),
				uuid.New().String(),
				event.Metadata(),
			)
		} else {
			resultEvent = payment.NewExternalPaymentFailedEvent(
				externalEvent.PaymentID(),
				"GATEWAY_REJECTED",
				"ERR_RANDOM_FAILURE",
				event.Metadata(),
			)
		}
	}

	// Store event
	if err := g.eventStore.Append(ctx, resultEvent, externalEvent.PaymentID()); err != nil {
		return err
	}

	// Publish result event
	return g.eventPublisher.Publish(ctx, resultEvent, g.topicArn)
}
