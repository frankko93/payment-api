package unit

import (
	"context"
	"testing"

	"github.com/franco/payment-api/internal/application/command"
	vo "github.com/franco/payment-api/internal/domain/shared/valueobjects"
	"github.com/franco/payment-api/internal/domain/wallet"
	"github.com/franco/payment-api/tests/unit/fakes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePayment_HappyPath(t *testing.T) {
	// Arrange
	paymentRepo := fakes.NewPaymentRepositoryFake()
	walletRepo := fakes.NewWalletRepositoryFake()
	idempotencyStore := fakes.NewIdempotencyStoreFake()
	eventStore := fakes.NewEventStoreFake()
	eventPublisher := fakes.NewEventPublisherFake()

	// Create wallet with sufficient balance
	userID, _ := vo.NewUserID("user-123")
	balance := vo.MustNewMoney("500.00", "ARS")
	wlt, _ := wallet.NewWallet(userID, balance)
	walletRepo.SetWallet(wlt)

	service := command.NewCreatePaymentService(
		paymentRepo,
		walletRepo,
		idempotencyStore,
		eventStore,
		eventPublisher,
		"test-topic-arn",
	)

	req := command.CreatePaymentRequest{
		UserID:         "user-123",
		Amount:         100.50,
		Currency:       "ARS",
		ServiceID:      "service-123",
		IdempotencyKey: "unique-key-123",
		ClientID:       "web-app",
	}

	// Act
	result, err := service.Execute(context.Background(), req)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, result.PaymentID)
	assert.Equal(t, vo.PaymentStatusPending.String(), result.Status)

	// Verify payment was saved
	payment, err := paymentRepo.FindByID(context.Background(), result.PaymentID)
	require.NoError(t, err)
	assert.Equal(t, req.UserID, payment.UserID().String())
	assert.Equal(t, req.Amount, payment.Money().AmountFloat())
	assert.True(t, payment.Status().IsPending())

	// Verify event was published
	events := eventPublisher.GetEventsByType("PaymentRequested")
	assert.Len(t, events, 1)
}

func TestCreatePayment_Idempotency(t *testing.T) {
	// Arrange
	paymentRepo := fakes.NewPaymentRepositoryFake()
	walletRepo := fakes.NewWalletRepositoryFake()
	idempotencyStore := fakes.NewIdempotencyStoreFake()
	eventStore := fakes.NewEventStoreFake()
	eventPublisher := fakes.NewEventPublisherFake()

	// Create wallet with sufficient balance
	userID, _ := vo.NewUserID("user-123")
	balance := vo.MustNewMoney("500.00", "ARS")
	wlt, _ := wallet.NewWallet(userID, balance)
	walletRepo.SetWallet(wlt)

	service := command.NewCreatePaymentService(
		paymentRepo,
		walletRepo,
		idempotencyStore,
		eventStore,
		eventPublisher,
		"test-topic-arn",
	)

	req := command.CreatePaymentRequest{
		UserID:         "user-123",
		Amount:         100.50,
		Currency:       "ARS",
		ServiceID:      "service-123",
		IdempotencyKey: "unique-key-123",
		ClientID:       "web-app",
	}

	// Act - First request
	result1, err := service.Execute(context.Background(), req)
	require.NoError(t, err)

	// Act - Second request with same idempotency key
	result2, err := service.Execute(context.Background(), req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, result1.PaymentID, result2.PaymentID)
	assert.Equal(t, "ALREADY_PROCESSED", result2.Status)

	// Verify only one event was published
	events := eventPublisher.GetEventsByType("PaymentRequested")
	assert.Len(t, events, 1)
}

func TestCreatePayment_ValidationErrors(t *testing.T) {
	// Arrange
	service := command.NewCreatePaymentService(
		fakes.NewPaymentRepositoryFake(),
		fakes.NewWalletRepositoryFake(),
		fakes.NewIdempotencyStoreFake(),
		fakes.NewEventStoreFake(),
		fakes.NewEventPublisherFake(),
		"test-topic-arn",
	)

	tests := []struct {
		name        string
		req         command.CreatePaymentRequest
		expectedErr string
	}{
		{
			name: "Missing UserID",
			req: command.CreatePaymentRequest{
				Amount:         100,
				Currency:       "ARS",
				ServiceID:      "service-123",
				IdempotencyKey: "key-123",
			},
			expectedErr: "userID is required",
		},
		{
			name: "Invalid Amount",
			req: command.CreatePaymentRequest{
				UserID:         "user-123",
				Amount:         0,
				Currency:       "ARS",
				ServiceID:      "service-123",
				IdempotencyKey: "key-123",
			},
			expectedErr: "amount must be greater than zero",
		},
		{
			name: "Missing Currency",
			req: command.CreatePaymentRequest{
				UserID:         "user-123",
				Amount:         100,
				ServiceID:      "service-123",
				IdempotencyKey: "key-123",
			},
			expectedErr: "currency is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			_, err := service.Execute(context.Background(), tt.req)

			// Assert
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestCreatePayment_InsufficientFunds(t *testing.T) {
	// Arrange
	paymentRepo := fakes.NewPaymentRepositoryFake()
	walletRepo := fakes.NewWalletRepositoryFake()
	idempotencyStore := fakes.NewIdempotencyStoreFake()
	eventStore := fakes.NewEventStoreFake()
	eventPublisher := fakes.NewEventPublisherFake()

	// Create wallet with LOW balance
	userID, _ := vo.NewUserID("user-123")
	balance := vo.MustNewMoney("50.00", "ARS")
	wlt, _ := wallet.NewWallet(userID, balance)
	walletRepo.SetWallet(wlt)

	service := command.NewCreatePaymentService(
		paymentRepo,
		walletRepo,
		idempotencyStore,
		eventStore,
		eventPublisher,
		"test-topic-arn",
	)

	req := command.CreatePaymentRequest{
		UserID:         "user-123",
		Amount:         100.00, // More than balance
		Currency:       "ARS",
		ServiceID:      "service-123",
		IdempotencyKey: "unique-key-123",
		ClientID:       "web-app",
	}

	// Act
	result, err := service.Execute(context.Background(), req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "INSUFFICIENT_FUNDS")

	// Verify payment was NOT created
	assert.Equal(t, 0, len(paymentRepo.GetAll()))

	// Verify event was NOT published
	events := eventPublisher.GetEventsByType("PaymentRequested")
	assert.Len(t, events, 0)
}

func TestCreatePayment_WalletNotFound(t *testing.T) {
	// Arrange
	paymentRepo := fakes.NewPaymentRepositoryFake()
	walletRepo := fakes.NewWalletRepositoryFake()
	idempotencyStore := fakes.NewIdempotencyStoreFake()
	eventStore := fakes.NewEventStoreFake()
	eventPublisher := fakes.NewEventPublisherFake()

	// DO NOT create wallet - user doesn't exist

	service := command.NewCreatePaymentService(
		paymentRepo,
		walletRepo,
		idempotencyStore,
		eventStore,
		eventPublisher,
		"test-topic-arn",
	)

	req := command.CreatePaymentRequest{
		UserID:         "user-999", // Non-existent user
		Amount:         100.00,
		Currency:       "ARS",
		ServiceID:      "service-123",
		IdempotencyKey: "unique-key-123",
		ClientID:       "web-app",
	}

	// Act
	result, err := service.Execute(context.Background(), req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "WALLET_NOT_FOUND")

	// Verify payment was NOT created
	assert.Equal(t, 0, len(paymentRepo.GetAll()))
}
