package recycle

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ASGConfig stores the original Auto Scaling Group configuration
type ASGConfig struct {
	Name        string
	MinSize     int32
	MaxSize     int32
	DesiredSize int32
}

// NewRecycleCmd creates the recycle subcommand
func NewRecycleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recycle [node-group-name...]",
		Short: "Recycle EKS node groups by scaling down to zero and back up",
		Long: `Scale down identified node groups to zero, wait for instances to terminate, 
then scale back up to original values and wait for new instances to start.`,
		RunE: runRecycle,
		Example: `  # Recycle a single node group
  kaws aws ngs recycle ng-workers-1
  
  # Recycle multiple node groups
  kaws aws ngs recycle ng-workers-1 ng-workers-2
  
  # With custom region
  kaws aws ngs recycle ng-workers-1 --region us-west-2
  
  # With custom polling interval
  kaws aws ngs recycle ng-workers-1 --poll-interval 10s`,
	}

	cmd.Flags().StringP("region", "r", "", "AWS region (default: from AWS config)")
	cmd.Flags().DurationP("poll-interval", "p", 15*time.Second, "polling interval for status checks")
	cmd.Flags().Duration("timeout", 20*time.Minute, "maximum time to wait for recycle to complete")

	return cmd
}

// runRecycle executes the node group recycle command
func runRecycle(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	region, _ := cmd.Flags().GetString("region")
	pollInterval, _ := cmd.Flags().GetDuration("poll-interval")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	// Get node group names from args
	nodeGroupNames := args
	if len(nodeGroupNames) == 0 {
		return fmt.Errorf("no node group names provided. Use: kaws aws ngs recycle <node-group-name> [node-group-name...]")
	}

	if verbose {
		fmt.Printf("Recycling %d node group(s)\n", len(nodeGroupNames))
		fmt.Printf("Poll interval: %s\n", pollInterval)
		fmt.Printf("Timeout: %s\n", timeout)
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
	asgClient := autoscaling.NewFromConfig(cfg)
	ec2Client := ec2.NewFromConfig(cfg)

	// Process each node group
	for _, ngName := range nodeGroupNames {
		fmt.Printf("\n=== Recycling node group: %s ===\n", ngName)

		if err := recycleNodeGroup(ctx, asgClient, ec2Client, ngName, pollInterval, timeout, verbose); err != nil {
			return fmt.Errorf("failed to recycle node group %s: %w", ngName, err)
		}

		fmt.Printf("✓ Successfully recycled node group: %s\n", ngName)
	}

	return nil
}

// recycleNodeGroup performs the full recycle operation for a single node group
func recycleNodeGroup(ctx context.Context, asgClient *autoscaling.Client, ec2Client *ec2.Client, ngName string, pollInterval, timeout time.Duration, verbose bool) error {
	// Step 1: Get current ASG configuration
	fmt.Println("\n[1/5] Getting current node group configuration...")
	originalConfig, instanceIDs, err := getASGConfig(ctx, asgClient, ngName)
	if err != nil {
		return err
	}

	fmt.Printf("  Current config: Min=%d, Max=%d, Desired=%d\n", originalConfig.MinSize, originalConfig.MaxSize, originalConfig.DesiredSize)
	fmt.Printf("  Current instances: %d\n", len(instanceIDs))

	// Step 2: Scale down to zero
	fmt.Println("\n[2/5] Scaling down to zero...")
	if err := scaleASG(ctx, asgClient, ngName, 0, 0, 0); err != nil {
		return err
	}

	// Step 3: Wait for instances to terminate
	fmt.Println("\n[3/5] Waiting for instances to terminate...")
	if err := waitForInstanceStates(ctx, ec2Client, instanceIDs, []ec2types.InstanceStateName{
		ec2types.InstanceStateNameShuttingDown,
		ec2types.InstanceStateNameTerminated,
	}, pollInterval, timeout, verbose); err != nil {
		return err
	}

	fmt.Println("  All instances terminated")

	// Step 4: Scale back up to original values
	fmt.Println("\n[4/5] Scaling back up to original configuration...")
	if err := scaleASG(ctx, asgClient, ngName, originalConfig.MinSize, originalConfig.MaxSize, originalConfig.DesiredSize); err != nil {
		return err
	}

	// Step 5: Wait for new instances to start (pending state)
	fmt.Println("\n[5/5] Waiting for new instances to start...")
	if err := waitForNewInstances(ctx, asgClient, ec2Client, ngName, int(originalConfig.DesiredSize), pollInterval, timeout, verbose); err != nil {
		return err
	}

	fmt.Println("  All new instances starting")

	return nil
}

// getASGConfig retrieves the current ASG configuration and instance IDs
func getASGConfig(ctx context.Context, client *autoscaling.Client, asgName string) (*ASGConfig, []string, error) {
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{asgName},
	}

	result, err := client.DescribeAutoScalingGroups(ctx, input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to describe ASG: %w", err)
	}

	if len(result.AutoScalingGroups) == 0 {
		return nil, nil, fmt.Errorf("ASG not found: %s", asgName)
	}

	asg := result.AutoScalingGroups[0]

	config := &ASGConfig{
		Name:        *asg.AutoScalingGroupName,
		MinSize:     *asg.MinSize,
		MaxSize:     *asg.MaxSize,
		DesiredSize: *asg.DesiredCapacity,
	}

	// Extract instance IDs
	instanceIDs := make([]string, 0, len(asg.Instances))
	for _, instance := range asg.Instances {
		if instance.InstanceId != nil {
			instanceIDs = append(instanceIDs, *instance.InstanceId)
		}
	}

	return config, instanceIDs, nil
}

