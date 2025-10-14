package container

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gofr.dev/pkg/gofr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/pischarti/nix/pkg/config"
	"github.com/pischarti/nix/pkg/print"
)

// ImagesOptions represents the parsed command line options for the images command
type ImagesOptions struct {
	Namespace     string
	AllNamespaces bool
	ByPod         bool
	TableOutput   bool
	TableStyle    string
	SortBy        string
}

// ParseImagesArgs parses command line arguments for the images command
func ParseImagesArgs(args []string) (*ImagesOptions, error) {
	opts := &ImagesOptions{
		TableStyle: "colored",
		SortBy:     "namespace",
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--namespace", "-n":
			if i+1 < len(args) {
				i++
				opts.Namespace = args[i]
			}
		case "--all-namespaces", "-A":
			opts.AllNamespaces = true
		case "--by-pod":
			opts.ByPod = true
		case "--table", "-t":
			opts.TableOutput = true
		case "--style":
			if i+1 < len(args) {
				i++
				opts.TableStyle = args[i]
			}
		case "--sort":
			if i+1 < len(args) {
				i++
				opts.SortBy = args[i]
			}
		}
	}

	// Apply defaults
	if opts.Namespace == "" && !opts.AllNamespaces {
		opts.AllNamespaces = true
	}

	// Validate options
	if opts.Namespace != "" && opts.AllNamespaces {
		return nil, fmt.Errorf("cannot use --namespace and --all-namespaces together")
	}
	if opts.TableOutput && opts.ByPod {
		return nil, fmt.Errorf("cannot use --table with --by-pod (table output is only for unique images)")
	}

	// Validate sort option
	validSorts := map[string]bool{"namespace": true, "image": true, "none": true}
	if !validSorts[opts.SortBy] {
		return nil, fmt.Errorf("invalid sort option '%s'. Valid options: namespace, image, none", opts.SortBy)
	}

	return opts, nil
}

// ImagesHandler handles the images command
func ImagesHandler(ctx *gofr.Context) (any, error) {
	args := os.Args[1:] // Get command line args for parsing flags

	// Check for help flag first
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			print.PrintImagesHelp()
			return nil, nil
		}
	}

	// Parse arguments
	opts, err := ParseImagesArgs(args)
	if err != nil {
		return nil, err
	}

	// Get Kubernetes client
	cfg, err := config.GetKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	// Determine namespace for query
	ns := opts.Namespace
	if opts.AllNamespaces {
		ns = metav1.NamespaceAll
	}

	// List pods
	pods, err := clientset.CoreV1().Pods(ns).List(ctx.Context, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}

	// Handle different output modes
	if opts.ByPod {
		return handleByPodOutput(pods, opts)
	}

	if opts.TableOutput && opts.AllNamespaces {
		return handleTableWithNamespacesOutput(pods, opts)
	}

	return handleStandardOutput(pods, opts)
}

// handleByPodOutput handles the --by-pod output format
func handleByPodOutput(pods *corev1.PodList, opts *ImagesOptions) (any, error) {
	// Sort pods by namespace then name
	sort.Slice(pods.Items, func(i, j int) bool {
		a := pods.Items[i]
		b := pods.Items[j]
		if a.Namespace == b.Namespace {
			return a.Name < b.Name
		}
		return a.Namespace < b.Namespace
	})

	for _, pod := range pods.Items {
		var images []string

		// Collect container images
		for _, c := range pod.Spec.Containers {
			if c.Image != "" {
				images = append(images, c.Image)
			}
		}
		for _, c := range pod.Spec.InitContainers {
			if c.Image != "" {
				images = append(images, c.Image)
			}
		}
		for _, c := range pod.Spec.EphemeralContainers {
			if c.Image != "" {
				images = append(images, c.Image)
			}
		}

		if len(images) == 0 {
			continue
		}

		// Remove duplicates
		seen := map[string]struct{}{}
		uniq := make([]string, 0, len(images))
		for _, img := range images {
			if _, ok := seen[img]; ok {
				continue
			}
			seen[img] = struct{}{}
			uniq = append(uniq, img)
		}

		fmt.Printf("%s/%s: %s\n", pod.Namespace, pod.Name, strings.Join(uniq, ", "))
	}

	return nil, nil
}

// handleTableWithNamespacesOutput handles table output with namespace information
func handleTableWithNamespacesOutput(pods *corev1.PodList, opts *ImagesOptions) (any, error) {
	imageNamespaceMap := make(map[string]string)

	for _, pod := range pods.Items {
		// Collect container images with their namespaces
		for _, c := range pod.Spec.Containers {
			if c.Image != "" {
				imageNamespaceMap[c.Image] = pod.Namespace
			}
		}
		for _, c := range pod.Spec.InitContainers {
			if c.Image != "" {
				imageNamespaceMap[c.Image] = pod.Namespace
			}
		}
		for _, c := range pod.Spec.EphemeralContainers {
			if c.Image != "" {
				imageNamespaceMap[c.Image] = pod.Namespace
			}
		}
	}

	print.PrintImagesTableWithNamespaces(imageNamespaceMap, opts.TableStyle, opts.SortBy)
	return nil, nil
}

