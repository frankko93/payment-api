package fakes

import (
	"context"
	"errors"
	"sync"

	"github.com/franco/payment-api/internal/domain/wallet"
)

// WalletRepositoryFake is a fake implementation of WalletRepository for testing
type WalletRepositoryFake struct {
	mu      sync.RWMutex
	wallets map[string]*wallet.Wallet
}

// NewWalletRepositoryFake creates a new WalletRepositoryFake
func NewWalletRepositoryFake() *WalletRepositoryFake {
	return &WalletRepositoryFake{
		wallets: make(map[string]*wallet.Wallet),
	}
}

// GetByUserID retrieves a wallet by user ID
func (f *WalletRepositoryFake) GetByUserID(ctx context.Context, userID string) (*wallet.Wallet, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	wallet, exists := f.wallets[userID]
	if !exists {
		return nil, errors.New("wallet not found")
	}

	return wallet, nil
}

// Save stores a wallet
func (f *WalletRepositoryFake) Save(ctx context.Context, wallet *wallet.Wallet) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.wallets[wallet.UserID().String()] = wallet
	return nil
}

// Update updates a wallet (same as Save in this fake)
func (f *WalletRepositoryFake) Update(ctx context.Context, wallet *wallet.Wallet) error {
	return f.Save(ctx, wallet)
}

// SetWallet is a helper method for tests to pre-populate wallets
func (f *WalletRepositoryFake) SetWallet(wallet *wallet.Wallet) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.wallets[wallet.UserID().String()] = wallet
}
