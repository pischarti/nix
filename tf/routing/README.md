# Network Routing with Transit Gateway and AWS Network Firewall

This Terraform configuration creates a comprehensive network architecture with two VPCs connected via Transit Gateway, featuring mandatory traffic inspection through AWS Network Firewall.

## 🚀 Transit Gateway Architecture with Network Firewall Inspection

### **🌐 Traffic Flow Architecture**

```
Public Subnet → TGW → Firewall VPC → Network Firewall → TGW → Private Subnet
```

### **🏗️ Infrastructure Components**

#### **🔄 Transit Gateway (`aws_ec2_transit_gateway.main`)**
- **Centralized routing hub** connecting main VPC and firewall VPC
- **Custom route tables** for granular traffic control
- **Disabled default propagation** for security control
- **DNS and VPN ECMP support** enabled

#### **🔌 VPC Attachments**
- **Main VPC Attachment**: Connects to dedicated TGW subnets
- **Firewall VPC Attachment**: Connects to firewall TGW subnets
- **Multi-AZ deployment** for high availability

#### **🛣️ Smart Routing Configuration**

##### **Main VPC Public Subnets:**
- **Internet traffic** → Internet Gateway
- **Private subnet traffic** → Transit Gateway (for firewall inspection)

##### **Firewall VPC:**
- **TGW subnets** → Route traffic to Network Firewall endpoints
- **Firewall subnets** → Route inspected traffic back to TGW

##### **TGW Route Tables:**
- **Main VPC Route Table**: Routes private subnet traffic to firewall VPC
- **Firewall VPC Route Table**: Routes inspected traffic back to main VPC

## 📊 Network Architecture

### **Main VPC (`10.0.0.0/16`)**
- **Public Subnets**: `10.0.1.0/24`, `10.0.2.0/24`
- **Private Subnets**: `10.0.10.0/24`, `10.0.20.0/24`
- **TGW Subnets**: `10.0.100.0/24`, `10.0.200.0/24`
- **Internet Gateway** for public internet access
- **NAT Gateways** in each public subnet for private subnet outbound traffic

### **Firewall VPC (`10.1.0.0/16`)**
- **Firewall Subnets**: `10.1.1.0/24`, `10.1.2.0/24`
- **TGW Subnets**: `10.1.10.0/24`, `10.1.20.0/24`
- **AWS Network Firewall** with stateless rule groups
- **Multi-AZ deployment** across `us-east-1a` and `us-east-1b`

## 🔒 Security Features

### **🛡️ AWS Network Firewall Components**
- **Stateless Rule Group**: Forward all traffic to stateful engine
- **Firewall Policy**: Configures default actions and rule group references
- **Network Firewall**: Main firewall resource with multi-AZ deployment
- **Smart Routing**: TGW subnets route traffic through firewall endpoints

### **Security Benefits:**
1. **🛡️ Mandatory Inspection**: All east-west traffic inspected by Network Firewall
2. **🔍 Centralized Policy**: Single firewall for all inter-subnet communication
3. **📊 Visibility**: Complete traffic monitoring and logging
4. **🚫 Threat Prevention**: Real-time threat detection and blocking
5. **🔧 Granular Control**: Customizable rules per traffic type

## 📁 File Structure

```
tf/routing/
├── main.tf                 # Main VPC with public/private subnets
├── firewall.tf            # Network Firewall VPC and resources
├── tgw.tf                 # Transit Gateway configuration
├── variables.tf           # All variable definitions
├── outputs.tf             # Resource outputs
├── env/
│   └── default.tfvars     # Configuration values
└── README.md              # This file
```

## 🎯 Key Architecture Features

- **🔄 Inspection Enforcement**: Traffic cannot bypass firewall
- **⚡ High Performance**: Multi-AZ Network Firewall deployment
- **📈 Scalable**: Easy to add more VPCs to the architecture
- **🏷️ Well-Tagged**: Consistent naming and tagging
- **🔧 Configurable**: Customizable routing and firewall rules

## 📋 Traffic Flow Summary

1. **📤 Outbound from Public**: Direct to Internet via IGW
2. **🔄 Public to Private**: Public → TGW → Firewall → Inspection → TGW → Private
3. **🔒 Private to Public**: Private → TGW → Firewall → Inspection → TGW → Public
4. **🛡️ All Inter-VPC**: Mandatory firewall inspection via TGW

## 🚀 Deployment

### Prerequisites
- AWS CLI configured with appropriate permissions
- Terraform >= 1.4.0 installed

### Deploy the Infrastructure

```bash
# Navigate to the routing directory
cd tf/routing

# Initialize Terraform
terraform init

# Review the plan
terraform plan -var-file="env/default.tfvars"

# Apply the configuration
terraform apply -var-file="env/default.tfvars"
```

### Configuration Variables

Key variables can be customized in `env/default.tfvars`:

```hcl
# Main VPC Configuration
vpc_cidr = "10.0.0.0/16"
public_subnet_cidrs = ["10.0.1.0/24", "10.0.2.0/24"]
private_subnet_cidrs = ["10.0.10.0/24", "10.0.20.0/24"]
main_vpc_tgw_subnet_cidrs = ["10.0.100.0/24", "10.0.200.0/24"]

# Firewall VPC Configuration
firewall_vpc_cidr = "10.1.0.0/16"
firewall_subnet_cidrs = ["10.1.1.0/24", "10.1.2.0/24"]
firewall_tgw_subnet_cidrs = ["10.1.10.0/24", "10.1.20.0/24"]

# Transit Gateway Configuration
tgw_name = "network-routing-tgw"
tgw_description = "Transit Gateway for routing traffic through Network Firewall inspection"
```

## 🔧 Customization

### Adding More VPCs
1. Create additional VPC resources
2. Add TGW attachments for new VPCs
3. Update TGW route tables to include new VPC CIDRs
4. Configure firewall rules as needed

### Modifying Firewall Rules
Update the `aws_networkfirewall_rule_group` resource in `firewall.tf` to customize:
- Stateless rules
- Stateful rules
- Custom actions
- Rule priorities

### Scaling Considerations
- **Network Firewall**: Automatically scales within capacity limits
- **Transit Gateway**: Supports up to 5,000 VPC attachments
- **Subnets**: Can be expanded or additional subnets added per AZ

## 🏷️ Resource Tagging

All resources are consistently tagged with:
- **Name**: Descriptive resource names
- **Custom tags**: From `var.tags` in tfvars

## 📊 Outputs

The configuration provides comprehensive outputs including:
- VPC and subnet IDs
- Transit Gateway information
- Network Firewall details
- Route table IDs
- Endpoint information

## 🔍 Monitoring and Troubleshooting

### Network Firewall Logs
- Configure VPC Flow Logs for traffic analysis
- Enable Network Firewall logging for rule evaluation
- Use CloudWatch for monitoring firewall metrics

### Transit Gateway Monitoring
- Monitor TGW route tables for proper propagation
- Check attachment states and associations
- Review TGW flow logs for traffic patterns

## 🛡️ Best Practices Implemented

1. **Security**: Mandatory firewall inspection for all inter-subnet traffic
2. **High Availability**: Multi-AZ deployment across all components
3. **Scalability**: Modular design for easy expansion
4. **Monitoring**: Comprehensive outputs for observability
5. **Cost Optimization**: Efficient resource sizing and placement

This architecture provides a robust foundation for enterprise-grade network security with centralized inspection and control.
