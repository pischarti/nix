
resource "aws_networkfirewall_rule_group" "default_rule_group" {
  capacity = 100
  name     = "${local.name}-default-stateless-rule-group"
  type     = "STATELESS"
  rule_group {
    rules_source {
      stateless_rules_and_custom_actions {
        stateless_rule {
          priority = 1
          rule_definition {
            actions = ["aws:forward_to_sfe"]
            # actions = ["aws:drop"]
            match_attributes {
              source {
                address_definition = "0.0.0.0/0"
              }
              source {
                address_definition = "0.0.0.0/0"
              }
            }
          }
        }
      }
    }
  }
  tags = merge(
    {
      Name = "${local.name}-default-stateless-rule-group"
    },
    var.tags
  )
}

resource "aws_networkfirewall_firewall_policy" "default_policy" {
  name = "default-policy"
  firewall_policy {
    stateless_default_actions          = ["aws:forward_to_sfe"]
    stateless_fragment_default_actions = ["aws:forward_to_sfe"]
    stateless_rule_group_reference {
      priority     = 1
      resource_arn = aws_networkfirewall_rule_group.default_rule_group.arn
    }
  }
  tags = merge(
    {
      Name = "${local.name}-default-policy"
    },
    var.tags
  )
}

resource "aws_networkfirewall_firewall" "aws_network_firewall" {
  name                = "${local.name}-aws-network-firewall"
  description         = "AWS Network Firewall to test the traffic"
  firewall_policy_arn = aws_networkfirewall_firewall_policy.default_policy.arn
  vpc_id              = module.inspection_vpc.vpc_id
  subnet_mapping {
    subnet_id = aws_subnet.inspection_vpc_firewall_subnet.id
  }
  tags = merge(
    {
      Name = "${local.name}-aws-network-firewall"
    },
    var.tags
  )
}
