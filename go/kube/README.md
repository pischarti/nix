# kube CLI

A Kubernetes command-line tool built with GoFr that provides utilities for managing and inspecting Kubernetes clusters.

## Features

### Images Subcommand

List all container images running in your Kubernetes cluster with various filtering and display options.

#### Usage

```bash
# List unique images across all namespaces (default)
./kube images

# List images in a specific namespace
./kube images --namespace default

# List images across all namespaces (explicit)
./kube images --all-namespaces

# Group images by pod instead of showing unique list
./kube images --by-pod

# Display output in table format
./kube images --table

# Display with different table styles
./kube images --table --style simple
./kube images --table --style box
./kube images --table --style rounded
./kube images --table --style colored

# Sort output by different criteria
./kube images --sort namespace    # Default: sort by namespace
./kube images --sort image        # Sort by image name alphabetically
./kube images --sort none         # No sorting (original order)

# Show help
./kube images --help
```

#### Options

- `--namespace, -n`: Query a specific namespace (default: all namespaces)
- `--all-namespaces, -A`: Query across all namespaces (default behavior)
- `--by-pod`: Show images grouped by pod instead of unique list
- `--table, -t`: Display output in table format with namespace and image columns (cannot be used with --by-pod). Shows actual namespace names when using --all-namespaces.
- `--style`: Table style - `simple`, `box`, `rounded`, or `colored` (default: colored)
- `--sort`: Sort order - `namespace` (default), `image`, or `none`
- `--help, -h`: Show help information

#### Examples

```bash
# Get unique images across the entire cluster
./kube images --all-namespaces

# List images in the kube-system namespace
./kube images --namespace kube-system

# See which images each pod is using
./kube images --all-namespaces --by-pod

# Display images in table format
./kube images --table --namespace default

# Different table styles
./kube images --table --style simple
./kube images --table --style box
./kube images --table --style rounded

# Combine sorting with other options
./kube images --table --sort image --style box
./kube images --sort image --namespace default

# Output format when using --by-pod:
# namespace/pod-name: image1, image2, image3
# default/my-app: nginx:1.21, busybox:1.34

# Output format when using --table (colored style):
# ┌───────────┬─────────────────┐
# │ NAMESPACE │ IMAGE           │
# ├───────────┼─────────────────┤
# │ default   │ nginx:1.21      │
# │ default   │ busybox:1.34    │
# │ kube-system │ kube-proxy:1.28 │
# │ monitoring │ prometheus:2.40 │
# └───────────┴─────────────────┘
```

## Building

```bash
# Build the CLI
go build -o kube .

# Run with help
./kube --help
```

## Requirements

- Go 1.21+
- Access to a Kubernetes cluster (via kubeconfig or in-cluster config)
- Appropriate RBAC permissions to list pods

## Configuration

The CLI automatically detects Kubernetes configuration in this order:

1. In-cluster configuration (if running inside a Kubernetes pod)
2. `KUBECONFIG` environment variable
3. Default kubeconfig location (`~/.kube/config`)

## Dependencies

- [GoFr](https://gofr.dev/) - CLI framework
- [Kubernetes client-go](https://github.com/kubernetes/client-go) - Kubernetes API client
- [go-pretty](https://github.com/jedib0t/go-pretty) - Beautiful table formatting

## Architecture

Built using GoFr's subcommand framework for clean CLI structure:

- `main()`: Initializes GoFr CMD app and registers subcommands
- `imagesHandler()`: Handles the images subcommand with flag parsing and K8s API calls
- `getKubeConfig()`: Handles kubeconfig resolution and client setup

## Contributing

1. Add new subcommands using `app.SubCommand(name, handler, options...)`
2. Follow GoFr's Handler signature: `func(ctx *gofr.Context) (any, error)`
3. Use GoFr's built-in help system with `AddDescription()` and `AddHelp()`

## License

See LICENSE file for details.
