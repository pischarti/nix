#!/bin/bash

# Automated deployment script for GWLB with Network Firewall endpoint registration
# This script performs a two-stage deployment to handle the dynamic endpoint discovery

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}GWLB with Network Firewall - Automated Deployment${NC}"
echo "=================================================="

# Stage 1: Initial deployment
echo -e "${YELLOW}Stage 1: Deploying initial infrastructure...${NC}"
terraform apply -auto-approve

echo -e "${GREEN}Initial deployment complete!${NC}"

# Wait for Network Firewall to be ready
echo -e "${YELLOW}Waiting for Network Firewall to be fully ready...${NC}"
sleep 30

# Stage 2: Discover endpoints and re-deploy
echo -e "${YELLOW}Stage 2: Discovering Network Firewall endpoints...${NC}"

# Try multiple times with increasing wait periods
MAX_ATTEMPTS=5
ATTEMPT=1

while [[ $ATTEMPT -le $MAX_ATTEMPTS ]]; do
    echo -e "${YELLOW}Attempt $ATTEMPT of $MAX_ATTEMPTS...${NC}"
    
    ENDPOINT_IPS=$(./register_firewall_endpoints.sh 2>/dev/null || echo "[]")
    
    if [[ "$ENDPOINT_IPS" != "[]" ]]; then
        break
    fi
    
    if [[ $ATTEMPT -lt $MAX_ATTEMPTS ]]; then
        WAIT_TIME=$((ATTEMPT * 30))
        echo -e "${YELLOW}No endpoints found yet. Waiting ${WAIT_TIME} seconds before retry...${NC}"
        sleep $WAIT_TIME
    fi
    
    ATTEMPT=$((ATTEMPT + 1))
done

if [[ "$ENDPOINT_IPS" == "[]" ]]; then
    echo -e "${RED}Warning: No firewall endpoints discovered after $MAX_ATTEMPTS attempts.${NC}"
    echo -e "${YELLOW}The Network Firewall may still be deploying. You can:${NC}"
    echo -e "${YELLOW}1. Wait longer and run this script again${NC}"
    echo -e "${YELLOW}2. Run './debug_firewall_endpoints.sh' for detailed diagnostics${NC}"
    echo -e "${YELLOW}3. Register endpoints manually once they're available${NC}"
    exit 0
fi

echo -e "${GREEN}Discovered endpoints: $ENDPOINT_IPS${NC}"

# Re-apply with discovered endpoints
echo -e "${YELLOW}Stage 3: Registering endpoints with GWLB...${NC}"
terraform apply -auto-approve -var="firewall_endpoint_ips=$ENDPOINT_IPS"

echo -e "${GREEN}Deployment complete!${NC}"

# Check target health
echo -e "${YELLOW}Checking target health...${NC}"
TARGET_GROUP_ARN=$(terraform output -raw gwlb_target_group_arn)
sleep 10

TARGET_HEALTH=$(aws elbv2 describe-target-health \
    --target-group-arn "$TARGET_GROUP_ARN" \
    --query 'TargetHealthDescriptions[*].{Target:Target.Id,Health:TargetHealth.State}' \
    --output table 2>/dev/null || echo "Could not retrieve target health")

echo "$TARGET_HEALTH"

echo -e "${GREEN}All done! Your GWLB with Network Firewall is ready.${NC}"
echo -e "${YELLOW}Note: It may take a few minutes for targets to become healthy.${NC}"
