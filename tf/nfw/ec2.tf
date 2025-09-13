data "aws_ami" "amazon_linux_2" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "tls_private_key" "ssh_generated" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "ssh_generated" {
  key_name   = var.app_key_name
  public_key = tls_private_key.ssh_generated.public_key_openssh
}

resource "aws_instance" "app" {
  ami                         = data.aws_ami.amazon_linux_2.id
  instance_type               = var.tgw_instance_type
  subnet_id                   = aws_subnet.app_vpc_workload_subnet.id
  vpc_security_group_ids      = [aws_security_group.app.id]
  associate_public_ip_address = true
  key_name                    = coalesce(var.app_key_name, aws_key_pair.ssh_generated.key_name)

  tags = merge(
    {
      Name = "${var.app_vpc_name}-app-ec2"
    },
    var.tags
  )

  depends_on = [
    module.app_vpc.app_vpc_workload_subnet,
    aws_security_group.app
  ]
}

resource "aws_instance" "egress" {
  ami                         = data.aws_ami.amazon_linux_2.id
  instance_type               = var.tgw_instance_type
  subnet_id                   = aws_subnet.egress_vpc_igw_subnet.id
  vpc_security_group_ids      = [aws_security_group.egress.id]
  associate_public_ip_address = true
  key_name                    = coalesce(var.app_key_name, aws_key_pair.ssh_generated.key_name)

  tags = merge(
    {
      Name = "${var.app_vpc_name}-egress-ec2"
    },
    var.tags
  )

  depends_on = [
    module.egress_vpc.egress_vpc_workload_subnet,
    aws_security_group.egress
  ]
}
