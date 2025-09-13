#!/bin/bash

# Setup script for Network Firewall testing
# Prepares SSH keys and environment for traffic testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    local status=$1
    local message=$2
    case $status in
        "INFO")
            echo -e "${BLUE}[INFO]${NC} $message"
            ;;
        "SUCCESS")
            echo -e "${GREEN}[SUCCESS]${NC} $message"
            ;;
        "WARNING")
            echo -e "${YELLOW}[WARNING]${NC} $message"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} $message"
            ;;
    esac
}

# Function to check prerequisites
check_prerequisites() {
    print_status "INFO" "Checking prerequisites..."
    
    # Check AWS CLI
    if ! command -v aws &> /dev/null; then
        print_status "ERROR" "AWS CLI is not installed"
        exit 1
    fi
    
    # Check AWS credentials
    if ! aws sts get-caller-identity >/dev/null 2>&1; then
        print_status "ERROR" "AWS credentials not configured"
        print_status "INFO" "Run 'aws configure' to set up credentials"
        exit 1
    fi
    
    # Check Terraform
    if ! command -v terraform &> /dev/null; then
        print_status "ERROR" "Terraform is not installed"
        exit 1
    fi
    
    # Check if Terraform has been applied
    if ! terraform output app_instance_id >/dev/null 2>&1; then
        print_status "ERROR" "Terraform has not been applied or outputs are missing"
        print_status "INFO" "Run 'terraform apply -var-file=env.tfvars' first"
        exit 1
    fi
    
    print_status "SUCCESS" "All prerequisites met"
}

# Function to setup SSH keys
setup_ssh_keys() {
    print_status "INFO" "Setting up SSH keys..."
    
    # Check if SSH key already exists
    if [ -f ~/.ssh/id_rsa ]; then
        print_status "WARNING" "SSH key already exists at ~/.ssh/id_rsa"
        read -p "Do you want to use the existing key? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_status "INFO" "Using existing SSH key"
            return 0
        fi
    fi
    
    # Generate SSH key if it doesn't exist
    print_status "INFO" "Generating SSH key..."
    ssh-keygen -t rsa -b 4096 -f ~/.ssh/id_rsa -N "" -C "network-firewall-test"
    
    print_status "SUCCESS" "SSH key generated"
}

# Function to get SSH public key
get_ssh_public_key() {
    if [ -f ~/.ssh/id_rsa.pub ]; then
        cat ~/.ssh/id_rsa.pub
    else
        print_status "ERROR" "SSH public key not found"
        exit 1
    fi
}

# Function to check instance status
check_instances() {
    print_status "INFO" "Checking instance status..."
    
    local app_instance_id=$(terraform output -raw app_instance_id)
    local egress_instance_id=$(terraform output -raw egress_instance_id)
    
    # Check App instance
    local app_state=$(aws ec2 describe-instances \
        --instance-ids "$app_instance_id" \
        --query 'Reservations[0].Instances[0].State.Name' \
        --output text)
    
    if [ "$app_state" = "running" ]; then
        print_status "SUCCESS" "App instance is running"
    else
        print_status "ERROR" "App instance is in state: $app_state"
        return 1
    fi
    
    # Check Egress instance
    local egress_state=$(aws ec2 describe-instances \
        --instance-ids "$egress_instance_id" \
        --query 'Reservations[0].Instances[0].State.Name' \
        --output text)
    
    if [ "$egress_state" = "running" ]; then
        print_status "SUCCESS" "Egress instance is running"
    else
        print_status "ERROR" "Egress instance is in state: $egress_state"
        return 1
    fi
}

