package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SqsQueueSpec defines the desired state of SqsQueue.
type SqsQueueSpec struct {
	// QueueName is the name of the SQS queue to create in AWS.
	// For FIFO queues, the name must end with ".fifo" — the operator enforces this automatically.
	QueueName string `json:"queueName"`

	// Region is the AWS region where the queue will be created (e.g., "us-east-1").
	Region string `json:"region"`

	// Fifo indicates whether the queue should be a FIFO (first-in, first-out) queue.
	// FIFO queues guarantee ordering and exactly-once processing.
	// +optional
	Fifo bool `json:"fifo,omitempty"`

	// VisibilityTimeoutSeconds is the duration (in seconds) that a received message is
	// hidden from other consumers. Defaults to 30 seconds. Valid range: 0–43200.
	// +optional
	VisibilityTimeoutSeconds int32 `json:"visibilityTimeoutSeconds,omitempty"`

	// MessageRetentionSeconds is how long (in seconds) SQS retains a message.
	// Defaults to 345600 (4 days). Valid range: 60–1209600 (14 days).
	// +optional
	MessageRetentionSeconds int32 `json:"messageRetentionSeconds,omitempty"`

	// DelaySeconds is the time (in seconds) that the delivery of all messages
	// in the queue is delayed. Valid range: 0–900. Default: 0.
	// +optional
	DelaySeconds int32 `json:"delaySeconds,omitempty"`

	// MaxMessageSizeBytes is the limit of how many bytes a message can contain.
	// Valid range: 1024–262144 (256 KiB). Default: 262144.
	// +optional
	MaxMessageSizeBytes int32 `json:"maxMessageSizeBytes,omitempty"`

	// Tags are AWS resource tags applied to the queue.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

// SqsQueueStatus defines the observed state of SqsQueue.
type SqsQueueStatus struct {
	// QueueURL is the URL of the created SQS queue, returned by AWS after creation.
	// This is the primary identifier used for all subsequent API calls.
	QueueURL string `json:"queueURL,omitempty"`

	// QueueARN is the Amazon Resource Name of the queue.
	// Useful for attaching IAM policies or SNS subscriptions.
	QueueARN string `json:"queueARN,omitempty"`

	// State reflects whether the queue currently exists in AWS.
	// Possible values: "Active", "NotFound", "Unknown".
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="QueueName",type="string",JSONPath=".spec.queueName",description="The SQS queue name"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="The current state of the queue"
// +kubebuilder:printcolumn:name="FIFO",type="boolean",JSONPath=".spec.fifo",description="Whether the queue is FIFO"
// +kubebuilder:printcolumn:name="QueueURL",type="string",JSONPath=".status.queueURL",description="The SQS queue URL"

// SqsQueue is the Schema for the sqsqueues API.
type SqsQueue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SqsQueueSpec   `json:"spec,omitempty"`
	Status SqsQueueStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SqsQueueList contains a list of SqsQueue.
type SqsQueueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SqsQueue `json:"items"`
}

// CreatedQueueInfo holds the result of a successful queue creation.
// Used to pass data from the AWS layer back to the reconciler.
type CreatedQueueInfo struct {
	QueueURL string
	QueueARN string
}

func init() {
	SchemeBuilder.Register(&SqsQueue{}, &SqsQueueList{})
}
