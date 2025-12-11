package dynamodb

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/franco/payment-api/internal/domain/shared"
	"github.com/google/uuid"
)

// DynamoDBEventStore implements EventStore using DynamoDB
type DynamoDBEventStore struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBEventStore creates a new DynamoDBEventStore
func NewDynamoDBEventStore(client *dynamodb.Client, tableName string) *DynamoDBEventStore {
	return &DynamoDBEventStore{
		client:    client,
		tableName: tableName,
	}
}

type eventItem struct {
	EventID    string `dynamodbav:"eventId"`
	PaymentID  string `dynamodbav:"paymentId"`
	EventType  string `dynamodbav:"eventType"`
	Payload    string `dynamodbav:"payload"`
	Metadata   string `dynamodbav:"metadata"`
	OccurredAt string `dynamodbav:"occurredAt"`
}

// Append stores an event in DynamoDB
func (s *DynamoDBEventStore) Append(ctx context.Context, event shared.Event, paymentID string) error {
	eventID := uuid.New().String()

	// Marshal event to JSON (simplified - in production would use proper serialization)
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	metadataBytes, err := json.Marshal(event.Metadata())
	if err != nil {
		return err
	}

	item := eventItem{
		EventID:    eventID,
		PaymentID:  paymentID,
		EventType:  event.EventType(),
		Payload:    string(payloadBytes),
		Metadata:   string(metadataBytes),
		OccurredAt: event.OccurredAt().Format(time.RFC3339),
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      av,
	})

	return err
}

// ListByPaymentID retrieves all events for a payment
func (s *DynamoDBEventStore) ListByPaymentID(ctx context.Context, paymentID string) ([]shared.StoredEvent, error) {
	result, err := s.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		KeyConditionExpression: aws.String("paymentId = :paymentId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":paymentId": &types.AttributeValueMemberS{Value: paymentID},
		},
	})

	if err != nil {
		return nil, err
	}

	events := make([]shared.StoredEvent, 0, len(result.Items))
	for _, item := range result.Items {
		var eventItem eventItem
		if err := attributevalue.UnmarshalMap(item, &eventItem); err != nil {
			return nil, err
		}

		events = append(events, shared.StoredEvent{
			EventID:    eventItem.EventID,
			EventType:  eventItem.EventType,
			PaymentID:  eventItem.PaymentID,
			Payload:    eventItem.Payload,
			OccurredAt: eventItem.OccurredAt,
			Metadata:   eventItem.Metadata,
		})
	}

	return events, nil
}
