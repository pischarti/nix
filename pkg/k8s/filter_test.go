package k8s

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFilterEvents(t *testing.T) {
	now := metav1.NewTime(time.Now())

	events := []corev1.Event{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "event-1",
				Namespace: "default",
			},
			Message:        "Failed to pull image: rpc error: code = Unknown desc = failed to get sandbox image",
			Type:           "Warning",
			Reason:         "FailedCreatePodSandBox",
			Count:          5,
			FirstTimestamp: now,
			LastTimestamp:  now,
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "event-2",
				Namespace: "kube-system",
			},
			Message:        "Successfully pulled image",
			Type:           "Normal",
			Reason:         "Pulled",
			Count:          1,
			FirstTimestamp: now,
			LastTimestamp:  now,
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "event-3",
				Namespace: "default",
			},
			Message:        "Error: failed to get sandbox image for container",
			Type:           "Warning",
			Reason:         "FailedCreatePodSandBox",
			Count:          3,
			FirstTimestamp: now,
			LastTimestamp:  now,
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "event-4",
				Namespace: "production",
			},
			Message:        "ImagePullBackOff: Back-off pulling image",
			Type:           "Warning",
			Reason:         "BackOff",
			Count:          10,
			FirstTimestamp: now,
			LastTimestamp:  now,
		},
	}

	tests := []struct {
		name          string
		events        []corev1.Event
		searchTerm    string
		expectedCount int
		expectedNames []string
	}{
		{
			name:          "filter by 'failed to get sandbox image'",
			events:        events,
			searchTerm:    "failed to get sandbox image",
			expectedCount: 2,
			expectedNames: []string{"event-1", "event-3"},
		},
		{
			name:          "filter by 'ImagePullBackOff'",
			events:        events,
			searchTerm:    "ImagePullBackOff",
			expectedCount: 1,
			expectedNames: []string{"event-4"},
		},
		{
			name:          "filter by 'Successfully'",
			events:        events,
			searchTerm:    "Successfully",
			expectedCount: 1,
			expectedNames: []string{"event-2"},
		},
		{
			name:          "no matches",
			events:        events,
			searchTerm:    "nonexistent",
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name:          "empty search term",
			events:        events,
			searchTerm:    "",
			expectedCount: 4,
			expectedNames: []string{"event-1", "event-2", "event-3", "event-4"},
		},
		{
			name:          "case sensitive search",
			events:        events,
			searchTerm:    "Failed",
			expectedCount: 1,
			expectedNames: []string{"event-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterEvents(tt.events, tt.searchTerm)

			if len(result) != tt.expectedCount {
				t.Errorf("FilterEvents() returned %d events, want %d", len(result), tt.expectedCount)
			}

			// Check that the correct events were returned
			for i, expectedName := range tt.expectedNames {
				if i >= len(result) {
					t.Errorf("Expected event %q at index %d, but result only has %d events", expectedName, i, len(result))
					continue
				}
				if result[i].Name != expectedName {
					t.Errorf("Event at index %d has name %q, want %q", i, result[i].Name, expectedName)
				}
			}
		})
	}
}

func TestFilterEvents_EmptyList(t *testing.T) {
	events := []corev1.Event{}
	result := FilterEvents(events, "test")

	if len(result) != 0 {
		t.Errorf("FilterEvents() on empty list returned %d events, want 0", len(result))
	}
}

func TestFilterEvents_RealWorldExample(t *testing.T) {
	now := metav1.NewTime(time.Now())

	// Real-world example with the specific error message mentioned in the requirement
	events := []corev1.Event{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pod.17a6b8c9d3e1f2a4",
				Namespace: "default",
			},
			InvolvedObject: corev1.ObjectReference{
				Kind: "Pod",
				Name: "my-pod",
			},
			Message:        "Failed to create pod sandbox: rpc error: code = Unknown desc = failed to get sandbox image \"registry.k8s.io/pause:3.9\": failed to pull image",
			Type:           "Warning",
			Reason:         "FailedCreatePodSandBox",
			Count:          15,
			FirstTimestamp: now,
			LastTimestamp:  now,
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "another-pod.abc123",
				Namespace: "kube-system",
			},
			InvolvedObject: corev1.ObjectReference{
				Kind: "Pod",
				Name: "another-pod",
			},
			Message:        "Pod started successfully",
			Type:           "Normal",
			Reason:         "Started",
			Count:          1,
			FirstTimestamp: now,
			LastTimestamp:  now,
		},
	}

	// This is the real-world use case: filtering for sandbox image issues
	searchTerm := "failed to get sandbox image"
	result := FilterEvents(events, searchTerm)

	if len(result) != 1 {
		t.Errorf("Expected 1 event matching %q, got %d", searchTerm, len(result))
	}

	if len(result) > 0 {
		if result[0].Name != "my-pod.17a6b8c9d3e1f2a4" {
			t.Errorf("Expected event name 'my-pod.17a6b8c9d3e1f2a4', got %q", result[0].Name)
		}
		if result[0].Reason != "FailedCreatePodSandBox" {
			t.Errorf("Expected reason 'FailedCreatePodSandBox', got %q", result[0].Reason)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		substr   string
		expected bool
	}{
		{
			name:     "exact match",
			str:      "hello",
			substr:   "hello",
			expected: true,
		},
		{
			name:     "substring at start",
			str:      "hello world",
			substr:   "hello",
			expected: true,
		},
		{
			name:     "substring in middle",
			str:      "hello world",
			substr:   "lo wo",
			expected: true,
		},
		{
			name:     "substring at end",
			str:      "hello world",
			substr:   "world",
			expected: true,
		},
		{
			name:     "substring not found",
			str:      "hello world",
			substr:   "goodbye",
			expected: false,
		},
		{
			name:     "empty substring",
			str:      "hello",
			substr:   "",
			expected: true,
		},
		{
			name:     "substring longer than string",
			str:      "hi",
			substr:   "hello",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.str, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.str, tt.substr, result, tt.expected)
			}
		})
	}
}
