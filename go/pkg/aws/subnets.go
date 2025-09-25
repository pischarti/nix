package aws

import (
	"context"
	"fmt"
	"os"
	"strings"

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

	// Check if subnet exists first and get its details
	describeInput := &ec2.DescribeSubnetsInput{
		SubnetIds: []string{subnetID},
	}

	describeResult, err := ec2Client.DescribeSubnets(context.TODO(), describeInput)
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnet %s: %w", subnetID, err)
	}

	if len(describeResult.Subnets) == 0 {
		return nil, fmt.Errorf("subnet %s not found", subnetID)
	}

	subnet := describeResult.Subnets[0]

	// Check for dependencies that might prevent deletion
	if err := checkSubnetDependencies(ec2Client, subnet); err != nil {
		return nil, fmt.Errorf("cannot delete subnet %s: %w", subnetID, err)
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
		// Provide more helpful error messages for common dependency issues
		if strings.Contains(err.Error(), "has dependencies") {
			return nil, fmt.Errorf("subnet %s has dependencies and cannot be deleted. Use 'aws subnets check-dependencies --subnet-id %s' to see what resources are preventing deletion", subnetID, subnetID)
		}
		if strings.Contains(err.Error(), "network_load_balancer") {
			return nil, fmt.Errorf("subnet %s has Network Load Balancer dependencies that cannot be manually detached. Use 'aws subnets check-dependencies --subnet-id %s' for details, then delete the NLB services via kubectl", subnetID, subnetID)
		}
		if strings.Contains(err.Error(), "InvalidSubnetID.NotFound") {
			return nil, fmt.Errorf("subnet %s not found or may have already been deleted", subnetID)
		}
		if strings.Contains(err.Error(), "InvalidSubnetState") {
			return nil, fmt.Errorf("subnet %s is in an invalid state for deletion. It may have dependencies or be in use", subnetID)
		}
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

// checkSubnetDependencies checks for resources that might prevent subnet deletion
func checkSubnetDependencies(ec2Client *ec2.Client, subnet types.Subnet) error {
	ctx := context.TODO()
	subnetID := aws.ToString(subnet.SubnetId)

	// Check for EC2 instances
	instancesInput := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("subnet-id"),
				Values: []string{subnetID},
			},
		},
	}

	instancesResult, err := ec2Client.DescribeInstances(ctx, instancesInput)
	if err != nil {
		return fmt.Errorf("failed to check for EC2 instances: %w", err)
	}

	var runningInstances []string
	for _, reservation := range instancesResult.Reservations {
		for _, instance := range reservation.Instances {
			if instance.State != nil && instance.State.Name != types.InstanceStateNameTerminated {
				runningInstances = append(runningInstances, aws.ToString(instance.InstanceId))
			}
		}
	}

	if len(runningInstances) > 0 {
		instanceList := strings.Join(runningInstances, "\n   ")
		return fmt.Errorf("subnet has running EC2 instances:\n   %s\nPlease terminate these instances first", instanceList)
	}

	// Check for Network Interfaces (ENIs)
	eniInput := &ec2.DescribeNetworkInterfacesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("subnet-id"),
				Values: []string{subnetID},
			},
		},
	}

	eniResult, err := ec2Client.DescribeNetworkInterfaces(ctx, eniInput)
	if err != nil {
		return fmt.Errorf("failed to check for network interfaces: %w", err)
	}

	var attachedENIs []string
	for _, eni := range eniResult.NetworkInterfaces {
		if eni.Attachment != nil && eni.Attachment.Status != types.AttachmentStatusDetached {
			attachedENIs = append(attachedENIs, aws.ToString(eni.NetworkInterfaceId))
		}
	}

	if len(attachedENIs) > 0 {
		eniList := strings.Join(attachedENIs, "\n   ")
		return fmt.Errorf("subnet has attached network interfaces:\n   %s\nPlease detach these interfaces first", eniList)
	}

	// Check for VPC Endpoints
	endpointsInput := &ec2.DescribeVpcEndpointsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("subnet-id"),
				Values: []string{subnetID},
			},
		},
	}

	endpointsResult, err := ec2Client.DescribeVpcEndpoints(ctx, endpointsInput)
	if err != nil {
		return fmt.Errorf("failed to check for VPC endpoints: %w", err)
	}

	var vpcEndpoints []string
	for _, endpoint := range endpointsResult.VpcEndpoints {
		if endpoint.State != types.StateDeleted {
			vpcEndpoints = append(vpcEndpoints, aws.ToString(endpoint.VpcEndpointId))
		}
	}

	if len(vpcEndpoints) > 0 {
		endpointList := strings.Join(vpcEndpoints, "\n   ")
		return fmt.Errorf("subnet has VPC endpoints:\n   %s\nPlease delete these endpoints first", endpointList)
	}

	// Check for Load Balancers (via Network Interfaces)
	// This is a simplified check - in practice, you might want to use ELBv2 client for more detailed checks
	var loadBalancerENIs []string
	var nlbENIs []string

	for _, eni := range eniResult.NetworkInterfaces {
		if eni.Description != nil {
			desc := aws.ToString(eni.Description)
			eniID := aws.ToString(eni.NetworkInterfaceId)

			if strings.Contains(desc, "ELB") || strings.Contains(desc, "load balancer") {
				loadBalancerENIs = append(loadBalancerENIs, eniID)
			}

			if strings.Contains(desc, "network_load_balancer") || strings.Contains(desc, "net/") {
				// Extract service name from description if possible
				serviceInfo := extractServiceInfoFromDescription(desc)
				if serviceInfo != "" {
					nlbENIs = append(nlbENIs, fmt.Sprintf("%s (%s)", eniID, serviceInfo))
				} else {
					nlbENIs = append(nlbENIs, eniID)
				}
			}
		}
	}

	if len(nlbENIs) > 0 {
		nlbList := strings.Join(nlbENIs, "\n   ")
		return fmt.Errorf("subnet has Network Load Balancer (NLB) network interfaces:\n   %s\nThese ENIs are managed by Kubernetes services and cannot be manually detached.\nPlease delete the associated NLB services first (e.g., via kubectl delete service <service-name>)", nlbList)
	}

	if len(loadBalancerENIs) > 0 {
		lbList := strings.Join(loadBalancerENIs, "\n   ")
		return fmt.Errorf("subnet has load balancer network interfaces:\n   %s\nPlease delete the associated load balancers first", lbList)
	}

	return nil
}

