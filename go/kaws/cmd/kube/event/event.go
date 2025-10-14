package event

import (
	"context"
	"fmt"

	"github.com/pischarti/nix/pkg/k8s"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
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
  
  # Filter events with case-insensitive search
  kaws kube event --search "error"`,
	}

	// Add event-specific flags
	cmd.Flags().StringP("search", "s", "", "search term to filter events (required)")
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
	matchingEvents := FilterEvents(events, searchTerm)

	// Display results
	if len(matchingEvents) == 0 {
		fmt.Printf("No events found matching %q\n", searchTerm)
		return nil
	}

	fmt.Printf("Found %d event(s) matching %q:\n\n", len(matchingEvents), searchTerm)

	for _, event := range matchingEvents {
		DisplayEvent(event)
	}

	return nil
}

// FilterEvents filters events by search term in the message field
func FilterEvents(events []corev1.Event, searchTerm string) []corev1.Event {
	matchingEvents := []corev1.Event{}

	for _, event := range events {
		if Contains(event.Message, searchTerm) {
			matchingEvents = append(matchingEvents, event)
		}
	}

	return matchingEvents
}

// DisplayEvent prints event details in a formatted way
func DisplayEvent(event corev1.Event) {
	fmt.Printf("Namespace: %s\n", event.Namespace)
	fmt.Printf("Name: %s\n", event.Name)
	fmt.Printf("Type: %s\n", event.Type)
	fmt.Printf("Reason: %s\n", event.Reason)
	fmt.Printf("Object: %s/%s\n", event.InvolvedObject.Kind, event.InvolvedObject.Name)
	fmt.Printf("Count: %d\n", event.Count)
	fmt.Printf("First Seen: %s\n", event.FirstTimestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Last Seen: %s\n", event.LastTimestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Message: %s\n", event.Message)
	fmt.Println("---")
}

// Contains checks if a string contains a substring
func Contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(findSubstring(s, substr) != -1))
}

// findSubstring finds the index of a substring in a string
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
