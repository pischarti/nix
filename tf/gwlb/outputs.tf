# ============================================================================
# MAIN VPC OUTPUTS
# ============================================================================

output "main_vpc_id" {
  description = "ID of the main VPC"
  value       = aws_vpc.main.id
}

output "main_vpc_cidr" {
  description = "CIDR block of the main VPC"
  value       = aws_vpc.main.cidr_block
}

output "public_subnet_ids" {
  description = "IDs of the public subnets"
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "IDs of the private subnets"
  value       = aws_subnet.private[*].id
}

# ============================================================================
# FIREWALL VPC OUTPUTS
# ============================================================================

output "firewall_vpc_id" {
  description = "ID of the firewall VPC"
  value       = aws_vpc.firewall.id
}

output "firewall_vpc_cidr" {
  description = "CIDR block of the firewall VPC"
  value       = aws_vpc.firewall.cidr_block
}

output "firewall_subnet_ids" {
  description = "IDs of the firewall subnets"
  value       = aws_subnet.firewall[*].id
}

output "gwlb_subnet_ids" {
  description = "IDs of the GWLB subnets"
  value       = aws_subnet.gwlb[*].id
}

# ============================================================================
# NETWORK FIREWALL OUTPUTS
# ============================================================================

output "network_firewall_name" {
  description = "Name of the Network Firewall"
  value       = aws_networkfirewall_firewall.main.name
}

output "network_firewall_id" {
  description = "ID of the Network Firewall"
  value       = aws_networkfirewall_firewall.main.id
}

output "network_firewall_arn" {
  description = "ARN of the Network Firewall"
  value       = aws_networkfirewall_firewall.main.arn
}

output "firewall_subnet_ids_for_manual_registration" {
  description = "Firewall subnet IDs - use these to find and manually register firewall endpoint IPs with the GWLB target group"
  value       = aws_subnet.firewall[*].id
}

output "firewall_endpoint_registration_note" {
  description = "Instructions for registering Network Firewall endpoints with GWLB"
  value = <<-EOT
    Network Firewall endpoints must be registered manually with the GWLB target group.
    
    Steps:
    1. Wait for Network Firewall to be fully deployed
    2. Find the Network Firewall endpoint IPs in each firewall subnet
    3. Register these IPs with the GWLB target group: ${aws_lb_target_group.firewall_endpoints.arn}
    
    Use AWS CLI:
    aws elbv2 register-targets --target-group-arn ${aws_lb_target_group.firewall_endpoints.arn} --targets Id=<ENDPOINT_IP>,Port=6081
  EOT
}

# ============================================================================
# GATEWAY LOAD BALANCER OUTPUTS
# ============================================================================

output "gwlb_id" {
  description = "ID of the Gateway Load Balancer"
  value       = aws_lb.gwlb.id
}

output "gwlb_arn" {
  description = "ARN of the Gateway Load Balancer"
  value       = aws_lb.gwlb.arn
}

output "gwlb_dns_name" {
  description = "DNS name of the Gateway Load Balancer"
  value       = aws_lb.gwlb.dns_name
}

output "gwlb_target_group_arn" {
  description = "ARN of the GWLB target group"
  value       = aws_lb_target_group.firewall_endpoints.arn
}

output "gwlb_endpoint_service_name" {
  description = "Service name of the GWLB endpoint service"
  value       = aws_vpc_endpoint_service.gwlb.service_name
}

# ============================================================================
# VPC ENDPOINT OUTPUTS
# ============================================================================

output "gwlb_endpoint_public_ids" {
  description = "IDs of the GWLB endpoints in public subnets"
  value       = aws_vpc_endpoint.gwlb_public[*].id
}

output "gwlb_endpoint_private_ids" {
  description = "IDs of the GWLB endpoints in private subnets"
  value       = aws_vpc_endpoint.gwlb_private[*].id
}

# ============================================================================
# ROUTING OUTPUTS
# ============================================================================

output "public_route_table_id" {
  description = "ID of the public route table"
  value       = aws_route_table.public.id
}

output "private_route_table_ids" {
  description = "IDs of the private route tables"
  value       = aws_route_table.private[*].id
}

output "firewall_route_table_ids" {
  description = "IDs of the firewall subnet route tables"
  value       = aws_route_table.firewall_subnet[*].id
}

output "gwlb_route_table_ids" {
  description = "IDs of the GWLB subnet route tables"
  value       = aws_route_table.gwlb_subnet[*].id
}

# ==============================================================================
# TEST INFRASTRUCTURE OUTPUTS
# ==============================================================================

output "test_instance_id" {
  description = "ID of the test EC2 instance in private subnet"
  value       = aws_instance.test_private.id
}

output "test_instance_private_ip" {
  description = "Private IP address of the test EC2 instance"
  value       = aws_instance.test_private.private_ip
}

output "test_nlb_dns_name" {
  description = "DNS name of the test Network Load Balancer"
  value       = aws_lb.test_nlb.dns_name
}

output "test_nlb_zone_id" {
  description = "Zone ID of the test Network Load Balancer"
  value       = aws_lb.test_nlb.zone_id
}

output "test_nlb_arn" {
  description = "ARN of the test Network Load Balancer"
  value       = aws_lb.test_nlb.arn
}

output "test_endpoints" {
  description = "Test endpoints for traffic validation"
  value = {
    http_endpoint  = "http://${aws_lb.test_nlb.dns_name}"
    https_endpoint = "https://${aws_lb.test_nlb.dns_name}"
    health_check   = "http://${aws_lb.test_nlb.dns_name}/health"
  }
}

output "test_key_pair" {
  description = "Information about the test key pair"
  value = {
    key_name    = aws_key_pair.test_key.key_name
    fingerprint = aws_key_pair.test_key.fingerprint
  }
}

output "test_private_key" {
  description = "Private key for SSH access to test instance (sensitive)"
  value       = tls_private_key.test_key.private_key_pem
  sensitive   = true
}

output "test_security_groups" {
  description = "Security group IDs for test infrastructure"
  value = {
    private_instance_sg = aws_security_group.test_private_instance.id
    nlb_sg              = aws_security_group.test_nlb.id
  }
}

output "traffic_flow_summary" {
  description = "Summary of expected traffic flow for testing"
  value = {
    description = "Traffic flow through Gateway Load Balancer and Network Firewall"
    flow_steps = [
      "1. Internet → Network Load Balancer (Public Subnet)",
      "2. NLB → Gateway Load Balancer Endpoint",
      "3. GWLB → Network Firewall (Inspection VPC)",
      "4. Firewall → GWLB → Private Subnet",
      "5. Private Subnet → EC2 Instance (Web Server)"
    ]
    test_commands = [
      "curl http://${aws_lb.test_nlb.dns_name}",
      "curl -k https://${aws_lb.test_nlb.dns_name}",
      "curl http://${aws_lb.test_nlb.dns_name}/health"
    ]
  }
}
