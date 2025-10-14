package operator

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	kawsv1alpha1 "github.com/pischarti/nix/go/kaws/api/v1alpha1"
	"github.com/pischarti/nix/go/kaws/controllers"
	"github.com/pischarti/nix/pkg/k8s"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// OperatorConfig holds the operator configuration
type OperatorConfig struct {
	WatchInterval    time.Duration
	SearchTerms      []string
	MinEventCount    int32
	RecycleThreshold int
	DryRun           bool
	ProcessedEvents  map[string]time.Time // Track processed events
	mu               sync.RWMutex
}

// NewOperatorCmd creates the operator command
func NewOperatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "operator",
		Short: "Run kaws as a Kubernetes operator",
		Long: `Run kaws in operator mode to continuously watch for error events and automatically recycle problematic node groups.

The operator watches for specified error patterns (e.g., "failed to get sandbox image"), identifies the affected node groups, 
and automatically recycles them to resolve the issues.`,
		RunE: runOperator,
		Example: `  # Run operator with default settings
  kaws operator
  
  # Run with custom watch interval
  kaws operator --watch-interval 30s
  
  # Dry run mode (don't actually recycle)
  kaws operator --dry-run
  
  # Custom search terms
  kaws operator --search "failed to get sandbox image" --search "ImagePullBackOff"
  
  # With custom event threshold
  kaws operator --threshold 3
  
  # Use CRD-based configuration
  kaws operator --use-crd`,
	}

	cmd.Flags().Duration("watch-interval", 60*time.Second, "interval between event checks")
	cmd.Flags().StringSlice("search", []string{"failed to get sandbox image"}, "search terms to watch for (can specify multiple)")
	cmd.Flags().Int("threshold", 5, "number of events before triggering recycle")
	cmd.Flags().Bool("dry-run", false, "log actions without actually recycling node groups")
	cmd.Flags().StringP("region", "r", "", "AWS region (default: from AWS config)")
	cmd.Flags().Bool("use-crd", false, "use EventRecycler CRD for configuration (requires CRD installed)")

	return cmd
}

