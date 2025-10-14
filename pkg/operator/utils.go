package operator

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/pischarti/nix/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
)

// OperatorConfig holds the operator configuration
type OperatorConfig struct {
	WatchInterval    time.Duration
	SearchTerms      []string
	RecycleThreshold int
	DryRun           bool
	ProcessedEvents  map[string]time.Time
}

// CheckAndRecycle checks for error events and recycles affected node groups
func CheckAndRecycle(ctx context.Context, k8sClient *k8s.Client, ec2Client *ec2.Client, asgClient *autoscaling.Client, opConfig *OperatorConfig, verbose bool) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	if verbose {
		fmt.Printf("[%s] Checking for error events...\n", timestamp)
	}

	// Query all events
	events, err := k8sClient.QueryEvents(ctx, k8s.EventQueryOptions{
		Namespace: "", // All namespaces
	})
	if err != nil {
		return fmt.Errorf("failed to query events: %w", err)
	}

	// Track node groups that need recycling
	nodeGroupsToRecycle := make(map[string]int) // key: nodeGroupName, value: event count

	// Check each search term
	for _, searchTerm := range opConfig.SearchTerms {
		matchingEvents := k8s.FilterEvents(events, searchTerm)

		if len(matchingEvents) == 0 {
			continue
		}

		// Filter out recently processed events (within last hour)
		recentEvents := FilterRecentEvents(matchingEvents, opConfig)

		if len(recentEvents) == 0 {
			continue
		}

		fmt.Printf("[%s] Found %d recent event(s) matching %q\n", timestamp, len(recentEvents), searchTerm)

		// Enrich with node information
		enrichedEvents, err := k8sClient.EnrichEventsWithNodeInfo(ctx, recentEvents, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: Could not enrich events: %v\n", err)
			continue
		}

		// Find affected node groups
		for _, enriched := range enrichedEvents {
			if enriched.InstanceID != "" && enriched.InstanceID != "N/A" {
				// Query node group for this instance
				nodeGroups, err := FindNodeGroupForInstance(ctx, ec2Client, enriched.InstanceID)
				if err != nil {
					if verbose {
						fmt.Fprintf(os.Stderr, "  Warning: Could not find node group for instance %s: %v\n", enriched.InstanceID, err)
					}
					continue
				}

				for _, ng := range nodeGroups {
					if ng != "" && ng != "Unknown" {
						nodeGroupsToRecycle[ng]++
					}
				}
			}
		}
	}

	// Recycle node groups that exceed threshold
	for ngName, count := range nodeGroupsToRecycle {
		if count >= opConfig.RecycleThreshold {
			fmt.Printf("[%s] 🔄 Node group %s has %d problematic events (threshold: %d)\n",
				timestamp, ngName, count, opConfig.RecycleThreshold)

			if opConfig.DryRun {
				fmt.Printf("  [DRY RUN] Would recycle node group: %s\n", ngName)
			} else {
				fmt.Printf("  Recycling node group: %s\n", ngName)
				// Note: Implement recycling logic here or call the recycle function
				fmt.Printf("  ⚠️  Automated recycling not yet implemented - manual intervention required\n")
			}
		} else if verbose {
			fmt.Printf("[%s] Node group %s has %d events (below threshold of %d)\n",
				timestamp, ngName, count, opConfig.RecycleThreshold)
		}
	}

	if len(nodeGroupsToRecycle) == 0 && verbose {
		fmt.Printf("[%s] ✓ No problematic node groups detected\n", timestamp)
	}

	return nil
}

// FilterRecentEvents filters out events that have been processed recently
func FilterRecentEvents(events []corev1.Event, opConfig *OperatorConfig) []corev1.Event {
	recentEvents := []corev1.Event{}
	now := time.Now()

	for _, event := range events {
		eventKey := fmt.Sprintf("%s/%s", event.Namespace, event.Name)

		// Check if we've processed this event recently (within last hour)
		if lastProcessed, found := opConfig.ProcessedEvents[eventKey]; found {
			if now.Sub(lastProcessed) < time.Hour {
				continue // Skip recently processed events
			}
		}

		recentEvents = append(recentEvents, event)
	}

	// Update processed events
	for _, event := range recentEvents {
		eventKey := fmt.Sprintf("%s/%s", event.Namespace, event.Name)
		opConfig.ProcessedEvents[eventKey] = now
	}

	// Clean up old entries (older than 2 hours)
	for key, timestamp := range opConfig.ProcessedEvents {
		if now.Sub(timestamp) > 2*time.Hour {
			delete(opConfig.ProcessedEvents, key)
		}
	}

	return recentEvents
}

// FindNodeGroupForInstance queries AWS to find node group for an instance
func FindNodeGroupForInstance(ctx context.Context, ec2Client *ec2.Client, instanceID string) ([]string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	result, err := ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, err
	}

	nodeGroups := []string{}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// Extract node group from tags
			for _, tag := range instance.Tags {
				if tag.Key == nil || tag.Value == nil {
					continue
				}

				if *tag.Key == "eks:nodegroup-name" || *tag.Key == "alpha.eksctl.io/nodegroup-name" {
					nodeGroups = append(nodeGroups, *tag.Value)
				}
			}
		}
	}

	return nodeGroups, nil
}

// ExtractInstanceIDFromProviderID extracts EC2 instance ID from Kubernetes provider ID
// Provider ID format: aws:///us-east-1a/i-1234567890abcdef0
func ExtractInstanceIDFromProviderID(providerID string) string {
	if providerID == "" {
		return "N/A"
	}

	// Split by '/' to parse the provider ID
	parts := []string{}
	current := ""
	for _, ch := range providerID {
		if ch == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}

	// Get the last part (instance ID)
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		// Verify it looks like an EC2 instance ID (format: i-xxxxxxxxxxxxxxxxx)
		if len(lastPart) > 2 && lastPart[0] == 'i' && lastPart[1] == '-' {
			return lastPart
		}
	}

	return "N/A"
}
