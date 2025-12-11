package dynamodb

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBIdempotencyStore implements IdempotencyStore using DynamoDB
type DynamoDBIdempotencyStore struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBIdempotencyStore creates a new DynamoDBIdempotencyStore
func NewDynamoDBIdempotencyStore(client *dynamodb.Client, tableName string) *DynamoDBIdempotencyStore {
	return &DynamoDBIdempotencyStore{
		client:    client,
		tableName: tableName,
	}
}

type idempotencyItem struct {
	IdempotencyKey string `dynamodbav:"idempotencyKey"`
	PaymentID      string `dynamodbav:"paymentId"`
}

// GetPaymentIDByKey retrieves the payment ID for an idempotency key
func (s *DynamoDBIdempotencyStore) GetPaymentIDByKey(ctx context.Context, idempotencyKey string) (string, error) {
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"idempotencyKey": &types.AttributeValueMemberS{Value: idempotencyKey},
		},
	})

	if err != nil {
		return "", err
	}

	if result.Item == nil {
		return "", errors.New("key not found")
	}

	var item idempotencyItem
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return "", err
	}

	return item.PaymentID, nil
}

// SaveKey saves an idempotency key with its payment ID
func (s *DynamoDBIdempotencyStore) SaveKey(ctx context.Context, idempotencyKey, paymentID string) error {
	item := idempotencyItem{
		IdempotencyKey: idempotencyKey,
		PaymentID:      paymentID,
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
