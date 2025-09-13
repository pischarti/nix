#!/bin/bash

# Network Firewall Traffic Test Script
# Tests traffic flow between App VPC and Egress VPC through the Network Firewall

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
APP_VPC_CIDR="12.101.0.0/16"
EGRESS_VPC_CIDR="10.2.0.0/16"
INSPECTION_VPC_CIDR="11.3.0.0/16"

# Test targets
INTERNET_TARGET="8.8.8.8"
DNS_TARGET="1.1.1.1"
HTTP_TARGET="httpbin.org"

# Function to print colored output
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

# Function to check if instance is running
check_instance_status() {
    local instance_id=$1
    local instance_name=$2
    
    print_status "INFO" "Checking status of $instance_name ($instance_id)"
    
    local state=$(aws ec2 describe-instances \
        --instance-ids "$instance_id" \
        --query 'Reservations[0].Instances[0].State.Name' \
        --output text)
    
    if [ "$state" = "running" ]; then
        print_status "SUCCESS" "$instance_name is running"
        return 0
    else
        print_status "ERROR" "$instance_name is in state: $state"
        return 1
    fi
}

# Function to get instance IP addresses
get_instance_ips() {
    local instance_id=$1
    local instance_name=$2
    
    print_status "INFO" "Getting IP addresses for $instance_name"
    
    local private_ip=$(aws ec2 describe-instances \
        --instance-ids "$instance_id" \
        --query 'Reservations[0].Instances[0].PrivateIpAddress' \
        --output text)
    
    local public_ip=$(aws ec2 describe-instances \
        --instance-ids "$instance_id" \
        --query 'Reservations[0].Instances[0].PublicIpAddress' \
        --output text)
    
    echo "$private_ip,$public_ip"
}

# Function to execute SSH command on instance
ssh_execute() {
    local instance_ip=$1
    local command=$2
    local instance_name=$3
    
    print_status "INFO" "Executing command on $instance_name ($instance_ip): $command"
    
    ssh -o StrictHostKeyChecking=no -o ConnectTimeout=10 \
        -i ~/.ssh/id_rsa ec2-user@"$instance_ip" "$command"
}

# Function to test basic connectivity
test_basic_connectivity() {
    local instance_ip=$1
    local instance_name=$2
    
    print_status "INFO" "Testing basic connectivity for $instance_name"
    
    # Test ping to internet
    ssh_execute "$instance_ip" "ping -c 3 $INTERNET_TARGET" "$instance_name"
    if [ $? -eq 0 ]; then
        print_status "SUCCESS" "$instance_name can reach internet via ping"
    else
        print_status "ERROR" "$instance_name cannot reach internet via ping"
        return 1
    fi
    
    # Test DNS resolution
    ssh_execute "$instance_ip" "nslookup google.com" "$instance_name"
    if [ $? -eq 0 ]; then
        print_status "SUCCESS" "$instance_name can resolve DNS"
    else
        print_status "ERROR" "$instance_name cannot resolve DNS"
        return 1
    fi
}

# Function to test HTTP connectivity
test_http_connectivity() {
    local instance_ip=$1
    local instance_name=$2
    
    print_status "INFO" "Testing HTTP connectivity for $instance_name"
    
    # Test HTTP GET request
    ssh_execute "$instance_ip" "curl -s -o /dev/null -w '%{http_code}' http://$HTTP_TARGET/get" "$instance_name"
    local http_code=$(ssh_execute "$instance_ip" "curl -s -o /dev/null -w '%{http_code}' http://$HTTP_TARGET/get" "$instance_name")
    
    if [ "$http_code" = "200" ]; then
        print_status "SUCCESS" "$instance_name can make HTTP requests"
    else
        print_status "ERROR" "$instance_name HTTP test failed with code: $http_code"
        return 1
    fi
}

# Function to test inter-VPC connectivity
test_inter_vpc_connectivity() {
    local app_instance_ip=$1
    local egress_instance_ip=$2
    
    print_status "INFO" "Testing inter-VPC connectivity between App and Egress instances"
    
    # Test ping from App to Egress
    ssh_execute "$app_instance_ip" "ping -c 3 $egress_instance_ip" "App Instance"
    if [ $? -eq 0 ]; then
        print_status "SUCCESS" "App instance can reach Egress instance"
    else
        print_status "ERROR" "App instance cannot reach Egress instance"
        return 1
    fi
    
    # Test ping from Egress to App
    ssh_execute "$egress_instance_ip" "ping -c 3 $app_instance_ip" "Egress Instance"
    if [ $? -eq 0 ]; then
        print_status "SUCCESS" "Egress instance can reach App instance"
    else
        print_status "ERROR" "Egress instance cannot reach App instance"
        return 1
    fi
}

# Function to test network firewall inspection
test_firewall_inspection() {
    local instance_ip=$1
    local instance_name=$2
    
    print_status "INFO" "Testing Network Firewall inspection for $instance_name"
    
    # Test various protocols that should be inspected
    local protocols=("http" "https" "dns" "ssh")
    
    for protocol in "${protocols[@]}"; do
        case $protocol in
            "http")
                ssh_execute "$instance_ip" "curl -s -o /dev/null -w '%{http_code}' http://$HTTP_TARGET/get" "$instance_name"
                ;;
            "https")
                ssh_execute "$instance_ip" "curl -s -o /dev/null -w '%{http_code}' https://$HTTP_TARGET/get" "$instance_name"
                ;;
            "dns")
                ssh_execute "$instance_ip" "dig google.com" "$instance_name"
                ;;
            "ssh")
                ssh_execute "$instance_ip" "nc -z -v $INTERNET_TARGET 22" "$instance_name"
                ;;
        esac
        
        if [ $? -eq 0 ]; then
            print_status "SUCCESS" "$instance_name $protocol traffic passed through firewall"
        else
            print_status "WARNING" "$instance_name $protocol traffic may be blocked by firewall"
        fi
    done
}

