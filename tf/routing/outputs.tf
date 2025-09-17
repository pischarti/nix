# VPC Outputs
output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.main.id
}

output "vpc_cidr_block" {
  description = "CIDR block of the VPC"
  value       = aws_vpc.main.cidr_block
}

# Internet Gateway
output "internet_gateway_id" {
  description = "ID of the Internet Gateway"
  value       = aws_internet_gateway.main.id
}

# Public Subnet Outputs
output "public_subnet_ids" {
  description = "IDs of the public subnets"
  value       = aws_subnet.public[*].id
}

output "public_subnet_cidrs" {
  description = "CIDR blocks of the public subnets"
  value       = aws_subnet.public[*].cidr_block
}

# Private Subnet Outputs
output "private_subnet_ids" {
  description = "IDs of the private subnets"
  value       = aws_subnet.private[*].id
}

output "private_subnet_cidrs" {
  description = "CIDR blocks of the private subnets"
  value       = aws_subnet.private[*].cidr_block
}

# NAT Gateway Outputs
output "nat_gateway_ids" {
  description = "IDs of the NAT Gateways"
  value       = aws_nat_gateway.main[*].id
}

output "nat_gateway_public_ips" {
  description = "Public IPs of the NAT Gateways"
  value       = aws_eip.nat[*].public_ip
}

# Route Table Outputs
output "public_route_table_id" {
  description = "ID of the public route table"
  value       = aws_route_table.public.id
}

output "private_route_table_ids" {
  description = "IDs of the private route tables"
  value       = aws_route_table.private[*].id
}

# ============================================================================
# NETWORK FIREWALL VPC OUTPUTS
# ============================================================================

# Firewall VPC Outputs
output "firewall_vpc_id" {
  description = "ID of the Network Firewall VPC"
  value       = aws_vpc.firewall.id
}

output "firewall_vpc_cidr_block" {
  description = "CIDR block of the Network Firewall VPC"
  value       = aws_vpc.firewall.cidr_block
}

# Firewall Subnet Outputs
output "firewall_subnet_ids" {
  description = "IDs of the Network Firewall subnets"
  value       = aws_subnet.firewall[*].id
}

output "firewall_subnet_cidrs" {
  description = "CIDR blocks of the Network Firewall subnets"
  value       = aws_subnet.firewall[*].cidr_block
}

# TGW Subnet Outputs for Firewall VPC
output "firewall_tgw_subnet_ids" {
  description = "IDs of the Transit Gateway subnets in firewall VPC"
  value       = aws_subnet.firewall_tgw[*].id
}

output "firewall_tgw_subnet_cidrs" {
  description = "CIDR blocks of the Transit Gateway subnets in firewall VPC"
  value       = aws_subnet.firewall_tgw[*].cidr_block
}

# Network Firewall Outputs
output "network_firewall_id" {
  description = "ID of the Network Firewall"
  value       = aws_networkfirewall_firewall.main.id
}

output "network_firewall_arn" {
  description = "ARN of the Network Firewall"
  value       = aws_networkfirewall_firewall.main.arn
}

output "network_firewall_endpoint_ids" {
  description = "IDs of the Network Firewall endpoints"
  value       = local.firewall_endpoint_ids
}

output "network_firewall_policy_arn" {
  description = "ARN of the Network Firewall policy"
  value       = aws_networkfirewall_firewall_policy.policy.arn
}

output "network_firewall_rule_group_arn" {
  description = "ARN of the Network Firewall rule group"
  value       = aws_networkfirewall_rule_group.stateless_rule_group.arn
}

# Firewall Route Table Outputs
output "firewall_subnet_route_table_ids" {
  description = "IDs of the firewall subnet route tables"
  value       = aws_route_table.firewall_subnet[*].id
}

output "firewall_tgw_route_table_ids" {
  description = "IDs of the firewall TGW subnet route tables"
  value       = aws_route_table.firewall_tgw[*].id
}

# ============================================================================
# TRANSIT GATEWAY OUTPUTS
# ============================================================================

# Transit Gateway Outputs
output "transit_gateway_id" {
  description = "ID of the Transit Gateway"
  value       = aws_ec2_transit_gateway.main.id
}

output "transit_gateway_arn" {
  description = "ARN of the Transit Gateway"
  value       = aws_ec2_transit_gateway.main.arn
}

# Main VPC TGW Subnet Outputs
output "main_vpc_tgw_subnet_ids" {
  description = "IDs of the main VPC Transit Gateway subnets"
  value       = aws_subnet.main_vpc_tgw[*].id
}

output "main_vpc_tgw_subnet_cidrs" {
  description = "CIDR blocks of the main VPC Transit Gateway subnets"
  value       = aws_subnet.main_vpc_tgw[*].cidr_block
}

# TGW VPC Attachment Outputs
output "main_vpc_tgw_attachment_id" {
  description = "ID of the main VPC Transit Gateway attachment"
  value       = aws_ec2_transit_gateway_vpc_attachment.main_vpc.id
}

output "firewall_vpc_tgw_attachment_id" {
  description = "ID of the firewall VPC Transit Gateway attachment"
  value       = aws_ec2_transit_gateway_vpc_attachment.firewall_vpc.id
}

# TGW Route Table Outputs
output "main_vpc_tgw_route_table_id" {
  description = "ID of the main VPC Transit Gateway route table"
  value       = aws_ec2_transit_gateway_route_table.main_vpc.id
}

output "firewall_vpc_tgw_route_table_id" {
  description = "ID of the firewall VPC Transit Gateway route table"
  value       = aws_ec2_transit_gateway_route_table.firewall_vpc.id
}

# Main VPC TGW Route Table Outputs
output "main_vpc_tgw_subnet_route_table_ids" {
  description = "IDs of the main VPC TGW subnet route tables"
  value       = aws_route_table.main_vpc_tgw[*].id
}
