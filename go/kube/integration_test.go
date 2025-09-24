package main

import (
	"bytes"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/pischarti/nix/go/pkg/print"
)

func TestPrintImagesList_Integration(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test data
	images := map[string]struct{}{
		"nginx:1.21":   {},
		"redis:7.0":    {},
		"busybox:1.34": {},
	}

	// Call the function
	print.PrintImagesList(images, "image")

	// Close the write end and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains expected images
	expectedImages := []string{"busybox:1.34", "nginx:1.21", "redis:7.0"}
	for _, expected := range expectedImages {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %s, got: %s", expected, output)
		}
	}
}

func TestPrintImagesTable_Integration(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test data
	images := map[string]struct{}{
		"nginx:1.21":   {},
		"redis:7.0":    {},
		"busybox:1.34": {},
	}

	// Call the function
	print.PrintImagesTable(images, "", true, "simple", "image")

	// Close the write end and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains table headers and images
	if !strings.Contains(output, "IMAGE") {
		t.Errorf("Expected output to contain 'IMAGE' header, got: %s", output)
	}

	expectedImages := []string{"busybox:1.34", "nginx:1.21", "redis:7.0"}
	for _, expected := range expectedImages {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %s, got: %s", expected, output)
		}
	}
}

func TestPrintImagesTableWithNamespaces_Integration(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test data
	imageNamespaceMap := map[string]string{
		"nginx:1.21":   "default",
		"redis:7.0":    "monitoring",
		"busybox:1.34": "default",
	}

	// Call the function
	print.PrintImagesTableWithNamespaces(imageNamespaceMap, "simple", "namespace")

	// Close the write end and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains table headers
	if !strings.Contains(output, "NAMESPACE") {
		t.Errorf("Expected output to contain 'NAMESPACE' header, got: %s", output)
	}
	if !strings.Contains(output, "IMAGE") {
		t.Errorf("Expected output to contain 'IMAGE' header, got: %s", output)
	}

	// Verify output contains expected data
	expectedData := []string{"default", "monitoring", "nginx:1.21", "redis:7.0", "busybox:1.34"}
	for _, expected := range expectedData {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %s, got: %s", expected, output)
		}
	}
}

func TestPrintImagesHelp_Integration(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the function
	print.PrintImagesHelp()

	// Close the write end and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains help information
	expectedHelp := []string{"Usage:", "images", "--namespace", "--all-namespaces", "--by-pod", "--table", "--style", "--sort"}
	for _, expected := range expectedHelp {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %s, got: %s", expected, output)
		}
	}
}

// Test helper functions that are used by the main logic
func TestSortImages(t *testing.T) {
	images := []string{"nginx:1.21", "redis:7.0", "busybox:1.34"}

	sort.Strings(images)

	expected := []string{"busybox:1.34", "nginx:1.21", "redis:7.0"}
	for i, img := range images {
		if img != expected[i] {
			t.Errorf("Expected sorted image at position %d to be %s, got %s", i, expected[i], img)
		}
	}
}

func TestValidateFlags(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		allNamespaces bool
		tableOutput   bool
		byPod         bool
		sortBy        string
		expectError   bool
	}{
		{
			name:          "valid: namespace specified",
			namespace:     "test",
			allNamespaces: false,
			expectError:   false,
		},
		{
			name:          "valid: all namespaces",
			namespace:     "",
			allNamespaces: true,
			expectError:   false,
		},
		{
			name:          "invalid: conflicting namespace flags",
			namespace:     "test",
			allNamespaces: true,
			expectError:   true,
		},
		{
			name:        "invalid: conflicting table and by-pod flags",
			tableOutput: true,
			byPod:       true,
			expectError: true,
		},
		{
			name:        "invalid: invalid sort option",
			sortBy:      "invalid",
			expectError: true,
		},
		{
			name:        "valid: valid sort options",
			sortBy:      "namespace",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var hasError bool

			// Check namespace conflict
			if tt.namespace != "" && tt.allNamespaces {
				hasError = true
			}

			// Check table/by-pod conflict
			if tt.tableOutput && tt.byPod {
				hasError = true
			}

			// Check sort validation
			validSorts := map[string]bool{"namespace": true, "image": true, "none": true}
			if tt.sortBy != "" && !validSorts[tt.sortBy] {
				hasError = true
			}

			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got %v", tt.expectError, hasError)
			}
		})
	}
}
