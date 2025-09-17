#!/bin/bash

# Debug script to help diagnose Network Firewall endpoint discovery issues

set -e

echo "=== Network Firewall Endpoint Debug Script ==="
echo

# Check if AWS CLI is available
if ! command -v aws &> /dev/null; then
    echo "Error: AWS CLI is not installed"
    exit 1
fi

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo "Error: jq is not installed"
    exit 1
fi

# Get basic info
echo "1. Getting Terraform outputs..."
FIREWALL_VPC_ID=$(terraform output -raw firewall_vpc_id 2>/dev/null || echo "")
FIREWALL_SUBNET_IDS=$(terraform output -json firewall_subnet_ids_for_manual_registration 2>/dev/null || echo "[]")
FIREWALL_NAME=$(terraform output -raw network_firewall_name 2>/dev/null || echo "")

echo "   Firewall VPC ID: $FIREWALL_VPC_ID"
echo "   Firewall Name: $FIREWALL_NAME"
echo "   Firewall Subnet IDs: $FIREWALL_SUBNET_IDS"
echo

if [[ -z "$FIREWALL_VPC_ID" ]]; then
    echo "Error: Could not get firewall VPC ID. Make sure Terraform has been applied."
    exit 1
fi

# Check Network Firewall status
echo "2. Checking Network Firewall status..."
if [[ -n "$FIREWALL_NAME" ]]; then
    FIREWALL_STATUS=$(aws network-firewall describe-firewall --firewall-name "$FIREWALL_NAME" --query 'Firewall.FirewallStatus.Status' --output text 2>/dev/null || echo "UNKNOWN")
    echo "   Network Firewall Status: $FIREWALL_STATUS"
    
    if [[ "$FIREWALL_STATUS" == "UNKNOWN" ]]; then
        echo "   Warning: Could not determine firewall status (check AWS CLI permissions and firewall name)"
    elif [[ "$FIREWALL_STATUS" != "READY" ]]; then
        echo "   Warning: Network Firewall is not in READY state. Current state: $FIREWALL_STATUS"
        echo "   Please wait for the firewall to be fully deployed before discovering endpoints."
    fi
else
    echo "   Warning: Could not get firewall name from Terraform outputs"
    echo "   Make sure Terraform has been applied and the firewall resource exists"
fi
echo

# Parse subnet IDs
SUBNET_IDS=$(echo "$FIREWALL_SUBNET_IDS" | jq -r '.[]' 2>/dev/null || echo "")

if [[ -z "$SUBNET_IDS" ]]; then
    echo "Error: Could not parse firewall subnet IDs"
    exit 1
fi

# Check each subnet for network interfaces
echo "3. Searching for network interfaces in firewall subnets..."
TOTAL_INTERFACES=0

for subnet_id in $SUBNET_IDS; do
    echo "   Subnet: $subnet_id"
    
    # Get all network interfaces in this subnet
    ALL_INTERFACES=$(aws ec2 describe-network-interfaces \
        --filters "Name=subnet-id,Values=$subnet_id" \
        --query 'NetworkInterfaces[*].{Id:NetworkInterfaceId,IP:PrivateIpAddress,Description:Description,Status:Status}' \
        --output json 2>/dev/null || echo "[]")
    
    INTERFACE_COUNT=$(echo "$ALL_INTERFACES" | jq length)
    TOTAL_INTERFACES=$((TOTAL_INTERFACES + INTERFACE_COUNT))
    
    echo "   Found $INTERFACE_COUNT network interface(s):"
    
    if [[ "$INTERFACE_COUNT" -gt 0 ]]; then
        echo "$ALL_INTERFACES" | jq -r '.[] | "     - \(.Id): \(.IP) (\(.Status)) - \(.Description)"'
    fi
    
    # Try different description patterns
    echo "   Testing different firewall description patterns:"
    
    PATTERNS=("*firewall*" "*Firewall*" "*FIREWALL*" "*Network*Firewall*" "*AWS*Network*Firewall*" "*vpce-*")
    
    for pattern in "${PATTERNS[@]}"; do
        MATCHES=$(aws ec2 describe-network-interfaces \
            --filters \
                "Name=subnet-id,Values=$subnet_id" \
                "Name=description,Values=$pattern" \
            --query 'NetworkInterfaces[*].PrivateIpAddress' \
            --output text 2>/dev/null || echo "")
        
        if [[ -n "$MATCHES" && "$MATCHES" != "None" ]]; then
            echo "     Pattern '$pattern': $MATCHES"
        fi
    done
    
    echo
done

echo "4. Summary:"
echo "   Total network interfaces found: $TOTAL_INTERFACES"

if [[ $TOTAL_INTERFACES -eq 0 ]]; then
    echo "   No network interfaces found in firewall subnets."
    echo "   This suggests the Network Firewall endpoints haven't been created yet."
    echo
    echo "   Troubleshooting steps:"
    echo "   1. Verify the Network Firewall is in READY state"
    echo "   2. Wait a few more minutes for endpoint creation"
    echo "   3. Check AWS Console for Network Firewall endpoint status"
    echo "   4. Verify the firewall subnets are correct"
fi

# Check if there are any VPC endpoints (alternative approach)
echo
echo "5. Checking for VPC endpoints in firewall VPC (alternative discovery)..."
VPC_ENDPOINTS=$(aws ec2 describe-vpc-endpoints \
    --filters "Name=vpc-id,Values=$FIREWALL_VPC_ID" \
    --query 'VpcEndpoints[*].{Id:VpcEndpointId,Type:VpcEndpointType,Service:ServiceName,State:State}' \
    --output json 2>/dev/null || echo "[]")

VPC_ENDPOINT_COUNT=$(echo "$VPC_ENDPOINTS" | jq length)
echo "   Found $VPC_ENDPOINT_COUNT VPC endpoint(s):"

if [[ "$VPC_ENDPOINT_COUNT" -gt 0 ]]; then
    echo "$VPC_ENDPOINTS" | jq -r '.[] | "     - \(.Id): \(.Type) (\(.State)) - \(.Service)"'
fi

echo
echo "=== Debug Complete ==="
