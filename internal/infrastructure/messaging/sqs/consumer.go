package sqs

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/franco/payment-api/internal/application/orchestrator"
	"github.com/franco/payment-api/internal/domain/shared"
	"github.com/franco/payment-api/internal/observability"
)

// SQSConsumer implements EventConsumer using AWS SQS
type SQSConsumer struct {
	client *sqs.Client
}

// NewSQSConsumer creates a new SQSConsumer
func NewSQSConsumer(client *sqs.Client) *SQSConsumer {
	return &SQSConsumer{
		client: client,
	}
}

// StartConsuming starts consuming messages from an SQS queue
func (c *SQSConsumer) StartConsuming(queueURL string, handler func(ctx context.Context, event shared.Event) error) {
	go func() {
		for {
			result, err := c.client.ReceiveMessage(context.Background(), &sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(queueURL),
				MaxNumberOfMessages: 10,
				WaitTimeSeconds:     20,
				MessageAttributeNames: []string{
					"All",
				},
			})

			if err != nil {
				log.Printf("Error receiving messages from queue %s: %v", queueURL, err)
				continue
			}

			for _, message := range result.Messages {
				if err := c.processMessage(message, handler); err != nil {
					log.Printf("Error processing message: %v", err)
					// DO NOT delete message - let it return to queue
					// SQS will increment ReceiveCount
					// After maxReceiveCount (3), SQS will automatically move to DLQ
					continue
				}

				// Delete message ONLY after successful processing
				_, err := c.client.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
					QueueUrl:      aws.String(queueURL),
					ReceiptHandle: message.ReceiptHandle,
				})

				if err != nil {
					log.Printf("Error deleting message: %v", err)
					// If delete fails, message will be reprocessed (idempotent handlers)
				}
			}
		}
	}()
}

func (c *SQSConsumer) processMessage(message types.Message, handler func(ctx context.Context, event shared.Event) error) error {
	// Parse SNS message wrapper
	var snsMessage struct {
		Message           string `json:"Message"`
		MessageAttributes map[string]struct {
			Type  string `json:"Type"`
			Value string `json:"Value"`
		} `json:"MessageAttributes"`
	}

	if err := json.Unmarshal([]byte(*message.Body), &snsMessage); err != nil {
		return err
	}

	// Parse event data
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(snsMessage.Message), &eventData); err != nil {
		return err
	}

	eventType, ok := eventData["eventType"].(string)
	if !ok {
		log.Printf("Missing eventType in message")
		return nil
	}

	// Parse event using application parser
	eventBytes, _ := json.Marshal(eventData)
	event, err := orchestrator.ParseEvent(eventType, eventBytes)
	if err != nil {
		log.Printf("Error parsing event type %s: %v", eventType, err)
		return nil
	}

	// Record observability event
	observability.RecordCustomEvent("EventConsumed", map[string]interface{}{
		"eventType": eventType,
		"metadata":  event.Metadata(),
	})

	// Handle event
	return handler(context.Background(), event)
}
