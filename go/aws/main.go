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

	// Add nlb command with nested sub-commands
	app.SubCommand("nlb", aws.NLBRouter,
		gofr.AddDescription("Manage AWS Network Load Balancers - list NLBs in a VPC"),
		gofr.AddHelp("Usage: aws nlb [COMMAND]\n"+
			"Commands:\n"+
			"  list               List all Network Load Balancers in a VPC (default)\n\n"+
			"Examples:\n"+
			"  aws nlb --vpc vpc-12345678\n"+
			"  aws nlb list --vpc vpc-12345678\n"+
			"  aws nlb list --vpc vpc-12345678 --zone us-east-1a\n"+
			"  aws nlb list --vpc vpc-12345678 --sort state"),
	)

	app.Run()
}
