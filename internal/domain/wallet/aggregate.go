package wallet

import (
	"errors"
	"time"

	vo "github.com/franco/payment-api/internal/domain/shared/valueobjects"
)

// Wallet is an aggregate root representing a user's wallet
// Uses Value Objects for type safety and protection of invariants
type Wallet struct {
	userID    vo.UserID
	balance   vo.Money
	updatedAt time.Time
}

// NewWallet creates a new Wallet aggregate
func NewWallet(userID vo.UserID, initialBalance vo.Money) (*Wallet, error) {
	if userID.IsEmpty() {
		return nil, errors.New("user ID is required")
	}

	// Balance can be zero, but not negative
	if initialBalance.Amount().IsNegative() {
		return nil, errors.New("initial balance cannot be negative")
	}

	return &Wallet{
		userID:    userID,
		balance:   initialBalance,
		updatedAt: time.Now().UTC(),
	}, nil
}

// Getters

func (w *Wallet) UserID() vo.UserID {
	return w.userID
}

func (w *Wallet) Balance() vo.Money {
	return w.balance
}

func (w *Wallet) UpdatedAt() time.Time {
	return w.updatedAt
}

// Domain Behaviors

// CanDebit checks if the wallet has sufficient balance for a debit
func (w *Wallet) CanDebit(amount vo.Money) bool {
	// Currency must match
	if !w.balance.Currency().Equals(amount.Currency()) {
		return false
	}

	// Must have sufficient balance
	hasEnough, err := w.balance.IsGreaterThanOrEqual(amount)
	if err != nil {
		return false
	}

	return hasEnough
}

// Debit removes funds from the wallet
// Returns previous balance and new balance on success
func (w *Wallet) Debit(amount vo.Money) (previousBalance vo.Money, newBalance vo.Money, err error) {
	// Validate currency match
	if !w.balance.Currency().Equals(amount.Currency()) {
		return vo.Money{}, vo.Money{}, errors.New("currency mismatch: cannot debit different currency")
	}

	// Validate sufficient funds
	if !w.CanDebit(amount) {
		return vo.Money{}, vo.Money{}, errors.New("insufficient funds")
	}

	// Capture previous state
	previousBalance = w.balance

	// Perform debit
	w.balance, err = w.balance.Subtract(amount)
	if err != nil {
		return vo.Money{}, vo.Money{}, err
	}

	// Update timestamp
	w.updatedAt = time.Now().UTC()

	return previousBalance, w.balance, nil
}

// Credit adds funds to the wallet
// Returns previous balance and new balance on success
func (w *Wallet) Credit(amount vo.Money) (previousBalance vo.Money, newBalance vo.Money, err error) {
	// Validate currency match
	if !w.balance.Currency().Equals(amount.Currency()) {
		return vo.Money{}, vo.Money{}, errors.New("currency mismatch: cannot credit different currency")
	}

	// Validate amount is positive
	if !amount.IsPositive() {
		return vo.Money{}, vo.Money{}, errors.New("credit amount must be positive")
	}

	// Capture previous state
	previousBalance = w.balance

	// Perform credit
	w.balance, err = w.balance.Add(amount)
	if err != nil {
		return vo.Money{}, vo.Money{}, err
	}

	// Update timestamp
	w.updatedAt = time.Now().UTC()

	return previousBalance, w.balance, nil
}

// Query methods

// IsBalanceZero checks if balance is exactly zero
func (w *Wallet) IsBalanceZero() bool {
	return w.balance.IsZero()
}

// HasMinimumBalance checks if wallet has at least the minimum required amount
func (w *Wallet) HasMinimumBalance(minimum vo.Money) (bool, error) {
	// Currency must match
	if !w.balance.Currency().Equals(minimum.Currency()) {
		return false, errors.New("currency mismatch")
	}

	return w.balance.IsGreaterThanOrEqual(minimum)
}

// HasSufficientBalanceFor checks if wallet can cover a specific amount
func (w *Wallet) HasSufficientBalanceFor(amount vo.Money) bool {
	return w.CanDebit(amount)
}

// Reconstruction for repositories

// ReconstructWallet reconstructs a Wallet from persistence
func ReconstructWallet(
	userID vo.UserID,
	balance vo.Money,
	updatedAt time.Time,
) *Wallet {
	return &Wallet{
		userID:    userID,
		balance:   balance,
		updatedAt: updatedAt,
	}
}
