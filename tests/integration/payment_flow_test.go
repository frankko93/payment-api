package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/franco/payment-api/internal/domain/payment"
	vo "github.com/franco/payment-api/internal/domain/shared/valueobjects"
	"github.com/franco/payment-api/internal/domain/wallet"
	"github.com/franco/payment-api/internal/infrastructure"
	dynamodbRepo "github.com/franco/payment-api/internal/infrastructure/persistence/dynamodb"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPaymentFlowIntegration tests the complete payment flow with LocalStack
func TestPaymentFlowIntegration(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=true to run.")
	}

	ctx := context.Background()

	// Setup AWS clients
	os.Setenv("USE_LOCALSTACK", "true")
	os.Setenv("AWS_ENDPOINT", "http://localhost:4566")
	os.Setenv("AWS_REGION", "us-east-1")

	awsClients, err := infrastructure.NewAWSClients(ctx)
	require.NoError(t, err)

	// Ensure infrastructure
	err = infrastructure.EnsureInfrastructure(ctx, awsClients.DynamoDB, awsClients.SNS, awsClients.SQS)
	require.NoError(t, err)

	// Give time for setup
	time.Sleep(2 * time.Second)

	// Seed test wallet
	walletRepo := dynamodbRepo.NewDynamoDBWalletRepository(awsClients.DynamoDB, "Wallets")
	userID, _ := vo.NewUserID("integration-user-123")
	balance := vo.MustNewMoney("1000.00", "ARS")
	testWallet, _ := wallet.NewWallet(userID, balance)
	err = walletRepo.Save(ctx, testWallet)
	require.NoError(t, err)

	// Test creating a payment via HTTP
	paymentReq := map[string]interface{}{
		"userId":         "integration-user-123",
		"amount":         150.75,
		"currency":       "ARS",
		"serviceId":      "integration-service",
		"idempotencyKey": "integration-test-" + time.Now().Format("20060102150405"),
		"clientId":       "test-client",
	}

	reqBody, _ := json.Marshal(paymentReq)

	// Assuming the API is running on localhost:8080
	// For a complete integration test, you'd start the server programmatically
	resp, err := http.Post("http://localhost:8080/payments", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		t.Logf("API not running, skipping HTTP test: %v", err)
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	assert.NotEmpty(t, result["paymentId"])
	assert.Equal(t, "PENDING", result["status"])

	// Wait for event processing
	time.Sleep(3 * time.Second)

	// Verify payment was processed
	paymentRepo := dynamodbRepo.NewDynamoDBPaymentRepository(awsClients.DynamoDB, "Payments")
	pmt, err := paymentRepo.FindByID(ctx, result["paymentId"].(string))

	if err == nil {
		// Payment should be completed or failed
		status := pmt.Status().String()
		assert.Contains(t, []string{"COMPLETED", "FAILED"}, status)
		t.Logf("Payment status: %s", status)
	}
}

func TestDynamoDBRepositories(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=true to run.")
	}

	ctx := context.Background()

	os.Setenv("USE_LOCALSTACK", "true")
	os.Setenv("AWS_ENDPOINT", "http://localhost:4566")
	os.Setenv("AWS_REGION", "us-east-1")

	awsClients, err := infrastructure.NewAWSClients(ctx)
	require.NoError(t, err)

	// Test Payment Repository
	t.Run("PaymentRepository", func(t *testing.T) {
		repo := dynamodbRepo.NewDynamoDBPaymentRepository(awsClients.DynamoDB, "Payments")

		paymentID := vo.GeneratePaymentID()
		userID, _ := vo.NewUserID("user-456")
		serviceID, _ := vo.NewServiceID("service-789")
		money := vo.MustNewMoney("200.00", "ARS")
		idempKey, _ := vo.NewIdempotencyKey("key-test")

		pmt, _ := payment.NewPayment(paymentID, userID, serviceID, money, idempKey)

		// Save
		err := repo.Save(ctx, pmt)
		require.NoError(t, err)

		// Find
		retrieved, err := repo.FindByID(ctx, paymentID.String())
		require.NoError(t, err)
		assert.True(t, retrieved.ID().Equals(pmt.ID()))
		assert.True(t, retrieved.Money().Equals(pmt.Money()))

		// Update (mark as completed)
		err = retrieved.MarkCompleted("ext-tx-123")
		require.NoError(t, err)
		err = repo.Update(ctx, retrieved)
		require.NoError(t, err)

		updated, _ := repo.FindByID(ctx, paymentID.String())
		assert.True(t, updated.Status().IsCompleted())
	})

	// Test Wallet Repository
	t.Run("WalletRepository", func(t *testing.T) {
		repo := dynamodbRepo.NewDynamoDBWalletRepository(awsClients.DynamoDB, "Wallets")

		userID, _ := vo.NewUserID("wallet-user-789")
		balance := vo.MustNewMoney("5000.00", "ARS")
		wlt, _ := wallet.NewWallet(userID, balance)

		// Save
		err := repo.Save(ctx, wlt)
		require.NoError(t, err)

		// Get
		retrieved, err := repo.GetByUserID(ctx, "wallet-user-789")
		require.NoError(t, err)
		assert.True(t, retrieved.Balance().Equals(wlt.Balance()))

		// Update Balance (debit 500)
		debitAmount := vo.MustNewMoney("500.00", "ARS")
		_, _, err = retrieved.Debit(debitAmount)
		require.NoError(t, err)
		err = repo.Update(ctx, retrieved)
		require.NoError(t, err)

		updated, _ := repo.GetByUserID(ctx, "wallet-user-789")
		expectedBalance := decimal.NewFromFloat(4500.00)
		assert.True(t, updated.Balance().Amount().Equal(expectedBalance))
	})
}