// extractServiceInfoFromDescription extracts service information from ENI description
func extractServiceInfoFromDescription(desc string) string {
	// Look for patterns like "net/k8s-kernos-srtliste-6ed31c790f/e26bb4a60f1986ae"
	// or "k8s-kernos-srtliste-6ed31c790f"

	// Pattern 1: net/k8s-<service-name>-<uuid>/<uuid>
	if strings.Contains(desc, "net/k8s-") {
		parts := strings.Split(desc, "net/k8s-")
		if len(parts) > 1 {
			servicePart := parts[1]
			// Split by '/' to get the service part before the final UUID
			serviceParts := strings.Split(servicePart, "/")
			if len(serviceParts) >= 1 {
				serviceName := serviceParts[0]
				// Remove the UUID suffix (last part after the last dash)
				nameParts := strings.Split(serviceName, "-")
				if len(nameParts) >= 3 {
					// Join all parts except the last one (which is the UUID)
					cleanServiceName := strings.Join(nameParts[:len(nameParts)-1], "-")
					return fmt.Sprintf("k8s service: %s", cleanServiceName)
				}
			}
		}
	}

	// Pattern 2: k8s-<service-name>-<uuid>
	if strings.HasPrefix(desc, "k8s-") {
		// Extract service name from k8s-<name>-<uuid> pattern
		parts := strings.Split(desc, "-")
		if len(parts) >= 3 {
			// Remove the last part (uuid) and join the middle parts
			serviceName := strings.Join(parts[1:len(parts)-1], "-")
			return fmt.Sprintf("k8s service: %s", serviceName)
		}
	}

	return ""
}

// CheckSubnetDependencies handles the check-dependencies command
func CheckSubnetDependencies(ctx *gofr.Context) (any, error) {
	args := os.Args[1:] // Get command line args for parsing flags

	// Check for help flag first
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println("Usage: aws subnets check-dependencies --subnet-id SUBNET_ID")
			fmt.Println("Options:")
			fmt.Println("  --subnet-id SUBNET_ID  Subnet ID to check dependencies for (required)")
			fmt.Println()
			fmt.Println("This command checks what AWS resources are preventing a subnet from being deleted.")
			return nil, nil
		}
	}

	// Parse arguments
	subnetID, _, err := parseDeleteSubnetArgs(args)
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

	describeResult, err := ec2Client.DescribeSubnets(context.TODO(), describeInput)
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnet %s: %w", subnetID, err)
	}

	if len(describeResult.Subnets) == 0 {
		return nil, fmt.Errorf("subnet %s not found", subnetID)
	}

	subnet := describeResult.Subnets[0]

	// Display subnet information
	fmt.Printf("Checking dependencies for subnet: %s\n", subnetID)
	fmt.Printf("VPC: %s\n", aws.ToString(subnet.VpcId))
	fmt.Printf("CIDR: %s\n", aws.ToString(subnet.CidrBlock))
	fmt.Printf("AZ: %s\n", aws.ToString(subnet.AvailabilityZone))
	fmt.Printf("State: %s\n\n", string(subnet.State))

	// Check for dependencies
	if err := checkSubnetDependencies(ec2Client, subnet); err != nil {
		fmt.Printf("❌ Dependencies found that prevent deletion:\n")
		fmt.Printf("   %s\n", err.Error())
		return nil, nil
	}

	fmt.Printf("✅ No dependencies found. Subnet can be deleted.\n")
	return nil, nil
}

// SubnetsRouter routes subnets sub-commands
func SubnetsRouter(ctx *gofr.Context) (any, error) {
	args := os.Args[1:] // Get command line args for parsing flags

	// Check if there's a sub-command
	if len(args) >= 2 && args[1] == "delete" {
		// Route to delete command (it will handle its own help)
		return DeleteSubnet(ctx)
	}

	if len(args) >= 2 && args[1] == "check-dependencies" {
		// Route to check dependencies command
		return CheckSubnetDependencies(ctx)
	}

	// Check for help flag for main subnets command
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println("Usage: aws subnets [COMMAND]")
			fmt.Println("Commands:")
			fmt.Println("  list               List all subnets in a VPC (default)")
			fmt.Println("  delete             Delete a subnet by ID")
			fmt.Println("  check-dependencies Check what resources are preventing subnet deletion")
			fmt.Println()
			fmt.Println("Examples:")
			fmt.Println("  aws subnets --vpc vpc-12345678")
			fmt.Println("  aws subnets list --vpc vpc-12345678")
			fmt.Println("  aws subnets delete --subnet-id subnet-12345678")
			fmt.Println("  aws subnets check-dependencies --subnet-id subnet-12345678")
			return nil, nil
		}
	}

	// Default to list command
	return ListSubnets(ctx)
}
