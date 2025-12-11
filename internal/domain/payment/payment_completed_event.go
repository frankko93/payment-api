package payment

import "github.com/franco/payment-api/internal/domain/shared"

// PaymentCompletedEvent is emitted when a payment is successfully completed
type PaymentCompletedEvent struct {
	shared.BaseEvent
	paymentID             string
	userID                string
	amount                float64
	externalTransactionID string
}

// NewPaymentCompletedEvent creates a new PaymentCompletedEvent
func NewPaymentCompletedEvent(
	paymentID, userID string,
	amount float64,
	externalTransactionID string,
	metadata shared.Metadata,
) *PaymentCompletedEvent {
	return &PaymentCompletedEvent{
		BaseEvent:             shared.NewBaseEvent("PaymentCompleted", metadata),
		paymentID:             paymentID,
		userID:                userID,
		amount:                amount,
		externalTransactionID: externalTransactionID,
	}
}

func (e *PaymentCompletedEvent) PaymentID() string {
	return e.paymentID
}

func (e *PaymentCompletedEvent) UserID() string {
	return e.userID
}

func (e *PaymentCompletedEvent) Amount() float64 {
	return e.amount
}

func (e *PaymentCompletedEvent) ExternalTransactionID() string {
	return e.externalTransactionID
}
