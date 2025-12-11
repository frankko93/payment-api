package payment

import "github.com/franco/payment-api/internal/domain/shared"

// ExternalPaymentRequestedEvent is emitted when a payment is sent to external gateway
type ExternalPaymentRequestedEvent struct {
	shared.BaseEvent
	paymentID string
	userID    string
	amount    float64
	currency  string
	serviceID string
}

// NewExternalPaymentRequestedEvent creates a new ExternalPaymentRequestedEvent
func NewExternalPaymentRequestedEvent(
	paymentID, userID string,
	amount float64,
	currency, serviceID string,
	metadata shared.Metadata,
) *ExternalPaymentRequestedEvent {
	return &ExternalPaymentRequestedEvent{
		BaseEvent: shared.NewBaseEvent("ExternalPaymentRequested", metadata),
		paymentID: paymentID,
		userID:    userID,
		amount:    amount,
		currency:  currency,
		serviceID: serviceID,
	}
}

func (e *ExternalPaymentRequestedEvent) PaymentID() string {
	return e.paymentID
}

func (e *ExternalPaymentRequestedEvent) UserID() string {
	return e.userID
}

func (e *ExternalPaymentRequestedEvent) Amount() float64 {
	return e.amount
}

func (e *ExternalPaymentRequestedEvent) Currency() string {
	return e.currency
}

func (e *ExternalPaymentRequestedEvent) ServiceID() string {
	return e.serviceID
}
