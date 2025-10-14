package print

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEventsYAML(t *testing.T) {
	now := metav1.NewTime(time.Date(2024, 10, 14, 10, 30, 0, 0, time.UTC))

	events := []corev1.Event{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pod.abc123",
				Namespace: "default",
			},
			InvolvedObject: corev1.ObjectReference{
				Kind: "Pod",
				Name: "my-pod",
			},
			Message:        "Failed to create pod sandbox: failed to get sandbox image",
			Type:           "Warning",
			Reason:         "FailedCreatePodSandBox",
			Count:          5,
			FirstTimestamp: now,
			LastTimestamp:  now,
		},
	}

	// Test that EventsYAML doesn't error
	err := EventsYAML(events)
	if err != nil {
		t.Errorf("EventsYAML() returned error: %v", err)
	}
}

func TestEventsYAML_EmptyList(t *testing.T) {
	events := []corev1.Event{}

	// Test that EventsYAML handles empty list
	err := EventsYAML(events)
	if err != nil {
		t.Errorf("EventsYAML() with empty list returned error: %v", err)
	}
}

func TestEventsTable_EmptyList(t *testing.T) {
	events := []corev1.Event{}

	// Test that EventsTable doesn't panic with empty list
	// This is a smoke test - we can't easily capture stdout
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("EventsTable() panicked with empty list: %v", r)
		}
	}()

	EventsTable(events)
}
