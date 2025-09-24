package print

import (
	"bytes"
	"os"
	"sort"
	"strings"
	"testing"
)

func TestPrintImagesTable(t *testing.T) {
	tests := []struct {
		name          string
		imagesSet     map[string]struct{}
		namespace     string
		allNamespaces bool
		style         string
		sortBy        string
		expectOutput  []string
	}{
		{
			name: "simple table with all namespaces",
			imagesSet: map[string]struct{}{
				"nginx:1.21":   {},
				"redis:7.0":    {},
				"busybox:1.34": {},
			},
			namespace:     "",
			allNamespaces: true,
			style:         "simple",
			sortBy:        "image",
			expectOutput:  []string{"NAMESPACE", "IMAGE", "all", "busybox:1.34", "nginx:1.21", "redis:7.0"},
		},
		{
			name: "colored table with specific namespace",
			imagesSet: map[string]struct{}{
				"nginx:1.21": {},
				"redis:7.0":  {},
			},
			namespace:     "default",
			allNamespaces: false,
			style:         "colored",
			sortBy:        "image",
			expectOutput:  []string{"NAMESPACE", "IMAGE", "default", "nginx:1.21", "redis:7.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call the function
			PrintImagesTable(tt.imagesSet, tt.namespace, tt.allNamespaces, tt.style, tt.sortBy)

			// Close the write end and restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read the output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// Check that expected strings are in the output
			for _, expected := range tt.expectOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %s, got: %s", expected, output)
				}
			}
		})
	}
}

func TestPrintImagesTableWithNamespaces(t *testing.T) {
	tests := []struct {
		name              string
		imageNamespaceMap map[string]string
		style             string
		sortBy            string
		expectOutput      []string
	}{
		{
			name: "table with namespaces sorted by namespace",
			imageNamespaceMap: map[string]string{
				"nginx:1.21":   "default",
				"redis:7.0":    "monitoring",
				"busybox:1.34": "default",
			},
			style:        "simple",
			sortBy:       "namespace",
			expectOutput: []string{"NAMESPACE", "IMAGE", "default", "monitoring", "nginx:1.21", "redis:7.0", "busybox:1.34"},
		},
		{
			name: "table with namespaces sorted by image",
			imageNamespaceMap: map[string]string{
				"nginx:1.21":   "default",
				"redis:7.0":    "monitoring",
				"busybox:1.34": "default",
			},
			style:        "box",
			sortBy:       "image",
			expectOutput: []string{"NAMESPACE", "IMAGE", "busybox:1.34", "nginx:1.21", "redis:7.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call the function
			PrintImagesTableWithNamespaces(tt.imageNamespaceMap, tt.style, tt.sortBy)

			// Close the write end and restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read the output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// Check that expected strings are in the output
			for _, expected := range tt.expectOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %s, got: %s", expected, output)
				}
			}
		})
	}
}

func TestPrintImagesList(t *testing.T) {
	tests := []struct {
		name         string
		imagesSet    map[string]struct{}
		sortBy       string
		expectOutput []string
	}{
		{
			name: "list sorted by image",
			imagesSet: map[string]struct{}{
				"nginx:1.21":   {},
				"redis:7.0":    {},
				"busybox:1.34": {},
			},
			sortBy:       "image",
			expectOutput: []string{"busybox:1.34", "nginx:1.21", "redis:7.0"},
		},
		{
			name: "list with no sorting",
			imagesSet: map[string]struct{}{
				"nginx:1.21":   {},
				"redis:7.0":    {},
				"busybox:1.34": {},
			},
			sortBy:       "none",
			expectOutput: []string{"nginx:1.21", "redis:7.0", "busybox:1.34"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call the function
			PrintImagesList(tt.imagesSet, tt.sortBy)

			// Close the write end and restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read the output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// Check that expected strings are in the output
			for _, expected := range tt.expectOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %s, got: %s", expected, output)
				}
			}
		})
	}
}

func TestPrintImagesHelp(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the function
	PrintImagesHelp()

	// Close the write end and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Check that help information is present
	expectedHelp := []string{
		"Usage:",
		"images",
		"--namespace",
		"--all-namespaces",
		"--by-pod",
		"--table",
		"--style",
		"--sort",
		"Options:",
		"--help",
	}

	for _, expected := range expectedHelp {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %s, got: %s", expected, output)
		}
	}
}

func TestImageNamespaceStruct(t *testing.T) {
	// Test the ImageNamespace struct
	imgNs := ImageNamespace{
		Image:     "nginx:1.21",
		Namespace: "default",
	}

	if imgNs.Image != "nginx:1.21" {
		t.Errorf("Expected Image to be 'nginx:1.21', got '%s'", imgNs.Image)
	}

	if imgNs.Namespace != "default" {
		t.Errorf("Expected Namespace to be 'default', got '%s'", imgNs.Namespace)
	}
}

func TestSortingLogic(t *testing.T) {
	// Test sorting logic independently
	tests := []struct {
		name          string
		imageNsList   []ImageNamespace
		sortBy        string
		expectedOrder []string
	}{
		{
			name: "sort by namespace then image",
			imageNsList: []ImageNamespace{
				{Image: "nginx:1.21", Namespace: "default"},
				{Image: "redis:7.0", Namespace: "monitoring"},
				{Image: "busybox:1.34", Namespace: "default"},
			},
			sortBy:        "namespace",
			expectedOrder: []string{"busybox:1.34", "nginx:1.21", "redis:7.0"},
		},
		{
			name: "sort by image name",
			imageNsList: []ImageNamespace{
				{Image: "nginx:1.21", Namespace: "default"},
				{Image: "redis:7.0", Namespace: "monitoring"},
				{Image: "busybox:1.34", Namespace: "default"},
			},
			sortBy:        "image",
			expectedOrder: []string{"busybox:1.34", "nginx:1.21", "redis:7.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply the same sorting logic used in PrintImagesTableWithNamespaces
			switch tt.sortBy {
			case "image":
				sort.Slice(tt.imageNsList, func(i, j int) bool {
					return tt.imageNsList[i].Image < tt.imageNsList[j].Image
				})
			case "namespace":
				sort.Slice(tt.imageNsList, func(i, j int) bool {
					if tt.imageNsList[i].Namespace == tt.imageNsList[j].Namespace {
						return tt.imageNsList[i].Image < tt.imageNsList[j].Image
					}
					return tt.imageNsList[i].Namespace < tt.imageNsList[j].Namespace
				})
			}

			// Check the order
			for i, expected := range tt.expectedOrder {
				if tt.imageNsList[i].Image != expected {
					t.Errorf("Expected image at position %d to be %s, got %s",
						i, expected, tt.imageNsList[i].Image)
				}
			}
		})
	}
}
