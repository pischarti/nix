# kaws

A CLI tool for Kubernetes on AWS built with Cobra.

## Installation

```bash
cd /Users/steve/dev/nix/go/kaws
go build -o kaws
```

## Complete Troubleshooting Workflow

Here's a complete example workflow tracing a Kubernetes issue back to AWS infrastructure:

```bash
# Step 1: Find pods with sandbox image issues
./kaws kube event --search "failed to get sandbox image" --show-instance-id

# Output shows:
# - Which pods are affected
# - Which nodes they're on
# - Which EC2 instances those nodes are

# Step 2: Get the node groups for those EC2 instances
./kaws aws ngs i-1234567890abcdef0 i-0987654321fedcba0

# Output shows:
# - Which EKS cluster the instances belong to
# - Which node group they're in
# - Instance types

# Step 3: Take action based on findings
# - Update node group configuration
# - Replace problematic instances
# - Scale node groups
# - Check AWS CloudWatch metrics for those instances
```

This workflow enables you to quickly trace Kubernetes pod issues back to specific AWS EKS node groups, making it easy to identify and remediate infrastructure-level problems.

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

# Find node groups for instances from event command
./kaws kube event --search "failed to get sandbox image" --show-instance-id --output yaml | \
  grep instanceId | awk '{print $2}' | xargs ./kaws aws ngs

# Recycle a problematic node group
./kaws aws ngs recycle ng-workers-1
```

## Commands

### `kube`

Parent command for all Kubernetes-related operations.

### `aws`

Parent command for all AWS-related operations.

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

#### `aws ngs`

Manages EKS node groups. When called with instance IDs, finds the EKS node groups for those EC2 instances. This command queries AWS EC2 and EKS to determine which node group each instance belongs to, making it easy to trace issues from Kubernetes events back to AWS infrastructure.

**Flags:**
- `-r, --region`: AWS region (default: from AWS config or environment)
- `-c, --cluster`: EKS cluster name (if not specified, searches all clusters)

**Examples:**

Find node groups for specific instances:
```bash
./kaws aws ngs i-1234567890abcdef0 i-0987654321fedcba0
```

Pipe instance IDs from event command:
```bash
./kaws kube event --search "failed to get sandbox image" --show-instance-id --output yaml | \
  grep instanceId | awk '{print $2}' | xargs ./kaws aws ngs
```

Specify AWS region:
```bash
./kaws aws ngs i-1234567890abcdef0 --region us-west-2
```

**Example output:**
```
Found node group information for 2 instance(s):

┌─────────────────────────┬──────────────┬─────────────────┬───────────────┐
│ Instance ID             │ Cluster      │ Node Group      │ Instance Type │
├─────────────────────────┼──────────────┼─────────────────┼───────────────┤
│ i-1234567890abcdef0     │ my-cluster   │ ng-workers-1    │ t3.large      │
│ i-0987654321fedcba0     │ my-cluster   │ ng-workers-2    │ t3.xlarge     │
└─────────────────────────┴──────────────┴─────────────────┴───────────────┘
```

This command is particularly useful for:
- Tracing Kubernetes events back to specific EKS node groups
- Identifying which node group is experiencing issues
- Correlating pod problems with underlying AWS infrastructure
- Planning node group updates or replacements

#### `aws ngs recycle`

Recycles (restarts) EKS node groups by scaling them down to zero, waiting for all instances to terminate, then scaling back up to the original configuration. This is a subcommand of `ngs`. This is useful for:
- Recovering from container runtime issues (like "failed to get sandbox image")
- Forcing fresh instances to fix persistent node problems
- Clearing stuck containers or zombie processes
- Applying underlying OS or runtime updates

**Flags:**
- `-r, --region`: AWS region (default: from AWS config)
- `-p, --poll-interval`: Polling interval for status checks (default: 15s)
- `--timeout`: Maximum time to wait for recycle to complete (default: 20m)

**Examples:**

Recycle a single node group:
```bash
./kaws aws ngs recycle ng-workers-1
```

Recycle multiple node groups:
```bash
./kaws aws ngs recycle ng-workers-1 ng-workers-2
```

With custom polling and region:
```bash
./kaws aws ngs recycle ng-workers-1 --region us-west-2 --poll-interval 10s
```

**Example output:**
```
=== Recycling node group: ng-workers-1 ===

