package event

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// NewEventCmd creates the event subcommand
func NewEventCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "event",
		Short: "Query for Kubernetes events matching 'failed to get sandbox image'",
		Long:  `Query Kubernetes events across all namespaces (or a specific namespace) for events containing "failed to get sandbox image"`,
		RunE:  runEvent,
	}
}

// getKubeClient creates a Kubernetes clientset from the configured kubeconfig
func getKubeClient() (*kubernetes.Clientset, error) {
	// Try to get kubeconfig from flag first, then viper (config file), then default
	kubeconfig := viper.GetString("kubeconfig")
	if kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	// Build config from kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return clientset, nil
}

// runEvent executes the event query command
func runEvent(cmd *cobra.Command, args []string) error {
	// Get values from viper (which includes flag values, config file, and env vars)
	verbose := viper.GetBool("verbose")
	namespace := viper.GetString("namespace")

	// Get Kubernetes client
	clientset, err := getKubeClient()
	if err != nil {
		return err
	}

	if verbose {
		if namespace != "" {
			fmt.Printf("Querying events in namespace: %s\n", namespace)
		} else {
			fmt.Println("Querying events in all namespaces")
		}
	}

	// Query events
	events, err := clientset.CoreV1().Events(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list events: %w", err)
	}

	// Filter events matching "failed to get sandbox image"
	matchingEvents := []corev1.Event{}
	searchTerm := "failed to get sandbox image"

	for _, event := range events.Items {
		if contains(event.Message, searchTerm) {
			matchingEvents = append(matchingEvents, event)
		}
	}

	// Display results
	if len(matchingEvents) == 0 {
		fmt.Println("No events found matching 'failed to get sandbox image'")
		return nil
	}

	fmt.Printf("Found %d event(s) matching 'failed to get sandbox image':\n\n", len(matchingEvents))

	for _, event := range matchingEvents {
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

	return nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
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
