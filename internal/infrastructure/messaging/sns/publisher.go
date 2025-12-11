package sns

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/franco/payment-api/internal/domain/payment"
	"github.com/franco/payment-api/internal/domain/shared"
	"github.com/franco/payment-api/internal/domain/wallet"
	"github.com/franco/payment-api/internal/observability"
)

// SNSPublisher implements EventPublisher using AWS SNS
type SNSPublisher struct {
	client *sns.Client
}

// NewSNSPublisher creates a new SNSPublisher
func NewSNSPublisher(client *sns.Client) *SNSPublisher {
	return &SNSPublisher{
		client: client,
	}
}

// Publish publishes an event to SNS
func (p *SNSPublisher) Publish(ctx context.Context, event shared.Event, topicArn string) error {
	// Serialize event
	eventData := map[string]interface{}{
		"eventType":  event.EventType(),
		"occurredAt": event.OccurredAt(),
		"metadata":   event.Metadata(),
	}

	// Add event-specific data based on type
	switch e := event.(type) {
	case *payment.PaymentRequestedEvent:
		eventData["paymentID"] = e.PaymentID()
		eventData["userID"] = e.UserID()
		eventData["amount"] = e.Amount()
		eventData["currency"] = e.Currency()
		eventData["serviceID"] = e.ServiceID()
		eventData["idempotencyKey"] = e.IdempotencyKey()

	case *wallet.WalletDebitedEvent:
		eventData["paymentID"] = e.PaymentID()
		eventData["userID"] = e.UserID()
		eventData["amount"] = e.Amount()
		eventData["prevBalance"] = e.PrevBalance()
		eventData["newBalance"] = e.NewBalance()

	case *wallet.WalletCreditedEvent:
		eventData["paymentID"] = e.PaymentID()
		eventData["userID"] = e.UserID()
		eventData["amount"] = e.Amount()
		eventData["prevBalance"] = e.PrevBalance()
		eventData["newBalance"] = e.NewBalance()
		eventData["reason"] = e.Reason()

	case *payment.ExternalPaymentRequestedEvent:
		eventData["paymentID"] = e.PaymentID()
		eventData["userID"] = e.UserID()
		eventData["amount"] = e.Amount()
		eventData["currency"] = e.Currency()
		eventData["serviceID"] = e.ServiceID()

	case *payment.ExternalPaymentSucceededEvent:
		eventData["paymentID"] = e.PaymentID()
		eventData["externalTransactionID"] = e.ExternalTransactionID()

	case *payment.ExternalPaymentFailedEvent:
		eventData["paymentID"] = e.PaymentID()
		eventData["reason"] = e.Reason()
		eventData["errorCode"] = e.ErrorCode()

	case *payment.ExternalPaymentTimeoutEvent:
		eventData["paymentID"] = e.PaymentID()
		eventData["timeoutDuration"] = e.TimeoutDuration().String()

	case *payment.PaymentCompletedEvent:
		eventData["paymentID"] = e.PaymentID()
		eventData["userID"] = e.UserID()
		eventData["amount"] = e.Amount()
		eventData["externalTransactionID"] = e.ExternalTransactionID()

	case *payment.PaymentFailedEvent:
		eventData["paymentID"] = e.PaymentID()
		eventData["userID"] = e.UserID()
		eventData["amount"] = e.Amount()
		eventData["reason"] = e.Reason()

	case *payment.PaymentRefundRequestedEvent:
		eventData["paymentID"] = e.PaymentID()
		eventData["userID"] = e.UserID()
		eventData["amount"] = e.Amount()
		eventData["reason"] = e.Reason()
	}

	messageBytes, err := json.Marshal(eventData)
	if err != nil {
		return err
	}

	// Publish to SNS
	_, err = p.client.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(topicArn),
		Message:  aws.String(string(messageBytes)),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"eventType": {
				DataType:    aws.String("String"),
				StringValue: aws.String(event.EventType()),
			},
		},
	})

	// Record observability event
	observability.RecordCustomEvent("EventPublished", map[string]interface{}{
		"eventType": event.EventType(),
		"topicArn":  topicArn,
		"metadata":  event.Metadata(),
	})

	return err
}
