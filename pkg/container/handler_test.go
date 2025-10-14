package container

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestParseServicesArgs(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedOpts  *ServicesOptions
		expectedError bool
	}{
		{
			name: "default values",
			args: []string{"services"},
			expectedOpts: &ServicesOptions{
				Namespace:       "",
				AllNamespaces:   true,
				TableOutput:     false,
				TableStyle:      "colored",
				SortBy:          "namespace",
				AnnotationValue: "",
			},
			expectedError: false,
		},
		{
			name: "namespace flag",
			args: []string{"services", "--namespace", "test"},
			expectedOpts: &ServicesOptions{
				Namespace:       "test",
				AllNamespaces:   false,
				TableOutput:     false,
				TableStyle:      "colored",
				SortBy:          "namespace",
				AnnotationValue: "",
			},
			expectedError: false,
		},
		{
			name: "all namespaces flag",
			args: []string{"services", "--all-namespaces"},
			expectedOpts: &ServicesOptions{
				Namespace:       "",
				AllNamespaces:   true,
				TableOutput:     false,
				TableStyle:      "colored",
				SortBy:          "namespace",
				AnnotationValue: "",
			},
			expectedError: false,
		},
		{
			name: "table flag",
			args: []string{"services", "--table"},
			expectedOpts: &ServicesOptions{
				Namespace:       "",
				AllNamespaces:   true,
				TableOutput:     true,
				TableStyle:      "colored",
				SortBy:          "namespace",
				AnnotationValue: "",
			},
			expectedError: false,
		},
		{
			name: "style flag",
			args: []string{"services", "--style", "simple"},
			expectedOpts: &ServicesOptions{
				Namespace:       "",
				AllNamespaces:   true,
				TableOutput:     false,
				TableStyle:      "simple",
				SortBy:          "namespace",
				AnnotationValue: "",
			},
			expectedError: false,
		},
		{
			name: "sort flag",
			args: []string{"services", "--sort", "name"},
			expectedOpts: &ServicesOptions{
				Namespace:       "",
				AllNamespaces:   true,
				TableOutput:     false,
				TableStyle:      "colored",
				SortBy:          "name",
				AnnotationValue: "",
			},
			expectedError: false,
		},
		{
			name: "multiple flags",
			args: []string{"services", "--table", "--style", "box", "--sort", "name", "--namespace", "test"},
			expectedOpts: &ServicesOptions{
				Namespace:       "test",
				AllNamespaces:   false,
				TableOutput:     true,
				TableStyle:      "box",
				SortBy:          "name",
				AnnotationValue: "",
			},
			expectedError: false,
		},
		{
			name: "annotation value flag",
			args: []string{"services", "--annotation-value", "nlb"},
			expectedOpts: &ServicesOptions{
				Namespace:       "",
				AllNamespaces:   true,
				TableOutput:     false,
				TableStyle:      "colored",
				SortBy:          "namespace",
				AnnotationValue: "nlb",
			},
			expectedError: false,
		},
		{
			name: "annotation value with other flags",
			args: []string{"services", "--annotation-value", "internet-facing", "--table", "--sort", "name"},
			expectedOpts: &ServicesOptions{
				Namespace:       "",
				AllNamespaces:   true,
				TableOutput:     true,
				TableStyle:      "colored",
				SortBy:          "name",
				AnnotationValue: "internet-facing",
			},
			expectedError: false,
		},
		{
			name:          "invalid sort option",
			args:          []string{"services", "--sort", "invalid"},
			expectedError: true,
		},
		{
			name:          "conflicting namespace flags",
			args:          []string{"services", "--namespace", "test", "--all-namespaces"},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := ParseServicesArgs(tt.args)

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
				if opts.TableOutput != tt.expectedOpts.TableOutput {
					t.Errorf("Expected tableOutput %v, got %v", tt.expectedOpts.TableOutput, opts.TableOutput)
				}
				if opts.TableStyle != tt.expectedOpts.TableStyle {
					t.Errorf("Expected tableStyle %v, got %v", tt.expectedOpts.TableStyle, opts.TableStyle)
				}
				if opts.SortBy != tt.expectedOpts.SortBy {
					t.Errorf("Expected sortBy %v, got %v", tt.expectedOpts.SortBy, opts.SortBy)
				}
				if opts.AnnotationValue != tt.expectedOpts.AnnotationValue {
					t.Errorf("Expected annotationValue %v, got %v", tt.expectedOpts.AnnotationValue, opts.AnnotationValue)
				}
			}
		})
	}
}

