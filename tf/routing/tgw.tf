# ============================================================================
# TRANSIT GATEWAY
# ============================================================================

# Transit Gateway
resource "aws_ec2_transit_gateway" "main" {
  description                     = var.tgw_description
  default_route_table_association = "disable"
  default_route_table_propagation = "disable"
  dns_support                     = "enable"
  vpn_ecmp_support               = "enable"

  tags = merge(var.tags, {
    Name = var.tgw_name
  })
}

# ============================================================================
# MAIN VPC TGW SUBNETS
# ============================================================================

# Transit Gateway Subnets in Main VPC
resource "aws_subnet" "main_vpc_tgw" {
  count = length(var.main_vpc_tgw_subnet_cidrs)

  vpc_id            = aws_vpc.main.id
  cidr_block        = var.main_vpc_tgw_subnet_cidrs[count.index]
  availability_zone = var.availability_zones[count.index]

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-main-tgw-subnet-${count.index + 1}"
    Type = "TransitGateway"
  })
}

# ============================================================================
# TGW VPC ATTACHMENTS
# ============================================================================

# Main VPC Attachment
resource "aws_ec2_transit_gateway_vpc_attachment" "main_vpc" {
  subnet_ids         = aws_subnet.main_vpc_tgw[*].id
  transit_gateway_id = aws_ec2_transit_gateway.main.id
  vpc_id             = aws_vpc.main.id

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-main-vpc-attachment"
  })
}

# Firewall VPC Attachment
resource "aws_ec2_transit_gateway_vpc_attachment" "firewall_vpc" {
  subnet_ids         = aws_subnet.firewall_tgw[*].id
  transit_gateway_id = aws_ec2_transit_gateway.main.id
  vpc_id             = aws_vpc.firewall.id

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-firewall-vpc-attachment"
  })
}

# ============================================================================
# TGW ROUTE TABLES
# ============================================================================

# Route Table for Main VPC (sends traffic to firewall)
resource "aws_ec2_transit_gateway_route_table" "main_vpc" {
  transit_gateway_id = aws_ec2_transit_gateway.main.id

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-main-vpc-rt"
  })
}

# Route Table for Firewall VPC (sends traffic back to main VPC)
resource "aws_ec2_transit_gateway_route_table" "firewall_vpc" {
  transit_gateway_id = aws_ec2_transit_gateway.main.id

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-firewall-vpc-rt"
  })
}

# ============================================================================
# TGW ROUTE TABLE ASSOCIATIONS
# ============================================================================

# Associate Main VPC attachment with main route table
resource "aws_ec2_transit_gateway_route_table_association" "main_vpc" {
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.main_vpc.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.main_vpc.id
}

# Associate Firewall VPC attachment with firewall route table
resource "aws_ec2_transit_gateway_route_table_association" "firewall_vpc" {
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.firewall_vpc.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.firewall_vpc.id
}

# ============================================================================
# TGW ROUTES
# ============================================================================

# Route from Main VPC to Firewall VPC (for inspection)
resource "aws_ec2_transit_gateway_route" "main_to_firewall" {
  destination_cidr_block         = var.private_subnet_cidrs[0] # First private subnet as example
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.firewall_vpc.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.main_vpc.id
}

# Route from Firewall VPC back to Main VPC (after inspection)
resource "aws_ec2_transit_gateway_route" "firewall_to_main" {
  destination_cidr_block         = var.vpc_cidr
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.main_vpc.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.firewall_vpc.id
}

# ============================================================================
# UPDATED ROUTE TABLES FOR TGW INTEGRATION
# ============================================================================

# Route Table for Main VPC TGW Subnets
resource "aws_route_table" "main_vpc_tgw" {
  count = length(var.main_vpc_tgw_subnet_cidrs)

  vpc_id = aws_vpc.main.id

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-main-tgw-rt-${count.index + 1}"
  })
}

# Route Table Associations for Main VPC TGW Subnets
resource "aws_route_table_association" "main_vpc_tgw" {
  count = length(var.main_vpc_tgw_subnet_cidrs)

  subnet_id      = aws_subnet.main_vpc_tgw[count.index].id
  route_table_id = aws_route_table.main_vpc_tgw[count.index].id
}
