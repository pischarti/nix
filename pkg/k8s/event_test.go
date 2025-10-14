package k8s

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetInstanceIDFromNode(t *testing.T) {
	tests := []struct {
		name       string
		node       *corev1.Node
		expectedID string
	}{
		{
			name: "provider ID with instance ID",
			node: &corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "aws:///us-east-1a/i-1234567890abcdef0",
				},
			},
			expectedID: "i-1234567890abcdef0",
		},
		{
			name: "provider ID without availability zone",
			node: &corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "aws:///i-0987654321fedcba0",
				},
			},
			expectedID: "i-0987654321fedcba0",
		},
		{
			name: "instance ID from label node.kubernetes.io/instance-id",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"node.kubernetes.io/instance-id": "i-abcdef1234567890",
					},
				},
			},
			expectedID: "i-abcdef1234567890",
		},
		{
			name: "instance ID from eksctl label",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"alpha.eksctl.io/instance-id": "i-fedcba0987654321",
					},
				},
			},
			expectedID: "i-fedcba0987654321",
		},
		{
			name: "provider ID takes precedence over labels",
			node: &corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "aws:///us-west-2b/i-priority123456789",
				},
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"node.kubernetes.io/instance-id": "i-ignored123456789",
					},
				},
			},
			expectedID: "i-priority123456789",
		},
		{
			name: "no instance ID available",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"some-other-label": "value",
					},
				},
			},
			expectedID: "N/A",
		},
		{
			name: "empty provider ID",
			node: &corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "",
				},
			},
			expectedID: "N/A",
		},
		{
			name: "invalid provider ID format",
			node: &corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "invalid-format",
				},
			},
			expectedID: "N/A",
		},
		{
			name: "complex provider ID with multiple slashes",
			node: &corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "aws:///us-east-1a/zone-extra/i-complexid123456",
				},
			},
			expectedID: "i-complexid123456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getInstanceIDFromNode(tt.node)
			if result != tt.expectedID {
				t.Errorf("getInstanceIDFromNode() = %q, want %q", result, tt.expectedID)
			}
		})
	}
}

func TestEventWithNode_Struct(t *testing.T) {
	// Test that EventWithNode can be created with all fields
	event := corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-event",
			Namespace: "default",
		},
	}

	enriched := EventWithNode{
		Event:      event,
		NodeName:   "node-1",
		InstanceID: "i-1234567890abcdef0",
	}

	if enriched.NodeName != "node-1" {
		t.Errorf("NodeName = %q, want %q", enriched.NodeName, "node-1")
	}
	if enriched.InstanceID != "i-1234567890abcdef0" {
		t.Errorf("InstanceID = %q, want %q", enriched.InstanceID, "i-1234567890abcdef0")
	}
	if enriched.Event.Name != "test-event" {
		t.Errorf("Event.Name = %q, want %q", enriched.Event.Name, "test-event")
	}
}
