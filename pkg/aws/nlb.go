package aws

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
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
	var nlbs []elbv2types.LoadBalancer
	for _, lb := range result.LoadBalancers {
		// Only include Network Load Balancers
		if lb.Type != elbv2types.LoadBalancerTypeEnumNetwork {
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
func convertELBv2ToNLBInfo(lbs []elbv2types.LoadBalancer) []vpc.NLBInfo {
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
func getLoadBalancerTags(arn *string) []elbv2types.Tag {
	if arn == nil {
		return []elbv2types.Tag{}
	}

	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return []elbv2types.Tag{}
	}

	// Create ELBv2 client
	elbv2Client := elasticloadbalancingv2.NewFromConfig(cfg)

	input := &elasticloadbalancingv2.DescribeTagsInput{
		ResourceArns: []string{aws.ToString(arn)},
	}

	result, err := elbv2Client.DescribeTags(context.TODO(), input)
	if err != nil {
		// Return empty tags on error to avoid breaking the listing
		return []elbv2types.Tag{}
	}

	if len(result.TagDescriptions) > 0 {
		return result.TagDescriptions[0].Tags
	}

	return []elbv2types.Tag{}
}

// RemoveSubnetFromNLB handles the remove-subnet command for removing a subnet from an NLB
func RemoveSubnetFromNLB(ctx *gofr.Context) (any, error) {
	args := os.Args[1:] // Get command line args for parsing flags

	// Check for help flag first
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println("Usage: aws nlb remove-subnet --vpc VPC_ID --zone AZ [--nlb-name NLB_NAME] [--force]")
			fmt.Println("Options:")
			fmt.Println("  --vpc VPC_ID       VPC ID containing the NLB (required)")
			fmt.Println("  --zone AZ          Availability zone of the subnet to remove (required)")
			fmt.Println("  --nlb-name NAME    Specific NLB name to target (optional, removes from all NLBs if not specified)")
			fmt.Println("  --force           Skip confirmation prompt")
			fmt.Println()
			fmt.Println("This command removes a subnet from Network Load Balancers in the specified VPC and zone.")
			fmt.Println("If no NLB name is specified, it will remove the subnet from all NLBs in the VPC that have subnets in the specified zone.")
			return nil, nil
		}
	}

	// Parse arguments
	opts, err := parseRemoveSubnetArgs(args)
	if err != nil {
		return nil, err
	}

	if opts.VPCID == "" {
		return nil, fmt.Errorf("vpc parameter is required")
	}
	if opts.Zone == "" {
		return nil, fmt.Errorf("zone parameter is required")
	}

	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create ELBv2 client
	elbv2Client := elasticloadbalancingv2.NewFromConfig(cfg)

	// Find NLBs in the VPC
	nlbs, err := findNLBsInVPC(elbv2Client, opts.VPCID, opts.NLBName)
	if err != nil {
		return nil, fmt.Errorf("failed to find NLBs: %w", err)
	}

	if len(nlbs) == 0 {
		return nil, fmt.Errorf("no NLBs found in VPC %s", opts.VPCID)
	}

	// Filter NLBs that have subnets in the specified zone
	var targetNLBs []elbv2types.LoadBalancer
	for _, nlb := range nlbs {
		hasZone := false
		for _, az := range nlb.AvailabilityZones {
			if aws.ToString(az.ZoneName) == opts.Zone {
				hasZone = true
				break
			}
		}
		if hasZone {
			targetNLBs = append(targetNLBs, nlb)
		}
	}

	if len(targetNLBs) == 0 {
		return nil, fmt.Errorf("no NLBs found in VPC %s with subnets in zone %s", opts.VPCID, opts.Zone)
	}

	// Show what will be modified
	fmt.Printf("Found %d NLB(s) in VPC %s with subnets in zone %s:\n", len(targetNLBs), opts.VPCID, opts.Zone)
	for _, nlb := range targetNLBs {
		nlbName := getNLBName(nlb)
		fmt.Printf("  - %s (%s)\n", nlbName, aws.ToString(nlb.LoadBalancerArn))
	}

	// Check for potential issues before proceeding
	fmt.Printf("\n⚠️  Note: If NLBs are associated with Kubernetes services or ECS services, subnet removal may fail.\n")
	fmt.Printf("   Use 'kubectl get services -o wide' to check for Kubernetes service associations.\n")

	// Confirm removal unless --force is used
	if !opts.Force {
		fmt.Printf("\nAre you sure you want to remove subnets in zone %s from these NLBs? (yes/no): ", opts.Zone)
		var response string
		fmt.Scanln(&response)
		if response != "yes" {
			fmt.Println("Operation cancelled.")
			return nil, nil
		}
	}

	// Remove subnets from each NLB
	successCount := 0
	for _, nlb := range targetNLBs {
		nlbName := getNLBName(nlb)

		// Get current subnets
		currentSubnets := make([]string, 0, len(nlb.AvailabilityZones))
		subnetsToRemove := make([]string, 0)

		for _, az := range nlb.AvailabilityZones {
			subnetID := aws.ToString(az.SubnetId)
			currentSubnets = append(currentSubnets, subnetID)

			if aws.ToString(az.ZoneName) == opts.Zone {
				subnetsToRemove = append(subnetsToRemove, subnetID)
			}
		}

		if len(subnetsToRemove) == 0 {
			fmt.Printf("No subnets found in zone %s for NLB %s\n", opts.Zone, nlbName)
			continue
		}

		// Calculate new subnets (remove the ones in the specified zone)
		newSubnets := make([]string, 0)
		for _, subnet := range currentSubnets {
			shouldRemove := false
			for _, removeSubnet := range subnetsToRemove {
				if subnet == removeSubnet {
					shouldRemove = true
					break
				}
			}
			if !shouldRemove {
				newSubnets = append(newSubnets, subnet)
			}
		}

		if len(newSubnets) == 0 {
			fmt.Printf("❌ Cannot remove all subnets from NLB %s. NLB must have at least one subnet.\n", nlbName)
			fmt.Printf("   Current subnets in zone %s: %v\n", opts.Zone, subnetsToRemove)
			fmt.Printf("   💡 To resolve this:\n")
			fmt.Printf("   1. First add subnets from other zones to the NLB\n")
			fmt.Printf("   2. Or remove subnets from other zones first\n")
			fmt.Printf("   3. Then retry removing subnets from zone %s\n", opts.Zone)
			fmt.Printf("   🔍 Use 'aws nlb list --vpc %s' to see all current subnets\n", opts.VPCID)
			continue
		}

		// Update the NLB
		input := &elasticloadbalancingv2.SetSubnetsInput{
			LoadBalancerArn: nlb.LoadBalancerArn,
			Subnets:         newSubnets,
		}

		_, err = elbv2Client.SetSubnets(context.TODO(), input)
		if err != nil {
			// Provide specific guidance for common AWS errors
			if strings.Contains(err.Error(), "ResourceInUse") && strings.Contains(err.Error(), "Subnets cannot be removed") {
				fmt.Printf("❌ Cannot remove subnets from NLB %s: The load balancer is currently associated with another service (e.g., Kubernetes service, ECS service).\n", nlbName)
				fmt.Printf("   To resolve this:\n")
				fmt.Printf("   1. Check if the NLB is used by Kubernetes services: kubectl get services -o wide\n")
				fmt.Printf("   2. Check if the NLB is used by ECS services: aws ecs describe-services --cluster CLUSTER_NAME\n")
				fmt.Printf("   3. Delete or modify the associated service first\n")
				fmt.Printf("   4. Wait a few minutes for the association to be removed\n")
				fmt.Printf("   5. Then retry the subnet removal\n")
			} else if strings.Contains(err.Error(), "InvalidParameter") {
				fmt.Printf("❌ Invalid parameter for NLB %s: %v\n", nlbName, err)
			} else if strings.Contains(err.Error(), "LoadBalancerNotFound") {
				fmt.Printf("❌ NLB %s not found: The load balancer may have been deleted\n", nlbName)
			} else {
				fmt.Printf("❌ Failed to remove subnets from NLB %s: %v\n", nlbName, err)
			}
			continue
		}

		fmt.Printf("Successfully removed subnets from NLB %s\n", nlbName)
		successCount++
	}

	fmt.Printf("\nOperation completed. Successfully updated %d out of %d NLB(s).\n", successCount, len(targetNLBs))
	return nil, nil
}

