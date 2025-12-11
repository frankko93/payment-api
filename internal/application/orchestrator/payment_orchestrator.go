package orchestrator

import (
	"context"
	"fmt"
	"log"

	"github.com/franco/payment-api/internal/domain/payment"
	"github.com/franco/payment-api/internal/domain/shared"
	domerrors "github.com/franco/payment-api/internal/domain/shared/errors"
	vo "github.com/franco/payment-api/internal/domain/shared/valueobjects"
	"github.com/franco/payment-api/internal/domain/wallet"
)

// PaymentOrchestrator uses Domain Services and follows SRP
type PaymentOrchestrator struct {
	paymentRepo      PaymentRepository
	walletRepo       WalletRepository
	eventStore       shared.EventStore
	eventPublisher   EventPublisher
	paymentProcessor *payment.Processor
	topicArn         string
}

// PaymentRepository defines operations for Payment
type PaymentRepository interface {
	Save(ctx context.Context, pmt *payment.Payment) error
	FindByID(ctx context.Context, paymentID string) (*payment.Payment, error)
	Update(ctx context.Context, pmt *payment.Payment) error
}

// WalletRepository defines operations for Wallet
type WalletRepository interface {
	GetByUserID(ctx context.Context, userID string) (*wallet.Wallet, error)
	Save(ctx context.Context, wlt *wallet.Wallet) error
	Update(ctx context.Context, wlt *wallet.Wallet) error
}

// NewPaymentOrchestrator creates a new payment orchestrator
func NewPaymentOrchestrator(
	paymentRepo PaymentRepository,
	walletRepo WalletRepository,
	eventStore shared.EventStore,
	eventPublisher EventPublisher,
	topicArn string,
) *PaymentOrchestrator {
	return &PaymentOrchestrator{
		paymentRepo:      paymentRepo,
		walletRepo:       walletRepo,
		eventStore:       eventStore,
		eventPublisher:   eventPublisher,
		paymentProcessor: payment.NewProcessor(),
		topicArn:         topicArn,
	}
}

// HandlePaymentRequested processes PaymentRequested events
// Now much simpler - delegates to PaymentProcessor
func (o *PaymentOrchestrator) HandlePaymentRequested(ctx context.Context, event shared.Event) error {
	// Parse event
	paymentEvent, ok := event.(*payment.PaymentRequestedEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", event)
	}

	// Get payment
	pmt, err := o.paymentRepo.FindByID(ctx, paymentEvent.PaymentID())
	if err != nil {
		return domerrors.WrapError(domerrors.ErrCodePaymentNotFound, "payment not found", err)
	}

	// Get wallet
	wlt, err := o.walletRepo.GetByUserID(ctx, paymentEvent.UserID())
	if err != nil {
		// If wallet not found, fail the payment
		return o.failPayment(ctx, pmt, "WALLET_NOT_FOUND")
	}

	// Delegate to Domain Service
	result, err := o.paymentProcessor.Process(pmt, wlt)
	if err != nil {
		return err
	}

	// Check if processing failed
	if !result.Success {
		return o.failPayment(ctx, pmt, result.FailureReason)
	}

	// Save updated wallet
	if err := o.walletRepo.Update(ctx, wlt); err != nil {
		return domerrors.DatabaseError("update wallet", err)
	}

	// Publish WalletDebited event
	debitedEvent := wallet.NewWalletDebitedEvent(
		pmt.ID().String(),
		pmt.UserID().String(),
		pmt.Money().AmountFloat(),
		result.PreviousBalance.AmountFloat(),
		result.NewBalance.AmountFloat(),
		event.Metadata(),
	)

	if err := o.publishEvent(ctx, debitedEvent, pmt.ID().String()); err != nil {
		return err
	}

	// Publish ExternalPaymentRequested event
	externalEvent := payment.NewExternalPaymentRequestedEvent(
		pmt.ID().String(),
		pmt.UserID().String(),
		pmt.Money().AmountFloat(),
		pmt.Money().Currency().Code(),
		pmt.ServiceID().String(),
		event.Metadata(),
	)

	return o.publishEvent(ctx, externalEvent, pmt.ID().String())
}

// HandleExternalPaymentSucceeded processes successful external payments
func (o *PaymentOrchestrator) HandleExternalPaymentSucceeded(ctx context.Context, event shared.Event) error {
	successEvent, ok := event.(*payment.ExternalPaymentSucceededEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", event)
	}

	// Get payment
	pmt, err := o.paymentRepo.FindByID(ctx, successEvent.PaymentID())
	if err != nil {
		return err
	}

	// Mark as completed
	if err := pmt.MarkCompleted(successEvent.ExternalTransactionID()); err != nil {
		return err
	}

	// Save updated payment
	if err := o.paymentRepo.Update(ctx, pmt); err != nil {
		return err
	}

	// Publish PaymentCompleted event
	completedEvent := payment.NewPaymentCompletedEvent(
		pmt.ID().String(),
		pmt.UserID().String(),
		pmt.Money().AmountFloat(),
		pmt.ExternalTxID(),
		event.Metadata(),
	)

	return o.publishEvent(ctx, completedEvent, pmt.ID().String())
}

