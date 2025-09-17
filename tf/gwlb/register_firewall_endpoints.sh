#!/bin/bash

# Script to register Network Firewall endpoints with GWLB target group
# This script should be run after the Terraform deployment is complete

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Network Firewall Endpoint Registration Script${NC}"
echo "=============================================="

# Check if AWS CLI is installed
if ! command -v aws &> /dev/null; then
    echo -e "${RED}Error: AWS CLI is not installed${NC}"
    exit 1
fi

# Get Terraform outputs
echo -e "${YELLOW}Getting Terraform outputs...${NC}"

TARGET_GROUP_ARN=$(terraform output -raw gwlb_target_group_arn 2>/dev/null || echo "")
FIREWALL_SUBNET_IDS=$(terraform output -json firewall_subnet_ids_for_manual_registration 2>/dev/null || echo "[]")
FIREWALL_VPC_ID=$(terraform output -raw firewall_vpc_id 2>/dev/null || echo "")

if [[ -z "$TARGET_GROUP_ARN" || -z "$FIREWALL_VPC_ID" ]]; then
    echo -e "${RED}Error: Could not get required Terraform outputs. Make sure you're in the correct directory and Terraform has been applied.${NC}"
    exit 1
fi

echo "Target Group ARN: $TARGET_GROUP_ARN"
echo "Firewall VPC ID: $FIREWALL_VPC_ID"

# Parse subnet IDs from JSON output
SUBNET_IDS=$(echo "$FIREWALL_SUBNET_IDS" | jq -r '.[]' 2>/dev/null || echo "")

if [[ -z "$SUBNET_IDS" ]]; then
    echo -e "${RED}Error: Could not parse firewall subnet IDs${NC}"
    exit 1
fi

echo -e "${YELLOW}Finding Network Firewall endpoints...${NC}"

# Find Network Firewall endpoint network interfaces
ENDPOINT_IPS=()
for subnet_id in $SUBNET_IDS; do
    echo "Checking subnet: $subnet_id"
    
    # Find network interfaces in this subnet with firewall-related descriptions
    NETWORK_INTERFACES=$(aws ec2 describe-network-interfaces \
        --filters \
            "Name=subnet-id,Values=$subnet_id" \
            "Name=description,Values=*firewall*" \
        --query 'NetworkInterfaces[*].{NetworkInterfaceId:NetworkInterfaceId,PrivateIpAddress:PrivateIpAddress,Description:Description}' \
        --output json 2>/dev/null || echo "[]")
    
    # Extract private IP addresses
    PRIVATE_IPS=$(echo "$NETWORK_INTERFACES" | jq -r '.[].PrivateIpAddress' 2>/dev/null || echo "")
    
    if [[ -n "$PRIVATE_IPS" ]]; then
        for ip in $PRIVATE_IPS; do
            if [[ "$ip" != "null" && -n "$ip" ]]; then
                echo "Found firewall endpoint IP: $ip"
                ENDPOINT_IPS+=("$ip")
            fi
        done
    fi
done

if [[ ${#ENDPOINT_IPS[@]} -eq 0 ]]; then
    echo -e "${RED}Error: No Network Firewall endpoints found. Make sure the Network Firewall is fully deployed.${NC}"
    echo -e "${YELLOW}You can check the firewall status with:${NC}"
    echo "aws network-firewall describe-firewall --firewall-name <firewall-name>"
    exit 1
fi

echo -e "${GREEN}Found ${#ENDPOINT_IPS[@]} firewall endpoint(s)${NC}"

# Register each endpoint with the target group
echo -e "${YELLOW}Registering endpoints with GWLB target group...${NC}"

for ip in "${ENDPOINT_IPS[@]}"; do
    echo "Registering endpoint: $ip"
    
    aws elbv2 register-targets \
        --target-group-arn "$TARGET_GROUP_ARN" \
        --targets "Id=$ip,Port=6081" \
        2>/dev/null || {
            echo -e "${YELLOW}Warning: Failed to register $ip (it may already be registered)${NC}"
        }
done

# Wait a moment for registration to process
echo -e "${YELLOW}Waiting for target registration to process...${NC}"
sleep 5

# Check target health
echo -e "${YELLOW}Checking target health...${NC}"
TARGET_HEALTH=$(aws elbv2 describe-target-health \
    --target-group-arn "$TARGET_GROUP_ARN" \
    --query 'TargetHealthDescriptions[*].{Target:Target.Id,Health:TargetHealth.State}' \
    --output table 2>/dev/null || echo "Could not retrieve target health")

echo "$TARGET_HEALTH"

echo -e "${GREEN}Registration complete!${NC}"
echo -e "${YELLOW}Note: It may take a few minutes for targets to become healthy.${NC}"
echo -e "${YELLOW}You can monitor target health with:${NC}"
echo "aws elbv2 describe-target-health --target-group-arn $TARGET_GROUP_ARN"
