# Network Firewall Testing Guide

This guide provides comprehensive testing scripts and instructions for validating traffic flow between the App VPC and Egress VPC through the Network Firewall.

## Overview

The testing suite includes:
- **Setup Script**: Prepares environment and SSH keys
- **Quick Test**: Basic connectivity validation
- **Comprehensive Test**: Full traffic analysis and firewall inspection testing

## Prerequisites

1. **AWS CLI**: Configured with appropriate credentials
2. **Terraform**: Applied with all resources deployed
3. **SSH Access**: SSH keys configured for EC2 instances
4. **Network Access**: Ability to SSH to public IPs of EC2 instances

## Quick Start

### 1. Setup Environment

```bash
# Run the setup script to prepare testing environment
./setup_test.sh
```

This script will:
- Check prerequisites (AWS CLI, Terraform, instances)
- Set up SSH keys if needed
- Verify instance status
- Test SSH connectivity
- Display instance information

### 2. Run Quick Test

```bash
# Run basic connectivity tests
./quick_test.sh
```

This script tests:
- Internet connectivity from both instances
- HTTP requests through the firewall
- Inter-VPC connectivity between App and Egress instances

### 3. Run Comprehensive Test

```bash
# Run full traffic analysis
./test_traffic.sh
```

This script performs:
- Detailed network configuration analysis
- Protocol-specific testing (HTTP, HTTPS, DNS, SSH)
- Network Firewall inspection validation
- NAT Gateway functionality testing
- Bidirectional traffic flow analysis

## Manual Testing Commands

### Get Instance Information

```bash
# Get instance IDs and IPs
terraform output app_instance_id
terraform output egress_instance_id

# Get detailed instance information
aws ec2 describe-instances --instance-ids <instance_id>
```

### SSH to Instances

```bash
# SSH to App instance
ssh -i ~/.ssh/id_rsa ec2-user@<app_public_ip>

# SSH to Egress instance
ssh -i ~/.ssh/id_rsa ec2-user@<egress_public_ip>
```

### Test Internet Connectivity

```bash
# Test from App instance
ping 8.8.8.8
curl http://httpbin.org/get
nslookup google.com

# Test from Egress instance
ping 8.8.8.8
curl http://httpbin.org/get
nslookup google.com
```

### Test Inter-VPC Connectivity

```bash
# Get private IPs
APP_PRIVATE_IP=$(aws ec2 describe-instances --instance-ids $(terraform output -raw app_instance_id) --query 'Reservations[0].Instances[0].PrivateIpAddress' --output text)
EGRESS_PRIVATE_IP=$(aws ec2 describe-instances --instance-ids $(terraform output -raw egress_instance_id) --query 'Reservations[0].Instances[0].PrivateIpAddress' --output text)

# Test from App instance
ssh -i ~/.ssh/id_rsa ec2-user@<app_public_ip> "ping -c 3 $EGRESS_PRIVATE_IP"

# Test from Egress instance
ssh -i ~/.ssh/id_rsa ec2-user@<egress_public_ip> "ping -c 3 $APP_PRIVATE_IP"
```

### Test Network Firewall Inspection

```bash
# Test various protocols through the firewall
ssh -i ~/.ssh/id_rsa ec2-user@<instance_ip> "curl -v http://httpbin.org/get"
ssh -i ~/.ssh/id_rsa ec2-user@<instance_ip> "curl -v https://httpbin.org/get"
ssh -i ~/.ssh/id_rsa ec2-user@<instance_ip> "dig google.com"
ssh -i ~/.ssh/id_rsa ec2-user@<instance_ip> "nc -z -v 8.8.8.8 22"
```

### Test NAT Gateway Functionality

```bash
# Test from Egress instance (should show NAT'd public IP)
ssh -i ~/.ssh/id_rsa ec2-user@<egress_public_ip> "curl http://httpbin.org/ip"

# Compare with instance's actual public IP
ssh -i ~/.ssh/id_rsa ec2-user@<egress_public_ip> "curl ifconfig.me"
```

## Test Traffic Flow Diagrams

