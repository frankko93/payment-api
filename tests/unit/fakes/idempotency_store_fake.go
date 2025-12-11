package fakes

import (
	"context"
	"errors"
	"sync"
)

// IdempotencyStoreFake is a fake implementation of IdempotencyStore for testing
type IdempotencyStoreFake struct {
	mu   sync.RWMutex
	keys map[string]string
}

// NewIdempotencyStoreFake creates a new IdempotencyStoreFake
func NewIdempotencyStoreFake() *IdempotencyStoreFake {
	return &IdempotencyStoreFake{
		keys: make(map[string]string),
	}
}

// GetPaymentIDByKey retrieves a payment ID by idempotency key
func (f *IdempotencyStoreFake) GetPaymentIDByKey(ctx context.Context, idempotencyKey string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	paymentID, exists := f.keys[idempotencyKey]
	if !exists {
		return "", errors.New("key not found")
	}

	return paymentID, nil
}

// SaveKey saves an idempotency key
func (f *IdempotencyStoreFake) SaveKey(ctx context.Context, idempotencyKey, paymentID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.keys[idempotencyKey] = paymentID
	return nil
}

