# Network Firewall Subnets (3 AZs required)
resource "aws_subnet" "firewall_subnet_1" {
  vpc_id            = aws_vpc.inspection.id
  cidr_block        = var.firewall_subnet_1_cidr
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = merge(
    {
      Name = "${var.inspection_vpc_name}-firewall-subnet-1"
    },
    var.tags
  )
}

resource "aws_subnet" "firewall_subnet_2" {
  vpc_id            = aws_vpc.inspection.id
  cidr_block        = var.firewall_subnet_2_cidr
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = merge(
    {
      Name = "${var.inspection_vpc_name}-firewall-subnet-2"
    },
    var.tags
  )
}

resource "aws_subnet" "firewall_subnet_3" {
  vpc_id            = aws_vpc.inspection.id
  cidr_block        = var.firewall_subnet_3_cidr
  availability_zone = data.aws_availability_zones.available.names[2]

  tags = merge(
    {
      Name = "${var.inspection_vpc_name}-firewall-subnet-3"
    },
    var.tags
  )
}

# Data source for availability zones
data "aws_availability_zones" "available" {
  state = "available"
}

# Network Firewall Policy
resource "aws_networkfirewall_firewall_policy" "main" {
  name = "${var.inspection_vpc_name}-firewall-policy"

  firewall_policy {
    stateless_default_actions          = ["aws:forward_to_sfe"]
    stateless_fragment_default_actions = ["aws:forward_to_sfe"]

    stateless_rule_group_reference {
      priority     = 100
      resource_arn = aws_networkfirewall_rule_group.stateless.arn
    }
  }

  tags = merge(
    {
      Name = "${var.inspection_vpc_name}-firewall-policy"
    },
    var.tags
  )
}

# Stateless Rule Group (block edge to app traffic, allow others)
resource "aws_networkfirewall_rule_group" "stateless" {
  capacity    = 100
  name        = "${var.inspection_vpc_name}-stateless-rules"
  type        = "STATELESS"
  description = "Stateless rule group for traffic inspection - blocks edge to app traffic"

  rule_group {
    rules_source {
      stateless_rules_and_custom_actions {
        # Rule 1: Block traffic between edge VPC to app VPC
        # stateless_rule {
        #   priority = 1
        #   rule_definition {
        #     actions = ["aws:drop"]
        #     match_attributes {
        #       protocols = [6] # TCP
        #       source {
        #         # address_definition = var.vpc_cidr_edge
        #         address_definition = var.vpc_cidr_app
        #       }
        #       destination {
        #         # address_definition = var.vpc_cidr_app
        #         address_definition = var.vpc_cidr_edge
        #       }
        #     }
        #   }
        # }
        
        # Rule 2: Allow all other traffic
        stateless_rule {
          # priority = 2
          priority = 1
          rule_definition {
            actions = ["aws:forward_to_sfe"]
            match_attributes {
              protocols = [6] # TCP
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

  tags = merge(
    {
      Name = "${var.inspection_vpc_name}-stateless-rules"
    },
    var.tags
  )
}

# Network Firewall
resource "aws_networkfirewall_firewall" "main" {
  name                = "${var.inspection_vpc_name}-firewall"
  firewall_policy_arn = aws_networkfirewall_firewall_policy.main.arn
  vpc_id              = aws_vpc.inspection.id

  subnet_mapping {
    subnet_id = aws_subnet.firewall_subnet_1.id
  }

  subnet_mapping {
    subnet_id = aws_subnet.firewall_subnet_2.id
  }

  subnet_mapping {
    subnet_id = aws_subnet.firewall_subnet_3.id
  }

  tags = merge(
    {
      Name = "${var.inspection_vpc_name}-firewall"
    },
    var.tags
  )

  depends_on = [
    aws_networkfirewall_firewall_policy.main
  ]
}

# Firewall Route Table
resource "aws_route_table" "firewall" {
  vpc_id = aws_vpc.inspection.id

  tags = merge(
    {
      Name = "${var.inspection_vpc_name}-firewall-rt"
    },
    var.tags
  )
}

# Data source to get firewall endpoint IDs
data "aws_networkfirewall_firewall" "main" {
  arn = aws_networkfirewall_firewall.main.arn
}

# Local value to get the VPC endpoint ID
locals {
  firewall_endpoint_id = try(tolist(data.aws_networkfirewall_firewall.main.firewall_status[0].sync_states)[0].attachment[0].endpoint_id, null)
}

# Route from firewall to TGW for inter-VPC traffic
resource "aws_route" "firewall_to_tgw_edge" {
  route_table_id         = aws_route_table.firewall.id
  destination_cidr_block = var.vpc_cidr_edge
  transit_gateway_id     = aws_ec2_transit_gateway.main.id
}

resource "aws_route" "firewall_to_tgw_app" {
  route_table_id         = aws_route_table.firewall.id
  destination_cidr_block = var.vpc_cidr_app
  transit_gateway_id     = aws_ec2_transit_gateway.main.id
}

# Route from firewall to internet
resource "aws_route" "firewall_to_internet" {
  route_table_id         = aws_route_table.firewall.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.inspection.id
}

# Associate firewall subnets with firewall route table
resource "aws_route_table_association" "firewall_subnet_1" {
  subnet_id      = aws_subnet.firewall_subnet_1.id
  route_table_id = aws_route_table.firewall.id
}

resource "aws_route_table_association" "firewall_subnet_2" {
  subnet_id      = aws_subnet.firewall_subnet_2.id
  route_table_id = aws_route_table.firewall.id
}

resource "aws_route_table_association" "firewall_subnet_3" {
  subnet_id      = aws_subnet.firewall_subnet_3.id
  route_table_id = aws_route_table.firewall.id
}