// handleStandardOutput handles standard list or table output
func handleStandardOutput(pods *corev1.PodList, opts *ImagesOptions) (any, error) {
	imagesSet := map[string]struct{}{}

	for _, pod := range pods.Items {
		// Collect unique images
		for _, c := range pod.Spec.Containers {
			if c.Image != "" {
				imagesSet[c.Image] = struct{}{}
			}
		}
		for _, c := range pod.Spec.InitContainers {
			if c.Image != "" {
				imagesSet[c.Image] = struct{}{}
			}
		}
		for _, c := range pod.Spec.EphemeralContainers {
			if c.Image != "" {
				imagesSet[c.Image] = struct{}{}
			}
		}
	}

	// Output based on format
	if opts.TableOutput {
		print.PrintImagesTable(imagesSet, opts.Namespace, opts.AllNamespaces, opts.TableStyle, opts.SortBy)
	} else {
		print.PrintImagesList(imagesSet, opts.SortBy)
	}

	return nil, nil
}

// ServicesOptions represents the parsed command line options for the services command
type ServicesOptions struct {
	Namespace       string
	AllNamespaces   bool
	TableOutput     bool
	TableStyle      string
	SortBy          string
	AnnotationValue string
}

// ParseServicesArgs parses command line arguments for the services command
func ParseServicesArgs(args []string) (*ServicesOptions, error) {
	opts := &ServicesOptions{
		TableStyle: "colored",
		SortBy:     "namespace",
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--namespace", "-n":
			if i+1 < len(args) {
				i++
				opts.Namespace = args[i]
			}
		case "--all-namespaces", "-A":
			opts.AllNamespaces = true
		case "--table", "-t":
			opts.TableOutput = true
		case "--style":
			if i+1 < len(args) {
				i++
				opts.TableStyle = args[i]
			}
		case "--sort":
			if i+1 < len(args) {
				i++
				opts.SortBy = args[i]
			}
		case "--annotation-value":
			if i+1 < len(args) {
				i++
				opts.AnnotationValue = args[i]
			}
		}
	}

	// Apply defaults
	if opts.Namespace == "" && !opts.AllNamespaces {
		opts.AllNamespaces = true
	}

	// Validate options
	if opts.Namespace != "" && opts.AllNamespaces {
		return nil, fmt.Errorf("cannot use --namespace and --all-namespaces together")
	}

	// Validate sort option
	validSorts := map[string]bool{"namespace": true, "name": true, "none": true}
	if !validSorts[opts.SortBy] {
		return nil, fmt.Errorf("invalid sort option '%s'. Valid options: namespace, name, none", opts.SortBy)
	}

	return opts, nil
}

// ServicesHandler handles the services command
func ServicesHandler(ctx *gofr.Context) (any, error) {
	args := os.Args[1:] // Get command line args for parsing flags

	// Check for help flag first
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			print.PrintServicesHelp()
			return nil, nil
		}
	}

	// Parse arguments
	opts, err := ParseServicesArgs(args)
	if err != nil {
		return nil, err
	}

	// Get Kubernetes client
	cfg, err := config.GetKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	// Determine namespace for query
	ns := opts.Namespace
	if opts.AllNamespaces {
		ns = metav1.NamespaceAll
	}

	// List services
	services, err := clientset.CoreV1().Services(ns).List(ctx.Context, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}

	// Filter services with matching annotations
	var filteredServices []corev1.Service
	for _, service := range services.Items {
		if hasMatchingAnnotation(service, opts.AnnotationValue) {
			filteredServices = append(filteredServices, service)
		}
	}

	// Handle output
	if opts.TableOutput {
		print.PrintServicesTable(filteredServices, opts.TableStyle, opts.SortBy)
	} else {
		print.PrintServicesList(filteredServices, opts.SortBy)
	}

	return nil, nil
}

// hasMatchingAnnotation checks if a service has any annotation matching the specified value
// If annotationValue is empty, returns true if service has any annotations
// If annotationValue is provided, checks if any annotation key or value contains the specified value
func hasMatchingAnnotation(service corev1.Service, annotationValue string) bool {
	// If no specific annotation value is requested, return true if service has any annotations
	if annotationValue == "" {
		return len(service.Annotations) > 0
	}

	// Search for the annotation value in both keys and values (case-insensitive)
	searchValue := strings.ToLower(annotationValue)
	for key, value := range service.Annotations {
		if strings.Contains(strings.ToLower(key), searchValue) ||
			strings.Contains(strings.ToLower(value), searchValue) {
			return true
		}
	}
	return false
}
