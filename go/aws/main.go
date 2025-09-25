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
		gofr.AddDescription("Manage AWS Network Load Balancers - list NLBs and manage subnets"),
		gofr.AddHelp("Usage: aws nlb [COMMAND]\n"+
			"Commands:\n"+
			"  list               List all Network Load Balancers in a VPC (default)\n"+
			"  add-subnet         Add subnets from a zone to NLBs in a VPC\n"+
			"  remove-subnet      Remove a subnet from NLBs in a VPC and zone\n"+
			"  check-associations Check for service associations that might prevent subnet removal\n\n"+
			"Examples:\n"+
			"  aws nlb --vpc vpc-12345678\n"+
			"  aws nlb list --vpc vpc-12345678\n"+
			"  aws nlb list --vpc vpc-12345678 --zone us-east-1a\n"+
			"  aws nlb list --vpc vpc-12345678 --sort state\n"+
			"  aws nlb add-subnet --vpc vpc-12345678 --zone us-east-1b\n"+
			"  aws nlb check-associations --vpc vpc-12345678\n"+
			"  aws nlb remove-subnet --vpc vpc-12345678 --zone us-east-1a\n"+
			"  aws nlb remove-subnet --vpc vpc-12345678 --zone us-east-1a --nlb-name my-nlb"),
	)

	// Add ecr command with nested sub-commands
	app.SubCommand("ecr", aws.ECRRouter,
		gofr.AddDescription("Manage AWS ECR repositories - list image versions and tags"),
		gofr.AddHelp("Usage: aws ecr [COMMAND]\n"+
			"Commands:\n"+
			"  list               List all image versions in an ECR repository (default)\n\n"+
			"Examples:\n"+
			"  aws ecr --repository my-repo\n"+
			"  aws ecr list --repository my-repo\n"+
			"  aws ecr list --repository my-repo --tag latest\n"+
			"  aws ecr list --repository my-repo --sort tag\n"+
			"  aws ecr list --repository my-repo --sort size\n"+
			"  aws ecr --all\n"+
			"  aws ecr list --all --tag latest\n"+
			"  aws ecr --repository my-repo --older-than latest\n"+
			"  aws ecr --all --older-than v1.0\n"+
			"  aws ecr --repository my-repo --output yaml\n"+
			"  aws ecr --all --output yaml"),
	)

	app.Run()
}
