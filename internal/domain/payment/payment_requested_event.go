package payment

import "github.com/franco/payment-api/internal/domain/shared"

// PaymentRequestedEvent is emitted when a new payment is requested
type PaymentRequestedEvent struct {
	shared.BaseEvent
	paymentID      string
	userID         string
	amount         float64
	currency       string
	serviceID      string
	idempotencyKey string
}

// NewPaymentRequestedEvent creates a new PaymentRequestedEvent
func NewPaymentRequestedEvent(
	paymentID, userID string,
	amount float64,
	currency, serviceID, idempotencyKey string,
	metadata shared.Metadata,
) *PaymentRequestedEvent {
	return &PaymentRequestedEvent{
		BaseEvent:      shared.NewBaseEvent("PaymentRequested", metadata),
		paymentID:      paymentID,
		userID:         userID,
		amount:         amount,
		currency:       currency,
		serviceID:      serviceID,
		idempotencyKey: idempotencyKey,
	}
}

func (e *PaymentRequestedEvent) PaymentID() string {
	return e.paymentID
}

func (e *PaymentRequestedEvent) UserID() string {
	return e.userID
}

func (e *PaymentRequestedEvent) Amount() float64 {
	return e.amount
}

func (e *PaymentRequestedEvent) Currency() string {
	return e.currency
}

func (e *PaymentRequestedEvent) ServiceID() string {
	return e.serviceID
}

func (e *PaymentRequestedEvent) IdempotencyKey() string {
	return e.idempotencyKey
}
