

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

output "edge_generated_key_pair_name" {
  description = "Name of the generated EC2 key pair"
  value       = aws_key_pair.edge_generated.key_name
}

output "edge_generated_public_key" {
  description = "Public key of the generated key pair"
  value       = tls_private_key.edge.public_key_openssh
}

output "edge_generated_private_key_pem" {
  description = "PEM-encoded private key for the generated key pair"
  value       = tls_private_key.edge.private_key_pem
  sensitive   = true
}

