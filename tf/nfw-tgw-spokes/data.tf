data "aws_availability_zones" "available" {
  state = "available"
  # Filter to exclude constrained zones and limit to stable AZs
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

# Use only the first 2 AZs to avoid constrained zones that don't support Network Firewall
locals {
  # Limit to first 2 AZs which are typically the most stable and support all services
  selected_azs = slice(data.aws_availability_zones.available.names, 0, 2)
}

data "aws_ami" "amazon-linux-2" {
  most_recent = true
  owners      = ["amazon"]
  name_regex  = "amzn2-ami-hvm*"
}

data "aws_region" "current" {}

