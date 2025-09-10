resource "aws_vpc" "app" {
  cidr_block           = var.vpc_cidr_app
  enable_dns_hostnames = var.enable_dns_hostnames
  enable_dns_support   = var.enable_dns_support

  tags = merge(
    {
      Name = var.app_vpc_name
    },
    var.tags
  )
}

resource "aws_internet_gateway" "app" {
  vpc_id = aws_vpc.app.id

  tags = merge(
    {
      Name = "${var.app_vpc_name}-igw"
    },
    var.tags
  )
}

resource "aws_subnet" "app_public" {
  vpc_id                  = aws_vpc.app.id
  cidr_block              = var.app_public_subnet_cidr
  map_public_ip_on_launch = true

  tags = merge(
    {
      Name = "${var.app_vpc_name}-public-subnet"
    },
    var.tags
  )
}

resource "aws_route_table" "app_public" {
  vpc_id = aws_vpc.app.id

  tags = merge(
    {
      Name = "${var.app_vpc_name}-public-rt"
    },
    var.tags
  )
}

resource "aws_route" "app_public_internet_access" {
  route_table_id         = aws_route_table.app_public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.app.id
}

resource "aws_route" "app_public_to_inspection" {
  route_table_id         = aws_route_table.app_public.id
  destination_cidr_block = var.vpc_cidr_inspection
  transit_gateway_id     = aws_ec2_transit_gateway.main.id
}

resource "aws_route" "app_public_to_edge" {
  route_table_id         = aws_route_table.app_public.id
  destination_cidr_block = var.vpc_cidr_edge
  transit_gateway_id     = aws_ec2_transit_gateway.main.id
}

resource "aws_route_table_association" "app_public" {
  subnet_id      = aws_subnet.app_public.id
  route_table_id = aws_route_table.app_public.id
}

resource "aws_security_group" "app_public" {
  name        = "${var.app_vpc_name}-public-sg"
  description = "Allow SSH, HTTP and egress to internet"
  vpc_id      = aws_vpc.app.id

  ingress {
    description      = "SSH from allowed CIDR"
    from_port        = 22
    to_port          = 22
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
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

  ingress {
    description      = "HTTP from edge VPC"
    from_port        = 80
    to_port          = 80
    protocol         = "tcp"
    cidr_blocks      = [var.vpc_cidr_edge]
    ipv6_cidr_blocks = []
  }

  ingress {
    description      = "HTTP from inspection VPC"
    from_port        = 80
    to_port          = 80
    protocol         = "tcp"
    cidr_blocks      = [var.vpc_cidr_inspection]
    ipv6_cidr_blocks = []
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
      Name = "${var.app_vpc_name}-public-sg"
    },
    var.tags
  )
}

resource "aws_instance" "app_public" {
  ami                         = data.aws_ami.amazon_linux_2.id
  instance_type               = var.tgw_instance_type
  subnet_id                   = aws_subnet.app_public.id
  vpc_security_group_ids      = [aws_security_group.app_public.id]
  associate_public_ip_address = true
  key_name                    = coalesce(var.app_key_name, aws_key_pair.edge_generated.key_name)
  user_data_replace_on_change = true

  user_data = <<-EOF
              #!/bin/bash
              set -euxo pipefail
              sudo yum update -y || true
              sudo yum install -y httpd
              sudo systemctl enable httpd
              SVC=httpd
              PRIVATE_IP=$(curl -s http://169.254.169.254/latest/meta-data/local-ipv4)
              PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)
              echo "<h1>App Server Details</h1><p><strong>Hostname:</strong> $(hostname)</p><p><strong>VPC:</strong> App VPC</p><p><strong>Private IP:</strong> $PRIVATE_IP</p><p><strong>Public IP:</strong> $PUBLIC_IP</p>" | sudo tee /var/www/html/index.html           
              sudo systemctl restart "$SVC"
              EOF

  tags = merge(
    {
      Name = "${var.app_vpc_name}-public-ec2"
    },
    var.tags
  )

  depends_on = [
    aws_route_table_association.app_public,
    aws_security_group.app_public
  ]
}
