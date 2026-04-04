package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	messagingv1 "github.com/yourusername/kubernetes-sqs-operator/api/v1"
)

const sqsFinalizer = "sqsqueue.messaging.cloud.com/finalizer"

// SqsQueueReconciler reconciles SqsQueue custom resources.
// It embeds the Kubernetes client for CRUD operations on the cluster,
// and implements the full lifecycle: create, drift-detect, and delete.
type SqsQueueReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=messaging.cloud.com,resources=sqsqueues,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=messaging.cloud.com,resources=sqsqueues/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=messaging.cloud.com,resources=sqsqueues/finalizers,verbs=update

func (r *SqsQueueReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("=== RECONCILE LOOP STARTED ===", "namespace", req.Namespace, "name", req.Name)

	// 1. Fetch the SqsQueue resource from the cluster.
	sqsQueue := &messagingv1.SqsQueue{}
	if err := r.Get(ctx, req.NamespacedName, sqsQueue); err != nil {
		if errors.IsNotFound(err) {
			// Resource was deleted before we could act — nothing to do.
			l.Info("SqsQueue resource not found. Already deleted.")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 2. Handle deletion: DeletionTimestamp is set when `kubectl delete` is called.
	//    The finalizer prevents the object from being removed until we clean up AWS.
	if !sqsQueue.DeletionTimestamp.IsZero() {
		l.Info("DeletionTimestamp set — deleting queue in AWS", "queueURL", sqsQueue.Status.QueueURL)

		if err := deleteSqsQueue(ctx, sqsQueue); err != nil {
			l.Error(err, "Failed to delete SQS queue from AWS")
			return ctrl.Result{Requeue: true}, err
		}

		// AWS cleanup done — remove finalizer so Kubernetes can delete the object.
		controllerutil.RemoveFinalizer(sqsQueue, sqsFinalizer)
		if err := r.Update(ctx, sqsQueue); err != nil {
			l.Error(err, "Failed to remove finalizer")
			return ctrl.Result{Requeue: true}, err
		}

		l.Info("Finalizer removed — SqsQueue object will now be deleted")
		return ctrl.Result{}, nil
	}

	// 3. Add finalizer on first reconcile so we can run cleanup on deletion.
	if !controllerutil.ContainsFinalizer(sqsQueue, sqsFinalizer) {
		l.Info("Adding finalizer")
		controllerutil.AddFinalizer(sqsQueue, sqsFinalizer)
		if err := r.Update(ctx, sqsQueue); err != nil {
			l.Error(err, "Failed to add finalizer")
			return ctrl.Result{Requeue: true}, err
		}
		// The Update call will trigger another reconcile — return here.
		return ctrl.Result{}, nil
	}

	// 4. Drift detection: if the queue was already created (QueueURL is set),
	//    verify it still exists in AWS. This handles out-of-band deletions.
	if sqsQueue.Status.QueueURL != "" {
		l.Info("Queue URL known — checking for drift", "queueURL", sqsQueue.Status.QueueURL)

		result, err := checkSqsQueueExists(ctx, sqsQueue)
		if err != nil {
			l.Error(err, "Failed to check queue existence — will retry")
			return ctrl.Result{Requeue: true}, err
		}

		if !result.Exists {
			// Queue was deleted outside of Kubernetes — clear status so it gets recreated.
			l.Info("Drift detected: queue gone from AWS, clearing status for recreation")
			sqsQueue.Status.QueueURL = ""
			sqsQueue.Status.QueueARN = ""
			sqsQueue.Status.State = "NotFound"
			if err := r.Status().Update(ctx, sqsQueue); err != nil {
				return ctrl.Result{Requeue: true}, err
			}
			// Requeue immediately to trigger recreation in the next loop.
			return ctrl.Result{Requeue: true}, nil
		}

		// Queue exists and is healthy — update ARN if it was missing.
		if result.QueueARN != "" && sqsQueue.Status.QueueARN == "" {
			sqsQueue.Status.QueueARN = result.QueueARN
			sqsQueue.Status.State = "Active"
			if err := r.Status().Update(ctx, sqsQueue); err != nil {
				return ctrl.Result{Requeue: true}, err
			}
		}

		l.Info("Queue is healthy, nothing to do", "state", sqsQueue.Status.State)
		return ctrl.Result{}, nil
	}

	// 5. Queue does not exist yet — create it.
	l.Info("No QueueURL in status — creating new SQS queue", "queueName", sqsQueue.Spec.QueueName)

	info, err := createSqsQueue(ctx, sqsQueue)
	if err != nil {
		l.Error(err, "Failed to create SQS queue")
		return ctrl.Result{Requeue: true}, err
	}

	// 6. Persist the result to Status so subsequent reconciles know the queue exists.
	sqsQueue.Status.QueueURL = info.QueueURL
	sqsQueue.Status.QueueARN = info.QueueARN
	sqsQueue.Status.State = "Active"

	if err := r.Status().Update(ctx, sqsQueue); err != nil {
		l.Error(err, "Failed to update status after queue creation")
		return ctrl.Result{Requeue: true}, err
	}

	l.Info("=== QUEUE CREATED AND STATUS UPDATED ===",
		"queueURL", sqsQueue.Status.QueueURL,
		"queueARN", sqsQueue.Status.QueueARN)

	return ctrl.Result{}, nil
}

// SetupWithManager registers the SqsQueueReconciler with the controller manager
// and tells it to watch SqsQueue resources.
func (r *SqsQueueReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&messagingv1.SqsQueue{}).
		Named("sqsqueue").
		Complete(r)
}