// HandleExternalPaymentFailed processes failed external payments
func (o *PaymentOrchestrator) HandleExternalPaymentFailed(ctx context.Context, event shared.Event) error {
	failedEvent, ok := event.(*payment.ExternalPaymentFailedEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", event)
	}

	return o.initiateRefund(ctx, failedEvent.PaymentID(), failedEvent.Reason(), event.Metadata())
}

// HandleExternalPaymentTimeout processes payment timeouts
func (o *PaymentOrchestrator) HandleExternalPaymentTimeout(ctx context.Context, event shared.Event) error {
	timeoutEvent, ok := event.(*payment.ExternalPaymentTimeoutEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", event)
	}

	return o.initiateRefund(ctx, timeoutEvent.PaymentID(), "TIMEOUT", event.Metadata())
}

// HandlePaymentRefundRequested processes refund requests
func (o *PaymentOrchestrator) HandlePaymentRefundRequested(ctx context.Context, event shared.Event) error {
	refundEvent, ok := event.(*payment.PaymentRefundRequestedEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", event)
	}

	// Get payment
	paymentID, _ := vo.NewPaymentID(refundEvent.PaymentID())
	pmt, err := o.paymentRepo.FindByID(ctx, paymentID.String())
	if err != nil {
		return err
	}

	// Get wallet
	wlt, err := o.walletRepo.GetByUserID(ctx, refundEvent.UserID())
	if err != nil {
		return err
	}

	// Delegate to Domain Service
	result, err := o.paymentProcessor.Refund(pmt, wlt)
	if err != nil {
		return err
	}

	// Save updated wallet
	if err := o.walletRepo.Update(ctx, wlt); err != nil {
		return err
	}

	// Publish WalletCredited event
	creditedEvent := wallet.NewWalletCreditedEvent(
		refundEvent.PaymentID(),
		refundEvent.UserID(),
		refundEvent.Amount(),
		result.PreviousBalance.AmountFloat(),
		result.NewBalance.AmountFloat(),
		"REFUND",
		event.Metadata(),
	)

	return o.publishEvent(ctx, creditedEvent, refundEvent.PaymentID())
}

// Private helper methods

func (o *PaymentOrchestrator) failPayment(ctx context.Context, pmt *payment.Payment, reason string) error {
	// Mark payment as failed
	if err := pmt.MarkFailed(reason); err != nil {
		log.Printf("Warning: failed to mark payment as failed: %v", err)
		// Continue anyway to publish the event
	}

	// Save updated payment
	if err := o.paymentRepo.Update(ctx, pmt); err != nil {
		return domerrors.DatabaseError("update payment status", err)
	}

	// Publish PaymentFailed event
	failedEvent := payment.NewPaymentFailedEvent(
		pmt.ID().String(),
		pmt.UserID().String(),
		pmt.Money().AmountFloat(),
		reason,
		shared.Metadata{},
	)

	return o.publishEvent(ctx, failedEvent, pmt.ID().String())
}

func (o *PaymentOrchestrator) initiateRefund(ctx context.Context, paymentID, reason string, metadata shared.Metadata) error {
	// Get payment
	pmt, err := o.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return err
	}

	// Mark as failed
	if err := pmt.MarkFailed(reason); err != nil {
		log.Printf("Warning: failed to mark payment as failed: %v", err)
	}

	// Save updated payment
	if err := o.paymentRepo.Update(ctx, pmt); err != nil {
		return err
	}

	// Publish PaymentRefundRequested event
	refundEvent := payment.NewPaymentRefundRequestedEvent(
		pmt.ID().String(),
		pmt.UserID().String(),
		pmt.Money().AmountFloat(),
		reason,
		metadata,
	)

	return o.publishEvent(ctx, refundEvent, pmt.ID().String())
}

func (o *PaymentOrchestrator) publishEvent(ctx context.Context, event shared.Event, paymentID string) error {
	// Store event (event sourcing)
	if err := o.eventStore.Append(ctx, event, paymentID); err != nil {
		return domerrors.WrapError(domerrors.ErrCodeEventStoreError, "failed to store event", err)
	}

	// Publish event
	if err := o.eventPublisher.Publish(ctx, event, o.topicArn); err != nil {
		return domerrors.EventPublishError(event.EventType(), err)
	}

	return nil
}