func TestServicesOptions(t *testing.T) {
	// Test the ServicesOptions struct
	opts := &ServicesOptions{
		Namespace:       "test",
		AllNamespaces:   false,
		TableOutput:     true,
		TableStyle:      "simple",
		SortBy:          "name",
		AnnotationValue: "nlb",
	}

	if opts.Namespace != "test" {
		t.Errorf("Expected Namespace to be 'test', got '%s'", opts.Namespace)
	}
	if opts.AllNamespaces != false {
		t.Errorf("Expected AllNamespaces to be false, got %v", opts.AllNamespaces)
	}
	if opts.TableOutput != true {
		t.Errorf("Expected TableOutput to be true, got %v", opts.TableOutput)
	}
	if opts.TableStyle != "simple" {
		t.Errorf("Expected TableStyle to be 'simple', got '%s'", opts.TableStyle)
	}
	if opts.SortBy != "name" {
		t.Errorf("Expected SortBy to be 'name', got '%s'", opts.SortBy)
	}
	if opts.AnnotationValue != "nlb" {
		t.Errorf("Expected AnnotationValue to be 'nlb', got '%s'", opts.AnnotationValue)
	}
}

func TestHasMatchingAnnotation(t *testing.T) {
	tests := []struct {
		name            string
		service         corev1.Service
		annotationValue string
		expected        bool
	}{
		{
			name: "service with annotations (no filter)",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"service.beta.kubernetes.io/aws-load-balancer-type": "nlb",
					},
				},
			},
			annotationValue: "",
			expected:        true,
		},
		{
			name: "service with multiple annotations (no filter)",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"custom-annotation":  "some-value",
						"another-annotation": "another-value",
					},
				},
			},
			annotationValue: "",
			expected:        true,
		},
		{
			name: "service with no annotations (no filter)",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			annotationValue: "",
			expected:        false,
		},
		{
			name: "service with matching annotation key",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"service.beta.kubernetes.io/aws-load-balancer-type": "nlb",
					},
				},
			},
			annotationValue: "aws-load-balancer",
			expected:        true,
		},
		{
			name: "service with matching annotation value",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"custom-annotation": "nlb",
					},
				},
			},
			annotationValue: "nlb",
			expected:        true,
		},
		{
			name: "service with non-matching annotations",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.io/service-name": "my-service",
						"custom-annotation":          "some-value",
					},
				},
			},
			annotationValue: "aws-load-balancer",
			expected:        false,
		},
		{
			name: "service with case-insensitive matching annotation key",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"SERVICE.BETA.KUBERNETES.IO/AWS-LOAD-BALANCER-TYPE": "nlb",
					},
				},
			},
			annotationValue: "aws-load-balancer",
			expected:        true,
		},
		{
			name: "service with case-insensitive matching annotation value",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"custom-annotation": "INTERNET-FACING",
					},
				},
			},
			annotationValue: "internet-facing",
			expected:        true,
		},
		{
			name: "service with partial matching annotation value",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"custom-annotation": "http-protocol",
					},
				},
			},
			annotationValue: "http",
			expected:        true,
		},
		{
			name: "service with matching in both key and value",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"aws-load-balancer-type": "nlb",
						"custom-annotation":      "some-value",
					},
				},
			},
			annotationValue: "aws-load-balancer",
			expected:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasMatchingAnnotation(tt.service, tt.annotationValue)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestServicesValidationLogic(t *testing.T) {
	tests := []struct {
		name          string
		opts          *ServicesOptions
		expectedError bool
		errorContains string
	}{
		{
			name: "valid: namespace specified",
			opts: &ServicesOptions{
				Namespace:     "test",
				AllNamespaces: false,
			},
			expectedError: false,
		},
		{
			name: "valid: all namespaces",
			opts: &ServicesOptions{
				Namespace:     "",
				AllNamespaces: true,
			},
			expectedError: false,
		},
		{
			name: "invalid: conflicting namespace flags",
			opts: &ServicesOptions{
				Namespace:     "test",
				AllNamespaces: true,
			},
			expectedError: true,
			errorContains: "cannot use --namespace and --all-namespaces together",
		},
		{
			name: "invalid: invalid sort option",
			opts: &ServicesOptions{
				SortBy: "invalid",
			},
			expectedError: true,
			errorContains: "invalid sort option",
		},
		{
			name: "valid: valid sort options",
			opts: &ServicesOptions{
				SortBy: "namespace",
			},
			expectedError: false,
		},
		{
			name: "valid: name sort option",
			opts: &ServicesOptions{
				SortBy: "name",
			},
			expectedError: false,
		},
		{
			name: "valid: none sort option",
			opts: &ServicesOptions{
				SortBy: "none",
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply the same validation logic used in ParseServicesArgs
			var hasError bool
			var errMsg string

			// Check namespace conflict
			if tt.opts.Namespace != "" && tt.opts.AllNamespaces {
				hasError = true
				errMsg = "cannot use --namespace and --all-namespaces together"
			}

			// Check sort validation
			validSorts := map[string]bool{"namespace": true, "name": true, "none": true}
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