// parseRemoveSubnetArgs parses command line arguments for the remove-subnet command
func parseRemoveSubnetArgs(args []string) (*RemoveSubnetOptions, error) {
	opts := &RemoveSubnetOptions{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "nlb", "remove-subnet":
			// Skip command names
			continue
		case "--vpc":
			if i+1 < len(args) {
				i++
				opts.VPCID = args[i]
			}
		case "--zone":
			if i+1 < len(args) {
				i++
				opts.Zone = args[i]
			}
		case "--nlb-name":
			if i+1 < len(args) {
				i++
				opts.NLBName = args[i]
			}
		case "--force":
			opts.Force = true
		}
	}

	return opts, nil
}

// RemoveSubnetOptions represents the parsed command line options for the remove-subnet command
type RemoveSubnetOptions struct {
	VPCID   string
	Zone    string
	NLBName string
	Force   bool
}

// findNLBsInVPC finds NLBs in a VPC, optionally filtered by name
func findNLBsInVPC(client *elasticloadbalancingv2.Client, vpcID, nlbName string) ([]elbv2types.LoadBalancer, error) {
	input := &elasticloadbalancingv2.DescribeLoadBalancersInput{}

	result, err := client.DescribeLoadBalancers(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	var nlbs []elbv2types.LoadBalancer
	for _, lb := range result.LoadBalancers {
		// Only include Network Load Balancers
		if lb.Type != elbv2types.LoadBalancerTypeEnumNetwork {
			continue
		}

		// Filter by VPC
		if aws.ToString(lb.VpcId) != vpcID {
			continue
		}

		// Filter by name if specified
		if nlbName != "" {
			actualName := getNLBName(lb)
			if actualName != nlbName {
				continue
			}
		}

		nlbs = append(nlbs, lb)
	}

	return nlbs, nil
}

// getNLBName gets the name of an NLB from its tags
func getNLBName(lb elbv2types.LoadBalancer) string {
	// Get tags for this load balancer
	tags := getLoadBalancerTags(lb.LoadBalancerArn)

	for _, tag := range tags {
		if aws.ToString(tag.Key) == "Name" {
			return aws.ToString(tag.Value)
		}
	}

	// Fallback to ARN if no name tag
	return aws.ToString(lb.LoadBalancerArn)
}

// CheckNLBAssociations handles the check-associations command for checking NLB service associations
func CheckNLBAssociations(ctx *gofr.Context) (any, error) {
	args := os.Args[1:] // Get command line args for parsing flags

	// Check for help flag first
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println("Usage: aws nlb check-associations --vpc VPC_ID [--nlb-name NLB_NAME]")
			fmt.Println("Options:")
			fmt.Println("  --vpc VPC_ID       VPC ID containing the NLB (required)")
			fmt.Println("  --nlb-name NAME    Specific NLB name to check (optional, checks all NLBs if not specified)")
			fmt.Println()
			fmt.Println("This command checks for service associations that might prevent subnet removal from NLBs.")
			fmt.Println("It provides guidance on how to resolve common association issues.")
			return nil, nil
		}
	}

	// Parse arguments
	opts, err := parseCheckAssociationsArgs(args)
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

	// Find NLBs in the VPC
	nlbs, err := findNLBsInVPC(elbv2Client, opts.VPCID, opts.NLBName)
	if err != nil {
		return nil, fmt.Errorf("failed to find NLBs: %w", err)
	}

	if len(nlbs) == 0 {
		return nil, fmt.Errorf("no NLBs found in VPC %s", opts.VPCID)
	}

	fmt.Printf("Checking associations for %d NLB(s) in VPC %s:\n\n", len(nlbs), opts.VPCID)

	for _, nlb := range nlbs {
		nlbName := getNLBName(nlb)
		fmt.Printf("🔍 NLB: %s\n", nlbName)
		fmt.Printf("   ARN: %s\n", aws.ToString(nlb.LoadBalancerArn))
		fmt.Printf("   State: %s\n", string(nlb.State.Code))

		// Check for common association patterns
		hasAssociations := false

		// Check if NLB has listeners (indicates potential service usage)
		listenersInput := &elasticloadbalancingv2.DescribeListenersInput{
			LoadBalancerArn: nlb.LoadBalancerArn,
		}

		listenersResult, err := elbv2Client.DescribeListeners(context.TODO(), listenersInput)
		if err == nil && len(listenersResult.Listeners) > 0 {
			fmt.Printf("   ⚠️  Has %d listener(s) - may be in use by services\n", len(listenersResult.Listeners))
			hasAssociations = true
		}

		// Check for target groups
		targetGroupsInput := &elasticloadbalancingv2.DescribeTargetGroupsInput{
			LoadBalancerArn: nlb.LoadBalancerArn,
		}

		targetGroupsResult, err := elbv2Client.DescribeTargetGroups(context.TODO(), targetGroupsInput)
		if err == nil && len(targetGroupsResult.TargetGroups) > 0 {
			fmt.Printf("   ⚠️  Has %d target group(s) - may be in use by services\n", len(targetGroupsResult.TargetGroups))
			hasAssociations = true
		}

		if !hasAssociations {
			fmt.Printf("   ✅ No obvious service associations detected\n")
		}

		fmt.Printf("   💡 To check for Kubernetes services: kubectl get services -o wide | grep %s\n", aws.ToString(nlb.LoadBalancerArn))
		fmt.Printf("   💡 To check for ECS services: aws ecs describe-services --cluster CLUSTER_NAME\n")
		fmt.Println()
	}

	fmt.Printf("📋 Next steps if you encounter 'ResourceInUse' errors:\n")
	fmt.Printf("   1. Check Kubernetes services: kubectl get services -o wide\n")
	fmt.Printf("   2. Check ECS services: aws ecs describe-services --cluster CLUSTER_NAME\n")
	fmt.Printf("   3. Delete or modify associated services\n")
	fmt.Printf("   4. Wait a few minutes for associations to be removed\n")
	fmt.Printf("   5. Retry the subnet removal operation\n")

	return nil, nil
}

