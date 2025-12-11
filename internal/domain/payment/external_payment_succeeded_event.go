package payment

import "github.com/franco/payment-api/internal/domain/shared"

// ExternalPaymentSucceededEvent is emitted when external gateway confirms payment
type ExternalPaymentSucceededEvent struct {
	shared.BaseEvent
	paymentID             string
	externalTransactionID string
}

// NewExternalPaymentSucceededEvent creates a new ExternalPaymentSucceededEvent
func NewExternalPaymentSucceededEvent(
	paymentID, externalTransactionID string,
	metadata shared.Metadata,
) *ExternalPaymentSucceededEvent {
	return &ExternalPaymentSucceededEvent{
		BaseEvent:             shared.NewBaseEvent("ExternalPaymentSucceeded", metadata),
		paymentID:             paymentID,
		externalTransactionID: externalTransactionID,
	}
}

func (e *ExternalPaymentSucceededEvent) PaymentID() string {
	return e.paymentID
}

func (e *ExternalPaymentSucceededEvent) ExternalTransactionID() string {
	return e.externalTransactionID
}
