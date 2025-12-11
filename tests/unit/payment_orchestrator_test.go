package unit

import (
	"context"
	"testing"

	"github.com/franco/payment-api/internal/application/orchestrator"
	"github.com/franco/payment-api/internal/domain/payment"
	"github.com/franco/payment-api/internal/domain/shared"
	vo "github.com/franco/payment-api/internal/domain/shared/valueobjects"
	"github.com/franco/payment-api/internal/domain/wallet"
	"github.com/franco/payment-api/tests/unit/fakes"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentOrchestrator_InsufficientFunds(t *testing.T) {
	// Arrange
	paymentRepo := fakes.NewPaymentRepositoryFake()
	walletRepo := fakes.NewWalletRepositoryFake()
	eventStore := fakes.NewEventStoreFake()
	eventPublisher := fakes.NewEventPublisherFake()

	// Seed wallet with low balance
	userID, _ := vo.NewUserID("user-123")
	balance := vo.MustNewMoney("50.00", "ARS")
	wallet, _ := wallet.NewWallet(userID, balance)
	walletRepo.SetWallet(wallet)

	orch := orchestrator.NewPaymentOrchestrator(
		paymentRepo,
		walletRepo,
		eventStore,
		eventPublisher,
		"test-topic-arn",
	)

	// Create payment that requires more than available balance
	paymentID := vo.GeneratePaymentID()
	serviceID, _ := vo.NewServiceID("service-123")
	amount := vo.MustNewMoney("100.00", "ARS")
	idempKey, _ := vo.NewIdempotencyKey("key-123")
	pmt, _ := payment.NewPayment(paymentID, userID, serviceID, amount, idempKey)
	paymentRepo.Save(context.Background(), pmt)

	metadata := shared.Metadata{
		ClientID:  "test-client",
		RequestID: "req-123",
		Source:    "test",
	}

	event := payment.NewPaymentRequestedEvent(
		paymentID.String(),
		"user-123",
		100.00,
		"ARS",
		"service-123",
		"key-123",
		metadata,
	)

	// Act
	err := orch.HandlePaymentRequested(context.Background(), event)

	// Assert
	require.NoError(t, err)

	// Verify payment was marked as failed
	updatedPayment, _ := paymentRepo.FindByID(context.Background(), paymentID.String())
	assert.True(t, updatedPayment.Status().IsFailed())
	assert.Equal(t, "INSUFFICIENT_FUNDS", updatedPayment.FailureReason())

	// Verify PaymentFailed event was published
	events := eventPublisher.GetEventsByType("PaymentFailed")
	assert.Len(t, events, 1)
}

func TestPaymentOrchestrator_SuccessfulDebit(t *testing.T) {
	// Arrange
	paymentRepo := fakes.NewPaymentRepositoryFake()
	walletRepo := fakes.NewWalletRepositoryFake()
	eventStore := fakes.NewEventStoreFake()
	eventPublisher := fakes.NewEventPublisherFake()

	// Seed wallet with sufficient balance
	userID, _ := vo.NewUserID("user-123")
	balance := vo.MustNewMoney("500.00", "ARS")
	wallet, _ := wallet.NewWallet(userID, balance)
	walletRepo.SetWallet(wallet)

	orch := orchestrator.NewPaymentOrchestrator(
		paymentRepo,
		walletRepo,
		eventStore,
		eventPublisher,
		"test-topic-arn",
	)

	paymentID := vo.GeneratePaymentID()
	serviceID, _ := vo.NewServiceID("service-123")
	amount := vo.MustNewMoney("100.00", "ARS")
	idempKey, _ := vo.NewIdempotencyKey("key-123")
	pmt, _ := payment.NewPayment(paymentID, userID, serviceID, amount, idempKey)
	paymentRepo.Save(context.Background(), pmt)

	metadata := shared.Metadata{
		ClientID:  "test-client",
		RequestID: "req-123",
		Source:    "test",
	}

	event := payment.NewPaymentRequestedEvent(
		paymentID.String(),
		"user-123",
		100.00,
		"ARS",
		"service-123",
		"key-123",
		metadata,
	)

	// Act
	err := orch.HandlePaymentRequested(context.Background(), event)

	// Assert
	require.NoError(t, err)

	// Verify wallet was debited
	updatedWallet, _ := walletRepo.GetByUserID(context.Background(), "user-123")
	expectedBalance := decimal.NewFromFloat(400.00)
	assert.True(t, updatedWallet.Balance().Amount().Equal(expectedBalance))

	// Verify events were published
	debitedEvents := eventPublisher.GetEventsByType("WalletDebited")
	assert.Len(t, debitedEvents, 1)

	externalEvents := eventPublisher.GetEventsByType("ExternalPaymentRequested")
	assert.Len(t, externalEvents, 1)
}

func TestPaymentOrchestrator_ExternalPaymentSuccess(t *testing.T) {
	// Arrange
	paymentRepo := fakes.NewPaymentRepositoryFake()
	walletRepo := fakes.NewWalletRepositoryFake()
	eventStore := fakes.NewEventStoreFake()
	eventPublisher := fakes.NewEventPublisherFake()

	orch := orchestrator.NewPaymentOrchestrator(
		paymentRepo,
		walletRepo,
		eventStore,
		eventPublisher,
		"test-topic-arn",
	)

	paymentID := vo.GeneratePaymentID()
	userID, _ := vo.NewUserID("user-123")
	serviceID, _ := vo.NewServiceID("service-123")
	amount := vo.MustNewMoney("100.00", "ARS")
	idempKey, _ := vo.NewIdempotencyKey("key-123")
	pmt, _ := payment.NewPayment(paymentID, userID, serviceID, amount, idempKey)
	paymentRepo.Save(context.Background(), pmt)

	metadata := shared.Metadata{
		ClientID:  "test-client",
		RequestID: "req-123",
		Source:    "test",
	}

	event := payment.NewExternalPaymentSucceededEvent(
		paymentID.String(),
		"external-tx-456",
		metadata,
	)

	// Act
	err := orch.HandleExternalPaymentSucceeded(context.Background(), event)

	// Assert
	require.NoError(t, err)

	// Verify payment was marked as completed
	updatedPayment, _ := paymentRepo.FindByID(context.Background(), paymentID.String())
	assert.True(t, updatedPayment.Status().IsCompleted())
	assert.Equal(t, "external-tx-456", updatedPayment.ExternalTxID())

	// Verify PaymentCompleted event was published
	events := eventPublisher.GetEventsByType("PaymentCompleted")
	assert.Len(t, events, 1)
}

func TestPaymentOrchestrator_RefundFlow(t *testing.T) {
	// Arrange
	paymentRepo := fakes.NewPaymentRepositoryFake()
	walletRepo := fakes.NewWalletRepositoryFake()
	eventStore := fakes.NewEventStoreFake()
	eventPublisher := fakes.NewEventPublisherFake()

	userID, _ := vo.NewUserID("user-123")
	balance := vo.MustNewMoney("400.00", "ARS")
	wallet, _ := wallet.NewWallet(userID, balance)
	walletRepo.SetWallet(wallet)

	orch := orchestrator.NewPaymentOrchestrator(
		paymentRepo,
		walletRepo,
		eventStore,
		eventPublisher,
		"test-topic-arn",
	)

	paymentID := vo.GeneratePaymentID()
	serviceID, _ := vo.NewServiceID("service-123")
	amount := vo.MustNewMoney("100.00", "ARS")
	idempKey, _ := vo.NewIdempotencyKey("key-123")
	pmt, _ := payment.NewPayment(paymentID, userID, serviceID, amount, idempKey)

	// Mark payment as failed so it can be refunded
	pmt.MarkFailed("EXTERNAL_FAILURE")
	paymentRepo.Save(context.Background(), pmt)

	metadata := shared.Metadata{
		ClientID:  "test-client",
		RequestID: "req-123",
		Source:    "test",
	}

	refundEvent := payment.NewPaymentRefundRequestedEvent(
		paymentID.String(),
		"user-123",
		100.00,
		"EXTERNAL_FAILURE",
		metadata,
	)

	// Act
	err := orch.HandlePaymentRefundRequested(context.Background(), refundEvent)

	// Assert
	require.NoError(t, err)

	// Verify wallet was credited
	updatedWallet, _ := walletRepo.GetByUserID(context.Background(), "user-123")
	expectedBalance := decimal.NewFromFloat(500.00)
	assert.True(t, updatedWallet.Balance().Amount().Equal(expectedBalance))

	// Verify WalletCredited event was published
	events := eventPublisher.GetEventsByType("WalletCredited")
	assert.Len(t, events, 1)
}
