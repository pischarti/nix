module "inspection_vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "v6.0.1"
  name    = "${local.name}-inspection-vpc"
  cidr    = var.vpc_cidr_inspection

  azs = ["${var.aws_region}a"]

  create_database_subnet_group = false
  create_igw                   = false

  manage_default_route_table = true
  default_route_table_tags = {
    Name = "default_${local.name}_inspection_vpc"
  }
}

// TGW Subnet
resource "aws_subnet" "inspection_vpc_tgw_subnet" {
  availability_zone       = "${var.aws_region}a"
  cidr_block              = var.inspection_vpc_tgw_subnet_cidr
  vpc_id                  = module.inspection_vpc.vpc_id
  map_public_ip_on_launch = false

  tags = {
    Name = "${local.name}_tgw_subnet"
  }
}


// AWS Network Firewall Subnets
resource "aws_subnet" "inspection_vpc_firewall_subnet" {
  availability_zone       = "${var.aws_region}a"
  cidr_block              = var.inspection_vpc_firewall_subnet_cidr
  vpc_id                  = module.inspection_vpc.vpc_id
  map_public_ip_on_launch = false

  tags = {
    Name = "${local.name}_firewall_subnet"
  }
}
