
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
