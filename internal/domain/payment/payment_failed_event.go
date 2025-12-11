package payment

import "github.com/franco/payment-api/internal/domain/shared"

// PaymentFailedEvent is emitted when a payment fails
type PaymentFailedEvent struct {
	shared.BaseEvent
	paymentID string
	userID    string
	amount    float64
	reason    string
}

// NewPaymentFailedEvent creates a new PaymentFailedEvent
func NewPaymentFailedEvent(
	paymentID, userID string,
	amount float64,
	reason string,
	metadata shared.Metadata,
) *PaymentFailedEvent {
	return &PaymentFailedEvent{
		BaseEvent: shared.NewBaseEvent("PaymentFailed", metadata),
		paymentID: paymentID,
		userID:    userID,
		amount:    amount,
		reason:    reason,
	}
}

func (e *PaymentFailedEvent) PaymentID() string {
	return e.paymentID
}

func (e *PaymentFailedEvent) UserID() string {
	return e.userID
}

func (e *PaymentFailedEvent) Amount() float64 {
	return e.amount
}

func (e *PaymentFailedEvent) Reason() string {
	return e.reason
}
