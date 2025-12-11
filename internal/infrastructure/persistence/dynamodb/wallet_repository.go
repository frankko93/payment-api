package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	domerrors "github.com/franco/payment-api/internal/domain/shared/errors"
	"github.com/franco/payment-api/internal/domain/wallet"
	"github.com/franco/payment-api/internal/infrastructure/persistence/dynamodb/mappers"
)

// DynamoDBWalletRepository implements WalletRepository using DynamoDB
// This version uses mappers and Value Objects
type DynamoDBWalletRepository struct {
	client    *dynamodb.Client
	tableName string
	mapper    *mappers.WalletMapper
}

// NewDynamoDBWalletRepository creates a new DynamoDBWalletRepository
func NewDynamoDBWalletRepository(client *dynamodb.Client, tableName string) *DynamoDBWalletRepository {
	return &DynamoDBWalletRepository{
		client:    client,
		tableName: tableName,
		mapper:    mappers.NewWalletMapper(),
	}
}

// GetByUserID retrieves a wallet by user ID
func (r *DynamoDBWalletRepository) GetByUserID(ctx context.Context, userID string) (*wallet.Wallet, error) {
	if userID == "" {
		return nil, domerrors.NewDomainError(domerrors.ErrCodeValidationFailed, "user ID cannot be empty")
	}

	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: userID},
		},
	})

	if err != nil {
		return nil, domerrors.DatabaseError("get wallet", err)
	}

	if result.Item == nil {
		return nil, domerrors.WalletNotFoundError(userID)
	}

	// Unmarshal from DynamoDB
	var dbModel mappers.WalletDBModel
	if err := attributevalue.UnmarshalMap(result.Item, &dbModel); err != nil {
		return nil, domerrors.WrapError(domerrors.ErrCodeDatabaseError, "failed to unmarshal wallet", err)
	}

	// Convert to domain model
	wallet, err := r.mapper.ToDomain(&dbModel)
	if err != nil {
		return nil, domerrors.WrapError(domerrors.ErrCodeDatabaseError, "failed to convert to domain model", err)
	}

	return wallet, nil
}

// Save persists a wallet to DynamoDB
func (r *DynamoDBWalletRepository) Save(ctx context.Context, wallet *wallet.Wallet) error {
	if wallet == nil {
		return domerrors.NewDomainError(domerrors.ErrCodeValidationFailed, "wallet cannot be nil")
	}

	// Convert to DB model
	dbModel, err := r.mapper.ToDBModel(wallet)
	if err != nil {
		return domerrors.WrapError(domerrors.ErrCodeDatabaseError, "failed to convert wallet to DB model", err)
	}

	// Marshal to DynamoDB attributes
	av, err := attributevalue.MarshalMap(dbModel)
	if err != nil {
		return domerrors.WrapError(domerrors.ErrCodeDatabaseError, "failed to marshal wallet", err)
	}

	// Save to DynamoDB
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})

	if err != nil {
		return domerrors.DatabaseError("save wallet", err)
	}

	return nil
}

// Update saves changes to an existing wallet
func (r *DynamoDBWalletRepository) Update(ctx context.Context, wallet *wallet.Wallet) error {
	return r.Save(ctx, wallet)
}
