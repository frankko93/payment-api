package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/franco/payment-api/internal/application/command"
	"github.com/franco/payment-api/internal/application/orchestrator"
	"github.com/franco/payment-api/internal/domain/shared"
	"github.com/franco/payment-api/internal/infrastructure"
	httpHandler "github.com/franco/payment-api/internal/infrastructure/http"
	"github.com/franco/payment-api/internal/infrastructure/messaging/sns"
	"github.com/franco/payment-api/internal/infrastructure/messaging/sqs"
	dynamodbRepo "github.com/franco/payment-api/internal/infrastructure/persistence/dynamodb"
)

func main() {
	ctx := context.Background()

	// Load configuration
	config := loadConfig()

	// Initialize AWS clients
	awsClients, err := infrastructure.NewAWSClients(ctx)
	if err != nil {
		log.Fatalf("Failed to create AWS clients: %v", err)
	}

	// Setup infrastructure (LocalStack)
	if err := infrastructure.EnsureInfrastructure(ctx, awsClients.DynamoDB, awsClients.SNS, awsClients.SQS); err != nil {
		log.Fatalf("Failed to setup infrastructure: %v", err)
	}

	// Give LocalStack time to set up subscriptions
	time.Sleep(2 * time.Second)

	// Initialize repositories
	paymentRepo := dynamodbRepo.NewDynamoDBPaymentRepository(awsClients.DynamoDB, "Payments")
	walletRepo := dynamodbRepo.NewDynamoDBWalletRepository(awsClients.DynamoDB, "Wallets")
	idempotencyStore := dynamodbRepo.NewDynamoDBIdempotencyStore(awsClients.DynamoDB, "Idempotency")
	eventStore := dynamodbRepo.NewDynamoDBEventStore(awsClients.DynamoDB, "EventStore")

	// Initialize event bus
	eventPublisher := sns.NewSNSPublisher(awsClients.SNS)
	eventConsumer := sqs.NewSQSConsumer(awsClients.SQS)

	// Initialize services
	createPaymentService := command.NewCreatePaymentService(
		paymentRepo,
		walletRepo,
		idempotencyStore,
		eventStore,
		eventPublisher,
		config.PaymentsTopicArn,
	)

	paymentOrchestrator := orchestrator.NewPaymentOrchestrator(
		paymentRepo,
		walletRepo,
		eventStore,
		eventPublisher,
		config.PaymentsTopicArn,
	)

	externalGatewayMock := orchestrator.NewExternalGatewayMock(
		eventStore,
		eventPublisher,
		config.PaymentsTopicArn,
		true, // always success for demo
	)

	// Start event consumers
	startEventConsumers(eventConsumer, paymentOrchestrator, externalGatewayMock, config)

	// Initialize HTTP server
	handler := httpHandler.NewPaymentHandler(createPaymentService)

	http.HandleFunc("/payments", handler.HandleCreatePayment)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := config.Port
	log.Printf("Starting payment API on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

type Config struct {
	PaymentsTopicArn        string
	WalletQueueURL          string
	PaymentQueueURL         string
	ExternalGatewayQueueURL string
	Port                    string
}

func loadConfig() Config {
	return Config{
		PaymentsTopicArn:        getEnv("PAYMENTS_TOPIC_ARN", "arn:aws:sns:us-east-1:000000000000:payments-events"),
		WalletQueueURL:          getEnv("WALLET_QUEUE_URL", "http://localhost:4566/000000000000/wallet-service-queue"),
		PaymentQueueURL:         getEnv("PAYMENT_QUEUE_URL", "http://localhost:4566/000000000000/payment-service-queue"),
		ExternalGatewayQueueURL: getEnv("EXTERNAL_GATEWAY_QUEUE_URL", "http://localhost:4566/000000000000/external-gateway-queue"),
		Port:                    getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// EventConsumer defines the interface for consuming events
type EventConsumer interface {
	StartConsuming(queueURL string, handler func(ctx context.Context, event shared.Event) error)
}

func startEventConsumers(
	consumer EventConsumer,
	orch *orchestrator.PaymentOrchestrator,
	gateway *orchestrator.ExternalGatewayMock,
	config Config,
) {
	// Payment service queue - handles PaymentRequested events
	consumer.StartConsuming(config.PaymentQueueURL, func(ctx context.Context, event shared.Event) error {
		switch event.EventType() {
		case "PaymentRequested":
			return orch.HandlePaymentRequested(ctx, event)
		case "ExternalPaymentSucceeded":
			return orch.HandleExternalPaymentSucceeded(ctx, event)
		case "ExternalPaymentFailed":
			return orch.HandleExternalPaymentFailed(ctx, event)
		case "ExternalPaymentTimeout":
			return orch.HandleExternalPaymentTimeout(ctx, event)
		default:
			log.Printf("Unhandled event type in payment queue: %s", event.EventType())
			return nil
		}
	})

	// Wallet service queue - handles refund requests
	consumer.StartConsuming(config.WalletQueueURL, func(ctx context.Context, event shared.Event) error {
		switch event.EventType() {
		case "PaymentRefundRequested":
			return orch.HandlePaymentRefundRequested(ctx, event)
		default:
			log.Printf("Unhandled event type in wallet queue: %s", event.EventType())
			return nil
		}
	})

	// External gateway queue - simulates external payment processing
	consumer.StartConsuming(config.ExternalGatewayQueueURL, func(ctx context.Context, event shared.Event) error {
		switch event.EventType() {
		case "ExternalPaymentRequested":
			return gateway.HandleExternalPaymentRequested(ctx, event)
		default:
			log.Printf("Unhandled event type in gateway queue: %s", event.EventType())
			return nil
		}
	})

	log.Println("Event consumers started")
}
