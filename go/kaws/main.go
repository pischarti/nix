package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
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
	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	// Version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("kaws version %s\n", version)
		},
	}

	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