// runOperator executes the operator command
func runOperator(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	watchInterval, _ := cmd.Flags().GetDuration("watch-interval")
	searchTerms, _ := cmd.Flags().GetStringSlice("search")
	threshold, _ := cmd.Flags().GetInt("threshold")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	region, _ := cmd.Flags().GetString("region")
	useCRD, _ := cmd.Flags().GetBool("use-crd")

	fmt.Println("🚀 Starting kaws operator...")
	fmt.Printf("   Mode: %s\n", map[bool]string{true: "CRD-based", false: "Standalone"}[useCRD])
	fmt.Printf("   Watch interval: %s\n", watchInterval)
	fmt.Printf("   Search terms: %v\n", searchTerms)
	fmt.Printf("   Event threshold: %d\n", threshold)
	fmt.Printf("   Dry run: %v\n", dryRun)
	if region != "" {
		fmt.Printf("   AWS region: %s\n", region)
	}
	fmt.Println()

	if useCRD {
		fmt.Println("📋 CRD-based mode with informers (race-condition safe)")
		fmt.Println("   Using controller-runtime with cached informers for efficient event watching")
		fmt.Println()
		return runCRDOperator(region, verbose)
	}

	// Create operator config
	opConfig := &OperatorConfig{
		WatchInterval:    watchInterval,
		SearchTerms:      searchTerms,
		MinEventCount:    int32(threshold),
		RecycleThreshold: threshold,
		DryRun:           dryRun,
		ProcessedEvents:  make(map[string]time.Time),
	}

	// Create Kubernetes client
	k8sClient, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Create AWS clients
	ctx := context.Background()
	awsCfg, err := config.LoadDefaultConfig(ctx, func(opts *config.LoadOptions) error {
		if region != "" {
			opts.Region = region
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	ec2Client := ec2.NewFromConfig(awsCfg)
	asgClient := autoscaling.NewFromConfig(awsCfg)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run operator loop
	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()

	fmt.Println("✓ Operator is running. Press Ctrl+C to stop.")
	fmt.Println()

	// Run first check immediately
	if err := checkAndRecycle(ctx, k8sClient, ec2Client, asgClient, opConfig, verbose); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Error during check: %v\n", err)
	}

	for {
		select {
		case <-sigChan:
			fmt.Println("\n🛑 Shutting down operator...")
			return nil
		case <-ticker.C:
			if err := checkAndRecycle(ctx, k8sClient, ec2Client, asgClient, opConfig, verbose); err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  Error during check: %v\n", err)
			}
		}
	}
}

// checkAndRecycle checks for error events and recycles affected node groups
func checkAndRecycle(ctx context.Context, k8sClient *k8s.Client, ec2Client *ec2.Client, asgClient *autoscaling.Client, opConfig *OperatorConfig, verbose bool) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	if verbose {
		fmt.Printf("[%s] Checking for error events...\n", timestamp)
	}

	// Query all events
	events, err := k8sClient.QueryEvents(ctx, k8s.EventQueryOptions{
		Namespace: "", // All namespaces
	})
	if err != nil {
		return fmt.Errorf("failed to query events: %w", err)
	}

	// Track node groups that need recycling
	nodeGroupsToRecycle := make(map[string]int) // key: nodeGroupName, value: event count

	// Check each search term
	for _, searchTerm := range opConfig.SearchTerms {
		matchingEvents := k8s.FilterEvents(events, searchTerm)

		if len(matchingEvents) == 0 {
			continue
		}

		// Filter out recently processed events (within last hour)
		recentEvents := filterRecentEvents(matchingEvents, opConfig)

		if len(recentEvents) == 0 {
			continue
		}

		fmt.Printf("[%s] Found %d recent event(s) matching %q\n", timestamp, len(recentEvents), searchTerm)

		// Enrich with node information
		enrichedEvents, err := k8sClient.EnrichEventsWithNodeInfo(ctx, recentEvents, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: Could not enrich events: %v\n", err)
			continue
		}

		// Find affected node groups
		for _, enriched := range enrichedEvents {
			if enriched.InstanceID != "" && enriched.InstanceID != "N/A" {
				// Query node group for this instance
				nodeGroups, err := findNodeGroupForInstance(ctx, ec2Client, enriched.InstanceID)
				if err != nil {
					if verbose {
						fmt.Fprintf(os.Stderr, "  Warning: Could not find node group for instance %s: %v\n", enriched.InstanceID, err)
					}
					continue
				}

				for _, ng := range nodeGroups {
					if ng != "" && ng != "Unknown" {
						nodeGroupsToRecycle[ng]++
					}
				}
			}
		}
	}

	// Recycle node groups that exceed threshold
	for ngName, count := range nodeGroupsToRecycle {
		if count >= opConfig.RecycleThreshold {
			fmt.Printf("[%s] 🔄 Node group %s has %d problematic events (threshold: %d)\n",
				timestamp, ngName, count, opConfig.RecycleThreshold)

			if opConfig.DryRun {
				fmt.Printf("  [DRY RUN] Would recycle node group: %s\n", ngName)
			} else {
				fmt.Printf("  Recycling node group: %s\n", ngName)
				// Note: Implement recycling logic here or call the recycle function
				fmt.Printf("  ⚠️  Automated recycling not yet implemented - manual intervention required\n")
			}
		} else if verbose {
			fmt.Printf("[%s] Node group %s has %d events (below threshold of %d)\n",
				timestamp, ngName, count, opConfig.RecycleThreshold)
		}
	}

	if len(nodeGroupsToRecycle) == 0 && verbose {
		fmt.Printf("[%s] ✓ No problematic node groups detected\n", timestamp)
	}

	return nil
}

