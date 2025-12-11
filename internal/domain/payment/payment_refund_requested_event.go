package payment

import "github.com/franco/payment-api/internal/domain/shared"

// PaymentRefundRequestedEvent is emitted when a payment needs to be refunded
type PaymentRefundRequestedEvent struct {
	shared.BaseEvent
	paymentID string
	userID    string
	amount    float64
	reason    string
}

// NewPaymentRefundRequestedEvent creates a new PaymentRefundRequestedEvent
func NewPaymentRefundRequestedEvent(
	paymentID, userID string,
	amount float64,
	reason string,
	metadata shared.Metadata,
) *PaymentRefundRequestedEvent {
	return &PaymentRefundRequestedEvent{
		BaseEvent: shared.NewBaseEvent("PaymentRefundRequested", metadata),
		paymentID: paymentID,
		userID:    userID,
		amount:    amount,
		reason:    reason,
	}
}

func (e *PaymentRefundRequestedEvent) PaymentID() string {
	return e.paymentID
}

func (e *PaymentRefundRequestedEvent) UserID() string {
	return e.userID
}

func (e *PaymentRefundRequestedEvent) Amount() float64 {
	return e.amount
}

func (e *PaymentRefundRequestedEvent) Reason() string {
	return e.reason
}
