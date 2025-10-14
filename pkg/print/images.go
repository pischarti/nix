package print

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	corev1 "k8s.io/api/core/v1"
)

// PrintImagesTable prints images in a table format with namespace information
func PrintImagesTable(imagesSet map[string]struct{}, namespace string, allNamespaces bool, style string, sortBy string) {
	images := make([]string, 0, len(imagesSet))
	for img := range imagesSet {
		images = append(images, img)
	}

	// Sort images based on sortBy parameter
	switch sortBy {
	case "image":
		sort.Strings(images)
	case "namespace":
		// For namespace sort, we need to consider the actual namespace context
		// Since we're showing unique images, we'll sort by image name within namespace
		sort.Strings(images)
	case "none":
		// No sorting
	default:
		sort.Strings(images)
	}

	// Create table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Set table style based on parameter
	switch style {
	case "simple":
		t.SetStyle(table.StyleDefault)
	case "box":
		t.SetStyle(table.StyleDouble)
	case "rounded":
		t.SetStyle(table.StyleRounded)
	case "colored", "color":
		t.SetStyle(table.StyleColoredBright)
	default:
		t.SetStyle(table.StyleColoredBright)
	}

	// Add headers
	t.AppendHeader(table.Row{"NAMESPACE", "IMAGE"})

	// Determine namespace display
	nsDisplay := "all"
	if !allNamespaces && namespace != "" {
		nsDisplay = namespace
	}

	// Add rows
	for _, img := range images {
		t.AppendRow(table.Row{nsDisplay, img})
	}

	// Render table
	t.Render()
}

// ImageNamespace represents an image with its namespace
type ImageNamespace struct {
	Image     string
	Namespace string
}

// PrintImagesTableWithNamespaces prints images in a table format showing actual namespace values
func PrintImagesTableWithNamespaces(imageNamespaceMap map[string]string, style string, sortBy string) {
	// Convert map to slice of structs for sorting
	var imageNsList []ImageNamespace
	for img, ns := range imageNamespaceMap {
		imageNsList = append(imageNsList, ImageNamespace{Image: img, Namespace: ns})
	}

	// Sort based on sortBy parameter
	switch sortBy {
	case "image":
		sort.Slice(imageNsList, func(i, j int) bool {
			return imageNsList[i].Image < imageNsList[j].Image
		})
	case "namespace":
		sort.Slice(imageNsList, func(i, j int) bool {
			if imageNsList[i].Namespace == imageNsList[j].Namespace {
				return imageNsList[i].Image < imageNsList[j].Image
			}
			return imageNsList[i].Namespace < imageNsList[j].Namespace
		})
	case "none":
		// No sorting
	default:
		sort.Slice(imageNsList, func(i, j int) bool {
			if imageNsList[i].Namespace == imageNsList[j].Namespace {
				return imageNsList[i].Image < imageNsList[j].Image
			}
			return imageNsList[i].Namespace < imageNsList[j].Namespace
		})
	}

	// Create table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Set table style based on parameter
	switch style {
	case "simple":
		t.SetStyle(table.StyleDefault)
	case "box":
		t.SetStyle(table.StyleDouble)
	case "rounded":
		t.SetStyle(table.StyleRounded)
	case "colored", "color":
		t.SetStyle(table.StyleColoredBright)
	default:
		t.SetStyle(table.StyleColoredBright)
	}

	// Add headers
	t.AppendHeader(table.Row{"NAMESPACE", "IMAGE"})

	// Add rows with actual namespace values
	for _, item := range imageNsList {
		t.AppendRow(table.Row{item.Namespace, item.Image})
	}

	// Render table
	t.Render()
}

// PrintImagesList prints images in a simple list format
func PrintImagesList(imagesSet map[string]struct{}, sortBy string) {
	images := make([]string, 0, len(imagesSet))
	for img := range imagesSet {
		images = append(images, img)
	}

	// Sort images based on sortBy parameter
	switch sortBy {
	case "image":
		sort.Strings(images)
	case "namespace":
		// For namespace sort in list mode, sort by image name (alphabetical)
		sort.Strings(images)
	case "none":
		// No sorting - keep original order
	default:
		sort.Strings(images)
	}

	for _, img := range images {
		fmt.Println(img)
	}
}

// PrintImagesHelp prints the help information for the images command
func PrintImagesHelp() {
	fmt.Println("Usage: kube images [--namespace NAMESPACE | --all-namespaces] [--by-pod] [--table] [--style STYLE] [--sort SORT]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --namespace, -n    Query a specific namespace")
	fmt.Println("  --all-namespaces, -A  Query across all namespaces (default)")
	fmt.Println("  --by-pod          Show images grouped by pod")
	fmt.Println("  --table, -t       Display output in table format")
	fmt.Println("  --style           Table style: simple, box, rounded, colored (default)")
	fmt.Println("  --sort            Sort order: namespace (default), image, none")
	fmt.Println("  --help, -h        Show this help message")
}

// ServiceInfo represents a service with its key information
type ServiceInfo struct {
	Namespace   string
	Name        string
	Type        string
	Annotations []string
}

