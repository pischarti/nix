#!/bin/bash

# Quick health check fix script
# This script provides immediate solutions for health check failures

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚑 Quick Health Check Fix${NC}"
echo -e "${BLUE}=========================${NC}"
echo ""

echo -e "${YELLOW}The most common causes of health check failures on port 443:${NC}"
echo ""

echo -e "${RED}1. TCP Health Checks on SSL Port (Most Common Issue)${NC}"
echo -e "   Problem: NLB trying to do TCP health checks directly on port 443"
echo -e "   Solution: Change to HTTP health checks on port 80"
echo ""

echo -e "${RED}2. Instance Not Ready${NC}"
echo -e "   Problem: EC2 instance still initializing or web servers not started"
echo -e "   Solution: Wait 5-10 minutes after deployment"
echo ""

echo -e "${RED}3. Security Group Issues${NC}"
echo -e "   Problem: Security groups blocking health check traffic"
echo -e "   Solution: Ensure SG allows traffic from NLB subnets"
echo ""

echo -e "${BLUE}🔧 Immediate Fix Options:${NC}"
echo -e "${BLUE}========================${NC}"
echo ""

echo -e "${GREEN}Option 1: Apply the Updated Terraform Configuration${NC}"
echo -e "The main.tf file has been updated to fix the health check issue:"
echo -e ""
echo -e "   cd /Users/steve/dev/nix/tf/gwlb"
echo -e "   terraform plan"
echo -e "   terraform apply"
echo -e ""
echo -e "This will:"
echo -e "   • Change HTTPS target group to use HTTP health checks on port 80"
echo -e "   • Add proper health check endpoints"
echo -e "   • Improve SSL configuration"
echo ""

echo -e "${GREEN}Option 2: Manual AWS CLI Fix (Immediate)${NC}"
echo -e "If you know your target group ARN, you can fix it immediately:"
echo -e ""
echo -e '   # Get your HTTPS target group ARN'
echo -e '   HTTPS_TG_ARN="arn:aws:elasticloadbalancing:us-east-1:ACCOUNT:targetgroup/NAME/ID"'
echo -e ""
echo -e '   # Update health check to use HTTP on port 80'
echo -e '   aws elbv2 modify-target-group \'
echo -e '     --target-group-arn "$HTTPS_TG_ARN" \'
echo -e '     --health-check-protocol HTTP \'
echo -e '     --health-check-port 80 \'
echo -e '     --health-check-path "/health" \'
echo -e '     --health-check-interval-seconds 30 \'
echo -e '     --health-check-timeout-seconds 10 \'
echo -e '     --healthy-threshold-count 2 \'
echo -e '     --unhealthy-threshold-count 3'
echo ""

echo -e "${GREEN}Option 3: Find and Fix Existing Resources${NC}"
echo -e "Find your existing resources and apply the fix:"
echo -e ""
echo -e "   # Find your NLB"
echo -e '   aws elbv2 describe-load-balancers --query "LoadBalancers[?contains(LoadBalancerName, \`test-nlb\`)].{Name:LoadBalancerName,ARN:LoadBalancerArn}" --output table'
echo -e ""
echo -e "   # Find target groups for the NLB"
echo -e '   NLB_ARN="your-nlb-arn-from-above"'
echo -e '   aws elbv2 describe-target-groups --load-balancer-arn "$NLB_ARN" --output table'
echo -e ""
echo -e "   # Fix the HTTPS target group (port 443)"
echo -e '   HTTPS_TG_ARN="your-https-target-group-arn"'
echo -e '   aws elbv2 modify-target-group \'
echo -e '     --target-group-arn "$HTTPS_TG_ARN" \'
echo -e '     --health-check-protocol HTTP \'
echo -e '     --health-check-port 80 \'
echo -e '     --health-check-path "/health"'
echo ""

echo -e "${BLUE}🔍 Diagnostic Commands:${NC}"
echo -e "${BLUE}======================${NC}"
echo -e ""

echo -e "${YELLOW}Check current target group health:${NC}"
echo -e '   aws elbv2 describe-target-health --target-group-arn "YOUR_TARGET_GROUP_ARN"'
echo -e ""

echo -e "${YELLOW}Check target group configuration:${NC}"
echo -e '   aws elbv2 describe-target-groups --target-group-arns "YOUR_TARGET_GROUP_ARN"'
echo -e ""

echo -e "${YELLOW}Check EC2 instance status:${NC}"
echo -e '   aws ec2 describe-instances --filters "Name=tag:Name,Values=*test-private*" --query "Reservations[*].Instances[*].{ID:InstanceId,State:State.Name,IP:PrivateIpAddress}"'
echo -e ""

echo -e "${BLUE}💡 Why This Happens:${NC}"
echo -e "${BLUE}==================${NC}"
echo -e ""
echo -e "Network Load Balancers perform health checks differently than Application Load Balancers:"
echo -e ""
echo -e "• ${RED}TCP Health Checks on SSL ports${NC} can fail because:"
echo -e "  - SSL handshake issues"
echo -e "  - Certificate validation problems"
echo -e "  - Timing issues with SSL negotiation"
echo -e ""
echo -e "• ${GREEN}HTTP Health Checks${NC} are more reliable because:"
echo -e "  - Simple HTTP request/response"
echo -e "  - No SSL complexity"
echo -e "  - Faster and more predictable"
echo -e ""
echo -e "The fix changes the HTTPS target group to use HTTP health checks"
echo -e "on port 80 while still forwarding HTTPS traffic on port 443."
echo ""

echo -e "${GREEN}✅ Expected Result After Fix:${NC}"
echo -e "   • HTTP Target Group: Healthy (HTTP health checks on port 80)"
echo -e "   • HTTPS Target Group: Healthy (HTTP health checks on port 80)"
echo -e "   • Traffic flows: Internet → NLB → GWLB → Firewall → Instance"
echo ""

echo -e "${BLUE}🏁 Next Steps:${NC}"
echo -e "1. Choose one of the fix options above"
echo -e "2. Wait 2-3 minutes for health checks to update"
echo -e "3. Test the endpoints:"
echo -e "   curl http://YOUR-NLB-DNS-NAME"
echo -e "   curl -k https://YOUR-NLB-DNS-NAME"
echo ""

echo -e "${YELLOW}Need help finding your resources? Run:${NC}"
echo -e "   aws elbv2 describe-load-balancers --query 'LoadBalancers[*].{Name:LoadBalancerName,DNS:DNSName}' --output table"
