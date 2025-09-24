package container

import (
	"strings"
	"testing"
)

func TestParseImagesArgs(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedOpts  *ImagesOptions
		expectedError bool
	}{
		{
			name: "default values",
			args: []string{"images"},
			expectedOpts: &ImagesOptions{
				Namespace:     "",
				AllNamespaces: true,
				ByPod:         false,
				TableOutput:   false,
				TableStyle:    "colored",
				SortBy:        "namespace",
			},
			expectedError: false,
		},
		{
			name: "namespace flag",
			args: []string{"images", "--namespace", "test"},
			expectedOpts: &ImagesOptions{
				Namespace:     "test",
				AllNamespaces: false,
				ByPod:         false,
				TableOutput:   false,
				TableStyle:    "colored",
				SortBy:        "namespace",
			},
			expectedError: false,
		},
		{
			name: "all namespaces flag",
			args: []string{"images", "--all-namespaces"},
			expectedOpts: &ImagesOptions{
				Namespace:     "",
				AllNamespaces: true,
				ByPod:         false,
				TableOutput:   false,
				TableStyle:    "colored",
				SortBy:        "namespace",
			},
			expectedError: false,
		},
		{
			name: "by-pod flag",
			args: []string{"images", "--by-pod"},
			expectedOpts: &ImagesOptions{
				Namespace:     "",
				AllNamespaces: true,
				ByPod:         true,
				TableOutput:   false,
				TableStyle:    "colored",
				SortBy:        "namespace",
			},
			expectedError: false,
		},
		{
			name: "table flag",
			args: []string{"images", "--table"},
			expectedOpts: &ImagesOptions{
				Namespace:     "",
				AllNamespaces: true,
				ByPod:         false,
				TableOutput:   true,
				TableStyle:    "colored",
				SortBy:        "namespace",
			},
			expectedError: false,
		},
		{
			name: "style flag",
			args: []string{"images", "--style", "simple"},
			expectedOpts: &ImagesOptions{
				Namespace:     "",
				AllNamespaces: true,
				ByPod:         false,
				TableOutput:   false,
				TableStyle:    "simple",
				SortBy:        "namespace",
			},
			expectedError: false,
		},
		{
			name: "sort flag",
			args: []string{"images", "--sort", "image"},
			expectedOpts: &ImagesOptions{
				Namespace:     "",
				AllNamespaces: true,
				ByPod:         false,
				TableOutput:   false,
				TableStyle:    "colored",
				SortBy:        "image",
			},
			expectedError: false,
		},
		{
			name: "multiple flags",
			args: []string{"images", "--table", "--style", "box", "--sort", "image", "--namespace", "test"},
			expectedOpts: &ImagesOptions{
				Namespace:     "test",
				AllNamespaces: false,
				ByPod:         false,
				TableOutput:   true,
				TableStyle:    "box",
				SortBy:        "image",
			},
			expectedError: false,
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
			opts, err := ParseImagesArgs(tt.args)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tt.expectedError && tt.expectedOpts != nil {
				if opts.Namespace != tt.expectedOpts.Namespace {
					t.Errorf("Expected namespace %v, got %v", tt.expectedOpts.Namespace, opts.Namespace)
				}
				if opts.AllNamespaces != tt.expectedOpts.AllNamespaces {
					t.Errorf("Expected allNamespaces %v, got %v", tt.expectedOpts.AllNamespaces, opts.AllNamespaces)
				}
				if opts.ByPod != tt.expectedOpts.ByPod {
					t.Errorf("Expected byPod %v, got %v", tt.expectedOpts.ByPod, opts.ByPod)
				}
				if opts.TableOutput != tt.expectedOpts.TableOutput {
					t.Errorf("Expected tableOutput %v, got %v", tt.expectedOpts.TableOutput, opts.TableOutput)
				}
				if opts.TableStyle != tt.expectedOpts.TableStyle {
					t.Errorf("Expected tableStyle %v, got %v", tt.expectedOpts.TableStyle, opts.TableStyle)
				}
				if opts.SortBy != tt.expectedOpts.SortBy {
					t.Errorf("Expected sortBy %v, got %v", tt.expectedOpts.SortBy, opts.SortBy)
				}
			}
		})
	}
}

func TestImagesOptions(t *testing.T) {
	// Test the ImagesOptions struct
	opts := &ImagesOptions{
		Namespace:     "test",
		AllNamespaces: false,
		ByPod:         true,
		TableOutput:   false,
		TableStyle:    "simple",
		SortBy:        "image",
	}

	if opts.Namespace != "test" {
		t.Errorf("Expected Namespace to be 'test', got '%s'", opts.Namespace)
	}
	if opts.AllNamespaces != false {
		t.Errorf("Expected AllNamespaces to be false, got %v", opts.AllNamespaces)
	}
	if opts.ByPod != true {
		t.Errorf("Expected ByPod to be true, got %v", opts.ByPod)
	}
	if opts.TableOutput != false {
		t.Errorf("Expected TableOutput to be false, got %v", opts.TableOutput)
	}
	if opts.TableStyle != "simple" {
		t.Errorf("Expected TableStyle to be 'simple', got '%s'", opts.TableStyle)
	}
	if opts.SortBy != "image" {
		t.Errorf("Expected SortBy to be 'image', got '%s'", opts.SortBy)
	}
}

func TestValidationLogic(t *testing.T) {
	tests := []struct {
		name          string
		opts          *ImagesOptions
		expectedError bool
		errorContains string
	}{
		{
			name: "valid: namespace specified",
			opts: &ImagesOptions{
				Namespace:     "test",
				AllNamespaces: false,
			},
			expectedError: false,
		},
		{
			name: "valid: all namespaces",
			opts: &ImagesOptions{
				Namespace:     "",
				AllNamespaces: true,
			},
			expectedError: false,
		},
		{
			name: "invalid: conflicting namespace flags",
			opts: &ImagesOptions{
				Namespace:     "test",
				AllNamespaces: true,
			},
			expectedError: true,
			errorContains: "cannot use --namespace and --all-namespaces together",
		},
		{
			name: "invalid: conflicting table and by-pod flags",
			opts: &ImagesOptions{
				TableOutput: true,
				ByPod:       true,
			},
			expectedError: true,
			errorContains: "cannot use --table with --by-pod",
		},
		{
			name: "invalid: invalid sort option",
			opts: &ImagesOptions{
				SortBy: "invalid",
			},
			expectedError: true,
			errorContains: "invalid sort option",
		},
		{
			name: "valid: valid sort options",
			opts: &ImagesOptions{
				SortBy: "namespace",
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply the same validation logic used in ParseImagesArgs
			var hasError bool
			var errMsg string

			// Check namespace conflict
			if tt.opts.Namespace != "" && tt.opts.AllNamespaces {
				hasError = true
				errMsg = "cannot use --namespace and --all-namespaces together"
			}

			// Check table/by-pod conflict
			if tt.opts.TableOutput && tt.opts.ByPod {
				hasError = true
				errMsg = "cannot use --table with --by-pod (table output is only for unique images)"
			}

			// Check sort validation
			validSorts := map[string]bool{"namespace": true, "image": true, "none": true}
			if tt.opts.SortBy != "" && !validSorts[tt.opts.SortBy] {
				hasError = true
				errMsg = "invalid sort option"
			}

			if hasError != tt.expectedError {
				t.Errorf("Expected error=%v, got %v", tt.expectedError, hasError)
			}

			if tt.expectedError && tt.errorContains != "" && !strings.Contains(errMsg, tt.errorContains) {
				t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorContains, errMsg)
			}
		})
	}
}
