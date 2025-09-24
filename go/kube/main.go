package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"gofr.dev/pkg/gofr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	app := gofr.NewCMD()

	app.SubCommand("images", imagesHandler,
		gofr.AddDescription("List container images running in the cluster"),
		gofr.AddHelp("Usage: kube images [--namespace NAMESPACE | --all-namespaces] [--by-pod] [--table] [--style STYLE] [--sort SORT]"),
	)

	app.Run()
}

func imagesHandler(ctx *gofr.Context) (any, error) {
	args := os.Args[1:] // Get command line args for parsing flags
	// Simple arg parsing to support flags without relying on stdlib FlagSet,
	// so we keep full control inside GoFr's subcommand function.
	namespace := ""
	allNamespaces := false
	byPod := false
	tableOutput := false
	tableStyle := "colored"
	sortBy := "namespace" // default sort by namespace
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--namespace", "-n":
			if i+1 < len(args) {
				i++
				namespace = args[i]
			}
		case "--all-namespaces", "-A":
			allNamespaces = true
		case "--by-pod":
			byPod = true
		case "--table", "-t":
			tableOutput = true
		case "--style":
			if i+1 < len(args) {
				i++
				tableStyle = args[i]
			}
		case "--sort":
			if i+1 < len(args) {
				i++
				sortBy = args[i]
			}
		case "-h", "--help":
			printImagesHelp()
			return nil, nil
		}
	}

	if namespace != "" && allNamespaces {
		return nil, fmt.Errorf("cannot use --namespace and --all-namespaces together")
	}
	if namespace == "" && !allNamespaces {
		allNamespaces = true
	}
	if tableOutput && byPod {
		return nil, fmt.Errorf("cannot use --table with --by-pod (table output is only for unique images)")
	}

	// Validate sort option
	validSorts := map[string]bool{"namespace": true, "image": true, "none": true}
	if !validSorts[sortBy] {
		return nil, fmt.Errorf("invalid sort option '%s'. Valid options: namespace, image, none", sortBy)
	}

	cfg, err := getKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	ns := namespace
	if allNamespaces {
		ns = metav1.NamespaceAll
	}

	pods, err := clientset.CoreV1().Pods(ns).List(ctx.Context, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}

	if byPod {
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

	// For table output with all namespaces, we need to track images with their namespaces
	if tableOutput && allNamespaces {
		imageNamespaceMap := make(map[string]string)
		for _, pod := range pods.Items {
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
		printImagesTableWithNamespaces(imageNamespaceMap, tableStyle, sortBy)
	} else {
		// For single namespace or list output, use the original logic
		imagesSet := map[string]struct{}{}
		for _, pod := range pods.Items {
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

		if tableOutput {
			printImagesTable(imagesSet, namespace, allNamespaces, tableStyle, sortBy)
		} else {
			printImagesList(imagesSet, sortBy)
		}
	}
	return nil, nil
}

func printImagesTable(imagesSet map[string]struct{}, namespace string, allNamespaces bool, style string, sortBy string) {
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

func printImagesTableWithNamespaces(imageNamespaceMap map[string]string, style string, sortBy string) {
	// Convert map to slice of structs for sorting
	type imageNamespace struct {
		image     string
		namespace string
	}

	var imageNsList []imageNamespace
	for img, ns := range imageNamespaceMap {
		imageNsList = append(imageNsList, imageNamespace{image: img, namespace: ns})
	}

	// Sort based on sortBy parameter
	switch sortBy {
	case "image":
		sort.Slice(imageNsList, func(i, j int) bool {
			return imageNsList[i].image < imageNsList[j].image
		})
	case "namespace":
		sort.Slice(imageNsList, func(i, j int) bool {
			if imageNsList[i].namespace == imageNsList[j].namespace {
				return imageNsList[i].image < imageNsList[j].image
			}
			return imageNsList[i].namespace < imageNsList[j].namespace
		})
	case "none":
		// No sorting
	default:
		sort.Slice(imageNsList, func(i, j int) bool {
			if imageNsList[i].namespace == imageNsList[j].namespace {
				return imageNsList[i].image < imageNsList[j].image
			}
			return imageNsList[i].namespace < imageNsList[j].namespace
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
		t.AppendRow(table.Row{item.namespace, item.image})
	}

	// Render table
	t.Render()
}

func printImagesList(imagesSet map[string]struct{}, sortBy string) {
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

func printImagesHelp() {
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

func getKubeConfig() (*rest.Config, error) {
	// Try in-cluster first
	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	}

	// Fall back to kubeconfig from env or default path
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}
	kubeconfigPath := filepath.Join(home, ".kube", "config")
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}
