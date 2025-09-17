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

# VPC
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-vpc"
  })
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-igw"
  })
}

# Public Subnets
resource "aws_subnet" "public" {
  count = length(var.public_subnet_cidrs)

  vpc_id                  = aws_vpc.main.id
  cidr_block              = var.public_subnet_cidrs[count.index]
  availability_zone       = var.availability_zones[count.index]
  map_public_ip_on_launch = true

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-public-subnet-${count.index + 1}"
    Type = "Public"
  })
}

# Private Subnets
resource "aws_subnet" "private" {
  count = length(var.private_subnet_cidrs)

  vpc_id            = aws_vpc.main.id
  cidr_block        = var.private_subnet_cidrs[count.index]
  availability_zone = var.availability_zones[count.index]

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-private-subnet-${count.index + 1}"
    Type = "Private"
  })
}

# Elastic IP for NAT Gateway
resource "aws_eip" "nat" {
  count = length(var.public_subnet_cidrs)

  domain = "vpc"
  depends_on = [aws_internet_gateway.main]

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-nat-eip-${count.index + 1}"
  })
}

# NAT Gateway
resource "aws_nat_gateway" "main" {
  count = length(var.public_subnet_cidrs)

  allocation_id = aws_eip.nat[count.index].id
  subnet_id     = aws_subnet.public[count.index].id

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-nat-gw-${count.index + 1}"
  })

  depends_on = [aws_internet_gateway.main]
}

# Route Table for Public Subnets (with GWLB inspection for private subnet traffic)
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  # Route traffic destined for private subnets through GWLB endpoint for inspection
  dynamic "route" {
    for_each = var.private_subnet_cidrs
    content {
      cidr_block      = route.value
      vpc_endpoint_id = aws_vpc_endpoint.gwlb_public[0].id
    }
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-public-rt"
  })

  depends_on = [aws_vpc_endpoint.gwlb_public]
}

# Route Table for Private Subnets (with GWLB inspection for public subnet traffic)
resource "aws_route_table" "private" {
  count = length(var.private_subnet_cidrs)

  vpc_id = aws_vpc.main.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.main[count.index].id
  }

  # Route traffic destined for public subnets through GWLB endpoint for inspection
  dynamic "route" {
    for_each = var.public_subnet_cidrs
    content {
      cidr_block      = route.value
      vpc_endpoint_id = aws_vpc_endpoint.gwlb_private[count.index].id
    }
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-private-rt-${count.index + 1}"
  })

  depends_on = [aws_vpc_endpoint.gwlb_private]
}

# Route Table Association for Public Subnets
resource "aws_route_table_association" "public" {
  count = length(var.public_subnet_cidrs)

  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# Route Table Association for Private Subnets
resource "aws_route_table_association" "private" {
  count = length(var.private_subnet_cidrs)

  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private[count.index].id
}

# ============================================================================
# NETWORK FIREWALL VPC
# ============================================================================

# Network Firewall VPC
resource "aws_vpc" "firewall" {
  cidr_block           = var.firewall_vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-firewall-vpc"
  })
}

# Firewall Subnets (for Network Firewall endpoints)
resource "aws_subnet" "firewall" {
  count = length(var.firewall_subnet_cidrs)

  vpc_id            = aws_vpc.firewall.id
  cidr_block        = var.firewall_subnet_cidrs[count.index]
  availability_zone = var.availability_zones[count.index]

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-firewall-subnet-${count.index + 1}"
    Type = "Firewall"
  })
}

# GWLB Subnets (for Gateway Load Balancer)
resource "aws_subnet" "gwlb" {
  count = length(var.gwlb_subnet_cidrs)

  vpc_id            = aws_vpc.firewall.id
  cidr_block        = var.gwlb_subnet_cidrs[count.index]
  availability_zone = var.availability_zones[count.index]

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-gwlb-subnet-${count.index + 1}"
    Type = "GWLB"
  })
}

# ============================================================================
# NETWORK FIREWALL
# ============================================================================

# Network Firewall Rule Group
resource "aws_networkfirewall_rule_group" "stateless_rule_group" {
  capacity = 100
  name     = "${var.tags.Name}-stateless-rule-group"
  type     = "STATELESS"

  rule_group {
    rules_source {
      stateless_rules_and_custom_actions {
        stateless_rule {
          priority = 1
          rule_definition {
            actions = ["aws:forward_to_sfe"]
            match_attributes {
              source {
                address_definition = "0.0.0.0/0"
              }
              destination {
                address_definition = "0.0.0.0/0"
              }
            }
          }
        }
      }
    }
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-stateless-rule-group"
  })
}

# Network Firewall Policy
resource "aws_networkfirewall_firewall_policy" "policy" {
  name = "${var.tags.Name}-firewall-policy"

  firewall_policy {
    stateless_default_actions          = ["aws:forward_to_sfe"]
    stateless_fragment_default_actions = ["aws:forward_to_sfe"]

    stateless_rule_group_reference {
      priority     = 1
      resource_arn = aws_networkfirewall_rule_group.stateless_rule_group.arn
    }
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-firewall-policy"
  })
}

