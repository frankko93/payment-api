package main

import (
	"context"
	"log"

	"github.com/franco/payment-api/internal/infrastructure"
)

func main() {
	ctx := context.Background()

	log.Println("Creating AWS clients...")
	awsClients, err := infrastructure.NewAWSClients(ctx)
	if err != nil {
		log.Fatalf("Failed to create AWS clients: %v", err)
	}

	log.Println("Creating DynamoDB tables, SNS topics, and SQS queues...")
	if err := infrastructure.EnsureInfrastructure(ctx, awsClients.DynamoDB, awsClients.SNS, awsClients.SQS); err != nil {
		log.Fatalf("Failed to setup infrastructure: %v", err)
	}

	log.Println("✅ All tables and queues created successfully!")
}