// PrintServicesTable prints services in a table format
func PrintServicesTable(services []corev1.Service, style string, sortBy string) {
	// Convert services to ServiceInfo structs
	var serviceInfos []ServiceInfo
	for _, service := range services {
		var allAnnotations []string
		for key, value := range service.Annotations {
			// Exclude last-applied-configuration annotations
			if !strings.Contains(strings.ToLower(key), "last-applied-configuration") {
				allAnnotations = append(allAnnotations, fmt.Sprintf("%s=%s", key, value))
			}
		}

		serviceInfos = append(serviceInfos, ServiceInfo{
			Namespace:   service.Namespace,
			Name:        service.Name,
			Type:        string(service.Spec.Type),
			Annotations: allAnnotations,
		})
	}

	// Sort services based on sortBy parameter
	switch sortBy {
	case "name":
		sort.Slice(serviceInfos, func(i, j int) bool {
			return serviceInfos[i].Name < serviceInfos[j].Name
		})
	case "namespace":
		sort.Slice(serviceInfos, func(i, j int) bool {
			if serviceInfos[i].Namespace == serviceInfos[j].Namespace {
				return serviceInfos[i].Name < serviceInfos[j].Name
			}
			return serviceInfos[i].Namespace < serviceInfos[j].Namespace
		})
	case "none":
		// No sorting
	default:
		sort.Slice(serviceInfos, func(i, j int) bool {
			if serviceInfos[i].Namespace == serviceInfos[j].Namespace {
				return serviceInfos[i].Name < serviceInfos[j].Name
			}
			return serviceInfos[i].Namespace < serviceInfos[j].Namespace
		})
	}

	// Create table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Set table style based on parameter
	switch style {
	case "simple":
		t.SetStyle(table.StyleDefault)
	case "box":
		t.SetStyle(table.StyleDouble)
	case "rounded":
		t.SetStyle(table.StyleRounded)
	case "colored", "color":
		t.SetStyle(table.StyleColoredBright)
	default:
		t.SetStyle(table.StyleColoredBright)
	}

	// Add headers
	t.AppendHeader(table.Row{"NAMESPACE", "NAME", "TYPE", "ANNOTATIONS"})

	// Add rows
	for _, info := range serviceInfos {
		if len(info.Annotations) == 0 {
			t.AppendRow(table.Row{info.Namespace, info.Name, info.Type, "-"})
		} else {
			for i, annotation := range info.Annotations {
				if i == 0 {
					// First annotation includes namespace, name, and type
					t.AppendRow(table.Row{info.Namespace, info.Name, info.Type, annotation})
				} else {
					// Subsequent annotations have empty cells for namespace, name, type
					t.AppendRow(table.Row{"", "", "", annotation})
				}
			}
		}
	}

	// Render table
	t.Render()
}

// PrintServicesList prints services in a simple list format
func PrintServicesList(services []corev1.Service, sortBy string) {
	// Convert services to ServiceInfo structs for sorting
	var serviceInfos []ServiceInfo
	for _, service := range services {
		var allAnnotations []string
		for key, value := range service.Annotations {
			// Exclude last-applied-configuration annotations
			if !strings.Contains(strings.ToLower(key), "last-applied-configuration") {
				allAnnotations = append(allAnnotations, fmt.Sprintf("%s=%s", key, value))
			}
		}

		serviceInfos = append(serviceInfos, ServiceInfo{
			Namespace:   service.Namespace,
			Name:        service.Name,
			Type:        string(service.Spec.Type),
			Annotations: allAnnotations,
		})
	}

	// Sort services based on sortBy parameter
	switch sortBy {
	case "name":
		sort.Slice(serviceInfos, func(i, j int) bool {
			return serviceInfos[i].Name < serviceInfos[j].Name
		})
	case "namespace":
		sort.Slice(serviceInfos, func(i, j int) bool {
			if serviceInfos[i].Namespace == serviceInfos[j].Namespace {
				return serviceInfos[i].Name < serviceInfos[j].Name
			}
			return serviceInfos[i].Namespace < serviceInfos[j].Namespace
		})
	case "none":
		// No sorting - keep original order
	default:
		sort.Slice(serviceInfos, func(i, j int) bool {
			if serviceInfos[i].Namespace == serviceInfos[j].Namespace {
				return serviceInfos[i].Name < serviceInfos[j].Name
			}
			return serviceInfos[i].Namespace < serviceInfos[j].Namespace
		})
	}

	// Print services
	for _, info := range serviceInfos {
		if len(info.Annotations) == 0 {
			fmt.Printf("%s/%s (%s): -\n", info.Namespace, info.Name, info.Type)
		} else {
			fmt.Printf("%s/%s (%s):\n", info.Namespace, info.Name, info.Type)
			for _, annotation := range info.Annotations {
				fmt.Printf("  %s\n", annotation)
			}
		}
	}
}

// PrintServicesHelp prints the help information for the services command
func PrintServicesHelp() {
	fmt.Println("Usage: kube services [--namespace NAMESPACE | --all-namespaces] [--table] [--style STYLE] [--sort SORT] [--annotation-value VALUE]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --namespace, -n    Query a specific namespace")
	fmt.Println("  --all-namespaces, -A  Query across all namespaces (default)")
	fmt.Println("  --table, -t       Display output in table format")
	fmt.Println("  --style           Table style: simple, box, rounded, colored (default)")
	fmt.Println("  --sort            Sort order: namespace (default), name, none")
	fmt.Println("  --annotation-value  Filter by annotation key or value containing this text (case-insensitive)")
	fmt.Println("  --help, -h        Show this help message")
	fmt.Println()
	fmt.Println("Note: last-applied-configuration annotations are automatically excluded from output.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ./kube services                                    # Show all services with annotations")
	fmt.Println("  ./kube services --annotation-value aws-load-balancer  # Filter by annotation containing 'aws-load-balancer'")
	fmt.Println("  ./kube services --annotation-value nlb             # Filter by annotation containing 'nlb'")
}
