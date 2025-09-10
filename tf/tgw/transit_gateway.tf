# Transit Gateway
resource "aws_ec2_transit_gateway" "main" {
  description                     = "Transit Gateway for edge to inspection routing"
  default_route_table_association = "disable"
  default_route_table_propagation = "disable"
  auto_accept_shared_attachments  = "disable"
  dns_support                     = "enable"
  vpn_ecmp_support               = "enable"

  tags = merge(
    {
      Name = "main-tgw"
    },
    var.tags
  )
}

# TGW Route Tables
resource "aws_ec2_transit_gateway_route_table" "edge" {
  transit_gateway_id = aws_ec2_transit_gateway.main.id

  tags = merge(
    {
      Name = "edge-tgw-rt"
    },
    var.tags
  )
}

resource "aws_ec2_transit_gateway_route_table" "inspection" {
  transit_gateway_id = aws_ec2_transit_gateway.main.id

  tags = merge(
    {
      Name = "inspection-tgw-rt"
    },
    var.tags
  )
}

# TGW Attachments
resource "aws_ec2_transit_gateway_vpc_attachment" "edge" {
  subnet_ids                                      = [aws_subnet.edge_public.id]
  transit_gateway_id                              = aws_ec2_transit_gateway.main.id
  vpc_id                                          = aws_vpc.edge.id
  transit_gateway_default_route_table_association = false
  transit_gateway_default_route_table_propagation = false

  tags = merge(
    {
      Name = "edge-tgw-attachment"
    },
    var.tags
  )
}

resource "aws_ec2_transit_gateway_vpc_attachment" "inspection" {
  subnet_ids                                      = [aws_subnet.inspection_public.id]
  transit_gateway_id                              = aws_ec2_transit_gateway.main.id
  vpc_id                                          = aws_vpc.inspection.id
  transit_gateway_default_route_table_association = false
  transit_gateway_default_route_table_propagation = false

  tags = merge(
    {
      Name = "inspection-tgw-attachment"
    },
    var.tags
  )
}

# TGW Route Table Associations
resource "aws_ec2_transit_gateway_route_table_association" "edge" {
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.edge.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.edge.id
}

resource "aws_ec2_transit_gateway_route_table_association" "inspection" {
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.inspection.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.inspection.id
}

# TGW Routes
resource "aws_ec2_transit_gateway_route" "edge_to_inspection" {
  destination_cidr_block         = var.vpc_cidr_inspection
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.inspection.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.edge.id
}

resource "aws_ec2_transit_gateway_route" "inspection_to_edge" {
  destination_cidr_block         = var.vpc_cidr_edge
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.edge.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.inspection.id
}
