package infrastructure

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// AWSClients holds all AWS service clients
type AWSClients struct {
	DynamoDB *dynamodb.Client
	SNS      *sns.Client
	SQS      *sqs.Client
}

// NewAWSClients creates AWS clients configured for LocalStack or AWS
func NewAWSClients(ctx context.Context) (*AWSClients, error) {
	useLocalStack := os.Getenv("USE_LOCALSTACK") == "true"
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "us-east-1"
	}

	var cfg aws.Config
	var err error

	if useLocalStack {
		endpoint := os.Getenv("AWS_ENDPOINT")
		if endpoint == "" {
			endpoint = "http://localhost:4566"
		}

		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(awsRegion),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				"test",
				"test",
				"",
			)),
		)

		if err != nil {
			return nil, err
		}

		// Create clients with custom endpoint
		return &AWSClients{
			DynamoDB: dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
				o.BaseEndpoint = aws.String(endpoint)
			}),
			SNS: sns.NewFromConfig(cfg, func(o *sns.Options) {
				o.BaseEndpoint = aws.String(endpoint)
			}),
			SQS: sqs.NewFromConfig(cfg, func(o *sqs.Options) {
				o.BaseEndpoint = aws.String(endpoint)
			}),
		}, nil
	}

	// Production AWS configuration
	cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		return nil, err
	}

	return &AWSClients{
		DynamoDB: dynamodb.NewFromConfig(cfg),
		SNS:      sns.NewFromConfig(cfg),
		SQS:      sqs.NewFromConfig(cfg),
	}, nil
}
