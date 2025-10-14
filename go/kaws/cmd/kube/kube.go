package kube

import (
	"github.com/pischarti/nix/go/kaws/cmd/kube/event"
	"github.com/spf13/cobra"
)

// NewKubeCmd creates the kube command with all its subcommands
func NewKubeCmd() *cobra.Command {
	kubeCmd := &cobra.Command{
		Use:   "kube",
		Short: "Kubernetes related commands",
		Long:  `Commands for interacting with Kubernetes clusters`,
	}

	// Add subcommands
	kubeCmd.AddCommand(event.NewEventCmd())

	return kubeCmd
}
