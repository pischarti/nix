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
