package event

import (
	"context"
	"fmt"

	"github.com/pischarti/nix/pkg/k8s"
	"github.com/pischarti/nix/pkg/print"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewEventCmd creates the event subcommand
func NewEventCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event",
		Short: "Query and filter Kubernetes events",
		Long:  `Query Kubernetes events across all namespaces (or a specific namespace) and filter by message content`,
		RunE:  runEvent,
		Example: `  # Filter events containing "failed to get sandbox image"
  kaws kube event --search "failed to get sandbox image"
  
  # Filter events in a specific namespace
  kaws kube event --search "ImagePullBackOff" --namespace default
  
  # Output in YAML format
  kaws kube event --search "error" --output yaml`,
	}

	// Add event-specific flags
	cmd.Flags().StringP("search", "s", "", "search term to filter events (required)")
	cmd.Flags().StringP("output", "o", "table", "output format: table or yaml")
	cmd.MarkFlagRequired("search")

	return cmd
}

// runEvent executes the event query command
func runEvent(cmd *cobra.Command, args []string) error {
	// Get values from viper (which includes flag values, config file, and env vars)
	verbose := viper.GetBool("verbose")
	namespace := viper.GetString("namespace")

	// Get search term from flag
	searchTerm, err := cmd.Flags().GetString("search")
	if err != nil {
		return fmt.Errorf("failed to get search flag: %w", err)
	}

	// Get output format from flag
	outputFormat, err := cmd.Flags().GetString("output")
	if err != nil {
		return fmt.Errorf("failed to get output flag: %w", err)
	}

	// Get Kubernetes client
	client, err := k8s.NewClient()
	if err != nil {
		return err
	}

	if verbose {
		if namespace != "" {
			fmt.Printf("Querying events in namespace: %s\n", namespace)
		} else {
			fmt.Println("Querying events in all namespaces")
		}
		fmt.Printf("Filtering for events containing: %q\n", searchTerm)
	}

	// Query events using the common k8s package
	events, err := client.QueryEvents(context.Background(), k8s.EventQueryOptions{
		Namespace: namespace,
	})
	if err != nil {
		return err
	}

	// Filter events matching the search term
	matchingEvents := k8s.FilterEvents(events, searchTerm)

	// Display results
	if len(matchingEvents) == 0 {
		fmt.Printf("No events found matching %q\n", searchTerm)
		return nil
	}

	// Display based on output format
	switch outputFormat {
	case "yaml":
		return print.EventsYAML(matchingEvents)
	case "table":
		fmt.Printf("Found %d event(s) matching %q:\n\n", len(matchingEvents), searchTerm)
		print.EventsTable(matchingEvents)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s (supported: table, yaml)", outputFormat)
	}
}
