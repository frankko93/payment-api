package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/franco/payment-api/internal/domain/payment"
	domerrors "github.com/franco/payment-api/internal/domain/shared/errors"
	"github.com/franco/payment-api/internal/infrastructure/persistence/dynamodb/mappers"
)

// DynamoDBPaymentRepository implements PaymentRepository using DynamoDB
// This version uses mappers and Value Objects
type DynamoDBPaymentRepository struct {
	client    *dynamodb.Client
	tableName string
	mapper    *mappers.PaymentMapper
}

// NewDynamoDBPaymentRepository creates a new DynamoDBPaymentRepository
func NewDynamoDBPaymentRepository(client *dynamodb.Client, tableName string) *DynamoDBPaymentRepository {
	return &DynamoDBPaymentRepository{
		client:    client,
		tableName: tableName,
		mapper:    mappers.NewPaymentMapper(),
	}
}

// Save persists a payment to DynamoDB
func (r *DynamoDBPaymentRepository) Save(ctx context.Context, payment *payment.Payment) error {
	if payment == nil {
		return domerrors.NewDomainError(domerrors.ErrCodeValidationFailed, "payment cannot be nil")
	}

	// Convert to DB model
	dbModel, err := r.mapper.ToDBModel(payment)
	if err != nil {
		return domerrors.WrapError(domerrors.ErrCodeDatabaseError, "failed to convert payment to DB model", err)
	}

	// Marshal to DynamoDB attributes
	av, err := attributevalue.MarshalMap(dbModel)
	if err != nil {
		return domerrors.WrapError(domerrors.ErrCodeDatabaseError, "failed to marshal payment", err)
	}

	// Save to DynamoDB
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})

	if err != nil {
		return domerrors.DatabaseError("save payment", err)
	}

	return nil
}

// FindByID retrieves a payment by its ID
func (r *DynamoDBPaymentRepository) FindByID(ctx context.Context, paymentID string) (*payment.Payment, error) {
	if paymentID == "" {
		return nil, domerrors.NewDomainError(domerrors.ErrCodeValidationFailed, "payment ID cannot be empty")
	}

	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: paymentID},
		},
	})

	if err != nil {
		return nil, domerrors.DatabaseError("find payment", err)
	}

	if result.Item == nil {
		return nil, domerrors.PaymentNotFoundError(paymentID)
	}

	// Unmarshal from DynamoDB
	var dbModel mappers.PaymentDBModel
	if err := attributevalue.UnmarshalMap(result.Item, &dbModel); err != nil {
		return nil, domerrors.WrapError(domerrors.ErrCodeDatabaseError, "failed to unmarshal payment", err)
	}

	// Convert to domain model
	payment, err := r.mapper.ToDomain(&dbModel)
	if err != nil {
		return nil, domerrors.WrapError(domerrors.ErrCodeDatabaseError, "failed to convert to domain model", err)
	}

	return payment, nil
}

// Update saves changes to an existing payment
func (r *DynamoDBPaymentRepository) Update(ctx context.Context, payment *payment.Payment) error {
	// For simplicity, we just re-save the entire payment
	// In production, you might want more granular updates
	return r.Save(ctx, payment)
}
