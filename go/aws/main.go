package main

import (
	"github.com/pischarti/nix/go/pkg/aws"
	"gofr.dev/pkg/gofr"
)

func main() {
	app := gofr.NewCMD()

	// Add subnets sub-command
	app.SubCommand("subnets", aws.ListSubnets,
		gofr.AddDescription("List all subnets in a VPC with optional filtering and sorting"),
		gofr.AddHelp("Usage: aws subnets --vpc VPC_ID [--zone AZ] [--sort SORT_BY]\n"+
			"Options:\n"+
			"  --vpc VPC_ID    VPC ID to list subnets for (required)\n"+
			"  --zone AZ       Filter by availability zone (optional)\n"+
			"  --sort SORT_BY  Sort by: cidr (default), az, name, type"),
	)

	app.Run()
}
