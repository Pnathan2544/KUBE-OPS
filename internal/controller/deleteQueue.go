package controller

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	messagingv1 "github.com/yourusername/kubernetes-sqs-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// deleteSqsQueue deletes the SQS queue identified by Status.QueueURL.
// SQS deletion is synchronous — no waiters needed. The queue is gone
// within ~60 seconds, but the API call returns immediately on success.
func deleteSqsQueue(ctx context.Context, sqsQueue *messagingv1.SqsQueue) error {
	l := log.FromContext(ctx)

	if sqsQueue.Status.QueueURL == "" {
		// Queue was never successfully created — nothing to delete.
		l.Info("No QueueURL in status, skipping AWS deletion")
		return nil
	}

	l.Info("Deleting SQS queue", "queueURL", sqsQueue.Status.QueueURL)

	client := sqsClient(sqsQueue.Spec.Region)

	_, err := client.DeleteQueue(ctx, &sqs.DeleteQueueInput{
		QueueUrl: aws.String(sqsQueue.Status.QueueURL),
	})
	if err != nil {
		return fmt.Errorf("failed to delete SQS queue %q: %w", sqsQueue.Status.QueueURL, err)
	}

	l.Info("SQS queue deleted successfully", "queueURL", sqsQueue.Status.QueueURL)
	return nil
}
