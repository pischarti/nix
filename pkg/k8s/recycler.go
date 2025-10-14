package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// RecyclerConfig holds configuration for event checking and recycling
type RecyclerConfig struct {
	SearchTerms []string
	Threshold   int
	DryRun      bool
}

// NodeGroupEventCounts maps node group names to event counts
type NodeGroupEventCounts map[string]int

// RecyclerStatus holds the status information to be updated
type RecyclerStatus struct {
	EventCounts   NodeGroupEventCounts
	LastCheckTime metav1.Time
}

// CheckAndRecycleWithStatus checks for matching events and returns both counts and status
func CheckAndRecycleWithStatus(
	ctx context.Context,
	kubeClient client.Client,
	ec2Client *ec2.Client,
	config RecyclerConfig,
	processedEvents map[string]metav1.Time,
) (NodeGroupEventCounts, RecyclerStatus, error) {
	nodeGroupCounts, err := CheckAndRecycle(ctx, kubeClient, ec2Client, config, processedEvents)
	if err != nil {
		return nil, RecyclerStatus{}, err
	}

	status := RecyclerStatus{
		EventCounts:   nodeGroupCounts,
		LastCheckTime: metav1.Now(),
	}

	return nodeGroupCounts, status, nil
}

// CheckAndRecycle checks for matching events and determines which node groups need recycling
func CheckAndRecycle(
	ctx context.Context,
	kubeClient client.Client,
	ec2Client *ec2.Client,
	config RecyclerConfig,
	processedEvents map[string]metav1.Time,
) (NodeGroupEventCounts, error) {
	log := log.FromContext(ctx)

	// List all events using the client's cached informer
	eventList := &corev1.EventList{}
	if err := kubeClient.List(ctx, eventList); err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	log.Info("Checking events", "total", len(eventList.Items))

	// Track node groups that need recycling
	nodeGroupCounts := make(NodeGroupEventCounts)

	// Check each search term
	for _, searchTerm := range config.SearchTerms {
		matchingEvents := FilterEvents(eventList.Items, searchTerm)

		if len(matchingEvents) == 0 {
			continue
		}

		// Filter out recently processed events
		recentEvents := FilterRecentEvents(matchingEvents, processedEvents)

		if len(recentEvents) == 0 {
			continue
		}

		log.Info("Found matching events", "searchTerm", searchTerm, "count", len(recentEvents))

		// For each event, try to identify the node group
		for _, event := range recentEvents {
			if event.InvolvedObject.Kind != "Pod" {
				continue
			}

			// Get pod to find node
			var pod corev1.Pod
			podKey := client.ObjectKey{
				Namespace: event.InvolvedObject.Namespace,
				Name:      event.InvolvedObject.Name,
			}

			// Use cached pod lookup (thread-safe via informer cache)
			if err := kubeClient.Get(ctx, podKey, &pod); err != nil {
				log.V(1).Info("Could not get pod", "pod", event.InvolvedObject.Name, "error", err)
				continue
			}

			if pod.Spec.NodeName == "" {
				continue
			}

			// Get node to find instance ID
			var node corev1.Node
			nodeKey := client.ObjectKey{Name: pod.Spec.NodeName}

			// Use cached node lookup (thread-safe via informer cache)
			if err := kubeClient.Get(ctx, nodeKey, &node); err != nil {
				log.V(1).Info("Could not get node", "node", pod.Spec.NodeName, "error", err)
				continue
			}

			// Extract instance ID and find node group
			instanceID := extractInstanceIDFromProviderID(node.Spec.ProviderID)
			if instanceID == "" || instanceID == "N/A" {
				continue
			}

			// Find node group from instance tags
			nodeGroups, err := findNodeGroupByInstanceID(ctx, ec2Client, instanceID)
			if err != nil {
				log.V(1).Info("Could not find node group", "instance", instanceID, "error", err)
				continue
			}

			for _, ng := range nodeGroups {
				if ng != "" && ng != "Unknown" {
					nodeGroupCounts[ng]++
				}
			}
		}
	}

	// Log node groups that meet or exceed threshold
	for ng, count := range nodeGroupCounts {
		if count >= config.Threshold {
			log.Info("Node group exceeds threshold", "nodeGroup", ng, "count", count, "threshold", config.Threshold)

			if config.DryRun {
				log.Info("[DRY RUN] Would recycle node group", "nodeGroup", ng)
			} else {
				log.Info("Node group ready for recycling", "nodeGroup", ng)
				// Note: Actual recycling is done by the caller
			}
		}
	}

	return nodeGroupCounts, nil
}

// FilterRecentEvents filters out events that have been processed recently
// It marks new events as processed and cleans up old entries (>2 hours)
func FilterRecentEvents(events []corev1.Event, processedEvents map[string]metav1.Time) []corev1.Event {
	recentEvents := []corev1.Event{}

	for _, event := range events {
		eventKey := fmt.Sprintf("%s/%s", event.Namespace, event.Name)

		// Check if we've processed this event recently (within last hour)
		if lastProcessed, found := processedEvents[eventKey]; found {
			if metav1.Now().Time.Sub(lastProcessed.Time) < time.Hour {
				continue
			}
		}

		recentEvents = append(recentEvents, event)

		// Mark as processed
		processedEvents[eventKey] = metav1.Now()
	}

	// Clean up old entries (older than 2 hours)
	for key, timestamp := range processedEvents {
		if metav1.Now().Time.Sub(timestamp.Time) > 2*time.Hour {
			delete(processedEvents, key)
		}
	}

	return recentEvents
}

// findNodeGroupByInstanceID queries AWS EC2 to find the node group name for a given instance ID
// It looks for standard EKS node group tags on the instance
func findNodeGroupByInstanceID(ctx context.Context, ec2Client *ec2.Client, instanceID string) ([]string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	result, err := ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance %s: %w", instanceID, err)
	}

	nodeGroups := []string{}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			for _, tag := range instance.Tags {
				if tag.Key == nil || tag.Value == nil {
					continue
				}

				// Check for EKS node group tags
				// eks:nodegroup-name is used by EKS
				// alpha.eksctl.io/nodegroup-name is used by eksctl
				if *tag.Key == "eks:nodegroup-name" || *tag.Key == "alpha.eksctl.io/nodegroup-name" {
					nodeGroups = append(nodeGroups, *tag.Value)
				}
			}
		}
	}

	return nodeGroups, nil
}

// extractInstanceIDFromProviderID extracts EC2 instance ID from Kubernetes provider ID
// Provider ID format: aws:///us-east-1a/i-1234567890abcdef0
func extractInstanceIDFromProviderID(providerID string) string {
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
