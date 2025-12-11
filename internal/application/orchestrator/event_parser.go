package orchestrator

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/franco/payment-api/internal/domain/payment"
	"github.com/franco/payment-api/internal/domain/shared"
	"github.com/franco/payment-api/internal/domain/wallet"
)

// ParseEvent parses a JSON event into a domain event
// TODO: This should be refactored into an Event Serializer with Strategy Pattern
func ParseEvent(eventType string, payload []byte) (shared.Event, error) {
	switch eventType {
	case "PaymentRequested":
		var data struct {
			PaymentID      string          `json:"paymentID"`
			UserID         string          `json:"userID"`
			Amount         float64         `json:"amount"`
			Currency       string          `json:"currency"`
			ServiceID      string          `json:"serviceID"`
			IdempotencyKey string          `json:"idempotencyKey"`
			Metadata       shared.Metadata `json:"metadata"`
		}
		if err := json.Unmarshal(payload, &data); err != nil {
			return nil, err
		}
		return payment.NewPaymentRequestedEvent(
			data.PaymentID,
			data.UserID,
			data.Amount,
			data.Currency,
			data.ServiceID,
			data.IdempotencyKey,
			data.Metadata,
		), nil

	case "ExternalPaymentSucceeded":
		var data struct {
			PaymentID             string          `json:"paymentID"`
			ExternalTransactionID string          `json:"externalTransactionID"`
			Metadata              shared.Metadata `json:"metadata"`
		}
		if err := json.Unmarshal(payload, &data); err != nil {
			return nil, err
		}
		return payment.NewExternalPaymentSucceededEvent(
			data.PaymentID,
			data.ExternalTransactionID,
			data.Metadata,
		), nil

	case "ExternalPaymentFailed":
		var data struct {
			PaymentID string          `json:"paymentID"`
			Reason    string          `json:"reason"`
			ErrorCode string          `json:"errorCode"`
			Metadata  shared.Metadata `json:"metadata"`
		}
		if err := json.Unmarshal(payload, &data); err != nil {
			return nil, err
		}
		return payment.NewExternalPaymentFailedEvent(
			data.PaymentID,
			data.Reason,
			data.ErrorCode,
			data.Metadata,
		), nil

	case "ExternalPaymentRequested":
		var data struct {
			PaymentID string          `json:"paymentID"`
			UserID    string          `json:"userID"`
			Amount    float64         `json:"amount"`
			Currency  string          `json:"currency"`
			ServiceID string          `json:"serviceID"`
			Metadata  shared.Metadata `json:"metadata"`
		}
		if err := json.Unmarshal(payload, &data); err != nil {
			return nil, err
		}
		return payment.NewExternalPaymentRequestedEvent(
			data.PaymentID,
			data.UserID,
			data.Amount,
			data.Currency,
			data.ServiceID,
			data.Metadata,
		), nil

	case "PaymentRefundRequested":
		var data struct {
			PaymentID string          `json:"paymentID"`
			UserID    string          `json:"userID"`
			Amount    float64         `json:"amount"`
			Reason    string          `json:"reason"`
			Metadata  shared.Metadata `json:"metadata"`
		}
		if err := json.Unmarshal(payload, &data); err != nil {
			return nil, err
		}
		return payment.NewPaymentRefundRequestedEvent(
			data.PaymentID,
			data.UserID,
			data.Amount,
			data.Reason,
			data.Metadata,
		), nil

	case "PaymentCompleted":
		var data struct {
			PaymentID             string          `json:"paymentID"`
			UserID                string          `json:"userID"`
			Amount                float64         `json:"amount"`
			ExternalTransactionID string          `json:"externalTransactionID"`
			Metadata              shared.Metadata `json:"metadata"`
		}
		if err := json.Unmarshal(payload, &data); err != nil {
			return nil, err
		}
		return payment.NewPaymentCompletedEvent(
			data.PaymentID,
			data.UserID,
			data.Amount,
			data.ExternalTransactionID,
			data.Metadata,
		), nil

	case "PaymentFailed":
		var data struct {
			PaymentID string          `json:"paymentID"`
			UserID    string          `json:"userID"`
			Amount    float64         `json:"amount"`
			Reason    string          `json:"reason"`
			Metadata  shared.Metadata `json:"metadata"`
		}
		if err := json.Unmarshal(payload, &data); err != nil {
			return nil, err
		}
		return payment.NewPaymentFailedEvent(
			data.PaymentID,
			data.UserID,
			data.Amount,
			data.Reason,
			data.Metadata,
		), nil

	case "WalletDebited":
		var data struct {
			PaymentID   string          `json:"paymentID"`
			UserID      string          `json:"userID"`
			Amount      float64         `json:"amount"`
			PrevBalance float64         `json:"prevBalance"`
			NewBalance  float64         `json:"newBalance"`
			Metadata    shared.Metadata `json:"metadata"`
		}
		if err := json.Unmarshal(payload, &data); err != nil {
			return nil, err
		}
		return wallet.NewWalletDebitedEvent(
			data.PaymentID,
			data.UserID,
			data.Amount,
			data.PrevBalance,
			data.NewBalance,
			data.Metadata,
		), nil

	case "WalletCredited":
		var data struct {
			PaymentID   string          `json:"paymentID"`
			UserID      string          `json:"userID"`
			Amount      float64         `json:"amount"`
			PrevBalance float64         `json:"prevBalance"`
			NewBalance  float64         `json:"newBalance"`
			Reason      string          `json:"reason"`
			Metadata    shared.Metadata `json:"metadata"`
		}
		if err := json.Unmarshal(payload, &data); err != nil {
			return nil, err
		}
		return wallet.NewWalletCreditedEvent(
			data.PaymentID,
			data.UserID,
			data.Amount,
			data.PrevBalance,
			data.NewBalance,
			data.Reason,
			data.Metadata,
		), nil

	case "ExternalPaymentTimeout":
		var data struct {
			PaymentID       string          `json:"paymentID"`
			TimeoutDuration string          `json:"timeoutDuration"`
			Metadata        shared.Metadata `json:"metadata"`
		}
		if err := json.Unmarshal(payload, &data); err != nil {
			return nil, err
		}
		// Parse duration from string
		duration, _ := time.ParseDuration(data.TimeoutDuration)
		return payment.NewExternalPaymentTimeoutEvent(
			data.PaymentID,
			duration,
			data.Metadata,
		), nil

	default:
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}
}