[1/5] Getting current node group configuration...
  Current config: Min=2, Max=10, Desired=5
  Current instances: 5

[2/5] Scaling down to zero...
  Scaled to Min=0, Max=0, Desired=0

[3/5] Waiting for instances to terminate...
  [15s] Instance states: map[shutting-down:3 terminated:2]
  [30s] Instance states: map[terminated:5]
  All instances terminated

[4/5] Scaling back up to original configuration...
  Scaled to Min=2, Max=10, Desired=5

[5/5] Waiting for new instances to start...
  [15s] Waiting for instances to appear: 2/5
  [30s] Instances: 5/5, States: map[pending:5]
  5 instances are now starting (pending/running)

✓ Successfully recycled node group: ng-workers-1
```

**Complete Workflow:**

Identify problem → Find node group → Recycle:
```bash
# Step 1: Find events with instance IDs
./kaws kube event --search "failed to get sandbox image" --show-instance-id

# Step 2: Find which node groups have the problem
./kaws aws ngs i-1234567890abcdef0

# Step 3: Recycle the problematic node group
./kaws aws ngs recycle ng-workers-1 --verbose
```

**⚠️ Warning:** This command will temporarily reduce node group capacity to zero. Ensure you have:
- Multiple node groups for redundancy
- Pod disruption budgets configured
- Tested in non-production first

## Development

This CLI is built using the following packages:
- [Cobra](https://github.com/spf13/cobra) - Powerful framework for building CLI applications
- [Viper](https://github.com/spf13/viper) - Configuration management with support for config files, environment variables, and flags
- [go-pretty](https://github.com/jedib0t/go-pretty) - Beautiful table formatting for terminal output
- [client-go](https://github.com/kubernetes/client-go) - Kubernetes Go client library
- [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2) - AWS service integration

### Project Structure

```
nix/
├── go/kaws/
│   ├── main.go                      # Entry point, root command setup (70 lines)
│   ├── cmd/
│   │   ├── aws/
│   │   │   ├── aws.go               # AWS command setup (20 lines)
│   │   │   └── ngs/
│   │   │       ├── ngs.go           # Node groups management (126 lines)
│   │   │       └── recycle/
│   │   │           └── ngrecycle.go # Node group recycle (359 lines)
│   │   └── kube/
│   │       ├── kube.go              # Kube command setup (20 lines)
│   │       └── event/
│   │           ├── event.go         # Event subcommand (129 lines)
│   │           └── event_test.go    # Event tests (204 lines)
│   ├── .kaws.yaml.example           # Example configuration file
│   └── README.md
└── pkg/
    ├── aws/
    │   ├── nodegroups.go            # Node group queries (102 lines)
    │   ├── ecr.go                   # ECR utilities
    │   ├── nlb.go                   # NLB utilities
    │   └── subnets.go               # Subnet utilities
    ├── config/
    │   ├── viper.go                 # Viper configuration (42 lines)
    │   ├── viper_test.go            # Config tests (80 lines)
    │   ├── kubeconfig.go            # Kubeconfig utilities
    │   └── kubeconfig_test.go       # Kubeconfig tests
    ├── k8s/
    │   ├── client.go                # K8s client creation (42 lines)
    │   ├── client_test.go           # Client tests (45 lines)
    │   ├── event.go                 # Event query & enrichment (177 lines)
    │   └── filter_test.go           # Filter tests (264 lines)
    └── print/
        ├── events.go                # Event display functions (176 lines)
        ├── events_test.go           # Display tests (62 lines)
        └── [other print utilities]
```

Each subcommand has its own package for better organization and maintainability. Common functionality is extracted into reusable packages:
- **`pkg/aws`**: AWS service queries (node groups, ECR, NLB, subnets)
- **`pkg/config`**: Configuration management (Viper initialization, kubeconfig utilities)
- **`pkg/k8s`**: Kubernetes client, event query, and enrichment utilities
- **`pkg/print`**: Display and formatting functions for various output types

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

