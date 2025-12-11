package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/franco/payment-api/internal/domain/payment"
	"github.com/franco/payment-api/internal/domain/shared"
	domerrors "github.com/franco/payment-api/internal/domain/shared/errors"
	vo "github.com/franco/payment-api/internal/domain/shared/valueobjects"
	"github.com/franco/payment-api/internal/domain/wallet"
	"github.com/shopspring/decimal"
)

// CreatePaymentRequest represents a payment creation request
type CreatePaymentRequest struct {
	UserID         string
	Amount         float64
	Currency       string
	ServiceID      string
	IdempotencyKey string
	ClientID       string
}

// CreatePaymentResponse represents the response from payment creation
type CreatePaymentResponse struct {
	PaymentID string
	Status    string
}

// CreatePaymentService handles payment creation use case
type CreatePaymentService struct {
	paymentRepo      PaymentRepository
	walletRepo       WalletRepository
	idempotencyStore shared.IdempotencyStore
	eventStore       shared.EventStore
	eventPublisher   EventPublisher
	topicArn         string
}

// PaymentRepository defines payment persistence operations
type PaymentRepository interface {
	Save(ctx context.Context, pmt *payment.Payment) error
	FindByID(ctx context.Context, paymentID string) (*payment.Payment, error)
	Update(ctx context.Context, pmt *payment.Payment) error
}

// WalletRepository defines wallet operations
type WalletRepository interface {
	GetByUserID(ctx context.Context, userID string) (*wallet.Wallet, error)
}

// EventPublisher publishes events
type EventPublisher interface {
	Publish(ctx context.Context, event shared.Event, topicArn string) error
}

// NewCreatePaymentService creates a new CreatePaymentService
func NewCreatePaymentService(
	paymentRepo PaymentRepository,
	walletRepo WalletRepository,
	idempotencyStore shared.IdempotencyStore,
	eventStore shared.EventStore,
	eventPublisher EventPublisher,
	topicArn string,
) *CreatePaymentService {
	return &CreatePaymentService{
		paymentRepo:      paymentRepo,
		walletRepo:       walletRepo,
		idempotencyStore: idempotencyStore,
		eventStore:       eventStore,
		eventPublisher:   eventPublisher,
		topicArn:         topicArn,
	}
}

// Execute creates a new payment or returns existing one if idempotent
func (s *CreatePaymentService) Execute(ctx context.Context, req CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// Validate request
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	// Check idempotency
	existingPaymentID, err := s.idempotencyStore.GetPaymentIDByKey(ctx, req.IdempotencyKey)
	if err == nil && existingPaymentID != "" {
		return &CreatePaymentResponse{
			PaymentID: existingPaymentID,
			Status:    "ALREADY_PROCESSED",
		}, nil
	}

	// Validate wallet exists and has sufficient balance (SYNC)
	if err := s.validateWalletBalance(ctx, req); err != nil {
		return nil, err
	}

	// Create Value Objects
	paymentID := vo.GeneratePaymentID()

	userID, err := vo.NewUserID(req.UserID)
	if err != nil {
		return nil, err
	}

	serviceID, err := vo.NewServiceID(req.ServiceID)
	if err != nil {
		return nil, err
	}

	currency, err := vo.NewCurrency(req.Currency)
	if err != nil {
		return nil, err
	}

	amount := decimal.NewFromFloat(req.Amount)
	money, err := vo.NewMoney(amount, currency)
	if err != nil {
		return nil, err
	}

	idempKey, err := vo.NewIdempotencyKey(req.IdempotencyKey)
	if err != nil {
		return nil, err
	}

	// Create new payment aggregate
	pmt, err := payment.NewPayment(
		paymentID,
		userID,
		serviceID,
		money,
		idempKey,
	)
	if err != nil {
		return nil, err
	}

	// Save payment
	if err := s.paymentRepo.Save(ctx, pmt); err != nil {
		return nil, err
	}

	// Save idempotency key
	if err := s.idempotencyStore.SaveKey(ctx, req.IdempotencyKey, paymentID.String()); err != nil {
		return nil, err
	}

	// Create and publish PaymentRequested event
	metadata := shared.Metadata{
		ClientID:  req.ClientID,
		RequestID: vo.GeneratePaymentID().String(),
		Source:    "payment-api",
		Extra:     make(map[string]string),
	}

	event := payment.NewPaymentRequestedEvent(
		paymentID.String(),
		req.UserID,
		req.Amount,
		req.Currency,
		req.ServiceID,
		req.IdempotencyKey,
		metadata,
	)

	// Store event
	if err := s.eventStore.Append(ctx, event, paymentID.String()); err != nil {
		return nil, err
	}

	// Publish event
	if err := s.eventPublisher.Publish(ctx, event, s.topicArn); err != nil {
		return nil, err
	}

	return &CreatePaymentResponse{
		PaymentID: paymentID.String(),
		Status:    vo.PaymentStatusPending.String(),
	}, nil
}

func (s *CreatePaymentService) validateRequest(req CreatePaymentRequest) error {
	if req.UserID == "" {
		return errors.New("userID is required")
	}
	if req.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if req.Currency == "" {
		return errors.New("currency is required")
	}
	if req.ServiceID == "" {
		return errors.New("serviceID is required")
	}
	if req.IdempotencyKey == "" {
		return errors.New("idempotencyKey is required")
	}
	return nil
}

// validateWalletBalance checks wallet exists and has sufficient funds
// This is a SYNC validation before creating the payment
func (s *CreatePaymentService) validateWalletBalance(ctx context.Context, req CreatePaymentRequest) error {
	// Get wallet
	wlt, err := s.walletRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return domerrors.WalletNotFoundError(req.UserID)
	}

	// Create money value object for amount
	currency, err := vo.NewCurrency(req.Currency)
	if err != nil {
		return err
	}

	amount := decimal.NewFromFloat(req.Amount)
	money, err := vo.NewMoney(amount, currency)
	if err != nil {
		return err
	}

	// Check currency match
	if !wlt.Balance().Currency().Equals(money.Currency()) {
		return domerrors.CurrencyMismatchError(
			wlt.Balance().Currency().Code(),
			money.Currency().Code(),
		)
	}

	// Check sufficient balance
	if !wlt.CanDebit(money) {
		return domerrors.InsufficientFundsError(
			fmt.Sprintf("%.2f %s", money.AmountFloat(), money.Currency().Code()),
			fmt.Sprintf("%.2f %s", wlt.Balance().AmountFloat(), wlt.Balance().Currency().Code()),
		)
	}

	return nil
}
