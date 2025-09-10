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

resource "aws_route" "inspection_public_internet_access" {
  route_table_id         = aws_route_table.inspection_public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.inspection.id
}

# Route ALL traffic from inspection VPC through firewall
resource "aws_route" "inspection_public_to_firewall_all" {
  route_table_id         = aws_route_table.inspection_public.id
  destination_cidr_block = "0.0.0.0/0"
  vpc_endpoint_id        = local.firewall_endpoint_id

  depends_on = [aws_networkfirewall_firewall.main]
}

# Route traffic from inspection VPC directly to edge VPC (bypassing firewall)
# resource "aws_route" "inspection_public_to_edge" {
#   route_table_id         = aws_route_table.inspection_public.id
#   destination_cidr_block = var.vpc_cidr_edge
#   vpc_endpoint_id        = local.firewall_endpoint_id

#   depends_on = [aws_networkfirewall_firewall.main]
# }

# All traffic now goes through firewall - no direct routes needed

resource "aws_route_table_association" "inspection_public" {
  subnet_id      = aws_subnet.inspection_public.id
  route_table_id = aws_route_table.inspection_public.id
}

resource "aws_security_group" "inspection_public_egress" {
  name        = "${var.inspection_vpc_name}-public-sg"
  description = "Allow egress to internet"
  vpc_id      = aws_vpc.inspection.id

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  ingress {
    description      = "HTTP from anywhere"
    from_port        = 80
    to_port          = 80
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
  }

  ingress {
    description      = "SSH from allowed CIDR"
    from_port        = 22
    to_port          = 22
    protocol         = "tcp"
    cidr_blocks      = [var.edge_ssh_ingress_cidr]
    ipv6_cidr_blocks = []
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
    description      = "HTTP from app VPC"
    from_port        = 80
    to_port          = 80
    protocol         = "tcp"
    cidr_blocks      = [var.vpc_cidr_app]
    ipv6_cidr_blocks = []
  }  

  tags = merge(
    {
      Name = "${var.inspection_vpc_name}-public-egress-sg"
    },
    var.tags
  )
}

resource "aws_instance" "inspection_public" {
  ami                         = data.aws_ami.amazon_linux_2.id
  instance_type               = var.tgw_instance_type
  subnet_id                   = aws_subnet.inspection_public.id
  vpc_security_group_ids      = [aws_security_group.inspection_public_egress.id]
  associate_public_ip_address = true
  user_data_replace_on_change = true

  user_data = <<-EOF
              #!/bin/bash
              set -euxo pipefail
              sudo yum update -y || true
              sudo yum install -y httpd
              sudo systemctl enable httpd
              SVC=httpd
              echo "<h1>Server Details: Inspection</h1><p><strong>Hostname:</strong> $(hostname)</p>" | sudo tee /var/www/html/index.html           
              sudo systemctl restart "$SVC"
              EOF

  tags = merge(
    {
      Name = "${var.inspection_vpc_name}-public-ec2"
    },
    var.tags
  )

  depends_on = [
    aws_route_table_association.inspection_public
  ]
}
