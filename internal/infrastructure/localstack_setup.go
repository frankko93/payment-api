package infrastructure

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// EnsureInfrastructure creates all necessary AWS resources in LocalStack
func EnsureInfrastructure(
	ctx context.Context,
	dynamoClient *dynamodb.Client,
	snsClient *sns.Client,
	sqsClient *sqs.Client,
) error {
	log.Println("Setting up LocalStack infrastructure...")

	// Create DynamoDB tables
	if err := createDynamoDBTables(ctx, dynamoClient); err != nil {
		return fmt.Errorf("error creating DynamoDB tables: %w", err)
	}

	// Create SNS topic
	topicArn, err := createSNSTopic(ctx, snsClient, "payments-events")
	if err != nil {
		return fmt.Errorf("error creating SNS topic: %w", err)
	}

	// Create SQS queues with DLQ support
	queueNames := []string{
		"wallet-service-queue",
		"payment-service-queue",
		"external-gateway-queue",
	}

	for _, queueName := range queueNames {
		// Create DLQ for this queue
		dlqName := queueName + "-dlq"
		dlqURL, dlqArn, err := createDLQ(ctx, sqsClient, dlqName)
		if err != nil {
			return fmt.Errorf("error creating DLQ %s: %w", dlqName, err)
		}
		log.Printf("Created DLQ: %s (%s)", dlqName, dlqURL)

		// Create main queue with redrive policy
		queueURL, err := createSQSQueueWithDLQ(ctx, sqsClient, queueName, dlqArn)
		if err != nil {
			return fmt.Errorf("error creating SQS queue %s: %w", queueName, err)
		}
		log.Printf("Created SQS queue: %s (%s) with DLQ", queueName, queueURL)

		// Subscribe queue to SNS topic
		if err := subscribeSQSToSNS(ctx, snsClient, sqsClient, topicArn, queueURL); err != nil {
			return fmt.Errorf("error subscribing queue %s to SNS: %w", queueName, err)
		}
	}

	log.Println("LocalStack infrastructure setup complete!")
	return nil
}

func createDynamoDBTables(ctx context.Context, client *dynamodb.Client) error {
	tables := []struct {
		name      string
		keySchema []dynamodbtypes.KeySchemaElement
		attrDefs  []dynamodbtypes.AttributeDefinition
	}{
		{
			name: "Payments",
			keySchema: []dynamodbtypes.KeySchemaElement{
				{AttributeName: aws.String("id"), KeyType: dynamodbtypes.KeyTypeHash},
			},
			attrDefs: []dynamodbtypes.AttributeDefinition{
				{AttributeName: aws.String("id"), AttributeType: dynamodbtypes.ScalarAttributeTypeS},
			},
		},
		{
			name: "Wallets",
			keySchema: []dynamodbtypes.KeySchemaElement{
				{AttributeName: aws.String("userId"), KeyType: dynamodbtypes.KeyTypeHash},
			},
			attrDefs: []dynamodbtypes.AttributeDefinition{
				{AttributeName: aws.String("userId"), AttributeType: dynamodbtypes.ScalarAttributeTypeS},
			},
		},
		{
			name: "Idempotency",
			keySchema: []dynamodbtypes.KeySchemaElement{
				{AttributeName: aws.String("idempotencyKey"), KeyType: dynamodbtypes.KeyTypeHash},
			},
			attrDefs: []dynamodbtypes.AttributeDefinition{
				{AttributeName: aws.String("idempotencyKey"), AttributeType: dynamodbtypes.ScalarAttributeTypeS},
			},
		},
		{
			name: "EventStore",
			keySchema: []dynamodbtypes.KeySchemaElement{
				{AttributeName: aws.String("paymentId"), KeyType: dynamodbtypes.KeyTypeHash},
				{AttributeName: aws.String("eventId"), KeyType: dynamodbtypes.KeyTypeRange},
			},
			attrDefs: []dynamodbtypes.AttributeDefinition{
				{AttributeName: aws.String("paymentId"), AttributeType: dynamodbtypes.ScalarAttributeTypeS},
				{AttributeName: aws.String("eventId"), AttributeType: dynamodbtypes.ScalarAttributeTypeS},
			},
		},
	}

	for _, table := range tables {
		_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
			TableName:            aws.String(table.name),
			KeySchema:            table.keySchema,
			AttributeDefinitions: table.attrDefs,
			BillingMode:          dynamodbtypes.BillingModePayPerRequest,
		})

		if err != nil {
			// Ignore ResourceInUseException (table already exists)
			log.Printf("Table %s: %v", table.name, err)
		} else {
			log.Printf("Created DynamoDB table: %s", table.name)
		}
	}

	return nil
}

func createSNSTopic(ctx context.Context, client *sns.Client, topicName string) (string, error) {
	result, err := client.CreateTopic(ctx, &sns.CreateTopicInput{
		Name: aws.String(topicName),
	})

	if err != nil {
		return "", err
	}

	log.Printf("Created SNS topic: %s (%s)", topicName, *result.TopicArn)
	return *result.TopicArn, nil
}

func createSQSQueue(ctx context.Context, client *sqs.Client, queueName string) (string, error) {
	result, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	})

	if err != nil {
		return "", err
	}

	return *result.QueueUrl, nil
}

// createDLQ creates a Dead Letter Queue
func createDLQ(ctx context.Context, client *sqs.Client, dlqName string) (queueURL string, queueArn string, err error) {
	// Create the DLQ
	result, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(dlqName),
		Attributes: map[string]string{
			// DLQ doesn't need message retention limit, but we set it for safety
			"MessageRetentionPeriod": "1209600", // 14 days (max)
		},
	})

	if err != nil {
		return "", "", err
	}

	queueURL = *result.QueueUrl

	// Get queue ARN
	attrs, err := client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: []sqstypes.QueueAttributeName{sqstypes.QueueAttributeNameQueueArn},
	})

	if err != nil {
		return "", "", err
	}

	queueArn = attrs.Attributes["QueueArn"]
	return queueURL, queueArn, nil
}

// createSQSQueueWithDLQ creates a queue with Dead Letter Queue redrive policy
func createSQSQueueWithDLQ(ctx context.Context, client *sqs.Client, queueName string, dlqArn string) (string, error) {
	// Redrive policy: after 3 failed attempts, send to DLQ
	redrivePolicy := fmt.Sprintf(`{"deadLetterTargetArn":"%s","maxReceiveCount":3}`, dlqArn)

	result, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
		Attributes: map[string]string{
			"RedrivePolicy":                 redrivePolicy,
			"VisibilityTimeout":             "30",    // 30 seconds to process
			"MessageRetentionPeriod":        "86400", // 1 day
			"ReceiveMessageWaitTimeSeconds": "20",    // Long polling
		},
	})

	if err != nil {
		return "", err
	}

	return *result.QueueUrl, nil
}

func subscribeSQSToSNS(ctx context.Context, snsClient *sns.Client, sqsClient *sqs.Client, topicArn, queueURL string) error {
	// Get queue ARN
	attrs, err := sqsClient.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: []sqstypes.QueueAttributeName{sqstypes.QueueAttributeNameQueueArn},
	})

	if err != nil {
		return err
	}

	queueArn := attrs.Attributes["QueueArn"]

	// Subscribe queue to topic
	_, err = snsClient.Subscribe(ctx, &sns.SubscribeInput{
		Protocol: aws.String("sqs"),
		TopicArn: aws.String(topicArn),
		Endpoint: aws.String(queueArn),
	})

	return err
}
