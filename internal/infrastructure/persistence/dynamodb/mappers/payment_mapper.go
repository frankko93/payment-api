package mappers

import (
	"fmt"
	"time"

	"github.com/franco/payment-api/internal/domain/payment"
	vo "github.com/franco/payment-api/internal/domain/shared/valueobjects"
	"github.com/shopspring/decimal"
)

// PaymentDBModel represents the database persistence model for Payment
type PaymentDBModel struct {
	ID             string `dynamodbav:"id"`
	UserID         string `dynamodbav:"userId"`
	Amount         string `dynamodbav:"amount"` // Store as string for precision
	Currency       string `dynamodbav:"currency"`
	ServiceID      string `dynamodbav:"serviceId"`
	Status         string `dynamodbav:"status"`
	IdempotencyKey string `dynamodbav:"idempotencyKey"`
	FailureReason  string `dynamodbav:"failureReason,omitempty"`
	ExternalTxID   string `dynamodbav:"externalTxId,omitempty"`
	CreatedAt      string `dynamodbav:"createdAt"`
	UpdatedAt      string `dynamodbav:"updatedAt"`
}

// PaymentMapper handles mapping between domain and persistence models
type PaymentMapper struct{}

// NewPaymentMapper creates a new PaymentMapper
func NewPaymentMapper() *PaymentMapper {
	return &PaymentMapper{}
}

// ToDBModel converts domain Payment to database model
func (m *PaymentMapper) ToDBModel(pmt *payment.Payment) (*PaymentDBModel, error) {
	if pmt == nil {
		return nil, fmt.Errorf("payment cannot be nil")
	}

	return &PaymentDBModel{
		ID:             pmt.ID().String(),
		UserID:         pmt.UserID().String(),
		Amount:         pmt.Money().Amount().String(),
		Currency:       pmt.Money().Currency().Code(),
		ServiceID:      pmt.ServiceID().String(),
		Status:         pmt.Status().String(),
		IdempotencyKey: pmt.IdempotencyKey().String(),
		FailureReason:  pmt.FailureReason(),
		ExternalTxID:   pmt.ExternalTxID(),
		CreatedAt:      pmt.CreatedAt().Format(time.RFC3339),
		UpdatedAt:      pmt.UpdatedAt().Format(time.RFC3339),
	}, nil
}

// ToDomain converts database model to domain Payment
func (m *PaymentMapper) ToDomain(model *PaymentDBModel) (*payment.Payment, error) {
	if model == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}

	paymentID, err := vo.NewPaymentID(model.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}

	userID, err := vo.NewUserID(model.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	serviceID, err := vo.NewServiceID(model.ServiceID)
	if err != nil {
		return nil, fmt.Errorf("invalid service ID: %w", err)
	}

	idempotencyKey, err := vo.NewIdempotencyKey(model.IdempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("invalid idempotency key: %w", err)
	}

	amount, err := decimal.NewFromString(model.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	currency, err := vo.NewCurrency(model.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}

	money, err := vo.NewMoney(amount, currency)
	if err != nil {
		return nil, fmt.Errorf("invalid money: %w", err)
	}

	status, err := vo.ParsePaymentStatus(model.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid status: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, model.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid createdAt: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, model.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid updatedAt: %w", err)
	}

	pmt := payment.ReconstructPayment(
		paymentID,
		userID,
		serviceID,
		money,
		idempotencyKey,
		status,
		model.FailureReason,
		model.ExternalTxID,
		createdAt,
		updatedAt,
	)

	return pmt, nil
}
