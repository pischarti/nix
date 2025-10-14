package k8s

// import (
// 	"context"
// 	"fmt"
// 	"path/filepath"

// 	"github.com/spf13/viper"
// 	corev1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/client-go/kubernetes"
// 	"k8s.io/client-go/tools/clientcmd"
// 	"k8s.io/client-go/util/homedir"
// )

// // Client wraps a Kubernetes clientset with additional functionality
// type Client struct {
// 	Clientset *kubernetes.Clientset
// }

// // NewClient creates a new Kubernetes client from the configured kubeconfig
// func NewClient() (*Client, error) {
// 	// Try to get kubeconfig from viper (config file/flags), then default
// 	kubeconfig := viper.GetString("kubeconfig")
// 	if kubeconfig == "" {
// 		if home := homedir.HomeDir(); home != "" {
// 			kubeconfig = filepath.Join(home, ".kube", "config")
// 		}
// 	}

// 	// Build config from kubeconfig file
// 	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
// 	}

// 	// Create clientset
// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
// 	}

// 	return &Client{Clientset: clientset}, nil
// }

// // EventQueryOptions contains options for querying events
// type EventQueryOptions struct {
// 	Namespace string
// }

// // QueryEvents retrieves Kubernetes events based on the provided options
// func (c *Client) QueryEvents(ctx context.Context, opts EventQueryOptions) ([]corev1.Event, error) {
// 	eventList, err := c.Clientset.CoreV1().Events(opts.Namespace).List(ctx, metav1.ListOptions{})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to list events: %w", err)
// 	}

// 	return eventList.Items, nil
// }
