#!/bin/bash

# One-liner script to fix HTTPS health check issues
# This script automatically finds and fixes the HTTPS target group health check configuration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Auto-Fix HTTPS Health Check${NC}"
echo -e "${BLUE}==============================${NC}"

# Find NLBs with "test" in the name
echo -e "${YELLOW}🔍 Finding test NLB...${NC}"
NLB_ARNS=$(aws elbv2 describe-load-balancers --query 'LoadBalancers[?contains(LoadBalancerName, `test`)].LoadBalancerArn' --output text)

if [ -z "$NLB_ARNS" ]; then
    echo -e "${RED}❌ No test NLB found. Looking for any NLB...${NC}"
    NLB_ARNS=$(aws elbv2 describe-load-balancers --query 'LoadBalancers[0].LoadBalancerArn' --output text)
fi

if [ -z "$NLB_ARNS" ] || [ "$NLB_ARNS" = "None" ]; then
    echo -e "${RED}❌ No NLB found. Please check your AWS configuration.${NC}"
    exit 1
fi

# Use the first NLB found
NLB_ARN=$(echo $NLB_ARNS | cut -d' ' -f1)
NLB_NAME=$(aws elbv2 describe-load-balancers --load-balancer-arns "$NLB_ARN" --query 'LoadBalancers[0].LoadBalancerName' --output text)

echo -e "${GREEN}✅ Found NLB: ${NLB_NAME}${NC}"
echo -e "   ARN: ${NLB_ARN}"

# Find HTTPS target group (port 443)
echo -e "${YELLOW}🎯 Finding HTTPS target group...${NC}"
HTTPS_TG_ARN=$(aws elbv2 describe-target-groups --load-balancer-arn "$NLB_ARN" --query 'TargetGroups[?Port==`443`].TargetGroupArn' --output text)

if [ -z "$HTTPS_TG_ARN" ] || [ "$HTTPS_TG_ARN" = "None" ]; then
    echo -e "${RED}❌ No HTTPS target group (port 443) found.${NC}"
    echo -e "${YELLOW}Available target groups:${NC}"
    aws elbv2 describe-target-groups --load-balancer-arn "$NLB_ARN" --query 'TargetGroups[*].{Name:TargetGroupName,Port:Port,Protocol:Protocol}' --output table
    exit 1
fi

HTTPS_TG_NAME=$(aws elbv2 describe-target-groups --target-group-arns "$HTTPS_TG_ARN" --query 'TargetGroups[0].TargetGroupName' --output text)

echo -e "${GREEN}✅ Found HTTPS Target Group: ${HTTPS_TG_NAME}${NC}"
echo -e "   ARN: ${HTTPS_TG_ARN}"

# Check current health check configuration
echo -e "${YELLOW}📊 Current health check configuration:${NC}"
CURRENT_CONFIG=$(aws elbv2 describe-target-groups --target-group-arns "$HTTPS_TG_ARN" --query 'TargetGroups[0].{Protocol:HealthCheckProtocol,Port:HealthCheckPort,Path:HealthCheckPath}' --output table)
echo "$CURRENT_CONFIG"

# Check current health status
echo -e "${YELLOW}🩺 Current target health:${NC}"
HEALTH_STATUS=$(aws elbv2 describe-target-health --target-group-arn "$HTTPS_TG_ARN" --output table)
echo "$HEALTH_STATUS"

# Apply the fix
echo -e "${YELLOW}🔧 Applying health check fix...${NC}"
echo -e "   Changing to HTTP health checks on port 80 with /health path"

aws elbv2 modify-target-group \
    --target-group-arn "$HTTPS_TG_ARN" \
    --health-check-protocol HTTP \
    --health-check-port 80 \
    --health-check-path "/health" \
    --health-check-interval-seconds 30 \
    --health-check-timeout-seconds 10 \
    --healthy-threshold-count 2 \
    --unhealthy-threshold-count 3 \
    --matcher HttpCode=200

echo -e "${GREEN}✅ Health check configuration updated!${NC}"

# Show new configuration
echo -e "${YELLOW}📊 New health check configuration:${NC}"
NEW_CONFIG=$(aws elbv2 describe-target-groups --target-group-arns "$HTTPS_TG_ARN" --query 'TargetGroups[0].{Protocol:HealthCheckProtocol,Port:HealthCheckPort,Path:HealthCheckPath,Interval:HealthCheckIntervalSeconds,Timeout:HealthCheckTimeoutSeconds}' --output table)
echo "$NEW_CONFIG"

echo -e "${BLUE}⏱️  Waiting for health checks to update...${NC}"
echo -e "   This may take 2-3 minutes..."

# Wait and check health status
for i in {1..6}; do
    echo -e "   Checking... ($i/6)"
    sleep 30
    
    HEALTH_STATUS=$(aws elbv2 describe-target-health --target-group-arn "$HTTPS_TG_ARN" --query 'TargetHealthDescriptions[0].TargetHealth.State' --output text)
    
    if [ "$HEALTH_STATUS" = "healthy" ]; then
        echo -e "${GREEN}✅ Target is now HEALTHY!${NC}"
        break
    elif [ "$HEALTH_STATUS" = "unhealthy" ]; then
        echo -e "${YELLOW}   Still unhealthy, continuing to wait...${NC}"
    else
        echo -e "${YELLOW}   Status: ${HEALTH_STATUS}${NC}"
    fi
done

# Final status check
echo -e "${YELLOW}🏁 Final health status:${NC}"
FINAL_STATUS=$(aws elbv2 describe-target-health --target-group-arn "$HTTPS_TG_ARN" --output table)
echo "$FINAL_STATUS"

# Get NLB DNS for testing
NLB_DNS=$(aws elbv2 describe-load-balancers --load-balancer-arns "$NLB_ARN" --query 'LoadBalancers[0].DNSName' --output text)

echo -e "${BLUE}🧪 Test your endpoints:${NC}"
echo -e "   HTTP:  curl http://${NLB_DNS}"
echo -e "   HTTPS: curl -k https://${NLB_DNS}"
echo -e "   Health: curl http://${NLB_DNS}/health"

if [ "$HEALTH_STATUS" = "healthy" ]; then
    echo -e "${GREEN}🎉 SUCCESS! HTTPS health checks are now working!${NC}"
else
    echo -e "${YELLOW}⚠️  Health checks may still be updating. Wait a few more minutes and check again.${NC}"
    echo -e "${YELLOW}If issues persist, the instance may still be initializing or there may be other issues.${NC}"
fi

echo -e "${BLUE}📋 Summary of changes:${NC}"
echo -e "   • Changed HTTPS target group health checks from TCP to HTTP"
echo -e "   • Health check port changed from 443 to 80"
echo -e "   • Added health check path: /health"
echo -e "   • HTTPS traffic still flows on port 443 (only health checks changed)"
