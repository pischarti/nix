#!/bin/bash

# Quick Network Firewall Test Script
# Simple connectivity tests between App and Egress instances

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

# Get instance information
get_instance_info() {
    local instance_id=$1
    local name=$2
    
    local private_ip=$(aws ec2 describe-instances \
        --instance-ids "$instance_id" \
        --query 'Reservations[0].Instances[0].PrivateIpAddress' \
        --output text)
    
    local public_ip=$(aws ec2 describe-instances \
        --instance-ids "$instance_id" \
        --query 'Reservations[0].Instances[0].PublicIpAddress' \
        --output text)
    
    print_status "INFO" "$name - Private IP: $private_ip, Public IP: $public_ip"
    echo "$private_ip,$public_ip"
}

# Quick connectivity test
quick_test() {
    local instance_ip=$1
    local instance_name=$2
    
    print_status "INFO" "Testing connectivity for $instance_name"
    
    # Test ping to internet
    if ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 \
        -i ~/.ssh/id_rsa ec2-user@"$instance_ip" "ping -c 2 8.8.8.8" >/dev/null 2>&1; then
        print_status "SUCCESS" "$instance_name can reach internet"
    else
        print_status "ERROR" "$instance_name cannot reach internet"
        return 1
    fi
    
    # Test HTTP connectivity
    local http_code=$(ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 \
        -i ~/.ssh/id_rsa ec2-user@"$instance_ip" \
        "curl -s -o /dev/null -w '%{http_code}' http://httpbin.org/get" 2>/dev/null || echo "000")
    
    if [ "$http_code" = "200" ]; then
        print_status "SUCCESS" "$instance_name can make HTTP requests"
    else
        print_status "ERROR" "$instance_name HTTP test failed (code: $http_code)"
        return 1
    fi
}

# Test inter-VPC connectivity
test_inter_vpc() {
    local app_ip=$1
    local egress_ip=$2
    
    print_status "INFO" "Testing inter-VPC connectivity"
    
    # Get private IPs
    local app_private=$(aws ec2 describe-instances \
        --instance-ids $(terraform output -raw app_instance_id) \
        --query 'Reservations[0].Instances[0].PrivateIpAddress' \
        --output text)
    
    local egress_private=$(aws ec2 describe-instances \
        --instance-ids $(terraform output -raw egress_instance_id) \
        --query 'Reservations[0].Instances[0].PrivateIpAddress' \
        --output text)
    
    # Test ping from App to Egress
    if ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 \
        -i ~/.ssh/id_rsa ec2-user@"$app_ip" "ping -c 2 $egress_private" >/dev/null 2>&1; then
        print_status "SUCCESS" "App can reach Egress instance"
    else
        print_status "ERROR" "App cannot reach Egress instance"
    fi
    
    # Test ping from Egress to App
    if ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 \
        -i ~/.ssh/id_rsa ec2-user@"$egress_ip" "ping -c 2 $app_private" >/dev/null 2>&1; then
        print_status "SUCCESS" "Egress can reach App instance"
    else
        print_status "ERROR" "Egress cannot reach App instance"
    fi
}

# Main function
main() {
    print_status "INFO" "Starting Quick Network Firewall Test"
    print_status "INFO" "===================================="
    
    # Check prerequisites
    if ! aws sts get-caller-identity >/dev/null 2>&1; then
        print_status "ERROR" "AWS CLI not configured"
        exit 1
    fi
    
    if ! command -v terraform &> /dev/null; then
        print_status "ERROR" "Terraform not found"
        exit 1
    fi
    
    # Get instance IDs and IPs
    local app_instance_id=$(terraform output -raw app_instance_id)
    local egress_instance_id=$(terraform output -raw egress_instance_id)
    
    local app_info=$(get_instance_info "$app_instance_id" "App Instance")
    local egress_info=$(get_instance_info "$egress_instance_id" "Egress Instance")
    
    local app_public=$(echo "$app_info" | cut -d',' -f2)
    local egress_public=$(echo "$egress_info" | cut -d',' -f2)
    
    # Run tests
    quick_test "$app_public" "App Instance"
    quick_test "$egress_public" "Egress Instance"
    test_inter_vpc "$app_public" "$egress_public"
    
    print_status "SUCCESS" "Quick test completed!"
}

main "$@"