// scaleASG updates the ASG size
func scaleASG(ctx context.Context, client *autoscaling.Client, asgName string, min, max, desired int32) error {
	input := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: &asgName,
		MinSize:              &min,
		MaxSize:              &max,
		DesiredCapacity:      &desired,
	}

	_, err := client.UpdateAutoScalingGroup(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update ASG: %w", err)
	}

	fmt.Printf("  Scaled to Min=%d, Max=%d, Desired=%d\n", min, max, desired)
	return nil
}

// waitForInstanceStates waits for all instances to reach one of the specified states
func waitForInstanceStates(ctx context.Context, client *ec2.Client, instanceIDs []string, targetStates []ec2types.InstanceStateName, pollInterval, timeout time.Duration, verbose bool) error {
	if len(instanceIDs) == 0 {
		return nil
	}

	startTime := time.Now()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(timeout):
			return fmt.Errorf("timeout waiting for instances to reach target state")
		case <-ticker.C:
			// Check instance states
			input := &ec2.DescribeInstancesInput{
				InstanceIds: instanceIDs,
			}

			result, err := client.DescribeInstances(ctx, input)
			if err != nil {
				if verbose {
					fmt.Printf("  Warning: failed to describe instances: %v\n", err)
				}
				continue
			}

			allInTargetState := true
			stateCount := make(map[string]int)

			for _, reservation := range result.Reservations {
				for _, instance := range reservation.Instances {
					stateName := instance.State.Name
					stateCount[string(stateName)]++

					inTargetState := false
					for _, targetState := range targetStates {
						if stateName == targetState {
							inTargetState = true
							break
						}
					}

					if !inTargetState {
						allInTargetState = false
					}
				}
			}

			if verbose {
				fmt.Printf("  [%s] Instance states: %v\n", time.Since(startTime).Round(time.Second), stateCount)
			} else {
				fmt.Print(".")
			}

			if allInTargetState {
				if !verbose {
					fmt.Println()
				}
				return nil
			}
		}
	}
}

// waitForNewInstances waits for new instances to appear and reach pending state
func waitForNewInstances(ctx context.Context, asgClient *autoscaling.Client, ec2Client *ec2.Client, asgName string, expectedCount int, pollInterval, timeout time.Duration, verbose bool) error {
	startTime := time.Now()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(timeout):
			return fmt.Errorf("timeout waiting for new instances")
		case <-ticker.C:
			// Get current ASG instances
			input := &autoscaling.DescribeAutoScalingGroupsInput{
				AutoScalingGroupNames: []string{asgName},
			}

			result, err := asgClient.DescribeAutoScalingGroups(ctx, input)
			if err != nil {
				if verbose {
					fmt.Printf("  Warning: failed to describe ASG: %v\n", err)
				}
				continue
			}

			if len(result.AutoScalingGroups) == 0 {
				continue
			}

			asg := result.AutoScalingGroups[0]
			currentInstanceCount := len(asg.Instances)

			if currentInstanceCount >= expectedCount {
				// Check if instances are in pending state
				instanceIDs := make([]string, 0, len(asg.Instances))
				for _, instance := range asg.Instances {
					if instance.InstanceId != nil {
						instanceIDs = append(instanceIDs, *instance.InstanceId)
					}
				}

				if len(instanceIDs) > 0 {
					ec2Input := &ec2.DescribeInstancesInput{
						InstanceIds: instanceIDs,
					}

					ec2Result, err := ec2Client.DescribeInstances(ctx, ec2Input)
					if err == nil {
						pendingCount := 0
						stateCount := make(map[string]int)

						for _, reservation := range ec2Result.Reservations {
							for _, instance := range reservation.Instances {
								stateName := string(instance.State.Name)
								stateCount[stateName]++
								if instance.State.Name == ec2types.InstanceStateNamePending ||
									instance.State.Name == ec2types.InstanceStateNameRunning {
									pendingCount++
								}
							}
						}

						if verbose {
							fmt.Printf("  [%s] Instances: %d/%d, States: %v\n",
								time.Since(startTime).Round(time.Second),
								pendingCount, expectedCount, stateCount)
						} else {
							fmt.Print(".")
						}

						if pendingCount >= expectedCount {
							if !verbose {
								fmt.Println()
							}
							fmt.Printf("  %d instances are now starting (pending/running)\n", pendingCount)
							return nil
						}
					}
				}
			} else {
				if verbose {
					fmt.Printf("  [%s] Waiting for instances to appear: %d/%d\n",
						time.Since(startTime).Round(time.Second),
						currentInstanceCount, expectedCount)
				} else {
					fmt.Print(".")
				}
			}
		}
	}
}
