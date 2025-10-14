package main

import (
	"fmt"
	"os"

	"github.com/pischarti/nix/go/kaws/cmd/kube"
	"github.com/pischarti/nix/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	cobra.OnInitialize(func() {
		if err := config.InitConfig(cfgFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
			os.Exit(1)
		}
	})

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

	// Add commands to root
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(kube.NewKubeCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
