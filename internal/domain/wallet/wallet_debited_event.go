package wallet

import "github.com/franco/payment-api/internal/domain/shared"

// WalletDebitedEvent is emitted when funds are debited from a wallet
type WalletDebitedEvent struct {
	shared.BaseEvent
	paymentID   string
	userID      string
	amount      float64
	newBalance  float64
	prevBalance float64
}

// NewWalletDebitedEvent creates a new WalletDebitedEvent
func NewWalletDebitedEvent(
	paymentID, userID string,
	amount, prevBalance, newBalance float64,
	metadata shared.Metadata,
) *WalletDebitedEvent {
	return &WalletDebitedEvent{
		BaseEvent:   shared.NewBaseEvent("WalletDebited", metadata),
		paymentID:   paymentID,
		userID:      userID,
		amount:      amount,
		prevBalance: prevBalance,
		newBalance:  newBalance,
	}
}

func (e *WalletDebitedEvent) PaymentID() string {
	return e.paymentID
}

func (e *WalletDebitedEvent) UserID() string {
	return e.userID
}

func (e *WalletDebitedEvent) Amount() float64 {
	return e.amount
}

func (e *WalletDebitedEvent) PrevBalance() float64 {
	return e.prevBalance
}

func (e *WalletDebitedEvent) NewBalance() float64 {
	return e.newBalance
}
