package k8s

// import (
// 	"testing"
// )

// func TestEventQueryOptions(t *testing.T) {
// 	// Test that EventQueryOptions can be created with different namespaces
// 	tests := []struct {
// 		name      string
// 		namespace string
// 	}{
// 		{
// 			name:      "default namespace",
// 			namespace: "default",
// 		},
// 		{
// 			name:      "kube-system namespace",
// 			namespace: "kube-system",
// 		},
// 		{
// 			name:      "empty namespace (all namespaces)",
// 			namespace: "",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			opts := EventQueryOptions{
// 				Namespace: tt.namespace,
// 			}

// 			if opts.Namespace != tt.namespace {
// 				t.Errorf("EventQueryOptions.Namespace = %q, want %q", opts.Namespace, tt.namespace)
// 			}
// 		})
// 	}
// }

// Note: Testing NewClient and QueryEvents would require either:
// 1. A real Kubernetes cluster (integration test)
// 2. Mock Kubernetes client (using interfaces)
// 3. Fake clientset from k8s.io/client-go/kubernetes/fake
// These are typically done in integration tests rather than unit tests
