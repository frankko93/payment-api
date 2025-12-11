package valueobjects

import (
	"fmt"
)

// PaymentStatus represents the status of a payment
// This is a type-safe enum to prevent invalid status values
type PaymentStatus int

const (
	PaymentStatusPending PaymentStatus = iota
	PaymentStatusCompleted
	PaymentStatusFailed
)

// String returns the string representation
func (s PaymentStatus) String() string {
	switch s {
	case PaymentStatusPending:
		return "PENDING"
	case PaymentStatusCompleted:
		return "COMPLETED"
	case PaymentStatusFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

// ParsePaymentStatus parses a string into PaymentStatus
func ParsePaymentStatus(s string) (PaymentStatus, error) {
	switch s {
	case "PENDING":
		return PaymentStatusPending, nil
	case "COMPLETED":
		return PaymentStatusCompleted, nil
	case "FAILED":
		return PaymentStatusFailed, nil
	default:
		return 0, fmt.Errorf("unknown payment status: %s", s)
	}
}

// IsPending checks if status is pending
func (s PaymentStatus) IsPending() bool {
	return s == PaymentStatusPending
}

// IsCompleted checks if status is completed
func (s PaymentStatus) IsCompleted() bool {
	return s == PaymentStatusCompleted
}

// IsFailed checks if status is failed
func (s PaymentStatus) IsFailed() bool {
	return s == PaymentStatusFailed
}

// IsTerminal checks if the status is terminal (no more transitions allowed)
func (s PaymentStatus) IsTerminal() bool {
	return s == PaymentStatusCompleted || s == PaymentStatusFailed
}

// CanTransitionTo checks if transition to target status is valid
func (s PaymentStatus) CanTransitionTo(target PaymentStatus) bool {
	switch s {
	case PaymentStatusPending:
		// Pending can transition to Completed or Failed
		return target == PaymentStatusCompleted || target == PaymentStatusFailed
	case PaymentStatusCompleted:
		// Terminal state - no transitions allowed
		return false
	case PaymentStatusFailed:
		// Terminal state - no transitions allowed
		return false
	default:
		return false
	}
}

// ValidateTransition returns an error if transition is invalid
func (s PaymentStatus) ValidateTransition(target PaymentStatus) error {
	if !s.CanTransitionTo(target) {
		return fmt.Errorf("invalid status transition from %s to %s", s.String(), target.String())
	}
	return nil
}
