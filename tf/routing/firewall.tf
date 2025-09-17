
# ============================================================================
# NETWORK FIREWALL VPC
# ============================================================================

# Network Firewall VPC
resource "aws_vpc" "firewall" {
  cidr_block           = var.firewall_vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(var.tags, {
    Name = "${var.firewall_vpc_name}"
  })
}

# Firewall Subnets (for Network Firewall endpoints)
resource "aws_subnet" "firewall" {
  count = length(var.firewall_subnet_cidrs)

  vpc_id            = aws_vpc.firewall.id
  cidr_block        = var.firewall_subnet_cidrs[count.index]
  availability_zone = var.availability_zones[count.index]

  tags = merge(var.tags, {
    Name = "${var.firewall_vpc_name}-firewall-subnet-${count.index + 1}"
    Type = "Firewall"
  })
}

# Transit Gateway Subnets (for TGW attachments)
resource "aws_subnet" "firewall_tgw" {
  count = length(var.firewall_tgw_subnet_cidrs)

  vpc_id            = aws_vpc.firewall.id
  cidr_block        = var.firewall_tgw_subnet_cidrs[count.index]
  availability_zone = var.availability_zones[count.index]

  tags = merge(var.tags, {
    Name = "${var.firewall_vpc_name}-tgw-subnet-${count.index + 1}"
    Type = "TransitGateway"
  })
}

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
  description         = "AWS Network Firewall for traffic inspection"
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

# Locals for firewall endpoints
locals {
  firewall_endpoint_ids = flatten([
    for status in aws_networkfirewall_firewall.main.firewall_status :
    [for sync_state in status.sync_states : sync_state.attachment[0].endpoint_id]
  ])
}

# Route Tables for Firewall VPC

# Route Table for Firewall Subnets
resource "aws_route_table" "firewall_subnet" {
  count = length(var.firewall_subnet_cidrs)

  vpc_id = aws_vpc.firewall.id

  # Route traffic back to TGW after inspection
  route {
    cidr_block         = var.vpc_cidr
    transit_gateway_id = aws_ec2_transit_gateway.main.id
  }

  tags = merge(var.tags, {
    Name = "${var.firewall_vpc_name}-firewall-rt-${count.index + 1}"
  })

  depends_on = [aws_ec2_transit_gateway.main]
}

# Route Table for TGW Subnets
resource "aws_route_table" "firewall_tgw" {
  count = length(var.firewall_tgw_subnet_cidrs)

  vpc_id = aws_vpc.firewall.id

  # Route traffic to firewall endpoint
  route {
    cidr_block      = "0.0.0.0/0"
    vpc_endpoint_id = local.firewall_endpoint_ids[count.index]
  }

  tags = merge(var.tags, {
    Name = "${var.firewall_vpc_name}-tgw-rt-${count.index + 1}"
  })

  depends_on = [aws_networkfirewall_firewall.main]
}

# Route Table Associations for Firewall Subnets
resource "aws_route_table_association" "firewall_subnet" {
  count = length(var.firewall_subnet_cidrs)

  subnet_id      = aws_subnet.firewall[count.index].id
  route_table_id = aws_route_table.firewall_subnet[count.index].id
}

# Route Table Associations for TGW Subnets
resource "aws_route_table_association" "firewall_tgw" {
  count = length(var.firewall_tgw_subnet_cidrs)

  subnet_id      = aws_subnet.firewall_tgw[count.index].id
  route_table_id = aws_route_table.firewall_tgw[count.index].id
}

