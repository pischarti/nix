package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
		gofr.AddHelp("Usage: kube images [--namespace NAMESPACE | --all-namespaces] [--by-pod]"),
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

	images := make([]string, 0, len(imagesSet))
	for img := range imagesSet {
		images = append(images, img)
	}
	sort.Strings(images)
	for _, img := range images {
		fmt.Println(img)
	}
	return nil, nil
}

func printImagesHelp() {
	fmt.Println("Usage: kube images [--namespace NAMESPACE | --all-namespaces] [--by-pod]")
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
