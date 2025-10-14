package aws

import (
	"github.com/pischarti/nix/go/kaws/cmd/aws/ngs"
	"github.com/spf13/cobra"
)

// NewAWSCmd creates the aws command with all its subcommands
func NewAWSCmd() *cobra.Command {
	awsCmd := &cobra.Command{
		Use:   "aws",
		Short: "AWS related commands",
		Long:  `Commands for interacting with AWS services`,
	}

	// Add subcommands
	awsCmd.AddCommand(ngs.NewNgsCmd())

	return awsCmd
}
