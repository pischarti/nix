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
  }
}

provider "aws" {
  region = var.aws_region
}

locals {
  name = "${var.env}-${var.team}-nfw"

  firewall_endpoint_id = flatten(resource.aws_networkfirewall_firewall.aws_network_firewall.firewall_status[*].sync_states[*].attachment[*].endpoint_id)

  aws_network_firewall_endpoint_id = local.firewall_endpoint_id

}
