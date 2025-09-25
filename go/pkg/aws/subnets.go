package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	printpkg "github.com/pischarti/nix/go/pkg/print"
	"github.com/pischarti/nix/go/pkg/vpc"
	"gofr.dev/pkg/gofr"
)

// ListSubnets handles the subnets command for listing AWS subnets
func ListSubnets(ctx *gofr.Context) (any, error) {
	args := os.Args[1:] // Get command line args for parsing flags

	// Check for help flag first
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println("Usage: aws subnets --vpc VPC_ID [--zone AZ] [--sort SORT_BY]")
			fmt.Println("Options:")
			fmt.Println("  --vpc VPC_ID    VPC ID to list subnets for (required)")
			fmt.Println("  --zone AZ       Filter by availability zone (optional)")
			fmt.Println("  --sort SORT_BY  Sort by: cidr (default), az, name, type")
			return nil, nil
		}
	}

	// Parse arguments
	opts, err := vpc.ParseSubnetsArgs(args)
	if err != nil {
		return nil, err
	}

	if opts.VPCID == "" {
		return nil, fmt.Errorf("vpc parameter is required")
	}

	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// Describe subnets
	input := &ec2.DescribeSubnetsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{opts.VPCID},
			},
		},
	}

	if opts.Zone != "" {
		input.Filters = append(input.Filters, types.Filter{
			Name:   aws.String("availability-zone"),
			Values: []string{opts.Zone},
		})
	}

	result, err := ec2Client.DescribeSubnets(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnets: %w", err)
	}

	// Convert to SubnetInfo structs
	subnets := vpc.ConvertEC2SubnetsToSubnetInfo(result.Subnets)

	// Sort subnets
	vpc.SortSubnets(subnets, opts.SortBy)

	// Print table output
	printpkg.PrintSubnetsTable(subnets)

	return nil, nil
}

// DeleteSubnet handles the delete subnet command
func DeleteSubnet(ctx *gofr.Context) (any, error) {
	args := os.Args[1:] // Get command line args for parsing flags

	// Check for help flag first
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println("Usage: aws subnets delete --subnet-id SUBNET_ID [--force]")
			fmt.Println("Options:")
			fmt.Println("  --subnet-id SUBNET_ID  Subnet ID to delete (required)")
			fmt.Println("  --force               Skip confirmation prompt")
			return nil, nil
		}
	}

	// Parse arguments
	subnetID, force, err := parseDeleteSubnetArgs(args)
	if err != nil {
		return nil, err
	}

	if subnetID == "" {
		return nil, fmt.Errorf("subnet-id parameter is required")
	}

	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// Check if subnet exists first
	describeInput := &ec2.DescribeSubnetsInput{
		SubnetIds: []string{subnetID},
	}

	_, err = ec2Client.DescribeSubnets(context.TODO(), describeInput)
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnet %s: %w", subnetID, err)
	}

	// Confirm deletion unless --force is used
	if !force {
		fmt.Printf("Are you sure you want to delete subnet %s? (yes/no): ", subnetID)
		var response string
		fmt.Scanln(&response)
		if response != "yes" {
			fmt.Println("Deletion cancelled.")
			return nil, nil
		}
	}

	// Delete the subnet
	deleteInput := &ec2.DeleteSubnetInput{
		SubnetId: aws.String(subnetID),
	}

	_, err = ec2Client.DeleteSubnet(context.TODO(), deleteInput)
	if err != nil {
		return nil, fmt.Errorf("failed to delete subnet %s: %w", subnetID, err)
	}

	fmt.Printf("Successfully deleted subnet %s\n", subnetID)
	return nil, nil
}

// parseDeleteSubnetArgs parses command line arguments for the delete subnet command
func parseDeleteSubnetArgs(args []string) (subnetID string, force bool, err error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "subnets", "delete":
			// Skip command names
			continue
		case "--subnet-id":
			if i+1 < len(args) {
				i++
				subnetID = args[i]
			}
		case "--force":
			force = true
		}
	}
	return subnetID, force, nil
}

// SubnetsRouter routes subnets sub-commands
func SubnetsRouter(ctx *gofr.Context) (any, error) {
	args := os.Args[1:] // Get command line args for parsing flags

	// Check if there's a sub-command
	if len(args) >= 2 && args[1] == "delete" {
		// Route to delete command (it will handle its own help)
		return DeleteSubnet(ctx)
	}

	// Check for help flag for main subnets command
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println("Usage: aws subnets [COMMAND]")
			fmt.Println("Commands:")
			fmt.Println("  list    List all subnets in a VPC (default)")
			fmt.Println("  delete  Delete a subnet by ID")
			fmt.Println()
			fmt.Println("Examples:")
			fmt.Println("  aws subnets --vpc vpc-12345678")
			fmt.Println("  aws subnets list --vpc vpc-12345678")
			fmt.Println("  aws subnets delete --subnet-id subnet-12345678")
			return nil, nil
		}
	}

	// Default to list command
	return ListSubnets(ctx)
}
