package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestInitConfig_WithConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")

	content := []byte("verbose: true\nnamespace: test-namespace\n")
	if err := os.WriteFile(configFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Reset viper for clean test
	viper.Reset()

	// Initialize config with the test file
	err := InitConfig(configFile)
	if err != nil {
		t.Errorf("InitConfig() returned error: %v", err)
	}

	// Verify config was loaded
	if !viper.GetBool("verbose") {
		t.Error("Expected verbose to be true from config file")
	}

	if viper.GetString("namespace") != "test-namespace" {
		t.Errorf("Expected namespace to be 'test-namespace', got %q", viper.GetString("namespace"))
	}
}

func TestInitConfig_WithoutConfigFile(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()

	// Initialize config without a file (should not error)
	err := InitConfig("")
	if err != nil {
		t.Errorf("InitConfig() returned error: %v", err)
	}

	// Should have set up viper with defaults
	// (We can't test much more without actually creating a .kaws.yaml file)
}

func TestInitConfig_EnvironmentVariables(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()

	// Set environment variables
	os.Setenv("KAWS_VERBOSE", "true")
	os.Setenv("KAWS_NAMESPACE", "env-namespace")
	defer func() {
		os.Unsetenv("KAWS_VERBOSE")
		os.Unsetenv("KAWS_NAMESPACE")
	}()

	// Initialize config
	err := InitConfig("")
	if err != nil {
		t.Errorf("InitConfig() returned error: %v", err)
	}

	// Verify environment variables are read
	if !viper.GetBool("verbose") {
		t.Error("Expected verbose to be true from environment variable")
	}

	if viper.GetString("namespace") != "env-namespace" {
		t.Errorf("Expected namespace to be 'env-namespace', got %q", viper.GetString("namespace"))
	}
}
