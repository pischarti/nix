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
  backend "s3" {
    bucket  = "terraform-state-aws-us-east-1"
    key     = "network-firewall/gwlb/terraform.tfstate"
    region  = "us-east-1"
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
  name     = "${var.tags.Name}-stateless"
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

# ==============================================================================
# TEST INFRASTRUCTURE - EC2 Instance and Network Load Balancer
# ==============================================================================

# Key Pair for EC2 instances
resource "tls_private_key" "test_key" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "aws_key_pair" "test_key" {
  key_name   = "${var.tags.Name}-test-key"
  public_key = tls_private_key.test_key.public_key_openssh

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-key"
  })
}

# Security Group for Test EC2 Instance in Private Subnet
resource "aws_security_group" "test_private_instance" {
  name_prefix = "${var.tags.Name}-test-private-"
  vpc_id      = aws_vpc.main.id
  description = "Security group for test instance in private subnet"

  # Allow HTTP traffic from NLB
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = var.public_subnet_cidrs
    description = "HTTP from public subnets (NLB)"
  }

  # Allow HTTPS traffic from NLB
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = var.public_subnet_cidrs
    description = "HTTPS from public subnets (NLB)"
  }

  # Allow SSH from public subnets (for management)
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = var.public_subnet_cidrs
    description = "SSH from public subnets"
  }

  # Allow all outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "All outbound traffic"
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-private-sg"
  })
}

# Security Group for Network Load Balancer
resource "aws_security_group" "test_nlb" {
  name_prefix = "${var.tags.Name}-test-nlb-"
  vpc_id      = aws_vpc.main.id
  description = "Security group for test Network Load Balancer"

  # Allow HTTP traffic from internet
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTP from internet"
  }

  # Allow HTTPS traffic from internet
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTPS from internet"
  }

  # Allow all outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "All outbound traffic"
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-nlb-sg"
  })
}

# Get latest Amazon Linux 2023 AMI
data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-*-x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# Test EC2 Instance in Private Subnet
resource "aws_instance" "test_private" {
  ami                    = data.aws_ami.amazon_linux.id
  instance_type          = var.test_instance_type
  key_name              = aws_key_pair.test_key.key_name
  subnet_id             = aws_subnet.private[0].id
  vpc_security_group_ids = [aws_security_group.test_private_instance.id]

  user_data = base64encode(templatefile("${path.module}/user_data.sh", {
    instance_name = "${var.tags.Name}-test-private"
  }))

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-private-instance"
    Type = "Test"
  })

  depends_on = [aws_nat_gateway.main]
}

# Network Load Balancer in Public Subnets
resource "aws_lb" "test_nlb" {
  name               = "${var.tags.Name}-test-nlb"
  internal           = false
  load_balancer_type = "network"
  subnets            = aws_subnet.public[*].id

  enable_deletion_protection = false

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-nlb"
    Type = "Test"
  })
}

# Target Group for HTTP traffic
resource "aws_lb_target_group" "test_http" {
  name        = "${var.tags.Name}-test-http"
  port        = 80
  protocol    = "TCP"
  vpc_id      = aws_vpc.main.id
  target_type = "instance"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 6
    interval            = 30
    protocol            = "TCP"
    port                = "traffic-port"
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-http-tg"
  })
}

# Target Group for HTTPS traffic
resource "aws_lb_target_group" "test_https" {
  name        = "${var.tags.Name}-test-https"
  port        = 443
  protocol    = "TCP"
  vpc_id      = aws_vpc.main.id
  target_type = "instance"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 3
    timeout             = 10
    interval            = 30
    protocol            = "HTTP"
    port                = "80"
    path                = "/health"
    matcher             = "200"
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-https-tg"
  })
}

# Target Group Attachments
resource "aws_lb_target_group_attachment" "test_http" {
  target_group_arn = aws_lb_target_group.test_http.arn
  target_id        = aws_instance.test_private.id
  port             = 80
}

resource "aws_lb_target_group_attachment" "test_https" {
  target_group_arn = aws_lb_target_group.test_https.arn
  target_id        = aws_instance.test_private.id
  port             = 443
}

# NLB Listener for HTTP
resource "aws_lb_listener" "test_http" {
  load_balancer_arn = aws_lb.test_nlb.arn
  port              = "80"
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test_http.arn
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-http-listener"
  })
}

# NLB Listener for HTTPS
resource "aws_lb_listener" "test_https" {
  load_balancer_arn = aws_lb.test_nlb.arn
  port              = "443"
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test_https.arn
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-https-listener"
  })
}

