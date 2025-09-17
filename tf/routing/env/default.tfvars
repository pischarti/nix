
aws_region = "us-east-1"

tags = {
  Name = "network-routing"
}

vpc_cidr = "10.0.0.0/16"

availability_zones = ["us-east-1a", "us-east-1b"]

public_subnet_cidrs = ["10.0.1.0/24", "10.0.2.0/24"]

private_subnet_cidrs = ["10.0.10.0/24", "10.0.20.0/24"]

# Network Firewall VPC Configuration
firewall_vpc_cidr = "10.1.0.0/16"

firewall_vpc_name = "inspection-vpc"

firewall_subnet_cidrs = ["10.1.1.0/24", "10.1.2.0/24"]

firewall_tgw_subnet_cidrs = ["10.1.10.0/24", "10.1.20.0/24"]

# Transit Gateway Configuration
tgw_name = "network-routing-tgw"

tgw_description = "Transit Gateway for routing traffic through Network Firewall inspection"

main_vpc_tgw_subnet_cidrs = ["10.0.100.0/24", "10.0.200.0/24"]
