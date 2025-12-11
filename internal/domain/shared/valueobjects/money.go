package valueobjects

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

// Money represents a monetary amount with currency
// This is a Value Object - immutable and comparable by value
type Money struct {
	amount   decimal.Decimal
	currency Currency
}

// NewMoney creates a new Money value object
func NewMoney(amount decimal.Decimal, currency Currency) (Money, error) {
	if amount.IsNegative() {
		return Money{}, errors.New("amount cannot be negative")
	}

	if currency.IsEmpty() {
		return Money{}, errors.New("currency is required")
	}

	return Money{
		amount:   amount,
		currency: currency,
	}, nil
}

// NewMoneyFromFloat creates Money from float64 (use with caution)
func NewMoneyFromFloat(amount float64, currency Currency) (Money, error) {
	if amount < 0 {
		return Money{}, errors.New("amount cannot be negative")
	}

	return NewMoney(decimal.NewFromFloat(amount), currency)
}

// NewMoneyFromString creates Money from string
func NewMoneyFromString(amount string, currency Currency) (Money, error) {
	dec, err := decimal.NewFromString(amount)
	if err != nil {
		return Money{}, fmt.Errorf("invalid amount format: %w", err)
	}

	return NewMoney(dec, currency)
}

// MustNewMoney creates Money or panics (use only in tests)
func MustNewMoney(amount string, currencyCode string) Money {
	dec, err := decimal.NewFromString(amount)
	if err != nil {
		panic(err)
	}
	curr := MustNewCurrency(currencyCode)
	m, err := NewMoney(dec, curr)
	if err != nil {
		panic(err)
	}
	return m
}

// Zero returns a zero amount for the given currency
func Zero(currency Currency) Money {
	return Money{
		amount:   decimal.Zero,
		currency: currency,
	}
}

// Getters

// Amount returns the monetary amount as Decimal
func (m Money) Amount() decimal.Decimal {
	return m.amount
}

// Currency returns the currency
func (m Money) Currency() Currency {
	return m.currency
}

// AmountFloat returns amount as float64 (use only for display/legacy)
func (m Money) AmountFloat() float64 {
	f, _ := m.amount.Float64()
	return f
}

// Operations

// Add adds two Money values (must be same currency)
func (m Money) Add(other Money) (Money, error) {
	if !m.currency.Equals(other.currency) {
		return Money{}, fmt.Errorf("cannot add different currencies: %s and %s",
			m.currency.Code(), other.currency.Code())
	}

	return Money{
		amount:   m.amount.Add(other.amount),
		currency: m.currency,
	}, nil
}

// Subtract subtracts two Money values (must be same currency)
func (m Money) Subtract(other Money) (Money, error) {
	if !m.currency.Equals(other.currency) {
		return Money{}, fmt.Errorf("cannot subtract different currencies: %s and %s",
			m.currency.Code(), other.currency.Code())
	}

	result := m.amount.Sub(other.amount)
	if result.IsNegative() {
		return Money{}, errors.New("result would be negative")
	}

	return Money{
		amount:   result,
		currency: m.currency,
	}, nil
}

// Multiply multiplies money by a factor
func (m Money) Multiply(factor decimal.Decimal) (Money, error) {
	result := m.amount.Mul(factor)
	if result.IsNegative() {
		return Money{}, errors.New("result would be negative")
	}

	return Money{
		amount:   result,
		currency: m.currency,
	}, nil
}

// Comparisons

// IsGreaterThan checks if this amount is greater than another
func (m Money) IsGreaterThan(other Money) (bool, error) {
	if !m.currency.Equals(other.currency) {
		return false, errors.New("cannot compare different currencies")
	}
	return m.amount.GreaterThan(other.amount), nil
}

// IsGreaterThanOrEqual checks if this amount is >= another
func (m Money) IsGreaterThanOrEqual(other Money) (bool, error) {
	if !m.currency.Equals(other.currency) {
		return false, errors.New("cannot compare different currencies")
	}
	return m.amount.GreaterThanOrEqual(other.amount), nil
}

// IsLessThan checks if this amount is less than another
func (m Money) IsLessThan(other Money) (bool, error) {
	if !m.currency.Equals(other.currency) {
		return false, errors.New("cannot compare different currencies")
	}
	return m.amount.LessThan(other.amount), nil
}

// Equals checks if two Money values are equal
func (m Money) Equals(other Money) bool {
	return m.amount.Equal(other.amount) && m.currency.Equals(other.currency)
}

// IsZero checks if the amount is zero
func (m Money) IsZero() bool {
	return m.amount.IsZero()
}

// IsPositive checks if amount is greater than zero
func (m Money) IsPositive() bool {
	return m.amount.IsPositive()
}

// String returns a formatted string representation
func (m Money) String() string {
	return fmt.Sprintf("%s %s", m.amount.StringFixed(2), m.currency.Code())
}

// JSON Serialization

// MarshalJSON implements json.Marshaler
func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"amount":   m.amount.String(),
		"currency": m.currency.Code(),
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (m *Money) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	amount, err := decimal.NewFromString(tmp.Amount)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	currency, err := NewCurrency(tmp.Currency)
	if err != nil {
		return err
	}

	money, err := NewMoney(amount, currency)
	if err != nil {
		return err
	}

	*m = money
	return nil
}
