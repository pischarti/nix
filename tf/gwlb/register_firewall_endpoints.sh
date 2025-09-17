#!/bin/bash

# Script to discover Network Firewall endpoint IPs
# This script outputs the endpoint IPs that can be used for GWLB target registration

set -e

# Check if required tools are installed
if ! command -v aws &> /dev/null; then
    echo "Error: AWS CLI is not installed" >&2
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "Error: jq is not installed" >&2
    exit 1
fi

# Get Terraform outputs
FIREWALL_SUBNET_IDS=$(terraform output -json firewall_subnet_ids_for_manual_registration 2>/dev/null || echo "[]")
FIREWALL_NAME=$(terraform output -raw network_firewall_name 2>/dev/null || echo "")

# Parse subnet IDs from JSON output
SUBNET_IDS=$(echo "$FIREWALL_SUBNET_IDS" | jq -r '.[]' 2>/dev/null || echo "")

if [[ -z "$SUBNET_IDS" ]]; then
    echo "Error: Could not parse firewall subnet IDs" >&2
    exit 1
fi

# Check Network Firewall status first
if [[ -n "$FIREWALL_NAME" ]]; then
    FIREWALL_STATUS=$(aws network-firewall describe-firewall --firewall-name "$FIREWALL_NAME" --query 'Firewall.FirewallStatus.Status' --output text 2>/dev/null || echo "UNKNOWN")
    if [[ "$FIREWALL_STATUS" == "UNKNOWN" ]]; then
        echo "Warning: Could not determine Network Firewall status. Proceeding with endpoint discovery..." >&2
    elif [[ "$FIREWALL_STATUS" != "READY" ]]; then
        echo "Error: Network Firewall is not ready (Status: $FIREWALL_STATUS). Please wait for deployment to complete." >&2
        exit 1
    fi
else
    echo "Warning: Could not get Network Firewall name from Terraform outputs. Proceeding with endpoint discovery..." >&2
fi

# Find Network Firewall endpoint network interfaces
ENDPOINT_IPS=()

# Try multiple description patterns as AWS may use different formats
PATTERNS=(
    "*firewall*"
    "*Firewall*" 
    "*FIREWALL*"
    "*Network*Firewall*"
    "*AWS*Network*Firewall*"
    "*vpce-*"
)

for subnet_id in $SUBNET_IDS; do
    FOUND_IN_SUBNET=false
    
    for pattern in "${PATTERNS[@]}"; do
        # Find network interfaces in this subnet with firewall-related descriptions
        NETWORK_INTERFACES=$(aws ec2 describe-network-interfaces \
            --filters \
                "Name=subnet-id,Values=$subnet_id" \
                "Name=description,Values=$pattern" \
            --query 'NetworkInterfaces[?Status==`in-use`].PrivateIpAddress' \
            --output text 2>/dev/null || echo "")
        
        # Add found IPs to array
        if [[ -n "$NETWORK_INTERFACES" && "$NETWORK_INTERFACES" != "None" ]]; then
            for ip in $NETWORK_INTERFACES; do
                if [[ "$ip" != "None" && -n "$ip" ]]; then
                    ENDPOINT_IPS+=("$ip")
                    FOUND_IN_SUBNET=true
                fi
            done
        fi
        
        # If we found interfaces with this pattern, no need to try others for this subnet
        if [[ "$FOUND_IN_SUBNET" == true ]]; then
            break
        fi
    done
done

# Remove duplicates
if [[ ${#ENDPOINT_IPS[@]} -gt 0 ]]; then
    UNIQUE_IPS=($(printf '%s\n' "${ENDPOINT_IPS[@]}" | sort -u))
    ENDPOINT_IPS=("${UNIQUE_IPS[@]}")
fi

if [[ ${#ENDPOINT_IPS[@]} -eq 0 ]]; then
    echo "Error: No Network Firewall endpoints found. Make sure the Network Firewall is fully deployed." >&2
    echo "Tip: Run './debug_firewall_endpoints.sh' for detailed troubleshooting information." >&2
    exit 1
fi

# Output the IPs as a JSON array for Terraform consumption
printf '%s\n' "${ENDPOINT_IPS[@]}" | jq -R . | jq -s .
