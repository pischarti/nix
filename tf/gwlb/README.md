# Gateway Load Balancer with Network Firewall for Traffic Inspection

This Terraform configuration creates an AWS infrastructure that uses Gateway Load Balancer (GWLB) with AWS Network Firewall to inspect traffic between public and private subnets in a VPC.

## Architecture Overview

The solution consists of:

1. **Main VPC**: Contains public and private subnets for your applications
2. **Firewall VPC**: Contains AWS Network Firewall and Gateway Load Balancer
3. **Traffic Inspection**: All traffic between public and private subnets is routed through the firewall for inspection

## Components

### Main VPC
- **Public Subnets**: Internet-facing subnets with Internet Gateway access
- **Private Subnets**: Internal subnets with NAT Gateway for outbound access
- **GWLB Endpoints**: VPC endpoints that route traffic to the firewall VPC

### Firewall VPC
- **Network Firewall Subnets**: Host AWS Network Firewall endpoints
- **GWLB Subnets**: Host the Gateway Load Balancer
- **Internet Gateway**: For return traffic routing

### Network Security
- **AWS Network Firewall**: Provides stateless packet filtering
- **Gateway Load Balancer**: Distributes traffic to firewall endpoints
- **VPC Endpoints**: Enable transparent traffic routing to GWLB

## Traffic Flow

### Public to Private Traffic
1. Traffic from public subnet destined for private subnet
2. Route table directs traffic to GWLB endpoint in public subnet
3. GWLB endpoint forwards traffic to firewall VPC
4. Network Firewall inspects and processes traffic
5. Inspected traffic is returned to private subnet

### Private to Public Traffic
1. Traffic from private subnet destined for public subnet
2. Route table directs traffic to GWLB endpoint in private subnet
3. GWLB endpoint forwards traffic to firewall VPC
4. Network Firewall inspects and processes traffic
5. Inspected traffic is returned to public subnet

## Deployment

### Prerequisites
- AWS CLI configured with appropriate permissions
- Terraform >= 1.4.0

### Variables
Configure the following variables in `terraform.tfvars` or use defaults:

```hcl
aws_region = "us-east-1"

tags = {
  Name        = "gwlb-firewall"
  Environment = "dev"
  Project     = "network-security"
}

# Main VPC Configuration
vpc_cidr              = "10.0.0.0/16"
public_subnet_cidrs   = ["10.0.1.0/24", "10.0.2.0/24"]
private_subnet_cidrs  = ["10.0.10.0/24", "10.0.20.0/24"]
availability_zones    = ["us-east-1a", "us-east-1b"]

# Firewall VPC Configuration
firewall_vpc_cidr     = "10.1.0.0/16"
firewall_subnet_cidrs = ["10.1.1.0/24", "10.1.2.0/24"]
gwlb_subnet_cidrs     = ["10.1.10.0/24", "10.1.20.0/24"]
```

### Deploy
```bash
terraform init
terraform plan
terraform apply
```

### Post-Deployment: Register Network Firewall Endpoints

Due to the dynamic nature of Network Firewall endpoint creation, the firewall endpoints must be registered with the GWLB target group after the initial deployment.

#### Option 1: Two-Stage Terraform Deployment (Recommended)
1. Deploy the initial infrastructure:
```bash
terraform apply
```

2. Discover the firewall endpoint IPs:
```bash
ENDPOINT_IPS=$(./register_firewall_endpoints.sh)
echo "Discovered endpoints: $ENDPOINT_IPS"
```

3. Re-apply Terraform with the discovered IPs:
```bash
terraform apply -var="firewall_endpoint_ips=$ENDPOINT_IPS"
```

#### Option 2: Manual AWS CLI Registration
1. Discover endpoint IPs using the script:
```bash
ENDPOINT_IPS=$(./register_firewall_endpoints.sh)
TARGET_GROUP_ARN=$(terraform output -raw gwlb_target_group_arn)
```

2. Register each endpoint:
```bash
for ip in $(echo $ENDPOINT_IPS | jq -r '.[]'); do
  aws elbv2 register-targets \
    --target-group-arn $TARGET_GROUP_ARN \
    --targets Id=$ip,Port=6081
done
```

3. Verify target health:
```bash
aws elbv2 describe-target-health --target-group-arn $TARGET_GROUP_ARN
```

#### Option 3: Direct Variable Setting
If you already know the endpoint IPs, you can set them directly:
```bash
terraform apply -var='firewall_endpoint_ips=["10.1.1.100","10.1.2.100"]'
```

## Testing the Infrastructure

### Automated Traffic Testing

