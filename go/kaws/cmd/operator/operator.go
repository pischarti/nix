package operator

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	kawsv1alpha1 "github.com/pischarti/nix/go/kaws/api/v1alpha1"
	"github.com/pischarti/nix/go/kaws/controllers"
	"github.com/pischarti/nix/pkg/k8s"
	pkgoperator "github.com/pischarti/nix/pkg/operator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

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
	opConfig := &pkgoperator.OperatorConfig{
		WatchInterval:    watchInterval,
		SearchTerms:      searchTerms,
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
	if err := pkgoperator.CheckAndRecycle(ctx, k8sClient, ec2Client, asgClient, opConfig, verbose); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Error during check: %v\n", err)
	}

	for {
		select {
		case <-sigChan:
			fmt.Println("\n🛑 Shutting down operator...")
			return nil
		case <-ticker.C:
			if err := pkgoperator.CheckAndRecycle(ctx, k8sClient, ec2Client, asgClient, opConfig, verbose); err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  Error during check: %v\n", err)
			}
		}
	}
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