### Basic Connectivity Test Flow

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│                           Basic Connectivity Test Flow                                          │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│  App Instance (12.101.144.x) - Basic Internet Test                                              │
│                                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────────────────────┐    │
│  │                    Test Commands Executed                                               │    │
│  │                                                                                         │    │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐               │    │
│  │  │    Test     │    │  Expected   │    │    Path     │    │   Result    │               │    │
│  │  │             │    │  Result     │    │             │    │             │               │    │
│  │  │ ping 8.8.8.8│ →  │  SUCCESS    │ →  │ App→TGW→FW  │ →  │ ✅ PASS     │               │    │
│  │  │ curl http   │ →  │  HTTP 200   │ →  │ →Egress→NAT │ →  │ ✅ PASS     │               │    │
│  │  │ nslookup    │ →  │  DNS Reply  │ →  │ →Internet   │ →  │ ✅ PASS     │               │    │
│  │  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘               │    │
│  └─────────────────────────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘
                                           │
                                           ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│  Egress Instance (10.2.128.x) - Basic Internet Test                                             │
│                                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────────────────────┐    │
│  │                    Test Commands Executed                                               │    │
│  │                                                                                         │    │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐               │    │
│  │  │    Test     │    │  Expected   │    │    Path     │    │   Result    │               │    │
│  │  │             │    │  Result     │    │             │    │             │               │    │
│  │  │ ping 8.8.8.8│ →  │  SUCCESS    │ →  │ Egress→NAT  │ →  │ ✅ PASS     │               │    │
│  │  │ curl http   │ →  │  HTTP 200   │ →  │ →Internet   │ →  │ ✅ PASS     │               │    │
│  │  │ nslookup    │ →  │  DNS Reply  │ →  │             │ →  │ ✅ PASS     │               │    │
│  │  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘               │    │
│  └─────────────────────────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Inter-VPC Connectivity Test Flow

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│                        Inter-VPC Connectivity Test Flow                                         │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│  Test Scenario: App Instance → Egress Instance                                                  │
│                                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────────────────────┐    │
│  │  App Instance (12.101.144.x)                                                            │    │
│  │                                                                                         │    │
│  │  ┌─────────────────────────────────────────────────────────────────────────────────┐    │    │
│  │  │  Command: ping -c 3 10.2.128.x (Egress Private IP)                              │    │    │
│  │  │                                                                                 │    │    │
│  │  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐       │    │    │
│  │  │  │   Source    │    │ Destination │    │    Path     │    │   Result    │       │    │    │
│  │  │  │ 12.101.144.x│ →  │ 10.2.128.x  │ →  │ App→TGW→FW  │ →  │ ✅ SUCCESS  │       │    │    │
│  │  │  │             │    │             │ →  │ →Egress     │ →  │             │       │    │    │
│  │  │  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘       │    │    │
│  │  └─────────────────────────────────────────────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘
                                           │
                                           ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│  Test Scenario: Egress Instance → App Instance                                                  │
