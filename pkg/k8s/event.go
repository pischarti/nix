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

// QueryEvents retrieves Kubernetes events based on the provided options
func (c *Client) QueryEvents(ctx context.Context, opts EventQueryOptions) ([]corev1.Event, error) {
	eventList, err := c.Clientset.CoreV1().Events(opts.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	return eventList.Items, nil
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
