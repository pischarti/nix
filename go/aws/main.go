package main

import (
	"github.com/pischarti/nix/go/pkg/aws"
	"gofr.dev/pkg/gofr"
)

func main() {
	app := gofr.NewCMD()

	// Add subnets command with nested sub-commands
	app.SubCommand("subnets", aws.SubnetsRouter,
		gofr.AddDescription("Manage AWS subnets - list or delete"),
		gofr.AddHelp("Usage: aws subnets [COMMAND]\n"+
			"Commands:\n"+
			"  list    List all subnets in a VPC (default)\n"+
			"  delete  Delete a subnet by ID\n\n"+
			"Examples:\n"+
			"  aws subnets --vpc vpc-12345678\n"+
			"  aws subnets list --vpc vpc-12345678\n"+
			"  aws subnets delete --subnet-id subnet-12345678"),
	)

	app.Run()
}
