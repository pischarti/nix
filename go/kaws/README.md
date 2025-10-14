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

# Use verbose mode for more details
./kaws kube event --search "error" --verbose

# Use a custom kubeconfig file
./kaws kube event --search "OOMKilled" --kubeconfig ~/.kube/custom-config

# Use a custom config file
./kaws --config ~/.kaws-prod.yaml kube event
```

## Commands

### `kube`

Parent command for all Kubernetes-related operations.

#### `kube event`

Queries Kubernetes events across all namespaces (or a specific namespace) and filters them by message content. This is useful for troubleshooting various Kubernetes issues by searching for specific error messages or patterns.

**Flags:**
- `-s, --search`: Search term to filter events (required)
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

**Example output:**
```
Found 2 event(s) matching "failed to get sandbox image":

┌───────────┬─────────┬───────────────────────┬──────────────┬───────┬─────────────────────┬─────────────────────────────────────────────────────────────────────────────────┐
│ Namespace │ Type    │ Reason                │ Object       │ Count │ Last Seen           │ Message                                                                         │
├───────────┼─────────┼───────────────────────┼──────────────┼───────┼─────────────────────┼─────────────────────────────────────────────────────────────────────────────────┤
│ default   │ Warning │ FailedCreatePodSandBox│ Pod/my-pod   │     5 │ 2024-10-14 10:35:00 │ Failed to create pod sandbox: rpc error: code = Unknown desc = failed to ge... │
│ default   │ Warning │ FailedCreatePodSandBox│ Pod/my-app   │     3 │ 2024-10-14 10:32:00 │ Failed to create pod sandbox: rpc error: code = Unknown desc = failed to ge... │
└───────────┴─────────┴───────────────────────┴──────────────┴───────┴─────────────────────┴─────────────────────────────────────────────────────────────────────────────────┘
```

The output is formatted as a clean table using [go-pretty](https://github.com/jedib0t/go-pretty), making it easy to scan multiple events at once.

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
│   ├── main.go                      # Entry point, root command setup
│   ├── cmd/
│   │   └── kube/
│   │       ├── kube.go              # Kube command setup
│   │       └── event/
│   │           ├── event.go         # Event subcommand implementation
│   │           └── event_test.go    # Event command tests
│   ├── .kaws.yaml.example           # Example configuration file
│   └── README.md
└── pkg/
    └── k8s/
        ├── client.go                # Kubernetes client and query utilities
        └── client_test.go           # K8s package tests
```

Each subcommand has its own package for better organization and maintainability. Common Kubernetes functionality is extracted into the `pkg/k8s` package for reusability across multiple commands. This structure makes it easy to add new subcommands without cluttering the parent command files.

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

