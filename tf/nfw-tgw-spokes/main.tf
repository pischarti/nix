terraform {
  required_version = ">= 1.4.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">=3.7.2"
    }
  }
  backend "s3" {
    bucket = "terraform-state-aws-us-east-1"
    key    = "network-firewall/tgw-spokes/terraform.tfstate"
    region = "us-east-1"
  }
}

provider "aws" {
  region = var.aws_region
}

locals {
  tags = {
    Environment = "dev"
    Project     = "nfw-tgw-spokes"
  }
  spoke_vpc_a_cidr    = cidrsubnet(var.super_cidr_block, 8, 10)
  spoke_vpc_b_cidr    = cidrsubnet(var.super_cidr_block, 8, 11)
  inspection_vpc_cidr = cidrsubnet(var.super_cidr_block, 8, 255)
}

