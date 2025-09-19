# AWS Region
aws_region = "us-east-1"

# Resource Tags
tags = {
  Name        = "nlb-firewall"
  Environment = "dev"
  Project     = "network-security"
  Owner       = "devsecops"
}

# Main VPC Configuration
vpc_cidr = "10.0.0.0/16"

# Availability Zones (adjust based on your region)
availability_zones = ["us-east-1a", "us-east-1b"]

# Public Subnet Configuration
public_subnet_cidrs = ["10.0.1.0/24", "10.0.2.0/24"]

# Private Subnet Configuration
private_subnet_cidrs = ["10.0.10.0/24", "10.0.20.0/24"]

# Firewall VPC Configuration
firewall_vpc_cidr = "10.1.0.0/16"

# Network Firewall Subnet Configuration
firewall_subnet_cidrs = ["10.1.1.0/24", "10.1.2.0/24"]

# Network Firewall Rule Group Configuration
stateless_rule_capacity = 100
stateful_rule_capacity = 100
