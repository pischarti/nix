

output "edge_vpc_id" {
  description = "The ID of the edge VPC"
  value       = aws_vpc.edge.id
}

output "edge_vpc_arn" {
  description = "The ARN of the edge VPC"
  value       = aws_vpc.edge.arn
}

output "edge_vpc_cidr_block" {
  description = "The CIDR block of the edge VPC"
  value       = aws_vpc.edge.cidr_block
}

output "edge_igw_id" {
  description = "The ID of the internet gateway for the edge VPC"
  value       = aws_internet_gateway.edge.id
}

output "edge_public_subnet_id" {
  description = "The ID of the public subnet in the edge VPC"
  value       = aws_subnet.edge_public.id
}

output "edge_public_route_table_id" {
  description = "The ID of the public route table in the edge VPC"
  value       = aws_route_table.edge_public.id
}

output "edge_public_instance_id" {
  description = "The ID of the EC2 instance in the edge public subnet"
  value       = aws_instance.edge_public.id
}

output "edge_public_instance_public_ip" {
  description = "The public IP of the EC2 instance in the edge public subnet"
  value       = aws_instance.edge_public.public_ip
}

output "edge_public_instance_private_ip" {
  description = "The private IP of the EC2 instance in the edge public subnet"
  value       = aws_instance.edge_public.private_ip
}

output "ssh_generated_key_pair_name" {
  description = "Name of the generated EC2 key pair"
  value       = aws_key_pair.ssh_generated.key_name
}

output "ssh_generated_public_key" {
  description = "Public key of the generated key pair"
  value       = tls_private_key.ssh_generated.public_key_openssh
}

output "ssh_generated_private_key_pem" {
  description = "PEM-encoded private key for the generated key pair"
  value       = tls_private_key.ssh_generated.private_key_pem
  sensitive   = true
}

# Transit Gateway Outputs
output "tgw_id" {
  description = "The ID of the Transit Gateway"
  value       = aws_ec2_transit_gateway.main.id
}

output "tgw_arn" {
  description = "The ARN of the Transit Gateway"
  value       = aws_ec2_transit_gateway.main.arn
}

output "tgw_edge_attachment_id" {
  description = "The ID of the edge VPC attachment to TGW"
  value       = aws_ec2_transit_gateway_vpc_attachment.edge.id
}

output "tgw_inspection_attachment_id" {
  description = "The ID of the inspection VPC attachment to TGW"
  value       = aws_ec2_transit_gateway_vpc_attachment.inspection.id
}

# Inspection VPC Outputs
output "inspection_vpc_id" {
  description = "The ID of the inspection VPC"
  value       = aws_vpc.inspection.id
}

output "inspection_public_instance_id" {
  description = "The ID of the EC2 instance in the inspection public subnet"
  value       = aws_instance.inspection_public.id
}

output "inspection_public_instance_public_ip" {
  description = "The public IP of the EC2 instance in the inspection public subnet"
  value       = aws_instance.inspection_public.public_ip
}

output "inspection_public_instance_private_ip" {
  description = "The private IP of the EC2 instance in the inspection public subnet"
  value       = aws_instance.inspection_public.private_ip
}

# App VPC Outputs
output "app_vpc_id" {
  description = "The ID of the app VPC"
  value       = aws_vpc.app.id
}

output "app_public_instance_id" {
  description = "The ID of the EC2 instance in the app public subnet"
  value       = aws_instance.app_public.id
}

output "app_public_instance_public_ip" {
  description = "The public IP of the EC2 instance in the app public subnet"
  value       = aws_instance.app_public.public_ip
}

output "app_public_instance_private_ip" {
  description = "The private IP of the EC2 instance in the app public subnet"
  value       = aws_instance.app_public.private_ip
}

output "tgw_app_attachment_id" {
  description = "The ID of the app VPC attachment to TGW"
  value       = aws_ec2_transit_gateway_vpc_attachment.app.id
}

# Network Firewall Outputs
output "network_firewall_id" {
  description = "The ID of the Network Firewall"
  value       = aws_networkfirewall_firewall.main.id
}

output "network_firewall_arn" {
  description = "The ARN of the Network Firewall"
  value       = aws_networkfirewall_firewall.main.arn
}

output "firewall_subnet_ids" {
  description = "The IDs of the firewall subnets"
  value       = [
    aws_subnet.firewall_subnet_1.id,
    aws_subnet.firewall_subnet_2.id,
    aws_subnet.firewall_subnet_3.id
  ]
}

