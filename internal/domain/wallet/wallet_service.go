package wallet

import (
	"errors"

	vo "github.com/franco/payment-api/internal/domain/shared/valueobjects"
)

// Service is a Domain Service for wallet-related operations
type Service struct {
	minimumBalance vo.Money
}

// NewService creates a new wallet Service
// minimumBalance is the minimum balance that must be maintained
func NewService(minimumBalance vo.Money) *Service {
	return &Service{
		minimumBalance: minimumBalance,
	}
}

// ValidateDebit validates if a debit operation can be performed
func (s *Service) ValidateDebit(wlt *Wallet, amount vo.Money) error {
	// Check currency match
	if !wlt.Balance().Currency().Equals(amount.Currency()) {
		return errors.New("currency mismatch")
	}

	// Check sufficient balance
	if !wlt.CanDebit(amount) {
		return errors.New("insufficient funds")
	}

	// Check minimum balance requirement (if configured)
	if !s.minimumBalance.IsZero() {
		potentialBalance, err := wlt.Balance().Subtract(amount)
		if err != nil {
			return err
		}

		hasMinimum, err := potentialBalance.IsGreaterThanOrEqual(s.minimumBalance)
		if err != nil {
			return err
		}

		if !hasMinimum {
			return errors.New("operation would violate minimum balance requirement")
		}
	}

	return nil
}

// CanCoverPayment checks if a wallet can cover a payment amount
func (s *Service) CanCoverPayment(wlt *Wallet, paymentAmount vo.Money) bool {
	return s.ValidateDebit(wlt, paymentAmount) == nil
}
