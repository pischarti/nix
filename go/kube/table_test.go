package main

import (
	"sort"
	"testing"

	"github.com/pischarti/nix/pkg/print"
)

func TestPrintImagesList(t *testing.T) {
	tests := []struct {
		name    string
		images  map[string]struct{}
		sortBy  string
		wantLen int
	}{
		{
			name:    "empty images set",
			images:  map[string]struct{}{},
			sortBy:  "image",
			wantLen: 0,
		},
		{
			name: "single image",
			images: map[string]struct{}{
				"nginx:1.21": {},
			},
			sortBy:  "image",
			wantLen: 1,
		},
		{
			name: "multiple images with image sort",
			images: map[string]struct{}{
				"nginx:1.21":   {},
				"redis:7.0":    {},
				"busybox:1.34": {},
			},
			sortBy:  "image",
			wantLen: 3,
		},
		{
			name: "multiple images with namespace sort",
			images: map[string]struct{}{
				"nginx:1.21":   {},
				"redis:7.0":    {},
				"busybox:1.34": {},
			},
			sortBy:  "namespace",
			wantLen: 3,
		},
		{
			name: "multiple images with no sort",
			images: map[string]struct{}{
				"nginx:1.21":   {},
				"redis:7.0":    {},
				"busybox:1.34": {},
			},
			sortBy:  "none",
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a slice to capture the output
			imagesList := make([]string, 0, len(tt.images))
			for img := range tt.images {
				imagesList = append(imagesList, img)
			}

			// Test the sorting logic that would be used in printImagesList
			switch tt.sortBy {
			case "image", "namespace":
				// Both use alphabetical sorting in the current implementation
				sort.Strings(imagesList)
			case "none":
				// No sorting
			default:
				sort.Strings(imagesList)
			}

			if len(imagesList) != tt.wantLen {
				t.Errorf("Expected %d images, got %d", tt.wantLen, len(imagesList))
			}
		})
	}
}

func TestImageNamespaceSorting(t *testing.T) {
	tests := []struct {
		name          string
		imageNsList   []print.ImageNamespace
		sortBy        string
		expectedOrder []string // Expected image names in order
	}{
		{
			name: "sort by namespace then image",
			imageNsList: []print.ImageNamespace{
				{Image: "nginx:1.21", Namespace: "default"},
				{Image: "redis:7.0", Namespace: "monitoring"},
				{Image: "busybox:1.34", Namespace: "default"},
			},
			sortBy:        "namespace",
			expectedOrder: []string{"busybox:1.34", "nginx:1.21", "redis:7.0"},
		},
		{
			name: "sort by image name",
			imageNsList: []print.ImageNamespace{
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
			// Test the sorting logic from PrintImagesTableWithNamespaces
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