# Network Firewall
resource "aws_networkfirewall_firewall" "main" {
  name                = "${var.tags.Name}-network-firewall"
  description         = "AWS Network Firewall for traffic inspection via GWLB"
  firewall_policy_arn = aws_networkfirewall_firewall_policy.policy.arn
  vpc_id              = aws_vpc.firewall.id

  dynamic "subnet_mapping" {
    for_each = aws_subnet.firewall
    content {
      subnet_id = subnet_mapping.value.id
    }
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-network-firewall"
  })
}

# ============================================================================
# GATEWAY LOAD BALANCER
# ============================================================================

# Gateway Load Balancer
resource "aws_lb" "gwlb" {
  name               = "${var.tags.Name}-gwlb"
  load_balancer_type = "gateway"
  subnets            = aws_subnet.gwlb[*].id

  enable_deletion_protection = false

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-gwlb"
  })
}

# Target Group for Network Firewall endpoints
resource "aws_lb_target_group" "firewall_endpoints" {
  name        = "${var.tags.Name}-firewall-tg"
  port        = 6081
  protocol    = "GENEVE"
  target_type = "ip"
  vpc_id      = aws_vpc.firewall.id

  health_check {
    port     = 6081
    protocol = "TCP"
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-firewall-target-group"
  })
}

# Note: Network Firewall endpoint registration with GWLB requires manual configuration
# or external automation due to the dynamic nature of firewall endpoint creation.
# The endpoints will be automatically created when the Network Firewall is deployed,
# but their IP addresses are not available during Terraform planning phase.
#
# For production use, consider one of these approaches:
# 1. Use a two-stage deployment (first deploy firewall, then register with GWLB)
# 2. Use external automation (Lambda function, script) to register endpoints
# 3. Use terraform apply with -target to deploy incrementally
#
# For now, we'll create the target group without initial targets.
# Targets can be registered manually or via external automation after deployment.

# Optional: Register firewall endpoint IPs if provided
resource "aws_lb_target_group_attachment" "firewall_endpoints_optional" {
  count = length(var.firewall_endpoint_ips)

  target_group_arn = aws_lb_target_group.firewall_endpoints.arn
  target_id        = var.firewall_endpoint_ips[count.index]
  port             = 6081

  lifecycle {
    create_before_destroy = true
  }
}

# GWLB Listener
resource "aws_lb_listener" "gwlb" {
  load_balancer_arn = aws_lb.gwlb.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.firewall_endpoints.arn
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-gwlb-listener"
  })
}

# ============================================================================
# VPC ENDPOINTS FOR GWLB
# ============================================================================

# VPC Endpoint Service for GWLB
resource "aws_vpc_endpoint_service" "gwlb" {
  acceptance_required        = false
  gateway_load_balancer_arns = [aws_lb.gwlb.arn]

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-gwlb-endpoint-service"
  })
}

# VPC Endpoints in Main VPC Public Subnets
resource "aws_vpc_endpoint" "gwlb_public" {
  count = length(var.public_subnet_cidrs)

  vpc_id              = aws_vpc.main.id
  service_name        = aws_vpc_endpoint_service.gwlb.service_name
  vpc_endpoint_type   = "GatewayLoadBalancer"
  subnet_ids          = [aws_subnet.public[count.index].id]

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-gwlb-endpoint-public-${count.index + 1}"
  })
}

# VPC Endpoints in Main VPC Private Subnets
resource "aws_vpc_endpoint" "gwlb_private" {
  count = length(var.private_subnet_cidrs)

  vpc_id              = aws_vpc.main.id
  service_name        = aws_vpc_endpoint_service.gwlb.service_name
  vpc_endpoint_type   = "GatewayLoadBalancer"
  subnet_ids          = [aws_subnet.private[count.index].id]

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-gwlb-endpoint-private-${count.index + 1}"
  })
}

# ============================================================================
# FIREWALL VPC ROUTING
# ============================================================================

# Internet Gateway for Firewall VPC (for return traffic)
resource "aws_internet_gateway" "firewall" {
  vpc_id = aws_vpc.firewall.id

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-firewall-igw"
  })
}

# Route Table for Firewall Subnets
resource "aws_route_table" "firewall_subnet" {
  count = length(var.firewall_subnet_cidrs)

  vpc_id = aws_vpc.firewall.id

  # Route to Internet Gateway for return traffic
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.firewall.id
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-firewall-rt-${count.index + 1}"
  })
}

# Route Table for GWLB Subnets
resource "aws_route_table" "gwlb_subnet" {
  count = length(var.gwlb_subnet_cidrs)

  vpc_id = aws_vpc.firewall.id

  # Route to Internet Gateway for return traffic
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.firewall.id
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-gwlb-rt-${count.index + 1}"
  })
}

# Route Table Associations for Firewall Subnets
resource "aws_route_table_association" "firewall_subnet" {
  count = length(var.firewall_subnet_cidrs)

  subnet_id      = aws_subnet.firewall[count.index].id
  route_table_id = aws_route_table.firewall_subnet[count.index].id
}

# Route Table Associations for GWLB Subnets
resource "aws_route_table_association" "gwlb_subnet" {
  count = length(var.gwlb_subnet_cidrs)

  subnet_id      = aws_subnet.gwlb[count.index].id
  route_table_id = aws_route_table.gwlb_subnet[count.index].id
}

