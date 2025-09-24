package print

import (
	"fmt"
	"os"
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
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
