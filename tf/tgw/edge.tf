resource "aws_vpc" "edge" {
  cidr_block           = var.vpc_cidr_edge
  enable_dns_hostnames = var.enable_dns_hostnames
  enable_dns_support   = var.enable_dns_support

  tags = merge(
    {
      Name = var.edge_vpc_name
    },
    var.tags
  )
}

resource "aws_internet_gateway" "edge" {
  vpc_id = aws_vpc.edge.id

  tags = merge(
    {
      Name = "${var.edge_vpc_name}-igw"
    },
    var.tags
  )
}

resource "aws_subnet" "edge_public" {
  vpc_id                  = aws_vpc.edge.id
  cidr_block              = var.edge_public_subnet_cidr
  map_public_ip_on_launch = true

  tags = merge(
    {
      Name = "${var.edge_vpc_name}-public-subnet"
    },
    var.tags
  )
}

resource "aws_route_table" "edge_public" {
  vpc_id = aws_vpc.edge.id

  tags = merge(
    {
      Name = "${var.edge_vpc_name}-public-rt"
    },
    var.tags
  )
}

resource "aws_route" "edge_public_internet_access" {
  route_table_id         = aws_route_table.edge_public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.edge.id
}

resource "aws_route" "edge_public_to_inspection" {
  route_table_id         = aws_route_table.edge_public.id
  destination_cidr_block = var.vpc_cidr_inspection
  transit_gateway_id     = aws_ec2_transit_gateway.main.id
}

resource "aws_route_table_association" "edge_public" {
  subnet_id      = aws_subnet.edge_public.id
  route_table_id = aws_route_table.edge_public.id
}

resource "aws_security_group" "edge_public" {
  name        = "${var.edge_vpc_name}-public-sg"
  description = "Allow SSH and egress to internet"
  vpc_id      = aws_vpc.edge.id

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
    ipv6_cidr_blocks = ["::/0"]
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
      Name = "${var.edge_vpc_name}-public-ssh-sg"
    },
    var.tags
  )
}

resource "aws_instance" "edge_public" {
  ami                         = data.aws_ami.amazon_linux_2.id
  instance_type               = var.tgw_instance_type
  subnet_id                   = aws_subnet.edge_public.id
  vpc_security_group_ids      = [aws_security_group.edge_public.id]
  associate_public_ip_address = true
  key_name                    = coalesce(var.edge_key_name, aws_key_pair.edge_generated.key_name)
  user_data_replace_on_change = true

  user_data = <<-EOF
              #!/bin/bash
              set -euxo pipefail
              sudo yum update -y || true
              sudo yum install -y httpd
              sudo systemctl enable httpd
              SVC=httpd
              echo "<h1>Server Details</h1><p><strong>Hostname:</strong> $(hostname)</p>" | sudo tee /var/www/html/index.html           
              sudo systemctl restart "$SVC"
              EOF

  tags = merge(
    {
      Name = "${var.edge_vpc_name}-public-ec2"
    },
    var.tags
  )

  depends_on = [
    aws_route_table_association.edge_public,
    aws_security_group.edge_public
  ]
}

resource "tls_private_key" "edge" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "edge_generated" {
  key_name   = "${var.edge_vpc_name}-generated"
  public_key = tls_private_key.edge.public_key_openssh
}