│                                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────────────────────┐    │
│  │  Egress Instance (10.2.128.x)                                                           │    │
│  │                                                                                         │    │
│  │  ┌─────────────────────────────────────────────────────────────────────────────────┐    │    │
│  │  │  Command: ping -c 3 12.101.144.x (App Private IP)                               │    │    │
│  │  │                                                                                 │    │    │
│  │  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐       │    │    │
│  │  │  │   Source    │    │ Destination │    │    Path     │    │   Result    │       │    │    │
│  │  │  │ 10.2.128.x  │ →  │ 12.101.144.x│ →  │ Egress→TGW  │ →  │ ✅ SUCCESS  │       │    │    │
│  │  │  │             │    │             │ →  │ →FW→App     │ →  │             │       │    │    │
│  │  │  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘       │    │    │
│  │  └─────────────────────────────────────────────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Network Firewall Inspection Test Flow

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│                        Network Firewall Inspection Test Flow                                    │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│  Protocol Testing Through Network Firewall                                                      │
│                                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────────────────────┐    │
│  │  HTTP Test (Port 80)                                                                    │    │
│  │                                                                                         │    │
│  │  ┌─────────────────────────────────────────────────────────────────────────────────┐    │    │
│  │  │  Command: curl -s -o /dev/null -w '%{http_code}' http://httpbin.org/get         │    │    │
│  │  │                                                                                 │    │    │
│  │  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐       │    │    │
│  │  │  │   Source    │    │ Destination │    │    Path     │    │   Result    │       │    │    │
│  │  │  │ App Instance│ →  │ httpbin.org │ →  │ App→TGW→FW  │ →  │ ✅ HTTP 200 │       │    │    │
│  │  │  │             │    │    :80      │ →  │ →Egress→NAT │ →  │             │       │    │    │
│  │  │  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘       │    │    │
│  │  └─────────────────────────────────────────────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────────────────────────────────────────┘    │
│                                           │                                                     │
│                                           ▼                                                     │
│  ┌─────────────────────────────────────────────────────────────────────────────────────────┐    │
│  │  HTTPS Test (Port 443)                                                                  │    │
│  │                                                                                         │    │
│  │  ┌─────────────────────────────────────────────────────────────────────────────────┐    │    │
│  │  │  Command: curl -s -o /dev/null -w '%{http_code}' https://httpbin.org/get        │    │    │
│  │  │                                                                                 │    │    │
│  │  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐       │    │    │
│  │  │  │   Source    │    │ Destination │    │    Path     │    │   Result    │       │    │    │
│  │  │  │ App Instance│ →  │ httpbin.org │ →  │ App→TGW→FW  │ →  │ ✅ HTTP 200 │       │    │    │
│  │  │  │             │    │    :443     │ →  │ →Egress→NAT │ →  │             │       │    │    │
│  │  │  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘       │    │    │
│  │  └─────────────────────────────────────────────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────────────────────────────────────────┘    │
│                                           │                                                     │
│                                           ▼                                                     │
│  ┌─────────────────────────────────────────────────────────────────────────────────────────┐    │
│  │  DNS Test (Port 53)                                                                     │    │
│  │                                                                                         │    │
│  │  ┌─────────────────────────────────────────────────────────────────────────────────┐    │    │
│  │  │  Command: dig google.com                                                        │    │    │
│  │  │                                                                                 │    │    │
│  │  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐       │    │    │
│  │  │  │   Source    │    │ Destination │    │    Path     │    │   Result    │       │    │    │
│  │  │  │ App Instance│ →  │ 8.8.8.8:53  │ →  │ App→TGW→FW  │ →  │ ✅ DNS Reply│       │    │    │
│  │  │  │             │    │             │ →  │ →Egress→NAT │ →  │             │       │    │    │
│  │  │  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘       │    │    │
│  │  └─────────────────────────────────────────────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘
```

### NAT Gateway Test Flow

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│                              NAT Gateway Test Flow                                              │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│  Test Scenario: Verify NAT Translation from Egress Instance                                     │
│                                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────────────────────┐    │
│  │  Egress Instance (10.2.128.x) - Private IP                                              │    │
│  │                                                                                         │    │
│  │  ┌─────────────────────────────────────────────────────────────────────────────────┐    │    │
│  │  │  Test 1: Check Public IP Seen by External Service                               │    │    │
│  │  │                                                                                 │    │    │
│  │  │  Command: curl -s http://httpbin.org/ip                                         │    │    │
│  │  │                                                                                 │    │    │
│  │  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐       │    │    │
│  │  │  │   Source    │    │ Destination │    │    Path     │    │   Result    │       │    │    │
│  │  │  │ 10.2.128.x  │ →  │ httpbin.org │ →  │ Egress→NAT  │ →  │ 54.x.x.x    │       │    │    │
│  │  │  │ (Private)   │    │             │ →  │ →Internet   │ →  │ (Public)    │       │    │    │
│  │  │  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘       │    │    │
│  │  └─────────────────────────────────────────────────────────────────────────────────┘    │    │
│  │                                                                                         │    │
│  │  ┌─────────────────────────────────────────────────────────────────────────────────┐    │    │
│  │  │  Test 2: Compare with Instance's Actual Public IP                               │    │    │
│  │  │                                                                                 │    │    │
│  │  │  Command: curl -s ifconfig.me                                                   │    │    │
│  │  │                                                                                 │    │    │
│  │  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐       │    │    │
│  │  │  │   Source    │    │ Destination │    │    Path     │    │   Result    │       │    │    │
│  │  │  │ 10.2.128.x  │ →  │ ifconfig.me │ →  │ Egress→NAT  │ →  │ 54.x.x.x    │       │    │    │
│  │  │  │ (Private)   │    │             │ →  │ →Internet   │ →  │ (Public)    │       │    │    │
│  │  │  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘       │    │    │
│  │  └─────────────────────────────────────────────────────────────────────────────────┘    │    │
│  │                                                                                         │    │
│  │  ┌─────────────────────────────────────────────────────────────────────────────────┐    │    │
│  │  │  Expected Result: Both commands return same public IP (NAT Gateway IP)          │    │    │
│  │  │  ✅ NAT Gateway is working correctly                                             │   │    │
│  │  └─────────────────────────────────────────────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Expected Traffic Flow

### Outbound Traffic (App → Internet)

1. **App Instance** (12.101.144.x) → App VPC Route Table
2. **App VPC Route Table** → Transit Gateway (0.0.0.0/0 → TGW)
3. **TGW Inspection Route Table** → Inspection VPC (0.0.0.0/0 → Inspection VPC Attachment)
4. **Inspection VPC Firewall Subnet** → Network Firewall Endpoint
5. **Network Firewall** → Inspection VPC Firewall Route Table (0.0.0.0/0 → TGW)
6. **TGW Firewall Route Table** → Egress VPC (0.0.0.0/0 → Egress VPC Attachment)
7. **Egress VPC TGW Subnet** → Egress VPC Firewall Route Table
8. **Egress VPC Firewall Route Table** → NAT Gateway (0.0.0.0/0 → NAT Gateway)
9. **NAT Gateway** → Internet Gateway (IP translation)
10. **Internet Gateway** → Internet

### Return Traffic (Internet → App)

1. **Internet** → Internet Gateway
2. **Internet Gateway** → NAT Gateway (reverse IP translation)
3. **NAT Gateway** → Egress VPC Firewall Route Table
4. **Egress VPC Firewall Route Table** → Transit Gateway (12.0.0.0/8 → TGW)
5. **TGW Firewall Route Table** → App VPC (App VPC CIDR → App VPC Attachment)
6. **App VPC** → App Instance

## Troubleshooting

### Common Issues

1. **SSH Connection Failed**
   - Check security group rules (port 22)
   - Verify SSH key is correct
   - Ensure instance is running

2. **Internet Connectivity Failed**
   - Check route table configurations
   - Verify NAT Gateway is running
   - Check Network Firewall rules

3. **Inter-VPC Connectivity Failed**
   - Verify TGW route table associations
   - Check TGW route table routes
   - Ensure VPC attachments are active

4. **Firewall Blocking Traffic**
   - Check Network Firewall rule groups
   - Verify firewall policy configuration
   - Check firewall endpoint status

### Debugging Commands

```bash
# Check instance status
aws ec2 describe-instances --instance-ids <instance_id>

