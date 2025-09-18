

# ==============================================================================
# TEST INFRASTRUCTURE - EC2 Instance and Network Load Balancer
# ==============================================================================

# Key Pair for EC2 instances
resource "tls_private_key" "test_key" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "aws_key_pair" "test_key" {
  key_name   = "${var.tags.Name}-test-key"
  public_key = tls_private_key.test_key.public_key_openssh

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-key"
  })
}

# Security Group for Test EC2 Instance in Private Subnet
resource "aws_security_group" "test_private_instance" {
  name_prefix = "${var.tags.Name}-test-private-"
  vpc_id      = aws_vpc.main.id
  description = "Security group for test instance in private subnet"

  # Allow HTTP traffic from NLB
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = var.public_subnet_cidrs
    description = "HTTP from public subnets (NLB)"
  }

  # Allow HTTPS traffic from NLB
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = var.public_subnet_cidrs
    description = "HTTPS from public subnets (NLB)"
  }

  # Allow SSH from public subnets (for management)
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = var.public_subnet_cidrs
    description = "SSH from public subnets"
  }

  # Allow all outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "All outbound traffic"
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-private-sg"
  })
}

# Security Group for Network Load Balancer
resource "aws_security_group" "test_nlb" {
  name_prefix = "${var.tags.Name}-test-nlb-"
  vpc_id      = aws_vpc.main.id
  description = "Security group for test Network Load Balancer"

  # Allow HTTP traffic from internet
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTP from internet"
  }

  # Allow HTTPS traffic from internet
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTPS from internet"
  }

  # Allow all outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "All outbound traffic"
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-nlb-sg"
  })
}

# Get latest Amazon Linux 2023 AMI
data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-*-x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# Test EC2 Instance in Private Subnet
resource "aws_instance" "test_private" {
  ami                    = data.aws_ami.amazon_linux.id
  instance_type          = var.test_instance_type
  key_name              = aws_key_pair.test_key.key_name
  subnet_id             = aws_subnet.private[0].id
  vpc_security_group_ids = [aws_security_group.test_private_instance.id]

  user_data = base64encode(templatefile("${path.module}/user_data.sh", {
    instance_name = "${var.tags.Name}-test-private"
  }))

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-private-instance"
    Type = "Test"
  })

  depends_on = [aws_nat_gateway.main]
}

# Network Load Balancer in Public Subnets
resource "aws_lb" "test_nlb" {
  name               = "${var.tags.Name}-test-nlb"
  internal           = false
  load_balancer_type = "network"
  subnets            = aws_subnet.public[*].id

  enable_deletion_protection = false

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-nlb"
    Type = "Test"
  })
}

# Target Group for HTTP traffic
resource "aws_lb_target_group" "test_http" {
  name        = "${var.tags.Name}-test-http"
  port        = 80
  protocol    = "TCP"
  vpc_id      = aws_vpc.main.id
  target_type = "instance"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 6
    interval            = 30
    protocol            = "TCP"
    port                = "traffic-port"
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-http-tg"
  })
}

# Target Group for HTTPS traffic
resource "aws_lb_target_group" "test_https" {
  name        = "${var.tags.Name}-test-https"
  port        = 443
  protocol    = "TCP"
  vpc_id      = aws_vpc.main.id
  target_type = "instance"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 3
    timeout             = 10
    interval            = 30
    protocol            = "HTTP"
    port                = "80"
    path                = "/health"
    matcher             = "200"
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-https-tg"
  })
}

# Target Group Attachments
resource "aws_lb_target_group_attachment" "test_http" {
  target_group_arn = aws_lb_target_group.test_http.arn
  target_id        = aws_instance.test_private.id
  port             = 80
}

resource "aws_lb_target_group_attachment" "test_https" {
  target_group_arn = aws_lb_target_group.test_https.arn
  target_id        = aws_instance.test_private.id
  port             = 443
}

# NLB Listener for HTTP
resource "aws_lb_listener" "test_http" {
  load_balancer_arn = aws_lb.test_nlb.arn
  port              = "80"
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test_http.arn
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-http-listener"
  })
}

# NLB Listener for HTTPS
resource "aws_lb_listener" "test_https" {
  load_balancer_arn = aws_lb.test_nlb.arn
  port              = "443"
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test_https.arn
  }

  tags = merge(var.tags, {
    Name = "${var.tags.Name}-test-https-listener"
  })
}

