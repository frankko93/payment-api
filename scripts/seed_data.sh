#!/bin/bash

# Script to seed initial data in LocalStack DynamoDB

echo "Seeding initial wallet data..."

# Use AWS CLI with LocalStack endpoint
ENDPOINT_URL="http://localhost:4566"

# Create a test wallet with balance
aws dynamodb put-item \
    --table-name Wallets \
    --item '{
        "userId": {"S": "user-123"},
        "balance": {"N": "1000.00"},
        "currency": {"S": "ARS"},
        "updatedAt": {"S": "2024-01-01T00:00:00Z"}
    }' \
    --endpoint-url $ENDPOINT_URL \
    --region us-east-1 \
    --no-sign-request

aws dynamodb put-item \
    --table-name Wallets \
    --item '{
        "userId": {"S": "user-456"},
        "balance": {"N": "50.00"},
        "currency": {"S": "ARS"},
        "updatedAt": {"S": "2024-01-01T00:00:00Z"}
    }' \
    --endpoint-url $ENDPOINT_URL \
    --region us-east-1 \
    --no-sign-request

echo ""
echo "✅ Data seeding complete!"
echo ""
echo "Test users created:"
echo "  - user-123: 1000.00 ARS (sufficient balance)"
echo "  - user-456: 50.00 ARS (low balance for testing insufficient funds)"
echo ""