# Check route tables
aws ec2 describe-route-tables --filters "Name=vpc-id,Values=<vpc_id>"

# Check TGW attachments
aws ec2 describe-transit-gateway-attachments

# Check TGW route tables
aws ec2 describe-transit-gateway-route-tables

# Check Network Firewall status
aws network-firewall describe-firewall --firewall-arn <firewall_arn>

# Check NAT Gateway status
aws ec2 describe-nat-gateways
```

## Test Results Interpretation

### Successful Tests
- ✅ All connectivity tests pass
- ✅ HTTP/HTTPS requests return 200 status
- ✅ Inter-VPC ping succeeds
- ✅ NAT Gateway shows different public IP than instance

### Failed Tests
- ❌ Internet connectivity fails → Check routing and NAT Gateway
- ❌ Inter-VPC connectivity fails → Check TGW configuration
- ❌ HTTP requests fail → Check Network Firewall rules
- ❌ SSH fails → Check security groups and SSH keys

## Security Considerations

1. **SSH Keys**: Store securely and rotate regularly
2. **Security Groups**: Ensure minimal required access
3. **Network Firewall**: Review rules for security compliance
4. **Logging**: Enable VPC Flow Logs and CloudWatch Logs
5. **Monitoring**: Set up CloudWatch alarms for network issues

## Performance Testing

For performance testing, you can use additional tools:

```bash
# Bandwidth testing
ssh -i ~/.ssh/id_rsa ec2-user@<instance_ip> "wget -O /dev/null http://speedtest.tele2.net/100MB.zip"

# Latency testing
ssh -i ~/.ssh/id_rsa ec2-user@<instance_ip> "ping -c 100 8.8.8.8"

# Concurrent connection testing
ssh -i ~/.ssh/id_rsa ec2-user@<instance_ip> "for i in {1..10}; do curl -s http://httpbin.org/get & done; wait"
```

## Cleanup

After testing, you can clean up resources:

```bash
# Destroy Terraform resources
terraform destroy -var-file=env.tfvars

# Remove SSH keys (optional)
rm ~/.ssh/id_rsa ~/.ssh/id_rsa.pub
```

## Support

For issues or questions:
1. Check Terraform state: `terraform state list`
2. Review CloudWatch logs
3. Check AWS console for resource status
4. Review Network Firewall metrics in CloudWatch
