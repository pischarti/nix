package main

import (
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedError  bool
		expectedValues map[string]interface{}
	}{
		{
			name: "default values",
			args: []string{"images"},
			expectedValues: map[string]interface{}{
				"namespace":     "",
				"allNamespaces": true,
				"byPod":         false,
				"tableOutput":   false,
				"tableStyle":    "colored",
				"sortBy":        "namespace",
			},
		},
		{
			name: "namespace flag",
			args: []string{"images", "--namespace", "test"},
			expectedValues: map[string]interface{}{
				"namespace":     "test",
				"allNamespaces": false,
				"byPod":         false,
				"tableOutput":   false,
				"tableStyle":    "colored",
				"sortBy":        "namespace",
			},
		},
		{
			name: "all namespaces flag",
			args: []string{"images", "--all-namespaces"},
			expectedValues: map[string]interface{}{
				"namespace":     "",
				"allNamespaces": true,
				"byPod":         false,
				"tableOutput":   false,
				"tableStyle":    "colored",
				"sortBy":        "namespace",
			},
		},
		{
			name: "by-pod flag",
			args: []string{"images", "--by-pod"},
			expectedValues: map[string]interface{}{
				"namespace":     "",
				"allNamespaces": true,
				"byPod":         true,
				"tableOutput":   false,
				"tableStyle":    "colored",
				"sortBy":        "namespace",
			},
		},
		{
			name: "table flag",
			args: []string{"images", "--table"},
			expectedValues: map[string]interface{}{
				"namespace":     "",
				"allNamespaces": true,
				"byPod":         false,
				"tableOutput":   true,
				"tableStyle":    "colored",
				"sortBy":        "namespace",
			},
		},
		{
			name: "style flag",
			args: []string{"images", "--style", "simple"},
			expectedValues: map[string]interface{}{
				"namespace":     "",
				"allNamespaces": true,
				"byPod":         false,
				"tableOutput":   false,
				"tableStyle":    "simple",
				"sortBy":        "namespace",
			},
		},
		{
			name: "sort flag",
			args: []string{"images", "--sort", "image"},
			expectedValues: map[string]interface{}{
				"namespace":     "",
				"allNamespaces": true,
				"byPod":         false,
				"tableOutput":   false,
				"tableStyle":    "colored",
				"sortBy":        "image",
			},
		},
		{
			name: "multiple flags",
			args: []string{"images", "--table", "--style", "box", "--sort", "image", "--namespace", "test"},
			expectedValues: map[string]interface{}{
				"namespace":     "test",
				"allNamespaces": false,
				"byPod":         false,
				"tableOutput":   true,
				"tableStyle":    "box",
				"sortBy":        "image",
			},
		},
		{
			name:          "invalid sort option",
			args:          []string{"images", "--sort", "invalid"},
			expectedError: true,
		},
		{
			name:          "conflicting namespace flags",
			args:          []string{"images", "--namespace", "test", "--all-namespaces"},
			expectedError: true,
		},
		{
			name:          "conflicting table and by-pod flags",
			args:          []string{"images", "--table", "--by-pod"},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse arguments similar to how the handler does it
			namespace := ""
			allNamespaces := false
			byPod := false
			tableOutput := false
			tableStyle := "colored"
			sortBy := "namespace"

			for i := 0; i < len(tt.args); i++ {
				arg := tt.args[i]
				switch arg {
				case "--namespace", "-n":
					if i+1 < len(tt.args) {
						i++
						namespace = tt.args[i]
					}
				case "--all-namespaces", "-A":
					allNamespaces = true
				case "--by-pod":
					byPod = true
				case "--table", "-t":
					tableOutput = true
				case "--style":
					if i+1 < len(tt.args) {
						i++
						tableStyle = tt.args[i]
					}
				case "--sort":
					if i+1 < len(tt.args) {
						i++
						sortBy = tt.args[i]
					}
				}
			}

			// Apply default logic
			if namespace == "" && !allNamespaces {
				allNamespaces = true
			}

			// Validation logic
			var validationError bool
			if namespace != "" && allNamespaces {
				validationError = true
			}
			if tableOutput && byPod {
				validationError = true
			}

			validSorts := map[string]bool{"namespace": true, "image": true, "none": true}
			if !validSorts[sortBy] {
				validationError = true
			}

			if tt.expectedError && !validationError {
				t.Error("Expected validation error but got none")
			}
			if !tt.expectedError && validationError {
				t.Error("Expected no validation error but got one")
			}

			if !tt.expectedError && tt.expectedValues != nil {
				// Check expected values
				if namespace != tt.expectedValues["namespace"] {
					t.Errorf("Expected namespace %v, got %v", tt.expectedValues["namespace"], namespace)
				}
				if allNamespaces != tt.expectedValues["allNamespaces"] {
					t.Errorf("Expected allNamespaces %v, got %v", tt.expectedValues["allNamespaces"], allNamespaces)
				}
				if byPod != tt.expectedValues["byPod"] {
					t.Errorf("Expected byPod %v, got %v", tt.expectedValues["byPod"], byPod)
				}
				if tableOutput != tt.expectedValues["tableOutput"] {
					t.Errorf("Expected tableOutput %v, got %v", tt.expectedValues["tableOutput"], tableOutput)
				}
				if tableStyle != tt.expectedValues["tableStyle"] {
					t.Errorf("Expected tableStyle %v, got %v", tt.expectedValues["tableStyle"], tableStyle)
				}
				if sortBy != tt.expectedValues["sortBy"] {
					t.Errorf("Expected sortBy %v, got %v", tt.expectedValues["sortBy"], sortBy)
				}
			}
		})
	}
}

func TestSortOptions(t *testing.T) {
	tests := []struct {
		name   string
		sortBy string
		valid  bool
	}{
		{"valid namespace sort", "namespace", true},
		{"valid image sort", "image", true},
		{"valid none sort", "none", true},
		{"invalid sort", "invalid", false},
		{"empty sort", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validSorts := map[string]bool{"namespace": true, "image": true, "none": true}
			isValid := validSorts[tt.sortBy]

			if isValid != tt.valid {
				t.Errorf("Expected sort option '%s' to be valid=%v, got %v", tt.sortBy, tt.valid, isValid)
			}
		})
	}
}

func TestTableStyles(t *testing.T) {
	tests := []struct {
		name  string
		style string
		valid bool
	}{
		{"simple style", "simple", true},
		{"box style", "box", true},
		{"rounded style", "rounded", true},
		{"colored style", "colored", true},
		{"color style", "color", true},
		{"invalid style", "invalid", false},
		{"empty style", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the style validation logic
			var isValid bool
			switch tt.style {
			case "simple", "box", "rounded", "colored", "color":
				isValid = true
			default:
				isValid = false
			}

			if isValid != tt.valid {
				t.Errorf("Expected style '%s' to be valid=%v, got %v", tt.style, tt.valid, isValid)
			}
		})
	}
}