// parseCheckAssociationsArgs parses command line arguments for the check-associations command
func parseCheckAssociationsArgs(args []string) (*CheckAssociationsOptions, error) {
	opts := &CheckAssociationsOptions{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "nlb", "check-associations":
			// Skip command names
			continue
		case "--vpc":
			if i+1 < len(args) {
				i++
				opts.VPCID = args[i]
			}
		case "--nlb-name":
			if i+1 < len(args) {
				i++
				opts.NLBName = args[i]
			}
		}
	}

	return opts, nil
}

// CheckAssociationsOptions represents the parsed command line options for the check-associations command
type CheckAssociationsOptions struct {
	VPCID   string
	NLBName string
}

// AddSubnetToNLB handles the add-subnet command for adding subnets to an NLB
func AddSubnetToNLB(ctx *gofr.Context) (any, error) {
	args := os.Args[1:] // Get command line args for parsing flags

	// Check for help flag first
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println("Usage: aws nlb add-subnet --vpc VPC_ID --zone AZ [--nlb-name NLB_NAME] [--force]")
			fmt.Println("Options:")
			fmt.Println("  --vpc VPC_ID       VPC ID containing the NLB (required)")
			fmt.Println("  --zone AZ          Availability zone to add subnets from (required)")
			fmt.Println("  --nlb-name NAME    Specific NLB name to target (optional, adds to all NLBs if not specified)")
			fmt.Println("  --force           Skip confirmation prompt")
			fmt.Println()
			fmt.Println("This command adds subnets from the specified zone to NLBs in the VPC.")
			fmt.Println("This is useful when you need to add subnets before removing others.")
			return nil, nil
		}
	}

	// Parse arguments
	opts, err := parseAddSubnetArgs(args)
	if err != nil {
		return nil, err
	}

	if opts.VPCID == "" {
		return nil, fmt.Errorf("vpc parameter is required")
	}
	if opts.Zone == "" {
		return nil, fmt.Errorf("zone parameter is required")
	}

	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create ELBv2 client
	elbv2Client := elasticloadbalancingv2.NewFromConfig(cfg)

	// Find NLBs in the VPC
	nlbs, err := findNLBsInVPC(elbv2Client, opts.VPCID, opts.NLBName)
	if err != nil {
		return nil, fmt.Errorf("failed to find NLBs: %w", err)
	}

	if len(nlbs) == 0 {
		return nil, fmt.Errorf("no NLBs found in VPC %s", opts.VPCID)
	}

	// Find subnets in the specified zone
	subnets, err := findSubnetsInZone(elbv2Client, opts.VPCID, opts.Zone)
	if err != nil {
		return nil, fmt.Errorf("failed to find subnets in zone %s: %w", opts.Zone, err)
	}

	if len(subnets) == 0 {
		return nil, fmt.Errorf("no subnets found in VPC %s zone %s", opts.VPCID, opts.Zone)
	}

	// Show what will be modified
	fmt.Printf("Found %d NLB(s) in VPC %s:\n", len(nlbs), opts.VPCID)
	for _, nlb := range nlbs {
		nlbName := getNLBName(nlb)
		fmt.Printf("  - %s (%s)\n", nlbName, aws.ToString(nlb.LoadBalancerArn))
	}

	fmt.Printf("\nFound %d subnet(s) in zone %s:\n", len(subnets), opts.Zone)
	for _, subnet := range subnets {
		fmt.Printf("  - %s (%s)\n", aws.ToString(subnet.SubnetId), aws.ToString(subnet.CidrBlock))
	}

	// Confirm addition unless --force is used
	if !opts.Force {
		fmt.Printf("\nAre you sure you want to add subnets from zone %s to these NLBs? (yes/no): ", opts.Zone)
		var response string
		fmt.Scanln(&response)
		if response != "yes" {
			fmt.Println("Operation cancelled.")
			return nil, nil
		}
	}

	// Add subnets to each NLB
	successCount := 0
	for _, nlb := range nlbs {
		nlbName := getNLBName(nlb)

		// Get current subnets
		currentSubnets := make([]string, 0, len(nlb.AvailabilityZones))
		for _, az := range nlb.AvailabilityZones {
			currentSubnets = append(currentSubnets, aws.ToString(az.SubnetId))
		}

		// Add new subnets (avoid duplicates)
		subnetMap := make(map[string]bool)
		for _, subnet := range currentSubnets {
			subnetMap[subnet] = true
		}

		newSubnets := make([]string, 0, len(currentSubnets)+len(subnets))
		newSubnets = append(newSubnets, currentSubnets...)

		addedCount := 0
		for _, subnet := range subnets {
			subnetID := aws.ToString(subnet.SubnetId)
			if !subnetMap[subnetID] {
				newSubnets = append(newSubnets, subnetID)
				subnetMap[subnetID] = true
				addedCount++
			}
		}

		if addedCount == 0 {
			fmt.Printf("No new subnets to add to NLB %s (all subnets already present)\n", nlbName)
			continue
		}

		// Update the NLB
		input := &elasticloadbalancingv2.SetSubnetsInput{
			LoadBalancerArn: nlb.LoadBalancerArn,
			Subnets:         newSubnets,
		}

		_, err = elbv2Client.SetSubnets(context.TODO(), input)
		if err != nil {
			fmt.Printf("❌ Failed to add subnets to NLB %s: %v\n", nlbName, err)
			continue
		}

		fmt.Printf("✅ Successfully added %d subnet(s) to NLB %s\n", addedCount, nlbName)
		successCount++
	}

	fmt.Printf("\nOperation completed. Successfully updated %d out of %d NLB(s).\n", successCount, len(nlbs))
	return nil, nil
}

