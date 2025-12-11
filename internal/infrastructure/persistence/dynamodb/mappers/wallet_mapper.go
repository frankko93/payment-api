package mappers

import (
	"fmt"
	"time"

	vo "github.com/franco/payment-api/internal/domain/shared/valueobjects"
	"github.com/franco/payment-api/internal/domain/wallet"
	"github.com/shopspring/decimal"
)

// WalletDBModel represents the database persistence model for Wallet
type WalletDBModel struct {
	UserID    string `dynamodbav:"userId"`
	Balance   string `dynamodbav:"balance"` // Decimal as string
	Currency  string `dynamodbav:"currency"`
	UpdatedAt string `dynamodbav:"updatedAt"`
}

// WalletMapper handles mapping between domain and persistence models
type WalletMapper struct{}

// NewWalletMapper creates a new WalletMapper
func NewWalletMapper() *WalletMapper {
	return &WalletMapper{}
}

// ToDBModel converts domain Wallet to database model
func (m *WalletMapper) ToDBModel(wlt *wallet.Wallet) (*WalletDBModel, error) {
	if wlt == nil {
		return nil, fmt.Errorf("wallet cannot be nil")
	}

	return &WalletDBModel{
		UserID:    wlt.UserID().String(),
		Balance:   wlt.Balance().Amount().String(),
		Currency:  wlt.Balance().Currency().Code(),
		UpdatedAt: wlt.UpdatedAt().Format(time.RFC3339),
	}, nil
}

// ToDomain converts database model to domain Wallet
func (m *WalletMapper) ToDomain(model *WalletDBModel) (*wallet.Wallet, error) {
	if model == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}

	userID, err := vo.NewUserID(model.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	amount, err := decimal.NewFromString(model.Balance)
	if err != nil {
		return nil, fmt.Errorf("invalid balance amount: %w", err)
	}

	currency, err := vo.NewCurrency(model.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}

	money, err := vo.NewMoney(amount, currency)
	if err != nil {
		return nil, fmt.Errorf("invalid money: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, model.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid updatedAt: %w", err)
	}

	wlt := wallet.ReconstructWallet(userID, money, updatedAt)

	return wlt, nil
}
