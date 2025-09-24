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

	"github.com/pischarti/nix/go/pkg/config"
	"github.com/pischarti/nix/go/pkg/print"
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
