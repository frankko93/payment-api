package payment

import (
	"errors"

	vo "github.com/franco/payment-api/internal/domain/shared/valueobjects"
	"github.com/franco/payment-api/internal/domain/wallet"
)

// Processor is a Domain Service that encapsulates payment processing logic
// It coordinates between Payment and Wallet aggregates following business rules
type Processor struct{}

// NewProcessor creates a new Processor domain service
func NewProcessor() *Processor {
	return &Processor{}
}

// ProcessResult contains the result of processing a payment
type ProcessResult struct {
	Success         bool
	FailureReason   string
	WalletDebited   bool
	PreviousBalance vo.Money
	NewBalance      vo.Money
}

// Process validates and processes a payment against a wallet
// This is the core business logic for payment processing
func (p *Processor) Process(
	pmt *Payment,
	wlt *wallet.Wallet,
) (*ProcessResult, error) {
	// Business Rule 1: Payment must be in pending status
	if !pmt.CanBeProcessed() {
		return nil, errors.New("payment must be in pending status to be processed")
	}

	// Business Rule 2: User IDs must match
	if !pmt.UserID().Equals(wlt.UserID()) {
		return &ProcessResult{
			Success:       false,
			FailureReason: "USER_MISMATCH",
			WalletDebited: false,
		}, nil
	}

	// Business Rule 3: Currencies must match
	if !wlt.Balance().Currency().Equals(pmt.Money().Currency()) {
		return &ProcessResult{
			Success:       false,
			FailureReason: "CURRENCY_MISMATCH",
			WalletDebited: false,
		}, nil
	}

	// Business Rule 4: Sufficient funds required
	if !wlt.CanDebit(pmt.Money()) {
		return &ProcessResult{
			Success:       false,
			FailureReason: "INSUFFICIENT_FUNDS",
			WalletDebited: false,
		}, nil
	}

	// Business Rule 5: Debit wallet
	prevBalance, newBalance, err := wlt.Debit(pmt.Money())
	if err != nil {
		return &ProcessResult{
			Success:       false,
			FailureReason: "DEBIT_FAILED",
			WalletDebited: false,
		}, nil
	}

	return &ProcessResult{
		Success:         true,
		WalletDebited:   true,
		PreviousBalance: prevBalance,
		NewBalance:      newBalance,
	}, nil
}

// RefundResult contains the result of a refund operation
type RefundResult struct {
	Success         bool
	PreviousBalance vo.Money
	NewBalance      vo.Money
}

// Refund processes a refund by crediting the wallet
// This implements the compensating transaction for failed payments
func (p *Processor) Refund(
	pmt *Payment,
	wlt *wallet.Wallet,
) (*RefundResult, error) {
	// Business Rule 1: Can only refund payments that are eligible
	if !pmt.CanBeRefunded() {
		return nil, errors.New("payment cannot be refunded in current state")
	}

	// Business Rule 2: User IDs must match
	if !pmt.UserID().Equals(wlt.UserID()) {
		return nil, errors.New("user ID mismatch: payment and wallet belong to different users")
	}

	// Business Rule 3: Credit the wallet
	prevBalance, newBalance, err := wlt.Credit(pmt.Money())
	if err != nil {
		return nil, err
	}

	return &RefundResult{
		Success:         true,
		PreviousBalance: prevBalance,
		NewBalance:      newBalance,
	}, nil
}

// ValidateCreation validates business rules for creating a new payment
func (p *Processor) ValidateCreation(
	userID vo.UserID,
	money vo.Money,
	serviceID vo.ServiceID,
) error {
	if userID.IsEmpty() {
		return errors.New("user ID is required")
	}

	if money.IsZero() {
		return errors.New("payment amount must be greater than zero")
	}

	if serviceID.IsEmpty() {
		return errors.New("service ID is required")
	}

	// Additional business rules can be added here
	// For example: maximum transaction amount, service availability, etc.

	return nil
}
