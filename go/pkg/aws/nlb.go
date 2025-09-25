package aws

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	printpkg "github.com/pischarti/nix/go/pkg/print"
	"github.com/pischarti/nix/go/pkg/vpc"
	"gofr.dev/pkg/gofr"
)

// ListNLBs handles the nlb command for listing AWS Network Load Balancers
func ListNLBs(ctx *gofr.Context) (any, error) {
	args := os.Args[1:] // Get command line args for parsing flags

	// Check for help flag first
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println("Usage: aws nlb --vpc VPC_ID [--zone AZ] [--sort SORT_BY]")
			fmt.Println("Options:")
			fmt.Println("  --vpc VPC_ID    VPC ID to list NLBs for (required)")
			fmt.Println("  --zone AZ       Filter by availability zone (optional)")
			fmt.Println("  --sort SORT_BY  Sort by: name (default), state, type, scheme, created")
			return nil, nil
		}
	}

	// Parse arguments
	opts, err := vpc.ParseNLBArgs(args)
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

	// Create ELBv2 client
	elbv2Client := elasticloadbalancingv2.NewFromConfig(cfg)

	// Describe load balancers
	input := &elasticloadbalancingv2.DescribeLoadBalancersInput{}

	result, err := elbv2Client.DescribeLoadBalancers(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe load balancers: %w", err)
	}

	// Filter NLBs by VPC and optionally by zone
	var nlbs []types.LoadBalancer
	for _, lb := range result.LoadBalancers {
		// Only include Network Load Balancers
		if lb.Type != types.LoadBalancerTypeEnumNetwork {
			continue
		}

		// Filter by VPC
		if aws.ToString(lb.VpcId) != opts.VPCID {
			continue
		}

		// Filter by zone if specified
		if opts.Zone != "" {
			hasZone := false
			for _, az := range lb.AvailabilityZones {
				if aws.ToString(az.ZoneName) == opts.Zone {
					hasZone = true
					break
				}
			}
			if !hasZone {
				continue
			}
		}

		nlbs = append(nlbs, lb)
	}

	// Convert to NLBInfo structs
	nlbInfos := convertELBv2ToNLBInfo(nlbs)

	// Sort NLBs
	vpc.SortNLBs(nlbInfos, opts.SortBy)

	// Print table output
	printpkg.PrintNLBTable(nlbInfos)

	return nil, nil
}

// convertELBv2ToNLBInfo converts AWS ELBv2 load balancer types to NLBInfo structs
func convertELBv2ToNLBInfo(lbs []types.LoadBalancer) []vpc.NLBInfo {
	var nlbInfos []vpc.NLBInfo

	for _, lb := range lbs {
		// Extract name from tags
		name := ""
		var relevantTags []string

		// Get tags for this load balancer
		// Note: In a real implementation, you might want to batch tag requests
		// for better performance when dealing with many load balancers
		tags := getLoadBalancerTags(lb.LoadBalancerArn)

		for _, tag := range tags {
			key := aws.ToString(tag.Key)
			value := aws.ToString(tag.Value)

			switch key {
			case "Name":
				name = value
			default:
				// Include relevant tags
				if strings.HasPrefix(key, "kubernetes.io/") ||
					strings.HasPrefix(key, "aws:") ||
					key == "Environment" ||
					key == "Project" ||
					key == "Service" {
					relevantTags = append(relevantTags, key)
				}
			}
		}

		// Format availability zones
		var azs []string
		var subnets []string
		for _, az := range lb.AvailabilityZones {
			azs = append(azs, aws.ToString(az.ZoneName))
			subnets = append(subnets, aws.ToString(az.SubnetId))
		}

		// Format tags with each tag on a separate line
		tagsStr := strings.Join(relevantTags, "\n")

		// Format created time
		createdTime := ""
		if lb.CreatedTime != nil {
			createdTime = lb.CreatedTime.Format(time.RFC3339)
		}

		nlbInfo := vpc.NLBInfo{
			LoadBalancerArn:   aws.ToString(lb.LoadBalancerArn),
			Name:              name,
			DNSName:           aws.ToString(lb.DNSName),
			State:             string(lb.State.Code),
			Type:              string(lb.Type),
			Scheme:            string(lb.Scheme),
			VPCID:             aws.ToString(lb.VpcId),
			AvailabilityZones: strings.Join(azs, ", "),
			Subnets:           strings.Join(subnets, ", "),
			CreatedTime:       createdTime,
			Tags:              tagsStr,
		}
		nlbInfos = append(nlbInfos, nlbInfo)
	}

	return nlbInfos
}

// getLoadBalancerTags retrieves tags for a load balancer
// This is a simplified implementation - in production you might want to batch these requests
func getLoadBalancerTags(arn *string) []types.Tag {
	if arn == nil {
		return []types.Tag{}
	}

	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return []types.Tag{}
	}

	// Create ELBv2 client
	elbv2Client := elasticloadbalancingv2.NewFromConfig(cfg)

	input := &elasticloadbalancingv2.DescribeTagsInput{
		ResourceArns: []string{aws.ToString(arn)},
	}

	result, err := elbv2Client.DescribeTags(context.TODO(), input)
	if err != nil {
		// Return empty tags on error to avoid breaking the listing
		return []types.Tag{}
	}

	if len(result.TagDescriptions) > 0 {
		return result.TagDescriptions[0].Tags
	}

	return []types.Tag{}
}

// NLBRouter routes nlb sub-commands
func NLBRouter(ctx *gofr.Context) (any, error) {
	args := os.Args[1:] // Get command line args for parsing flags

	// Check for help flag for main nlb command
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println("Usage: aws nlb [COMMAND]")
			fmt.Println("Commands:")
			fmt.Println("  list               List all Network Load Balancers in a VPC (default)")
			fmt.Println()
			fmt.Println("Examples:")
			fmt.Println("  aws nlb --vpc vpc-12345678")
			fmt.Println("  aws nlb list --vpc vpc-12345678")
			fmt.Println("  aws nlb list --vpc vpc-12345678 --zone us-east-1a")
			fmt.Println("  aws nlb list --vpc vpc-12345678 --sort state")
			return nil, nil
		}
	}

	// Default to list command
	return ListNLBs(ctx)
}
