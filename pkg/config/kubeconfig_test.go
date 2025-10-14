package config

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/rest"
)

func TestGetKubeConfig(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() func() // setup function returns cleanup function
		expectError    bool
		validateConfig func(t *testing.T, config *rest.Config)
	}{
		{
			name: "KUBECONFIG environment variable",
			setup: func() func() {
				// Create a temporary kubeconfig file
				tmpDir := t.TempDir()
				kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")

				// Create a minimal kubeconfig file
				kubeconfigContent := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://test.example.com
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user: {}
`
				err := os.WriteFile(kubeconfigPath, []byte(kubeconfigContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create test kubeconfig: %v", err)
				}

				// Set KUBECONFIG environment variable
				originalKubeconfig := os.Getenv("KUBECONFIG")
				os.Setenv("KUBECONFIG", kubeconfigPath)

				return func() {
					// Cleanup
					if originalKubeconfig != "" {
						os.Setenv("KUBECONFIG", originalKubeconfig)
					} else {
						os.Unsetenv("KUBECONFIG")
					}
				}
			},
			expectError: false,
			validateConfig: func(t *testing.T, config *rest.Config) {
				if config.Host != "https://test.example.com" {
					t.Errorf("Expected host https://test.example.com, got %s", config.Host)
				}
			},
		},
		{
			name: "No KUBECONFIG, fallback to default location",
			setup: func() func() {
				// Unset KUBECONFIG to test fallback
				originalKubeconfig := os.Getenv("KUBECONFIG")
				os.Unsetenv("KUBECONFIG")

				return func() {
					if originalKubeconfig != "" {
						os.Setenv("KUBECONFIG", originalKubeconfig)
					}
				}
			},
			expectError: false, // Should not error even if file doesn't exist
			validateConfig: func(t *testing.T, config *rest.Config) {
				// Just verify we get a config object
				if config == nil {
					t.Error("Expected non-nil config")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			config, err := GetKubeConfig()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tt.expectError && config != nil {
				tt.validateConfig(t, config)
			}
		})
	}
}

func TestGetKubeConfig_ErrorHandling(t *testing.T) {
	// Test with invalid KUBECONFIG path
	originalKubeconfig := os.Getenv("KUBECONFIG")
	defer func() {
		if originalKubeconfig != "" {
			os.Setenv("KUBECONFIG", originalKubeconfig)
		} else {
			os.Unsetenv("KUBECONFIG")
		}
	}()

	// Set invalid KUBECONFIG path
	os.Setenv("KUBECONFIG", "/nonexistent/path/kubeconfig")

	config, err := GetKubeConfig()

	// client-go returns an error for invalid kubeconfig paths
	if err == nil {
		t.Error("Expected error with invalid KUBECONFIG path")
	}
	if config != nil {
		t.Error("Expected nil config with invalid KUBECONFIG path")
	}
}
