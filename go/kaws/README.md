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

# Query for Kubernetes events matching "failed to get sandbox image"
./kaws kube event

# Query events in a specific namespace
./kaws kube event --namespace kube-system

# Use verbose mode for more details
./kaws kube event --verbose

# Use a custom kubeconfig file
./kaws kube event --kubeconfig ~/.kube/custom-config

# Use a custom config file
./kaws --config ~/.kaws-prod.yaml kube event
```

## Commands

### `kube`

Parent command for all Kubernetes-related operations.

#### `kube event`

Queries Kubernetes events across all namespaces (or a specific namespace) and filters for events containing the message "failed to get sandbox image". This is useful for troubleshooting pod startup issues related to container runtime problems.

**Flags:**
- `-n, --namespace`: Specify a namespace to query (default: all namespaces)
- `-k, --kubeconfig`: Path to kubeconfig file (default: `$HOME/.kube/config`)
- `-v, --verbose`: Enable verbose output

**Example output:**
```
Found 2 event(s) matching 'failed to get sandbox image':

Namespace: default
Name: my-pod.17a6b8c9d3e1f2a4
Type: Warning
Reason: FailedCreatePodSandBox
Object: Pod/my-pod
Count: 5
First Seen: 2024-10-14 10:30:00
Last Seen: 2024-10-14 10:35:00
Message: Failed to create pod sandbox: rpc error: code = Unknown desc = failed to get sandbox image...
---
```

## Development

This CLI is built using the following packages:
- [Cobra](https://github.com/spf13/cobra) - Powerful framework for building CLI applications
- [Viper](https://github.com/spf13/viper) - Configuration management with support for config files, environment variables, and flags
- [client-go](https://github.com/kubernetes/client-go) - Kubernetes Go client library

### Adding New Commands

To add new commands, create new `cobra.Command` instances and add them to the root command using `rootCmd.AddCommand()`.

Example:
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

