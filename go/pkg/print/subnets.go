package print

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pischarti/nix/go/pkg/vpc"
)

// PrintSubnetsTable prints subnets in a formatted table
func PrintSubnetsTable(subnets []vpc.SubnetInfo) {
	// Create table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleColoredBright)

	// Add headers
	t.AppendHeader(table.Row{"Subnet ID", "VPC ID", "CIDR Block", "AZ", "Name", "State", "Type"})

	// Add rows
	for _, subnet := range subnets {
		t.AppendRow(table.Row{
			subnet.SubnetID,
			subnet.VPCID,
			subnet.CIDRBlock,
			subnet.AZ,
			subnet.Name,
			subnet.State,
			subnet.Type,
		})
	}

	// Render table
	t.Render()
}

// PrintSubnetsTableString returns the table as a string instead of printing to stdout
func PrintSubnetsTableString(subnets []vpc.SubnetInfo) string {
	// Create table
	t := table.NewWriter()
	t.SetStyle(table.StyleColoredBright)

	// Add headers
	t.AppendHeader(table.Row{"Subnet ID", "VPC ID", "CIDR Block", "AZ", "Name", "State", "Type"})

	// Add rows
	for _, subnet := range subnets {
		t.AppendRow(table.Row{
			subnet.SubnetID,
			subnet.VPCID,
			subnet.CIDRBlock,
			subnet.AZ,
			subnet.Name,
			subnet.State,
			subnet.Type,
		})
	}

	// Return table as string
	return t.Render()
}
