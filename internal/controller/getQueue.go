package controller

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	messagingv1 "github.com/yourusername/kubernetes-sqs-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// QueueCheckResult is returned by checkSqsQueueExists.
type QueueCheckResult struct {
	// Exists is true if the queue was found in AWS.
	Exists bool
	// QueueURL is the current URL of the queue (may differ if queue was recreated).
	QueueURL string
	// QueueARN is the ARN of the queue.
	QueueARN string
}

// checkSqsQueueExists queries AWS to verify whether the queue in Status still exists.
// This is the drift detection mechanism — called on every reconcile where Status.QueueURL is set.
//
// This is the key improvement over the EC2 operator reference: if someone deletes the queue
// directly in the AWS console, the operator detects it and marks the resource for recreation.
func checkSqsQueueExists(ctx context.Context, sqsQueue *messagingv1.SqsQueue) (QueueCheckResult, error) {
	l := log.FromContext(ctx)

	client := sqsClient(sqsQueue.Spec.Region)

	// GetQueueUrl is a cheap read call that returns 404 if the queue doesn't exist.
	result, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(sqsQueue.Spec.QueueName),
	})
	if err != nil {
		// Check if this is a "queue does not exist" error specifically.
		var notFound *sqstypes.QueueDoesNotExist
		if errors.As(err, &notFound) {
			l.Info("Drift detected: queue no longer exists in AWS", "queueName", sqsQueue.Spec.QueueName)
			return QueueCheckResult{Exists: false}, nil
		}
		// Any other error (auth, network, etc.) — treat as transient.
		return QueueCheckResult{}, err
	}

	// Queue exists — also refresh the ARN in case it changed.
	queueARN, err := getQueueARN(ctx, client, *result.QueueUrl)
	if err != nil {
		// Non-fatal: queue exists, ARN fetch failed transiently.
		l.Error(err, "Queue exists but failed to refresh ARN")
		return QueueCheckResult{Exists: true, QueueURL: *result.QueueUrl}, nil
	}

	return QueueCheckResult{
		Exists:   true,
		QueueURL: *result.QueueUrl,
		QueueARN: queueARN,
	}, nil
}
