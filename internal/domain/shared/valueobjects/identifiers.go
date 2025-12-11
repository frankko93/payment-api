package valueobjects

import (
	"errors"

	"github.com/google/uuid"
)

// PaymentID represents a payment identifier
type PaymentID struct {
	value string
}

// NewPaymentID creates a new PaymentID from a string
func NewPaymentID(value string) (PaymentID, error) {
	if value == "" {
		return PaymentID{}, errors.New("payment ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(value); err != nil {
		return PaymentID{}, errors.New("invalid payment ID format: must be valid UUID")
	}

	return PaymentID{value: value}, nil
}

// GeneratePaymentID generates a new random PaymentID
func GeneratePaymentID() PaymentID {
	return PaymentID{value: uuid.New().String()}
}

// String returns the string representation
func (id PaymentID) String() string {
	return id.value
}

// Equals checks equality
func (id PaymentID) Equals(other PaymentID) bool {
	return id.value == other.value
}

// IsEmpty checks if ID is zero value
func (id PaymentID) IsEmpty() bool {
	return id.value == ""
}

// UserID represents a user identifier
type UserID struct {
	value string
}

// NewUserID creates a new UserID
func NewUserID(value string) (UserID, error) {
	if value == "" {
		return UserID{}, errors.New("user ID cannot be empty")
	}
	// Add more validation if needed (length, format, etc)
	if len(value) > 255 {
		return UserID{}, errors.New("user ID too long")
	}
	return UserID{value: value}, nil
}

// String returns the string representation
func (id UserID) String() string {
	return id.value
}

// Equals checks equality
func (id UserID) Equals(other UserID) bool {
	return id.value == other.value
}

// IsEmpty checks if ID is zero value
func (id UserID) IsEmpty() bool {
	return id.value == ""
}

// ServiceID represents a service identifier
type ServiceID struct {
	value string
}

// NewServiceID creates a new ServiceID
func NewServiceID(value string) (ServiceID, error) {
	if value == "" {
		return ServiceID{}, errors.New("service ID cannot be empty")
	}
	if len(value) > 255 {
		return ServiceID{}, errors.New("service ID too long")
	}
	return ServiceID{value: value}, nil
}

// String returns the string representation
func (id ServiceID) String() string {
	return id.value
}

// Equals checks equality
func (id ServiceID) Equals(other ServiceID) bool {
	return id.value == other.value
}

// IsEmpty checks if ID is zero value
func (id ServiceID) IsEmpty() bool {
	return id.value == ""
}

// IdempotencyKey represents an idempotency key
type IdempotencyKey struct {
	value string
}

// NewIdempotencyKey creates a new IdempotencyKey
func NewIdempotencyKey(value string) (IdempotencyKey, error) {
	if value == "" {
		return IdempotencyKey{}, errors.New("idempotency key cannot be empty")
	}
	if len(value) > 255 {
		return IdempotencyKey{}, errors.New("idempotency key too long (max 255 characters)")
	}
	return IdempotencyKey{value: value}, nil
}

// String returns the string representation
func (k IdempotencyKey) String() string {
	return k.value
}

// Equals checks equality
func (k IdempotencyKey) Equals(other IdempotencyKey) bool {
	return k.value == other.value
}

// IsEmpty checks if key is zero value
func (k IdempotencyKey) IsEmpty() bool {
	return k.value == ""
}
