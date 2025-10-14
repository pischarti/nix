# kaws

A CLI tool for Kubernetes on AWS built with Cobra.

## Installation

```bash
cd /Users/steve/dev/nix/go/kaws
go build -o kaws
```

## Configuration

`kaws` supports configuration through multiple sources (in order of precedence):
1. Command-line flags (highest priority)
2. Environment variables (prefixed with `KAWS_`)
3. Configuration file (lowest priority)

### Configuration File

Create a `.kaws.yaml` file in your home directory (`~/.kaws.yaml`) or current directory:

```yaml
# kaws configuration file
kubeconfig: ~/.kube/config
namespace: ""  # Leave empty for all namespaces
verbose: false
```

You can also specify a custom config file location:
```bash
./kaws --config /path/to/config.yaml kube event
```

### Environment Variables

Set environment variables with the `KAWS_` prefix:
```bash
export KAWS_NAMESPACE=kube-system
export KAWS_VERBOSE=true
./kaws kube event
```

## Usage

```bash
# Run the CLI
./kaws

# Show version
./kaws version

# Show help
./kaws --help

# Query for Kubernetes events matching a specific term
./kaws kube event --search "failed to get sandbox image"

# Query events in a specific namespace
./kaws kube event --search "ImagePullBackOff" --namespace kube-system

# Output in YAML format
./kaws kube event --search "error" --output yaml

# Use verbose mode for more details
./kaws kube event --search "error" --verbose

# Use a custom kubeconfig file
./kaws kube event --search "OOMKilled" --kubeconfig ~/.kube/custom-config

# Use a custom config file
./kaws --config ~/.kaws-prod.yaml kube event --search "BackOff"
```

## Commands

### `kube`

Parent command for all Kubernetes-related operations.

#### `kube event`

Queries Kubernetes events across all namespaces (or a specific namespace) and filters them by message content. For pod-related events, the command automatically enriches the output with node information, showing which node the pod is scheduled on. This is useful for troubleshooting various Kubernetes issues by searching for specific error messages or patterns and identifying node-specific problems.

**Flags:**
- `-s, --search`: Search term to filter events (required)
- `-o, --output`: Output format: `table` or `yaml` (default: `table`)
- `--show-instance-id`: Include EC2 instance IDs from node labels (useful for AWS EKS clusters)
- `-n, --namespace`: Specify a namespace to query (default: all namespaces)
- `-k, --kubeconfig`: Path to kubeconfig file (default: `$HOME/.kube/config`)
- `-v, --verbose`: Enable verbose output

**Examples:**

Search for sandbox image issues:
```bash
./kaws kube event --search "failed to get sandbox image"
```

Search for image pull errors in a specific namespace:
```bash
./kaws kube event --search "ImagePullBackOff" --namespace default
```

Search for any error events:
```bash
./kaws kube event --search "error" --verbose
```

Output in YAML format:
```bash
./kaws kube event --search "ImagePullBackOff" --output yaml
```

Include EC2 instance IDs (useful for EKS troubleshooting):
```bash
./kaws kube event --search "failed to get sandbox image" --show-instance-id
```

**Example output (table format with node names):**
```
Found 2 event(s) matching "failed to get sandbox image":

┌───────────┬─────────┬───────────────────────┬──────────────┬──────────────────┬───────┬─────────────────────┬──────────────────────────────────────────────────────────────────┐
│ Namespace │ Type    │ Reason                │ Object       │ Node             │ Count │ Last Seen           │ Message                                                          │
├───────────┼─────────┼───────────────────────┼──────────────┼──────────────────┼───────┼─────────────────────┼──────────────────────────────────────────────────────────────────┤
│ default   │ Warning │ FailedCreatePodSandBox│ Pod/my-pod   │ ip-10-0-1-50.ec2 │     5 │ 2024-10-14 10:35:00 │ Failed to create pod sandbox: rpc error: code = Unknown desc... │
│ default   │ Warning │ FailedCreatePodSandBox│ Pod/my-app   │ ip-10-0-1-51.ec2 │     3 │ 2024-10-14 10:32:00 │ Failed to create pod sandbox: rpc error: code = Unknown desc... │
└───────────┴─────────┴───────────────────────┴──────────────┴──────────────────┴───────┴─────────────────────┴──────────────────────────────────────────────────────────────────┘
```

