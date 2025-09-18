
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
