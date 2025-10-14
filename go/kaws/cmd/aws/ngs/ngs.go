package ngs

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/jedib0t/go-pretty/v6/table"
	awspkg "github.com/pischarti/nix/pkg/aws"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewNgsCmd creates the ngs (node groups) subcommand
func NewNgsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ngs [instance-id...]",
		Short: "Find EKS node groups for EC2 instance IDs",
		Long:  `Query AWS EKS to find which node group each EC2 instance belongs to. Accepts instance IDs as arguments or reads from stdin (one per line).`,
		RunE:  runNgs,
		Example: `  # Find node groups for specific instances
  kaws aws ngs i-1234567890abcdef0 i-0987654321fedcba0
  
  # Pipe instance IDs from kube event command
  kaws kube event --search "failed to get sandbox image" --show-instance-id --output yaml | grep instanceId | awk '{print $2}' | kaws aws ngs
  
  # Find node groups for instances with custom region
  kaws aws ngs i-1234567890abcdef0 --region us-west-2`,
	}

	// Add ngs-specific flags
	cmd.Flags().StringP("region", "r", "", "AWS region (default: from AWS config)")
	cmd.Flags().StringP("cluster", "c", "", "EKS cluster name (if not specified, searches all clusters)")

	return cmd
}

// runNgs executes the node groups query command
func runNgs(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	region, _ := cmd.Flags().GetString("region")
	clusterName, _ := cmd.Flags().GetString("cluster")

	// Get instance IDs from args or stdin
	instanceIDs := args
	if len(instanceIDs) == 0 {
		// TODO: Read from stdin
		return fmt.Errorf("no instance IDs provided. Use: kaws aws ngs <instance-id> [instance-id...]")
	}

	if verbose {
		fmt.Printf("Querying node groups for %d instance(s)\n", len(instanceIDs))
		if region != "" {
			fmt.Printf("Region: %s\n", region)
		}
		if clusterName != "" {
			fmt.Printf("Cluster: %s\n", clusterName)
		}
	}

	// Load AWS config
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, func(opts *config.LoadOptions) error {
		if region != "" {
			opts.Region = region
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create AWS clients
	ec2Client := ec2.NewFromConfig(cfg)

	// Find node groups for instances using pkg/aws
	nodeGroups, err := awspkg.FindNodeGroups(ctx, ec2Client, instanceIDs, clusterName)
	if err != nil {
		return err
	}

	// Display results
	if len(nodeGroups) == 0 {
		fmt.Println("No node group information found for the provided instance IDs")
		return nil
	}

	fmt.Printf("Found node group information for %d instance(s):\n\n", len(nodeGroups))
	displayNodeGroupsTable(nodeGroups)

	return nil
}

// displayNodeGroupsTable displays node group information in a formatted table
func displayNodeGroupsTable(nodeGroups []awspkg.NodeGroupInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)

	t.AppendHeader(table.Row{
		"Instance ID",
		"Cluster",
		"Node Group",
		"Instance Type",
	})

	for _, ng := range nodeGroups {
		t.AppendRow(table.Row{
			ng.InstanceID,
			ng.ClusterName,
			ng.NodeGroupName,
			ng.InstanceType,
		})
	}

	t.Render()
}
