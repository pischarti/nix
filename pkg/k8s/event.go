package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EventQueryOptions contains options for querying events
type EventQueryOptions struct {
	Namespace string
}

// EventWithNode combines an event with node information for pods
type EventWithNode struct {
	Event      corev1.Event
	NodeName   string
	InstanceID string
}

// QueryEvents retrieves Kubernetes events based on the provided options
func (c *Client) QueryEvents(ctx context.Context, opts EventQueryOptions) ([]corev1.Event, error) {
	eventList, err := c.Clientset.CoreV1().Events(opts.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	return eventList.Items, nil
}

// EnrichEventsWithNodeInfo fetches pod information and adds node names to events
// If fetchInstanceID is true, also fetches EC2 instance IDs from node labels
func (c *Client) EnrichEventsWithNodeInfo(ctx context.Context, events []corev1.Event, fetchInstanceID bool) ([]EventWithNode, error) {
	enrichedEvents := make([]EventWithNode, 0, len(events))

	// Cache pods and nodes to avoid repeated queries
	podCache := make(map[string]string)  // key: namespace/podName, value: nodeName
	nodeCache := make(map[string]string) // key: nodeName, value: instanceID

	for _, event := range events {
		enriched := EventWithNode{
			Event:      event,
			NodeName:   "",
			InstanceID: "",
		}

		// Check if the event is related to a Pod
		if event.InvolvedObject.Kind == "Pod" {
			podKey := fmt.Sprintf("%s/%s", event.InvolvedObject.Namespace, event.InvolvedObject.Name)

			// Check cache first
			var nodeName string
			if cached, found := podCache[podKey]; found {
				nodeName = cached
			} else {
				// Query pod to get node name
				pod, err := c.Clientset.CoreV1().Pods(event.InvolvedObject.Namespace).Get(ctx, event.InvolvedObject.Name, metav1.GetOptions{})
				if err == nil {
					nodeName = pod.Spec.NodeName
					podCache[podKey] = nodeName
				} else {
					// Pod might not exist anymore, use "N/A"
					nodeName = "N/A"
					podCache[podKey] = "N/A"
				}
			}

			enriched.NodeName = nodeName

			// If requested, fetch EC2 instance ID from node labels
			if fetchInstanceID && nodeName != "" && nodeName != "N/A" {
				if instanceID, cached := nodeCache[nodeName]; cached {
					enriched.InstanceID = instanceID
				} else {
					// Query node to get instance ID from labels
					node, err := c.Clientset.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
					if err == nil {
						// Try common label keys for EC2 instance ID
						instanceID := getInstanceIDFromNode(node)
						enriched.InstanceID = instanceID
						nodeCache[nodeName] = instanceID
					} else {
						nodeCache[nodeName] = "N/A"
					}
				}
			}
		}

		enrichedEvents = append(enrichedEvents, enriched)
	}

	return enrichedEvents, nil
}

// getInstanceIDFromNode extracts EC2 instance ID from node labels
func getInstanceIDFromNode(node *corev1.Node) string {
	// Try different common label keys
	labelKeys := []string{
		"node.kubernetes.io/instance-id",
		"topology.kubernetes.io/zone",
		"alpha.eksctl.io/instance-id",
		"failure-domain.beta.kubernetes.io/zone",
	}

	// First try the spec.providerID which is most reliable
	if node.Spec.ProviderID != "" {
		// ProviderID format: aws:///us-east-1a/i-1234567890abcdef0
		// Extract instance ID from provider ID
		if len(node.Spec.ProviderID) > 0 {
			// Split by / and get the last part
			parts := []string{}
			current := ""
			for _, ch := range node.Spec.ProviderID {
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
				// Check if it looks like an instance ID (starts with i-)
				if len(lastPart) > 2 && lastPart[0] == 'i' && lastPart[1] == '-' {
					return lastPart
				}
			}
		}
	}

	// Fall back to labels
	for _, key := range labelKeys {
		if val, ok := node.Labels[key]; ok && val != "" {
			return val
		}
	}

	return "N/A"
}

// FilterEvents filters events by search term in the message field
func FilterEvents(events []corev1.Event, searchTerm string) []corev1.Event {
	matchingEvents := []corev1.Event{}

	for _, event := range events {
		if contains(event.Message, searchTerm) {
			matchingEvents = append(matchingEvents, event)
		}
	}

	return matchingEvents
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(findSubstring(s, substr) != -1))
}

// findSubstring finds the index of a substring in a string
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
