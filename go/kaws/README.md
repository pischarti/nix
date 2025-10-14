# kaws

A CLI tool for Kubernetes on AWS built with Cobra.

## Installation

```bash
cd /Users/steve/dev/nix/go/kaws
go build -o kaws
```

## Usage

```bash
# Run the CLI
./kaws

# Show version
./kaws version

# Show help
./kaws --help

# Use verbose mode
./kaws --verbose
```

## Development

This CLI is built using the [Cobra](https://github.com/spf13/cobra) package, which provides a powerful framework for building CLI applications.

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

