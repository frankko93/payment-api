package valueobjects

import (
	"errors"
	"fmt"
	"strings"
)

// Currency represents a currency code (ISO 4217)
type Currency struct {
	code string
}

var validCurrencies = map[string]bool{
	"ARS": true,
	"USD": true,
	"EUR": true,
	"BRL": true,
	"MXN": true,
	"CLP": true,
	"COP": true,
}

// NewCurrency creates a new Currency value object
func NewCurrency(code string) (Currency, error) {
	code = strings.ToUpper(strings.TrimSpace(code))

	if code == "" {
		return Currency{}, errors.New("currency code cannot be empty")
	}

	if len(code) != 3 {
		return Currency{}, errors.New("currency code must be 3 characters")
	}

	if !validCurrencies[code] {
		return Currency{}, fmt.Errorf("unsupported currency: %s", code)
	}

	return Currency{code: code}, nil
}

// MustNewCurrency creates Currency or panics (use only in tests or constants)
func MustNewCurrency(code string) Currency {
	c, err := NewCurrency(code)
	if err != nil {
		panic(err)
	}
	return c
}

// Code returns the currency code
func (c Currency) Code() string {
	return c.code
}

// Equals checks if two currencies are equal
func (c Currency) Equals(other Currency) bool {
	return c.code == other.code
}

// String returns the currency code
func (c Currency) String() string {
	return c.code
}

// IsEmpty checks if currency is zero value
func (c Currency) IsEmpty() bool {
	return c.code == ""
}

// Common currencies as constants
var (
	ARS = MustNewCurrency("ARS")
	USD = MustNewCurrency("USD")
	EUR = MustNewCurrency("EUR")
	BRL = MustNewCurrency("BRL")
	MXN = MustNewCurrency("MXN")
	CLP = MustNewCurrency("CLP")
	COP = MustNewCurrency("COP")
)
