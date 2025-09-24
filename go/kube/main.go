package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gofr.dev/pkg/gofr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/pischarti/nix/go/pkg/config"
	"github.com/pischarti/nix/go/pkg/print"
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
			print.PrintImagesHelp()
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

	cfg, err := config.GetKubeConfig()
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
		print.PrintImagesTableWithNamespaces(imageNamespaceMap, tableStyle, sortBy)
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
			print.PrintImagesTable(imagesSet, namespace, allNamespaces, tableStyle, sortBy)
		} else {
			print.PrintImagesList(imagesSet, sortBy)
		}
	}
	return nil, nil
}
