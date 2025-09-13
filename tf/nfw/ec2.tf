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
      Name = "${local.name}-app-ec2"
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
      Name = "${local.name}-egress-ec2"
    },
    var.tags
  )

  depends_on = [
    module.egress_vpc.egress_vpc_workload_subnet,
    aws_security_group.egress
  ]
}
