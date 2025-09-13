# EC2 Instance Outputs
output "app_instance_id" {
  description = "ID of the App VPC EC2 instance"
  value       = aws_instance.app.id
}

output "app_instance_private_ip" {
  description = "Private IP address of the App VPC EC2 instance"
  value       = aws_instance.app.private_ip
}

output "app_instance_public_ip" {
  description = "Public IP address of the App VPC EC2 instance"
  value       = aws_instance.app.public_ip
}

output "egress_instance_id" {
  description = "ID of the Egress VPC EC2 instance"
  value       = aws_instance.egress.id
}

output "egress_instance_private_ip" {
  description = "Private IP address of the Egress VPC EC2 instance"
  value       = aws_instance.egress.private_ip
}

output "egress_instance_public_ip" {
  description = "Public IP address of the Egress VPC EC2 instance"
  value       = aws_instance.egress.public_ip
}

# SSH Key Output
output "ssh_private_key" {
  description = "SSH private key for accessing EC2 instances"
  value       = tls_private_key.ssh_generated.private_key_pem
  sensitive   = true
}

output "ssh_public_key" {
  description = "SSH public key for accessing EC2 instances"
  value       = tls_private_key.ssh_generated.public_key_openssh
}

# VPC Information
output "app_vpc_id" {
  description = "ID of the App VPC"
  value       = module.app_vpc.vpc_id
}

output "egress_vpc_id" {
  description = "ID of the Egress VPC"
  value       = module.egress_vpc.vpc_id
}

output "inspection_vpc_id" {
  description = "ID of the Inspection VPC"
  value       = module.inspection_vpc.vpc_id
}

# Transit Gateway Information
output "transit_gateway_id" {
  description = "ID of the Transit Gateway"
  value       = aws_ec2_transit_gateway.tgw.id
}

# Network Firewall Information
output "network_firewall_arn" {
  description = "ARN of the Network Firewall"
  value       = aws_networkfirewall_firewall.aws_network_firewall.arn
}