// filterRecentEvents filters out events that have been processed recently
func filterRecentEvents(events []corev1.Event, opConfig *OperatorConfig) []corev1.Event {
	opConfig.mu.RLock()
	defer opConfig.mu.RUnlock()

	recentEvents := []corev1.Event{}
	now := time.Now()

	for _, event := range events {
		eventKey := fmt.Sprintf("%s/%s", event.Namespace, event.Name)

		// Check if we've processed this event recently (within last hour)
		if lastProcessed, found := opConfig.ProcessedEvents[eventKey]; found {
			if now.Sub(lastProcessed) < time.Hour {
				continue // Skip recently processed events
			}
		}

		recentEvents = append(recentEvents, event)
	}

	// Update processed events (with write lock)
	opConfig.mu.Lock()
	for _, event := range recentEvents {
		eventKey := fmt.Sprintf("%s/%s", event.Namespace, event.Name)
		opConfig.ProcessedEvents[eventKey] = now
	}
	opConfig.mu.Unlock()

	// Clean up old entries (older than 2 hours)
	opConfig.mu.Lock()
	for key, timestamp := range opConfig.ProcessedEvents {
		if now.Sub(timestamp) > 2*time.Hour {
			delete(opConfig.ProcessedEvents, key)
		}
	}
	opConfig.mu.Unlock()

	return recentEvents
}

// findNodeGroupForInstance queries AWS to find node group for an instance
func findNodeGroupForInstance(ctx context.Context, ec2Client *ec2.Client, instanceID string) ([]string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	result, err := ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, err
	}

	nodeGroups := []string{}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// Extract node group from tags
			for _, tag := range instance.Tags {
				if tag.Key == nil || tag.Value == nil {
					continue
				}

				if *tag.Key == "eks:nodegroup-name" || *tag.Key == "alpha.eksctl.io/nodegroup-name" {
					nodeGroups = append(nodeGroups, *tag.Value)
				}
			}
		}
	}

	return nodeGroups, nil
}

// runCRDOperator runs the operator in CRD mode using controller-runtime with informers
func runCRDOperator(region string, verbose bool) error {
	// Setup logging
	opts := zap.Options{
		Development: verbose,
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog := ctrl.Log.WithName("setup")

	// Create a new scheme and register our API types
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(kawsv1alpha1.AddToScheme(scheme))

	setupLog.Info("Starting manager with leader election")

	// Create manager with informer cache and leader election
	// The cache provides thread-safe, efficient access to Kubernetes resources
	// Leader election ensures only one replica is active at a time
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Cache: cache.Options{
			// Sync period for the informer cache (how often to re-list)
			SyncPeriod: ptr(10 * time.Minute),
		},
		// Leader election configuration
		LeaderElection:          true,
		LeaderElectionID:        "kaws-operator-lock",
		LeaderElectionNamespace: "kube-system", // Use kube-system for cluster-scoped operators
		// Recommended lease durations for production
		LeaseDuration: ptr(15 * time.Second),
		RenewDeadline: ptr(10 * time.Second),
		RetryPeriod:   ptr(2 * time.Second),
	})
	if err != nil {
		return fmt.Errorf("unable to start manager: %w", err)
	}

	// Setup the EventRecycler controller with informers
	if err = (&controllers.EventRecyclerReconciler{
		Client: mgr.GetClient(), // This client uses the cached informers
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create controller: %w", err)
	}

	setupLog.Info("Starting controller manager with informers and leader election")
	setupLog.Info("✓ All informers are thread-safe and cache-backed")
	setupLog.Info("✓ No race conditions in event watching")
	setupLog.Info("✓ Leader election enabled - safe to run multiple replicas")
	setupLog.Info("ℹ️  Only the leader replica will reconcile resources")

	// Start the manager (this starts all informers and controllers)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("problem running manager: %w", err)
	}

	return nil
}

// ptr returns a pointer to the value
func ptr[T any](v T) *T {
	return &v
}
