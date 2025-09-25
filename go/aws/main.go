package main

import (
	"github.com/pischarti/nix/go/pkg/aws"
	"gofr.dev/pkg/gofr"
)

func main() {
	app := gofr.NewCMD()

	// Add subnets command with nested sub-commands
	app.SubCommand("subnets", aws.SubnetsRouter,
		gofr.AddDescription("Manage AWS subnets - list, delete, or check dependencies"),
		gofr.AddHelp("Usage: aws subnets [COMMAND]\n"+
			"Commands:\n"+
			"  list               List all subnets in a VPC (default)\n"+
			"  delete             Delete a subnet by ID\n"+
			"  check-dependencies Check what resources are preventing subnet deletion\n\n"+
			"Examples:\n"+
			"  aws subnets --vpc vpc-12345678\n"+
			"  aws subnets list --vpc vpc-12345678\n"+
			"  aws subnets delete --subnet-id subnet-12345678\n"+
			"  aws subnets check-dependencies --subnet-id subnet-12345678"),
	)

	app.Run()
}
