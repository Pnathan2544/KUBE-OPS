package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	messagingv1 "github.com/yourusername/kubernetes-sqs-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// createSqsQueue creates a new SQS queue in AWS based on the given SqsQueue spec.
// Unlike EC2, SQS CreateQueue is synchronous — no waiters needed.
// Returns a CreatedQueueInfo with the QueueURL and QueueARN on success.
func createSqsQueue(ctx context.Context, sqsQueue *messagingv1.SqsQueue) (*messagingv1.CreatedQueueInfo, error) {
	l := log.FromContext(ctx)

	queueName := sqsQueue.Spec.QueueName

	// FIFO queues MUST have names ending in ".fifo" — enforce this automatically.
	if sqsQueue.Spec.Fifo && !strings.HasSuffix(queueName, ".fifo") {
		queueName = queueName + ".fifo"
	}

	l.Info("Creating SQS queue", "queueName", queueName, "region", sqsQueue.Spec.Region, "fifo", sqsQueue.Spec.Fifo)

	client := sqsClient(sqsQueue.Spec.Region)

	// Build the attributes map. SQS uses string values for all attributes.
	attributes := map[string]string{}

	if sqsQueue.Spec.VisibilityTimeoutSeconds > 0 {
		attributes[string(sqstypes.QueueAttributeNameVisibilityTimeout)] =
			fmt.Sprintf("%d", sqsQueue.Spec.VisibilityTimeoutSeconds)
	}
	if sqsQueue.Spec.MessageRetentionSeconds > 0 {
		attributes[string(sqstypes.QueueAttributeNameMessageRetentionPeriod)] =
			fmt.Sprintf("%d", sqsQueue.Spec.MessageRetentionSeconds)
	}
	if sqsQueue.Spec.DelaySeconds > 0 {
		attributes[string(sqstypes.QueueAttributeNameDelaySeconds)] =
			fmt.Sprintf("%d", sqsQueue.Spec.DelaySeconds)
	}
	if sqsQueue.Spec.MaxMessageSizeBytes > 0 {
		attributes[string(sqstypes.QueueAttributeNameMaximumMessageSize)] =
			fmt.Sprintf("%d", sqsQueue.Spec.MaxMessageSizeBytes)
	}
	if sqsQueue.Spec.Fifo {
		attributes[string(sqstypes.QueueAttributeNameFifoQueue)] = "true"
	}

	createInput := &sqs.CreateQueueInput{
		QueueName:  aws.String(queueName),
		Attributes: attributes,
		Tags:       sqsQueue.Spec.Tags,
	}

	createResult, err := client.CreateQueue(ctx, createInput)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQS queue %q: %w", queueName, err)
	}

	l.Info("SQS queue created successfully", "queueURL", *createResult.QueueUrl)

	// Fetch the ARN — it is not returned by CreateQueue, only by GetQueueAttributes.
	queueARN, err := getQueueARN(ctx, client, *createResult.QueueUrl)
	if err != nil {
		// Non-fatal: we have the URL, ARN can be fetched on the next reconcile.
		l.Error(err, "Queue created but failed to fetch ARN — will retry on next reconcile")
		return &messagingv1.CreatedQueueInfo{QueueURL: *createResult.QueueUrl}, nil
	}

	return &messagingv1.CreatedQueueInfo{
		QueueURL: *createResult.QueueUrl,
		QueueARN: queueARN,
	}, nil
}

// getQueueARN fetches the ARN for an existing queue by URL.
func getQueueARN(ctx context.Context, client *sqs.Client, queueURL string) (string, error) {
	attrResult, err := client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: []sqstypes.QueueAttributeName{sqstypes.QueueAttributeNameQueueArn},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get queue attributes: %w", err)
	}

	arn, ok := attrResult.Attributes[string(sqstypes.QueueAttributeNameQueueArn)]
	if !ok {
		return "", fmt.Errorf("QueueArn not found in attributes response")
	}

	return arn, nil
}