// parseAddSubnetArgs parses command line arguments for the add-subnet command
func parseAddSubnetArgs(args []string) (*AddSubnetOptions, error) {
	opts := &AddSubnetOptions{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "nlb", "add-subnet":
			// Skip command names
			continue
		case "--vpc":
			if i+1 < len(args) {
				i++
				opts.VPCID = args[i]
			}
		case "--zone":
			if i+1 < len(args) {
				i++
				opts.Zone = args[i]
			}
		case "--nlb-name":
			if i+1 < len(args) {
				i++
				opts.NLBName = args[i]
			}
		case "--force":
			opts.Force = true
		}
	}

	return opts, nil
}

// AddSubnetOptions represents the parsed command line options for the add-subnet command
type AddSubnetOptions struct {
	VPCID   string
	Zone    string
	NLBName string
	Force   bool
}

// findSubnetsInZone finds subnets in a specific VPC and zone
func findSubnetsInZone(client *elasticloadbalancingv2.Client, vpcID, zone string) ([]types.Subnet, error) {
	// We need to use EC2 client for subnet operations
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	ec2Client := ec2.NewFromConfig(cfg)

	input := &ec2.DescribeSubnetsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
			{
				Name:   aws.String("availability-zone"),
				Values: []string{zone},
			},
		},
	}

	result, err := ec2Client.DescribeSubnets(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return result.Subnets, nil
}

