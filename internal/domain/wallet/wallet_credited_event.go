package wallet

import "github.com/franco/payment-api/internal/domain/shared"

// WalletCreditedEvent is emitted when funds are credited to a wallet
type WalletCreditedEvent struct {
	shared.BaseEvent
	paymentID   string
	userID      string
	amount      float64
	newBalance  float64
	prevBalance float64
	reason      string
}

// NewWalletCreditedEvent creates a new WalletCreditedEvent
func NewWalletCreditedEvent(
	paymentID, userID string,
	amount, prevBalance, newBalance float64,
	reason string,
	metadata shared.Metadata,
) *WalletCreditedEvent {
	return &WalletCreditedEvent{
		BaseEvent:   shared.NewBaseEvent("WalletCredited", metadata),
		paymentID:   paymentID,
		userID:      userID,
		amount:      amount,
		prevBalance: prevBalance,
		newBalance:  newBalance,
		reason:      reason,
	}
}

func (e *WalletCreditedEvent) PaymentID() string {
	return e.paymentID
}

func (e *WalletCreditedEvent) UserID() string {
	return e.userID
}

func (e *WalletCreditedEvent) Amount() float64 {
	return e.amount
}

func (e *WalletCreditedEvent) PrevBalance() float64 {
	return e.prevBalance
}

func (e *WalletCreditedEvent) NewBalance() float64 {
	return e.newBalance
}

func (e *WalletCreditedEvent) Reason() string {
	return e.reason
}