# Function to test SSH connectivity
test_ssh_connectivity() {
    print_status "INFO" "Testing SSH connectivity..."
    
    local app_instance_id=$(terraform output -raw app_instance_id)
    local egress_instance_id=$(terraform output -raw egress_instance_id)
    
    local app_public_ip=$(aws ec2 describe-instances \
        --instance-ids "$app_instance_id" \
        --query 'Reservations[0].Instances[0].PublicIpAddress' \
        --output text)
    
    local egress_public_ip=$(aws ec2 describe-instances \
        --instance-ids "$egress_instance_id" \
        --query 'Reservations[0].Instances[0].PublicIpAddress' \
        --output text)
    
    # Test SSH to App instance
    if ssh -o StrictHostKeyChecking=no -o ConnectTimeout=10 \
        -i ~/.ssh/id_rsa ec2-user@"$app_public_ip" "echo 'SSH test successful'" >/dev/null 2>&1; then
        print_status "SUCCESS" "SSH to App instance working"
    else
        print_status "ERROR" "SSH to App instance failed"
        print_status "INFO" "Make sure the SSH key is configured in the EC2 instances"
        return 1
    fi
    
    # Test SSH to Egress instance
    if ssh -o StrictHostKeyChecking=no -o ConnectTimeout=10 \
        -i ~/.ssh/id_rsa ec2-user@"$egress_public_ip" "echo 'SSH test successful'" >/dev/null 2>&1; then
        print_status "SUCCESS" "SSH to Egress instance working"
    else
        print_status "ERROR" "SSH to Egress instance failed"
        print_status "INFO" "Make sure the SSH key is configured in the EC2 instances"
        return 1
    fi
}

# Function to display instance information
display_instance_info() {
    print_status "INFO" "Instance Information:"
    
    local app_instance_id=$(terraform output -raw app_instance_id)
    local egress_instance_id=$(terraform output -raw egress_instance_id)
    
    local app_private_ip=$(aws ec2 describe-instances \
        --instance-ids "$app_instance_id" \
        --query 'Reservations[0].Instances[0].PrivateIpAddress' \
        --output text)
    
    local app_public_ip=$(aws ec2 describe-instances \
        --instance-ids "$app_instance_id" \
        --query 'Reservations[0].Instances[0].PublicIpAddress' \
        --output text)
    
    local egress_private_ip=$(aws ec2 describe-instances \
        --instance-ids "$egress_instance_id" \
        --query 'Reservations[0].Instances[0].PrivateIpAddress' \
        --output text)
    
    local egress_public_ip=$(aws ec2 describe-instances \
        --instance-ids "$egress_instance_id" \
        --query 'Reservations[0].Instances[0].PublicIpAddress' \
        --output text)
    
    echo -e "\n${BLUE}App Instance:${NC}"
    echo "  Instance ID: $app_instance_id"
    echo "  Private IP:  $app_private_ip"
    echo "  Public IP:   $app_public_ip"
    
    echo -e "\n${BLUE}Egress Instance:${NC}"
    echo "  Instance ID: $egress_instance_id"
    echo "  Private IP:  $egress_private_ip"
    echo "  Public IP:   $egress_public_ip"
    
    echo -e "\n${BLUE}VPC CIDR Blocks:${NC}"
    echo "  App VPC:     12.101.0.0/16"
    echo "  Egress VPC:  10.2.0.0/16"
    echo "  Inspection:  11.3.0.0/16"
}

# Function to show usage instructions
show_usage() {
    print_status "INFO" "Setup completed! You can now run the following tests:"
    echo ""
    echo "1. Quick test (basic connectivity):"
    echo "   ./quick_test.sh"
    echo ""
    echo "2. Comprehensive test (full traffic analysis):"
    echo "   ./test_traffic.sh"
    echo ""
    echo "3. Manual testing commands:"
    echo "   # SSH to App instance"
    echo "   ssh -i ~/.ssh/id_rsa ec2-user@<app_public_ip>"
    echo ""
    echo "   # SSH to Egress instance"
    echo "   ssh -i ~/.ssh/id_rsa ec2-user@<egress_public_ip>"
    echo ""
    echo "4. Test internet connectivity from instances:"
    echo "   # From App instance"
    echo "   ping 8.8.8.8"
    echo "   curl http://httpbin.org/get"
    echo ""
    echo "   # From Egress instance"
    echo "   ping 8.8.8.8"
    echo "   curl http://httpbin.org/get"
    echo ""
    echo "5. Test inter-VPC connectivity:"
    echo "   # From App instance (replace with actual private IP)"
    echo "   ping 10.2.128.x"
    echo ""
    echo "   # From Egress instance (replace with actual private IP)"
    echo "   ping 12.101.144.x"
}

# Main function
main() {
    print_status "INFO" "Network Firewall Test Setup"
    print_status "INFO" "==========================="
    
    check_prerequisites
    setup_ssh_keys
    check_instances
    test_ssh_connectivity
    display_instance_info
    show_usage
    
    print_status "SUCCESS" "Setup completed successfully!"
}

main "$@"