// NLBRouter routes nlb sub-commands
func NLBRouter(ctx *gofr.Context) (any, error) {
	args := os.Args[1:] // Get command line args for parsing flags

	// Check for sub-commands first
	if len(args) >= 2 {
		// Check if this is a sub-command with help flag
		if len(args) >= 3 && (args[2] == "-h" || args[2] == "--help") {
			switch args[1] {
			case "add-subnet":
				return AddSubnetToNLB(ctx)
			case "remove-subnet":
				return RemoveSubnetFromNLB(ctx)
			case "check-associations":
				return CheckNLBAssociations(ctx)
			}
		}

		switch args[1] {
		case "add-subnet":
			return AddSubnetToNLB(ctx)
		case "remove-subnet":
			return RemoveSubnetFromNLB(ctx)
		case "check-associations":
			return CheckNLBAssociations(ctx)
		case "list":
			// Remove the "list" argument and pass the rest to ListNLBs
			os.Args = append(os.Args[:1], os.Args[2:]...)
			return ListNLBs(ctx)
		}
	}

	// Check for help flag for main nlb command
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println("Usage: aws nlb [COMMAND]")
			fmt.Println("Commands:")
			fmt.Println("  list               List all Network Load Balancers in a VPC (default)")
			fmt.Println("  add-subnet         Add subnets from a zone to NLBs in a VPC")
			fmt.Println("  remove-subnet      Remove a subnet from NLBs in a VPC and zone")
			fmt.Println("  check-associations Check for service associations that might prevent subnet removal")
			fmt.Println()
			fmt.Println("Examples:")
			fmt.Println("  aws nlb --vpc vpc-12345678")
			fmt.Println("  aws nlb list --vpc vpc-12345678")
			fmt.Println("  aws nlb list --vpc vpc-12345678 --zone us-east-1a")
			fmt.Println("  aws nlb list --vpc vpc-12345678 --sort state")
			fmt.Println("  aws nlb add-subnet --vpc vpc-12345678 --zone us-east-1b")
			fmt.Println("  aws nlb check-associations --vpc vpc-12345678")
			fmt.Println("  aws nlb remove-subnet --vpc vpc-12345678 --zone us-east-1a")
			fmt.Println("  aws nlb remove-subnet --vpc vpc-12345678 --zone us-east-1a --nlb-name my-nlb")
			return nil, nil
		}
	}

	// Default to list command
	return ListNLBs(ctx)
}