The infrastructure includes a complete test setup with an EC2 instance in a private subnet and a Network Load Balancer for testing traffic flow through the firewall:

```bash
# Run the comprehensive traffic test
./test_firewall_traffic.sh
```

This script will:
- Test HTTP and HTTPS traffic flow through the firewall
- Validate that traffic reaches the private instance
- Perform performance testing
- Provide troubleshooting information

### Test Infrastructure Components

- **EC2 Instance**: Amazon Linux 2023 instance in private subnet running HTTP (Apache) and HTTPS (Nginx) servers
- **Network Load Balancer**: External-facing NLB in public subnets that forwards traffic to the private instance
- **Security Groups**: Properly configured to allow traffic flow while maintaining security
- **Key Pair**: Auto-generated SSH key pair for instance access

### Manual Testing

You can also test manually using the terraform outputs:

```bash
# Get the NLB DNS name
NLB_DNS=$(terraform output -raw test_nlb_dns_name)

# Test HTTP traffic
curl http://$NLB_DNS

# Test HTTPS traffic (self-signed certificate)
curl -k https://$NLB_DNS

# Test health check endpoint
curl http://$NLB_DNS/health
```

### Traffic Flow Analysis

The test setup validates this traffic flow:
1. **Internet** → Network Load Balancer (Public Subnet)
2. **NLB** → Gateway Load Balancer Endpoint
3. **GWLB** → Network Firewall (Inspection VPC)
4. **Firewall** → GWLB → Private Subnet
5. **Private Subnet** → EC2 Instance (Web Server)

### Infrastructure Validation

For basic infrastructure validation, you can also use:

```bash
./test_traffic_inspection.sh
```

This script provides:
- Infrastructure status overview
- Routing table information
- GWLB endpoint status
- Manual testing instructions

## Security Considerations

### Network Firewall Rules
The default configuration includes basic stateless rules that forward all traffic to the stateful rule engine. You can customize the firewall rules by:

1. Adding more rule groups
2. Implementing stateful rules
3. Adding domain-based filtering
4. Implementing threat intelligence feeds

### High Availability
- Resources are deployed across multiple Availability Zones
- GWLB automatically distributes traffic across healthy firewall endpoints
- Network Firewall endpoints are automatically scaled

### Monitoring
Consider implementing:
- CloudWatch metrics for GWLB and Network Firewall
- VPC Flow Logs for traffic analysis
- AWS Config for compliance monitoring

## Costs

Key cost factors:
- Gateway Load Balancer: Hourly charges and data processing charges
- Network Firewall: Hourly charges and data processing charges
- VPC Endpoints: Hourly charges and data processing charges
- NAT Gateway: Hourly charges and data processing charges

## Cleanup

To destroy the infrastructure:
```bash
terraform destroy
```

## Troubleshooting

### Common Issues

1. **"No Network Firewall endpoints found" Error**
   - **Cause**: Network Firewall endpoints take time to create after deployment
   - **Solution**: 
     ```bash
     # Run the debug script for detailed information
     ./debug_firewall_endpoints.sh
     
     # Wait longer and retry
     ./register_firewall_endpoints.sh
     
     # Or check firewall status manually
     aws network-firewall describe-firewall --firewall-name <firewall-name>
     ```

2. **Network Firewall Not Ready**
   - **Cause**: Firewall is still in "PROVISIONING" state
   - **Solution**: Wait 5-10 minutes for deployment to complete
   - **Check Status**: 
     ```bash
     aws network-firewall describe-firewall --firewall-name $(terraform output -raw network_firewall_name)
     ```

3. **Connectivity Issues**
   - Verify GWLB endpoints are in "available" state
   - Check route table configurations
   - Ensure security groups allow required traffic

4. **Firewall Not Inspecting Traffic**
   - Verify Network Firewall is in "READY" state
   - Check firewall policy and rules
   - Confirm target group health checks are passing

5. **High Latency**
   - Review firewall rules complexity
   - Consider stateless vs stateful rule performance
   - Monitor GWLB target group health

### Useful Commands

```bash
# Check GWLB status
aws elbv2 describe-load-balancers --names gwlb-firewall-gwlb

# Check Network Firewall status
aws network-firewall describe-firewall --firewall-name gwlb-firewall-network-firewall

# Check VPC endpoints
aws ec2 describe-vpc-endpoints --filters Name=service-name,Values=com.amazonaws.vpce.*
```

## References

- [AWS Gateway Load Balancer Documentation](https://docs.aws.amazon.com/elasticloadbalancing/latest/gateway/)
- [AWS Network Firewall Documentation](https://docs.aws.amazon.com/network-firewall/)
- [VPC Endpoints Documentation](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-endpoints.html)
