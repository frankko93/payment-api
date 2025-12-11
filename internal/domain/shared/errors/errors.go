package errors

import (
	"fmt"
)

// ErrorCode represents a typed error code for better error handling
type ErrorCode string

const (
	// Domain errors - Payment
	ErrCodeInsufficientFunds    ErrorCode = "INSUFFICIENT_FUNDS"
	ErrCodeInvalidAmount        ErrorCode = "INVALID_AMOUNT"
	ErrCodeInvalidCurrency      ErrorCode = "INVALID_CURRENCY"
	ErrCodeCurrencyMismatch     ErrorCode = "CURRENCY_MISMATCH"
	ErrCodeInvalidTransition    ErrorCode = "INVALID_STATE_TRANSITION"
	ErrCodePaymentNotFound      ErrorCode = "PAYMENT_NOT_FOUND"
	ErrCodePaymentAlreadyExists ErrorCode = "PAYMENT_ALREADY_EXISTS"
	ErrCodePaymentNotPending    ErrorCode = "PAYMENT_NOT_PENDING"

	// Domain errors - Wallet
	ErrCodeWalletNotFound   ErrorCode = "WALLET_NOT_FOUND"
	ErrCodeWalletDebitError ErrorCode = "WALLET_DEBIT_ERROR"
	ErrCodeNegativeBalance  ErrorCode = "NEGATIVE_BALANCE"

	// Domain errors - User
	ErrCodeUserMismatch ErrorCode = "USER_MISMATCH"

	// Application errors
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeDuplicateRequest ErrorCode = "DUPLICATE_REQUEST"
	ErrCodeIdempotencyError ErrorCode = "IDEMPOTENCY_ERROR"

	// Infrastructure errors
	ErrCodeDatabaseError     ErrorCode = "DATABASE_ERROR"
	ErrCodeEventPublishError ErrorCode = "EVENT_PUBLISH_ERROR"
	ErrCodeEventStoreError   ErrorCode = "EVENT_STORE_ERROR"
	ErrCodeRepositoryError   ErrorCode = "REPOSITORY_ERROR"

	// External service errors
	ErrCodeExternalGatewayError ErrorCode = "EXTERNAL_GATEWAY_ERROR"
	ErrCodeExternalTimeout      ErrorCode = "EXTERNAL_TIMEOUT"

	// Generic errors
	ErrCodeInternal ErrorCode = "INTERNAL_ERROR"
	ErrCodeUnknown  ErrorCode = "UNKNOWN_ERROR"
)

// DomainError represents a domain-specific error with structured information
type DomainError struct {
	Code    ErrorCode
	Message string
	Cause   error
	Details map[string]interface{}
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap implements the unwrap interface for error chaining
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// WithDetail adds a detail to the error
func (e *DomainError) WithDetail(key string, value interface{}) *DomainError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// NewDomainError creates a new DomainError
func NewDomainError(code ErrorCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WrapError wraps an existing error with a domain error
func WrapError(code ErrorCode, message string, cause error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Details: make(map[string]interface{}),
	}
}

// Helper constructors for common errors

// InsufficientFundsError creates an insufficient funds error
func InsufficientFundsError(required, available string) *DomainError {
	return NewDomainError(
		ErrCodeInsufficientFunds,
		"Insufficient funds in wallet",
	).WithDetail("required", required).WithDetail("available", available)
}

// PaymentNotFoundError creates a payment not found error
func PaymentNotFoundError(paymentID string) *DomainError {
	return NewDomainError(
		ErrCodePaymentNotFound,
		fmt.Sprintf("Payment not found: %s", paymentID),
	).WithDetail("paymentId", paymentID)
}

// WalletNotFoundError creates a wallet not found error
func WalletNotFoundError(userID string) *DomainError {
	return NewDomainError(
		ErrCodeWalletNotFound,
		fmt.Sprintf("Wallet not found for user: %s", userID),
	).WithDetail("userId", userID)
}

// InvalidStateTransitionError creates an invalid state transition error
func InvalidStateTransitionError(from, to string) *DomainError {
	return NewDomainError(
		ErrCodeInvalidTransition,
		fmt.Sprintf("Invalid state transition from %s to %s", from, to),
	).WithDetail("fromStatus", from).WithDetail("toStatus", to)
}

// CurrencyMismatchError creates a currency mismatch error
func CurrencyMismatchError(expected, actual string) *DomainError {
	return NewDomainError(
		ErrCodeCurrencyMismatch,
		fmt.Sprintf("Currency mismatch: expected %s but got %s", expected, actual),
	).WithDetail("expected", expected).WithDetail("actual", actual)
}

// ValidationError creates a validation error
func ValidationError(field, reason string) *DomainError {
	return NewDomainError(
		ErrCodeValidationFailed,
		fmt.Sprintf("Validation failed for %s: %s", field, reason),
	).WithDetail("field", field).WithDetail("reason", reason)
}

// DuplicateRequestError creates a duplicate request error
func DuplicateRequestError(idempotencyKey string) *DomainError {
	return NewDomainError(
		ErrCodeDuplicateRequest,
		"Request already processed",
	).WithDetail("idempotencyKey", idempotencyKey)
}

// DatabaseError creates a database error
func DatabaseError(operation string, cause error) *DomainError {
	return WrapError(
		ErrCodeDatabaseError,
		fmt.Sprintf("Database error during %s", operation),
		cause,
	).WithDetail("operation", operation)
}

// EventPublishError creates an event publish error
func EventPublishError(eventType string, cause error) *DomainError {
	return WrapError(
		ErrCodeEventPublishError,
		fmt.Sprintf("Failed to publish event: %s", eventType),
		cause,
	).WithDetail("eventType", eventType)
}

// IsErrorCode checks if an error has a specific error code
func IsErrorCode(err error, code ErrorCode) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == code
	}
	return false
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) ErrorCode {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code
	}
	return ErrCodeUnknown
}
