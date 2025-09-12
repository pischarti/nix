resource "aws_vpc" "inspection" {
  cidr_block           = var.vpc_cidr_inspection
  enable_dns_hostnames = var.enable_dns_hostnames
  enable_dns_support   = var.enable_dns_support

  tags = merge(
    {
      Name = var.inspection_vpc_name
    },
    var.tags
  )
}

resource "aws_internet_gateway" "inspection" {
  vpc_id = aws_vpc.inspection.id

  tags = merge(
    {
      Name = "${var.inspection_vpc_name}-igw"
    },
    var.tags
  )
}

resource "aws_subnet" "inspection_public" {
  vpc_id                  = aws_vpc.inspection.id
  cidr_block              = var.inspection_public_subnet_cidr
  map_public_ip_on_launch = true

  tags = merge(
    {
      Name = "${var.inspection_vpc_name}-public-subnet"
    },
    var.tags
  )
}

resource "aws_route_table" "inspection_public" {
  vpc_id = aws_vpc.inspection.id

  tags = merge(
    {
      Name = "${var.inspection_vpc_name}-public-rt"
    },
    var.tags
  )
}

# resource "aws_route" "inspection_public_internet_access" {
#   route_table_id         = aws_route_table.inspection_public.id
#   destination_cidr_block = "0.0.0.0/0"
#   gateway_id             = aws_internet_gateway.inspection.id
# }

# Route traffic from inspection VPC to firewall endpoints (more specific routes first)
# resource "aws_route" "inspection_public_to_firewall_edge" {
#   route_table_id         = aws_route_table.inspection_public.id
#   destination_cidr_block = var.vpc_cidr_edge
#   vpc_endpoint_id        = local.firewall_endpoint_ids
# }

# resource "aws_route" "inspection_public_to_firewall_app" {
#   route_table_id         = aws_route_table.inspection_public.id
#   destination_cidr_block = var.vpc_cidr_app
#   vpc_endpoint_id        = local.firewall_endpoint_ids
# }

# # Route other traffic to TGW (less specific route)
# resource "aws_route" "inspection_public_to_tgw" {
#   route_table_id         = aws_route_table.inspection_public.id
#   destination_cidr_block = "0.0.0.0/0"
#   transit_gateway_id     = aws_ec2_transit_gateway.main.id
# }

# Route all traffic to NFW
resource "aws_route" "inspection_public_to_tgw" {
  route_table_id         = aws_route_table.inspection_public.id
  destination_cidr_block = "0.0.0.0/0"
  vpc_endpoint_id        = local.firewall_endpoint_ids  
}

resource "aws_route_table_association" "inspection_public" {
  subnet_id      = aws_subnet.inspection_public.id
  route_table_id = aws_route_table.inspection_public.id
}

resource "aws_security_group" "inspection_public" {
  name        = "${var.inspection_vpc_name}-public-sg"
  description = "Allow egress to internet"
  vpc_id      = aws_vpc.inspection.id

  ingress {
    description      = "SSH from allowed CIDR"
    from_port        = 22
    to_port          = 22
    protocol         = "tcp"
    cidr_blocks      = [var.edge_ssh_ingress_cidr]
    ipv6_cidr_blocks = []
  }

  ingress {
    description      = "HTTP from anywhere"
    from_port        = 80
    to_port          = 80
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }


  tags = merge(
    {
      Name = "${var.inspection_vpc_name}-public-sg"
    },
    var.tags
  )
}

resource "aws_instance" "inspection_public" {
  ami                         = data.aws_ami.amazon_linux_2.id
  instance_type               = var.tgw_instance_type
  subnet_id                   = aws_subnet.inspection_public.id
  vpc_security_group_ids      = [aws_security_group.inspection_public.id]
  associate_public_ip_address = true
  user_data_replace_on_change = true
  key_name                    = coalesce(var.edge_key_name, aws_key_pair.ssh_generated.key_name)

  user_data = <<-EOF
              #!/bin/bash
              set -euxo pipefail
              sudo yum update -y || true
              sudo yum install -y httpd
              sudo systemctl enable httpd
              SVC=httpd
              PRIVATE_IP=$(curl -s http://169.254.169.254/latest/meta-data/local-ipv4)
              PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)
              echo "<h1>Inspection Server Details</h1><p><strong>Hostname:</strong> $(hostname)</p><p><strong>Private IP:</strong> $PRIVATE_IP</p><p><strong>Public IP:</strong> $PUBLIC_IP</p>" | sudo tee /var/www/html/index.html           
              sudo systemctl restart "$SVC"
              EOF

  tags = merge(
    {
      Name = "${var.inspection_vpc_name}-public-ec2"
    },
    var.tags
  )

  depends_on = [
    aws_route_table_association.inspection_public,
    aws_security_group.inspection_public
  ]
}
