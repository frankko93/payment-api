package payment

import (
	"errors"
	"time"

	vo "github.com/franco/payment-api/internal/domain/shared/valueobjects"
)

// Payment is an aggregate root representing a payment transaction
// Uses Value Objects for type safety and rich domain behavior
type Payment struct {
	// Identifiers (Value Objects)
	id             vo.PaymentID
	userID         vo.UserID
	serviceID      vo.ServiceID
	idempotencyKey vo.IdempotencyKey

	// Value Objects
	money  vo.Money
	status vo.PaymentStatus

	// Optional fields
	failureReason string
	externalTxID  string

	// Timestamps
	createdAt time.Time
	updatedAt time.Time
}

// NewPayment creates a new Payment aggregate with proper validation
func NewPayment(
	id vo.PaymentID,
	userID vo.UserID,
	serviceID vo.ServiceID,
	money vo.Money,
	idempotencyKey vo.IdempotencyKey,
) (*Payment, error) {
	// Validate inputs
	if id.IsEmpty() {
		return nil, errors.New("payment ID is required")
	}
	if userID.IsEmpty() {
		return nil, errors.New("user ID is required")
	}
	if serviceID.IsEmpty() {
		return nil, errors.New("service ID is required")
	}
	if money.IsZero() {
		return nil, errors.New("payment amount must be greater than zero")
	}
	if idempotencyKey.IsEmpty() {
		return nil, errors.New("idempotency key is required")
	}

	now := time.Now().UTC()

	return &Payment{
		id:             id,
		userID:         userID,
		serviceID:      serviceID,
		money:          money,
		idempotencyKey: idempotencyKey,
		status:         vo.PaymentStatusPending,
		createdAt:      now,
		updatedAt:      now,
	}, nil
}

// Getters (read-only access to protect invariants)

func (p *Payment) ID() vo.PaymentID {
	return p.id
}

func (p *Payment) UserID() vo.UserID {
	return p.userID
}

func (p *Payment) ServiceID() vo.ServiceID {
	return p.serviceID
}

func (p *Payment) Money() vo.Money {
	return p.money
}

func (p *Payment) Status() vo.PaymentStatus {
	return p.status
}

func (p *Payment) IdempotencyKey() vo.IdempotencyKey {
	return p.idempotencyKey
}

func (p *Payment) FailureReason() string {
	return p.failureReason
}

func (p *Payment) ExternalTxID() string {
	return p.externalTxID
}

func (p *Payment) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Payment) UpdatedAt() time.Time {
	return p.updatedAt
}

// Domain Behaviors (protected state transitions)

// MarkCompleted transitions the payment to completed status
func (p *Payment) MarkCompleted(externalTxID string) error {
	// Validate state transition
	if err := p.status.ValidateTransition(vo.PaymentStatusCompleted); err != nil {
		return err
	}

	if externalTxID == "" {
		return errors.New("external transaction ID is required when completing payment")
	}

	// Apply state change
	p.status = vo.PaymentStatusCompleted
	p.externalTxID = externalTxID
	p.updatedAt = time.Now().UTC()

	return nil
}

// MarkFailed transitions the payment to failed status
func (p *Payment) MarkFailed(reason string) error {
	// Validate state transition
	if err := p.status.ValidateTransition(vo.PaymentStatusFailed); err != nil {
		return err
	}

	if reason == "" {
		return errors.New("failure reason is required when marking payment as failed")
	}

	// Apply state change
	p.status = vo.PaymentStatusFailed
	p.failureReason = reason
	p.updatedAt = time.Now().UTC()

	return nil
}

// Query methods

// IsPending checks if payment is in pending status
func (p *Payment) IsPending() bool {
	return p.status.IsPending()
}

// IsCompleted checks if payment is completed
func (p *Payment) IsCompleted() bool {
	return p.status.IsCompleted()
}

// IsFailed checks if payment has failed
func (p *Payment) IsFailed() bool {
	return p.status.IsFailed()
}

// IsTerminal checks if payment is in a terminal state
func (p *Payment) IsTerminal() bool {
	return p.status.IsTerminal()
}

// CanBeRefunded checks if this payment can be refunded
// Business rule: Only failed payments that had wallet debit can be refunded
func (p *Payment) CanBeRefunded() bool {
	return p.status.IsFailed()
}

// CanBeProcessed checks if payment can be processed
func (p *Payment) CanBeProcessed() bool {
	return p.status.IsPending()
}

// Reconstruction methods for repositories

// ReconstructPayment reconstructs a Payment from persistence
// This bypasses validation for data coming from the database
func ReconstructPayment(
	id vo.PaymentID,
	userID vo.UserID,
	serviceID vo.ServiceID,
	money vo.Money,
	idempotencyKey vo.IdempotencyKey,
	status vo.PaymentStatus,
	failureReason string,
	externalTxID string,
	createdAt time.Time,
	updatedAt time.Time,
) *Payment {
	return &Payment{
		id:             id,
		userID:         userID,
		serviceID:      serviceID,
		money:          money,
		idempotencyKey: idempotencyKey,
		status:         status,
		failureReason:  failureReason,
		externalTxID:   externalTxID,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}
