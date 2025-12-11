package payment

import "github.com/franco/payment-api/internal/domain/shared"

// ExternalPaymentFailedEvent is emitted when external gateway rejects payment
type ExternalPaymentFailedEvent struct {
	shared.BaseEvent
	paymentID string
	reason    string
	errorCode string
}

// NewExternalPaymentFailedEvent creates a new ExternalPaymentFailedEvent
func NewExternalPaymentFailedEvent(
	paymentID, reason, errorCode string,
	metadata shared.Metadata,
) *ExternalPaymentFailedEvent {
	return &ExternalPaymentFailedEvent{
		BaseEvent: shared.NewBaseEvent("ExternalPaymentFailed", metadata),
		paymentID: paymentID,
		reason:    reason,
		errorCode: errorCode,
	}
}

func (e *ExternalPaymentFailedEvent) PaymentID() string {
	return e.paymentID
}

func (e *ExternalPaymentFailedEvent) Reason() string {
	return e.reason
}

func (e *ExternalPaymentFailedEvent) ErrorCode() string {
	return e.errorCode
}