The output is formatted as a clean table using [go-pretty](https://github.com/jedib0t/go-pretty), making it easy to scan multiple events at once. **For pod-related events, the table includes the node name** where the pod is (or was) scheduled, making it easier to identify node-specific issues.

When using the `--show-instance-id` flag, the table includes an additional column showing the EC2 instance ID for each node. This is particularly useful for:
- Correlating Kubernetes events with AWS EC2 instance metrics
- Identifying problematic EC2 instances in EKS clusters
- Facilitating AWS CloudWatch log searches by instance ID
- Cross-referencing with AWS Systems Manager or CloudWatch dashboards

**Example output (YAML format):**
```yaml
- metadata:
    name: my-pod.17a6b8c9d3e1f2a4
    namespace: default
  involvedObject:
    kind: Pod
    name: my-pod
  reason: FailedCreatePodSandBox
  message: Failed to create pod sandbox: rpc error: code = Unknown desc = failed to get sandbox image "registry.k8s.io/pause:3.9"
  type: Warning
  count: 5
  firstTimestamp: "2024-10-14T10:30:00Z"
  lastTimestamp: "2024-10-14T10:35:00Z"
```

The YAML output format is useful for piping to other tools, storing event data, or integrating with automation scripts.

## Development

This CLI is built using the following packages:
- [Cobra](https://github.com/spf13/cobra) - Powerful framework for building CLI applications
- [Viper](https://github.com/spf13/viper) - Configuration management with support for config files, environment variables, and flags
- [go-pretty](https://github.com/jedib0t/go-pretty) - Beautiful table formatting for terminal output
- [client-go](https://github.com/kubernetes/client-go) - Kubernetes Go client library

### Project Structure

```
nix/
├── go/kaws/
│   ├── main.go                      # Entry point, root command setup (68 lines)
│   ├── cmd/
│   │   └── kube/
│   │       ├── kube.go              # Kube command setup (20 lines)
│   │       └── event/
│   │           ├── event.go         # Event subcommand implementation (173 lines)
│   │           └── event_test.go    # Event command tests (264 lines)
│   ├── .kaws.yaml.example           # Example configuration file
│   └── README.md
└── pkg/
    ├── config/
    │   ├── viper.go                 # Viper configuration initialization
    │   ├── viper_test.go            # Config tests
    │   ├── kubeconfig.go            # Kubeconfig utilities
    │   └── kubeconfig_test.go       # Kubeconfig tests
    └── k8s/
        ├── client.go                # Kubernetes client and query utilities
        └── client_test.go           # K8s package tests
```

Each subcommand has its own package for better organization and maintainability. Common functionality is extracted into reusable packages:
- **`pkg/config`**: Configuration management (Viper initialization, kubeconfig utilities)
- **`pkg/k8s`**: Kubernetes client and query utilities

This structure makes it easy to add new subcommands without cluttering the parent command files.

### Adding New Commands

#### Adding a Kube Subcommand

To add a new subcommand under `kube`, follow the pattern used by the `event` package:

1. Create a new package: `cmd/kube/mysubcmd/`
2. Create `mysubcmd.go` with the following structure:

```go
package mysubcmd

import (
    "github.com/spf13/cobra"
)

// NewMySubCmdCmd creates the mysubcmd subcommand
func NewMySubCmdCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "mysubcmd",
        Short: "Description of my subcommand",
        RunE:  run,
    }
}

func run(cmd *cobra.Command, args []string) error {
    // Command logic here
    return nil
}
```

3. Add it to `cmd/kube/kube.go`:

```go
import "github.com/pischarti/nix/go/kaws/cmd/kube/mysubcmd"

// In NewKubeCmd()
kubeCmd.AddCommand(mysubcmd.NewMySubCmdCmd())
```

#### Adding a Top-Level Command

To add a new top-level command, edit `main.go`:

```go
myCmd := &cobra.Command{
    Use:   "mycommand",
    Short: "Description of my command",
    Run: func(cmd *cobra.Command, args []string) {
        // Command logic here
    },
}
rootCmd.AddCommand(myCmd)
```

For complex commands with multiple subcommands, follow the pattern used in `cmd/kube/` and create a new package.

