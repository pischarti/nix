# kube CLI

A Kubernetes command-line tool built with GoFr that provides utilities for managing and inspecting Kubernetes clusters.

## Features

### Images Subcommand

List all container images running in your Kubernetes cluster with various filtering and display options.

### Services Subcommand

List all Kubernetes services with annotations matching specified criteria. By default shows all services with any annotations, or filter by specific annotation keys or values. Automatically excludes `last-applied-configuration` annotations for cleaner output.

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

#### Usage

```bash
# List services with any annotations across all namespaces (default)
./kube services

# List services in a specific namespace
./kube services --namespace default

# List services across all namespaces (explicit)
./kube services --all-namespaces

# Display output in table format
./kube services --table

# Display with different table styles
./kube services --table --style simple
./kube services --table --style box
./kube services --table --style rounded
./kube services --table --style colored

# Sort output by different criteria
./kube services --sort namespace    # Default: sort by namespace
./kube services --sort name         # Sort by service name alphabetically
./kube services --sort none         # No sorting (original order)

# Filter by specific annotation key or value
./kube services --annotation-value "aws-load-balancer"
./kube services --annotation-value "nlb"
./kube services --annotation-value "internet-facing"

# Show help
./kube services --help
```

#### Options

- `--namespace, -n`: Query a specific namespace (default: all namespaces)
- `--all-namespaces, -A`: Query across all namespaces (default behavior)
- `--table, -t`: Display output in table format with namespace, name, type, and annotations columns
- `--style`: Table style - `simple`, `box`, `rounded`, or `colored` (default: colored)
- `--sort`: Sort order - `namespace` (default), `name`, or `none`
- `--annotation-value`: Filter by annotation key or value containing this text (case-insensitive)
- `--help, -h`: Show help information

#### Examples

```bash
# Get services with any annotations across the entire cluster
./kube services --all-namespaces

# List services in the kube-system namespace
./kube services --namespace kube-system

# Display services in table format
./kube services --table --namespace default

# Different table styles
./kube services --table --style simple
./kube services --table --style box
./kube services --table --style rounded

# Combine sorting with other options
./kube services --table --sort name --style box
./kube services --sort name --namespace default

# Filter by specific annotation keys or values
./kube services --annotation-value "aws-load-balancer" --table
./kube services --annotation-value "nlb" --table
./kube services --annotation-value "internet-facing" --namespace production

# Output format when using --table (colored style):
# ┌───────────┬──────────┬──────┬──────────────────────────────────────────────┐
# │ NAMESPACE │ NAME     │ TYPE │ ANNOTATIONS                                 │
# ├───────────┼──────────┼──────┼──────────────────────────────────────────────┤
# │ default   │ my-svc   │ LoadBalancer │ service.beta.kubernetes.io/aws-load-balancer-type=nlb │
# │           │          │      │ service.beta.kubernetes.io/aws-load-balancer-scheme=internet-facing │
# │ default   │ api-svc  │ LoadBalancer │ custom.annotation=value │
# └───────────┴──────────┴──────┴──────────────────────────────────────────────┘

# Output format when using list mode:
# default/my-svc (LoadBalancer):
#   service.beta.kubernetes.io/aws-load-balancer-type=nlb
#   service.beta.kubernetes.io/aws-load-balancer-scheme=internet-facing
# default/api-svc (LoadBalancer):
#   custom.annotation=value
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
- Appropriate RBAC permissions to list pods and services

## Configuration

The CLI automatically detects Kubernetes configuration in this order:

1. In-cluster configuration (if running inside a Kubernetes pod)
2. `KUBECONFIG` environment variable
3. Default kubeconfig location (`~/.kube/config`)

## Dependencies

- [GoFr](https://gofr.dev/) - CLI framework
- [Kubernetes client-go](https://github.com/kubernetes/client-go) - Kubernetes API client
- [go-pretty](https://github.com/jedib0t/go-pretty) - Beautiful table formatting

## Testing

The project includes comprehensive unit tests covering:

- **Config Package** (`pkg/config`): Tests for Kubernetes configuration loading
- **Main Package** (`kube`): Tests for CLI argument parsing, validation, and output formatting
- **Container Package** (`pkg/container`): Tests for images and services command handlers, argument parsing, and AWS load balancer annotation filtering
- **Integration Tests**: Tests that verify actual function behavior and output

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run tests with verbose output
go test ./... -v

# Use the test runner script
./test.sh
```

### Test Coverage

- **Config Package**: 77.8% coverage
- **Main Package**: 28.3% coverage
- **Overall**: 30.8% coverage

The main package coverage is lower because the `main()` function and `imagesHandler()` are not directly testable without integration testing against a real Kubernetes cluster.

## Architecture

Built using GoFr's subcommand framework with a modular package structure:

### Package Organization

- **`main`**: CLI application entry point and subcommand registration
- **`pkg/config`**: Kubernetes configuration loading and management
- **`pkg/container`**: Container image listing logic and command handling
- **`pkg/print`**: Output formatting and display functions

### Clean Separation of Concerns

- **Main Package**: Initializes GoFr CMD app and registers subcommands
- **Container Package**: Handles images and services command logic, argument parsing, and Kubernetes API interactions
- **Config Package**: Manages Kubernetes configuration loading and client setup
- **Print Package**: Handles all output formatting and display logic

## Contributing

1. Add new subcommands using `app.SubCommand(name, handler, options...)`
2. Follow GoFr's Handler signature: `func(ctx *gofr.Context) (any, error)`
3. Use GoFr's built-in help system with `AddDescription()` and `AddHelp()`

## License

See LICENSE file for details.
