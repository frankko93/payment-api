package fakes

import (
	"context"
	"errors"
	"sync"

	"github.com/franco/payment-api/internal/domain/payment"
)

// PaymentRepositoryFake is a fake implementation of PaymentRepository for testing
type PaymentRepositoryFake struct {
	mu       sync.RWMutex
	payments map[string]*payment.Payment
}

// NewPaymentRepositoryFake creates a new PaymentRepositoryFake
func NewPaymentRepositoryFake() *PaymentRepositoryFake {
	return &PaymentRepositoryFake{
		payments: make(map[string]*payment.Payment),
	}
}

// Save stores a payment
func (f *PaymentRepositoryFake) Save(ctx context.Context, payment *payment.Payment) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Store using ID as key
	f.payments[payment.ID().String()] = payment
	return nil
}

// FindByID retrieves a payment by ID
func (f *PaymentRepositoryFake) FindByID(ctx context.Context, paymentID string) (*payment.Payment, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	payment, exists := f.payments[paymentID]
	if !exists {
		return nil, errors.New("payment not found")
	}

	return payment, nil
}

// Update updates a payment (same as Save in this fake)
func (f *PaymentRepositoryFake) Update(ctx context.Context, payment *payment.Payment) error {
	return f.Save(ctx, payment)
}

// GetAll returns all payments (helper for testing)
func (f *PaymentRepositoryFake) GetAll() []*payment.Payment {
	f.mu.RLock()
	defer f.mu.RUnlock()

	payments := make([]*payment.Payment, 0, len(f.payments))
	for _, p := range f.payments {
		payments = append(payments, p)
	}
	return payments
}
