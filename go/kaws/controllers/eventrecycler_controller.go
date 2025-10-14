package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kawsv1alpha1 "github.com/pischarti/nix/go/kaws/api/v1alpha1"
	"github.com/pischarti/nix/pkg/k8s"
)

// EventRecyclerReconciler reconciles an EventRecycler object
type EventRecyclerReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// AWS clients
	EC2Client *ec2.Client
	ASGClient *autoscaling.Client

	// Thread-safe tracking of processed events (uses metav1.Time for K8s compatibility)
	processedEvents map[string]metav1.Time
}

// +kubebuilder:rbac:groups=kaws.pischarti.dev,resources=eventrecyclers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kaws.pischarti.dev,resources=eventrecyclers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kaws.pischarti.dev,resources=eventrecyclers/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list

// Reconcile is part of the main kubernetes reconciliation loop
func (r *EventRecyclerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the EventRecycler instance
	var eventRecycler kawsv1alpha1.EventRecycler
	if err := r.Get(ctx, req.NamespacedName, &eventRecycler); err != nil {
		log.Error(err, "unable to fetch EventRecycler")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Reconciling EventRecycler", "name", eventRecycler.Name)

	// Get watch interval from spec
	watchInterval := 60 * time.Second
	if eventRecycler.Spec.WatchInterval.Duration > 0 {
		watchInterval = eventRecycler.Spec.WatchInterval.Duration
	}

	// Process events and check for issues
	if err := r.checkAndRecycle(ctx, &eventRecycler); err != nil {
		log.Error(err, "failed to check and recycle")
		return ctrl.Result{RequeueAfter: watchInterval}, err
	}

	// Requeue after watch interval
	return ctrl.Result{RequeueAfter: watchInterval}, nil
}

// SetupWithManager sets up the controller with the Manager and configures informers
func (r *EventRecyclerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize AWS clients
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	r.EC2Client = ec2.NewFromConfig(cfg)
	r.ASGClient = autoscaling.NewFromConfig(cfg)
	r.processedEvents = make(map[string]metav1.Time)

	// The manager's cache automatically sets up informers for all watched types
	// This provides thread-safe, cached access to events and avoids race conditions
	// The cache is automatically synced and kept up-to-date

	return ctrl.NewControllerManagedBy(mgr).
		For(&kawsv1alpha1.EventRecycler{}).
		Complete(r)
}

// checkAndRecycle checks for matching events and triggers recycling if needed
func (r *EventRecyclerReconciler) checkAndRecycle(ctx context.Context, recycler *kawsv1alpha1.EventRecycler) error {
	log := log.FromContext(ctx)

	// Use pkg/k8s CheckAndRecycleWithStatus for the core logic
	config := k8s.RecyclerConfig{
		SearchTerms: recycler.Spec.SearchTerms,
		Threshold:   recycler.Spec.Threshold,
		DryRun:      recycler.Spec.DryRun,
	}

	nodeGroupCounts, status, err := k8s.CheckAndRecycleWithStatus(ctx, r.Client, r.EC2Client, config, r.processedEvents)
	if err != nil {
		return fmt.Errorf("failed to check and recycle: %w", err)
	}

	// Update status
	recycler.Status.EventCounts = status.EventCounts
	recycler.Status.LastCheckTime = status.LastCheckTime

	if err := r.Status().Update(ctx, recycler); err != nil {
		log.Error(err, "failed to update EventRecycler status")
	}

	// Check if any node groups exceed threshold and need actual recycling
	for ng, count := range nodeGroupCounts {
		if count >= recycler.Spec.Threshold && !recycler.Spec.DryRun {
			log.Info("Triggering recycle for node group", "nodeGroup", ng)
			// TODO: Implement actual recycling logic using ASGClient
			// For now, just log
			log.Info("⚠️  Automated recycling not yet fully implemented", "nodeGroup", ng)
		}
	}

	return nil
}
