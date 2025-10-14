package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	cfgFile string

	// Root command
	rootCmd = &cobra.Command{
		Use:   "kaws",
		Short: "kaws - A CLI tool for Kubernetes on AWS",
		Long:  `kaws is a command-line tool for managing Kubernetes resources on AWS`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Welcome to kaws! Use --help to see available commands.")
		},
	}

	// Version flag
	version = "0.1.0"
)

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $HOME/.kaws.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringP("kubeconfig", "k", "", "path to kubeconfig file (default: $HOME/.kube/config)")
	rootCmd.PersistentFlags().StringP("namespace", "n", "", "namespace to query (default: all namespaces)")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig"))
	viper.BindPFlag("namespace", rootCmd.PersistentFlags().Lookup("namespace"))

	// Version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("kaws version %s\n", version)
		},
	}

	// Kube command (parent for Kubernetes-related subcommands)
	kubeCmd := &cobra.Command{
		Use:   "kube",
		Short: "Kubernetes related commands",
		Long:  `Commands for interacting with Kubernetes clusters`,
	}

	// Event subcommand under kube
	eventCmd := &cobra.Command{
		Use:   "event",
		Short: "Query for Kubernetes events matching 'failed to get sandbox image'",
		Long:  `Query Kubernetes events across all namespaces (or a specific namespace) for events containing "failed to get sandbox image"`,
		RunE:  runKubeEvent,
	}

	// Add event as subcommand of kube
	kubeCmd.AddCommand(eventCmd)

	// Add commands to root
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(kubeCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}

		// Search config in home directory with name ".kaws" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".kaws")
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("KAWS")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}

func getKubeClient(cmd *cobra.Command) (*kubernetes.Clientset, error) {
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

func runKubeEvent(cmd *cobra.Command, args []string) error {
	// Get values from viper (which includes flag values, config file, and env vars)
	verbose := viper.GetBool("verbose")
	namespace := viper.GetString("namespace")

	// Get Kubernetes client
	clientset, err := getKubeClient(cmd)
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

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(findSubstring(s, substr) != -1))
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
