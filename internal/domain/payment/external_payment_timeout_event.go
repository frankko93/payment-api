package payment

import (
	"time"

	"github.com/franco/payment-api/internal/domain/shared"
)

// ExternalPaymentTimeoutEvent is emitted when external gateway times out
type ExternalPaymentTimeoutEvent struct {
	shared.BaseEvent
	paymentID       string
	timeoutDuration time.Duration
}

// NewExternalPaymentTimeoutEvent creates a new ExternalPaymentTimeoutEvent
func NewExternalPaymentTimeoutEvent(
	paymentID string,
	timeoutDuration time.Duration,
	metadata shared.Metadata,
) *ExternalPaymentTimeoutEvent {
	return &ExternalPaymentTimeoutEvent{
		BaseEvent:       shared.NewBaseEvent("ExternalPaymentTimeout", metadata),
		paymentID:       paymentID,
		timeoutDuration: timeoutDuration,
	}
}

func (e *ExternalPaymentTimeoutEvent) PaymentID() string {
	return e.paymentID
}

func (e *ExternalPaymentTimeoutEvent) TimeoutDuration() time.Duration {
	return e.timeoutDuration
}