# Function to test NAT Gateway functionality
test_nat_gateway() {
    local egress_instance_ip=$1
    
    print_status "INFO" "Testing NAT Gateway functionality from Egress instance"
    
    # Get the public IP that the instance sees
    local public_ip=$(ssh_execute "$egress_instance_ip" "curl -s ifconfig.me" "Egress Instance")
    print_status "INFO" "Egress instance sees public IP: $public_ip"
    
    # Test that outbound traffic is NAT'd
    ssh_execute "$egress_instance_ip" "curl -s http://$HTTP_TARGET/ip" "Egress Instance"
    if [ $? -eq 0 ]; then
        print_status "SUCCESS" "Egress instance can make outbound requests through NAT Gateway"
    else
        print_status "ERROR" "Egress instance cannot make outbound requests through NAT Gateway"
        return 1
    fi
}

# Function to display network configuration
display_network_config() {
    local app_instance_ip=$1
    local egress_instance_ip=$2
    
    print_status "INFO" "Displaying network configuration"
    
    echo -e "\n${BLUE}=== App Instance Network Configuration ===${NC}"
    ssh_execute "$app_instance_ip" "ip route show" "App Instance"
    ssh_execute "$app_instance_ip" "cat /etc/resolv.conf" "App Instance"
    
    echo -e "\n${BLUE}=== Egress Instance Network Configuration ===${NC}"
    ssh_execute "$egress_instance_ip" "ip route show" "Egress Instance"
    ssh_execute "$egress_instance_ip" "cat /etc/resolv.conf" "Egress Instance"
}

# Function to run comprehensive traffic tests
run_traffic_tests() {
    local app_instance_id=$1
    local egress_instance_id=$2
    
    print_status "INFO" "Starting comprehensive traffic tests"
    
    # Get instance IPs
    local app_ips=$(get_instance_ips "$app_instance_id" "App Instance")
    local egress_ips=$(get_instance_ips "$egress_instance_id" "Egress Instance")
    
    local app_private_ip=$(echo "$app_ips" | cut -d',' -f1)
    local app_public_ip=$(echo "$app_ips" | cut -d',' -f2)
    local egress_private_ip=$(echo "$egress_ips" | cut -d',' -f1)
    local egress_public_ip=$(echo "$egress_ips" | cut -d',' -f2)
    
    print_status "INFO" "App Instance - Private: $app_private_ip, Public: $app_public_ip"
    print_status "INFO" "Egress Instance - Private: $egress_private_ip, Public: $egress_public_ip"
    
    # Display network configuration
    display_network_config "$app_public_ip" "$egress_public_ip"
    
    # Test basic connectivity
    test_basic_connectivity "$app_public_ip" "App Instance"
    test_basic_connectivity "$egress_public_ip" "Egress Instance"
    
    # Test HTTP connectivity
    test_http_connectivity "$app_public_ip" "App Instance"
    test_http_connectivity "$egress_public_ip" "Egress Instance"
    
    # Test inter-VPC connectivity
    test_inter_vpc_connectivity "$app_public_ip" "$egress_public_ip"
    
    # Test firewall inspection
    test_firewall_inspection "$app_public_ip" "App Instance"
    test_firewall_inspection "$egress_public_ip" "Egress Instance"
    
    # Test NAT Gateway functionality
    test_nat_gateway "$egress_public_ip"
    
    print_status "SUCCESS" "All traffic tests completed"
}

# Function to get instance IDs from Terraform
get_instance_ids() {
    print_status "INFO" "Getting instance IDs from Terraform state"
    
    local app_instance_id=$(terraform output -raw app_instance_id 2>/dev/null || echo "")
    local egress_instance_id=$(terraform output -raw egress_instance_id 2>/dev/null || echo "")
    
    if [ -z "$app_instance_id" ] || [ -z "$egress_instance_id" ]; then
        print_status "ERROR" "Could not get instance IDs from Terraform outputs"
        print_status "INFO" "Make sure you have run 'terraform apply' and have the following outputs:"
        print_status "INFO" "  - app_instance_id"
        print_status "INFO" "  - egress_instance_id"
        exit 1
    fi
    
    echo "$app_instance_id,$egress_instance_id"
}

# Main function
main() {
    print_status "INFO" "Starting Network Firewall Traffic Test"
    print_status "INFO" "========================================"
    
    # Check if AWS CLI is configured
    if ! aws sts get-caller-identity >/dev/null 2>&1; then
        print_status "ERROR" "AWS CLI is not configured or credentials are invalid"
        exit 1
    fi
    
    # Check if Terraform is available
    if ! command -v terraform &> /dev/null; then
        print_status "ERROR" "Terraform is not installed or not in PATH"
        exit 1
    fi
    
    # Get instance IDs
    local instance_ids=$(get_instance_ids)
    local app_instance_id=$(echo "$instance_ids" | cut -d',' -f1)
    local egress_instance_id=$(echo "$instance_ids" | cut -d',' -f2)
    
    print_status "INFO" "App Instance ID: $app_instance_id"
    print_status "INFO" "Egress Instance ID: $egress_instance_id"
    
    # Check instance status
    check_instance_status "$app_instance_id" "App Instance"
    check_instance_status "$egress_instance_id" "Egress Instance"
    
    # Run traffic tests
    run_traffic_tests "$app_instance_id" "$egress_instance_id"
    
    print_status "SUCCESS" "Network Firewall Traffic Test completed successfully!"
}

# Run main function
main "$@"
