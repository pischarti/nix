#!/bin/bash

# Debug script for Network Load Balancer health check issues
# This script helps diagnose why health checks are failing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 NLB Health Check Debug Tool${NC}"
echo -e "${BLUE}==============================${NC}"

# Check if terraform outputs are available
if ! command -v terraform &> /dev/null; then
    echo -e "${RED}❌ Terraform not found. Please install terraform.${NC}"
    exit 1
fi

if [ ! -f "terraform.tfstate" ]; then
    echo -e "${RED}❌ No terraform state found. Please run 'terraform apply' first.${NC}"
    exit 1
fi

# Get terraform outputs
echo -e "${YELLOW}📊 Getting infrastructure information...${NC}"

NLB_ARN=$(terraform output -raw test_nlb_arn 2>/dev/null)
INSTANCE_ID=$(terraform output -raw test_instance_id 2>/dev/null)
INSTANCE_IP=$(terraform output -raw test_instance_private_ip 2>/dev/null)

if [ -z "$NLB_ARN" ]; then
    echo -e "${RED}❌ Could not get NLB ARN from terraform output.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Infrastructure Information:${NC}"
echo -e "   🌐 NLB ARN: ${NLB_ARN}"
echo -e "   🖥️  Instance ID: ${INSTANCE_ID}"
echo -e "   🔒 Instance IP: ${INSTANCE_IP}"
echo ""

# Get target groups
echo -e "${YELLOW}🎯 Checking Target Groups...${NC}"
TARGET_GROUPS=$(aws elbv2 describe-target-groups --load-balancer-arn "$NLB_ARN" --query 'TargetGroups[*].{Name:TargetGroupName,ARN:TargetGroupArn,Port:Port,Protocol:Protocol,HealthCheck:HealthCheckProtocol}' --output table)
echo "$TARGET_GROUPS"
echo ""

# Get target group ARNs
HTTP_TG_ARN=$(aws elbv2 describe-target-groups --load-balancer-arn "$NLB_ARN" --query 'TargetGroups[?Port==`80`].TargetGroupArn' --output text)
HTTPS_TG_ARN=$(aws elbv2 describe-target-groups --load-balancer-arn "$NLB_ARN" --query 'TargetGroups[?Port==`443`].TargetGroupArn' --output text)

echo -e "${YELLOW}🩺 Checking Target Health...${NC}"

if [ -n "$HTTP_TG_ARN" ]; then
    echo -e "${BLUE}HTTP Target Group Health:${NC}"
    aws elbv2 describe-target-health --target-group-arn "$HTTP_TG_ARN" --output table
    echo ""
fi

if [ -n "$HTTPS_TG_ARN" ]; then
    echo -e "${BLUE}HTTPS Target Group Health:${NC}"
    aws elbv2 describe-target-health --target-group-arn "$HTTPS_TG_ARN" --output table
    echo ""
fi

# Check instance status
echo -e "${YELLOW}🖥️  Checking EC2 Instance Status...${NC}"
INSTANCE_STATE=$(aws ec2 describe-instances --instance-ids "$INSTANCE_ID" --query 'Reservations[0].Instances[0].State.Name' --output text)
echo -e "   Instance State: ${INSTANCE_STATE}"

if [ "$INSTANCE_STATE" != "running" ]; then
    echo -e "${RED}❌ Instance is not in running state!${NC}"
    exit 1
fi

# Check security groups
echo -e "${YELLOW}🛡️  Checking Security Groups...${NC}"
SECURITY_GROUPS=$(aws ec2 describe-instances --instance-ids "$INSTANCE_ID" --query 'Reservations[0].Instances[0].SecurityGroups[*].GroupId' --output text)
for sg in $SECURITY_GROUPS; do
    echo -e "${BLUE}Security Group: ${sg}${NC}"
    aws ec2 describe-security-groups --group-ids "$sg" --query 'SecurityGroups[0].{GroupName:GroupName,InboundRules:IpPermissions[*].{Port:FromPort,Protocol:IpProtocol,Source:IpRanges[0].CidrIp}}' --output table
    echo ""
done

# Test direct connectivity to the instance (if we have access)
echo -e "${YELLOW}🔌 Testing Direct Connectivity...${NC}"

# Check if we can reach the instance via Systems Manager
if aws ssm describe-instance-information --filters "Key=InstanceIds,Values=$INSTANCE_ID" --query 'InstanceInformationList[0].PingStatus' --output text 2>/dev/null | grep -q "Online"; then
    echo -e "${GREEN}✅ Instance is accessible via Systems Manager${NC}"
    
    # Run health check on the instance
    echo -e "${YELLOW}🧪 Running health checks on the instance...${NC}"
    
    COMMAND_ID=$(aws ssm send-command \
        --instance-ids "$INSTANCE_ID" \
        --document-name "AWS-RunShellScript" \
        --parameters 'commands=["/home/ec2-user/check_servers.sh"]' \
        --query 'Command.CommandId' \
        --output text)
    
    echo -e "   Command ID: $COMMAND_ID"
    echo -e "   Waiting for command to complete..."
    
    sleep 10
    
    # Get command output
    COMMAND_OUTPUT=$(aws ssm get-command-invocation \
        --command-id "$COMMAND_ID" \
        --instance-id "$INSTANCE_ID" \
        --query 'StandardOutputContent' \
        --output text 2>/dev/null)
    
    if [ -n "$COMMAND_OUTPUT" ]; then
        echo -e "${GREEN}✅ Command output:${NC}"
        echo "$COMMAND_OUTPUT"
    else
        echo -e "${YELLOW}⚠️  Could not get command output. Command may still be running.${NC}"
        echo -e "   Check manually: aws ssm get-command-invocation --command-id $COMMAND_ID --instance-id $INSTANCE_ID"
    fi
else
    echo -e "${YELLOW}⚠️  Instance not accessible via Systems Manager${NC}"
    echo -e "   You may need to SSH to the instance to debug further"
fi

echo ""

# Provide troubleshooting recommendations
echo -e "${BLUE}💡 Troubleshooting Recommendations${NC}"
echo -e "${BLUE}=================================${NC}"

echo -e "${YELLOW}1. Check Health Check Configuration:${NC}"
if [ -n "$HTTPS_TG_ARN" ]; then
    HEALTH_CHECK_INFO=$(aws elbv2 describe-target-groups --target-group-arns "$HTTPS_TG_ARN" --query 'TargetGroups[0].{Protocol:HealthCheckProtocol,Port:HealthCheckPort,Path:HealthCheckPath,Interval:HealthCheckIntervalSeconds,Timeout:HealthCheckTimeoutSeconds}' --output table)
    echo "$HEALTH_CHECK_INFO"
fi

echo ""
echo -e "${YELLOW}2. Common Issues and Solutions:${NC}"
echo -e "   • Health check protocol mismatch (TCP vs HTTP)"
echo -e "   • Security group not allowing health check traffic"
echo -e "   • Web server not responding on health check port/path"
echo -e "   • Instance not fully initialized"
echo ""

echo -e "${YELLOW}3. Manual Tests to Try:${NC}"
echo -e "   • SSH to instance and run: /home/ec2-user/check_servers.sh"
echo -e "   • Check logs: sudo tail -f /var/log/user-data.log"
echo -e "   • Test local health checks: curl -k https://localhost/health"
echo -e "   • Check Nginx status: sudo systemctl status nginx"
echo ""

echo -e "${YELLOW}4. Fix Health Check Configuration:${NC}"
echo -e "   The HTTPS target group now uses HTTP health checks on port 80"
echo -e "   This avoids SSL certificate issues with TCP health checks"
echo -e "   Run 'terraform apply' to update the configuration"

echo ""
echo -e "${GREEN}🔧 Quick Fix Command:${NC}"
echo -e "   terraform apply -auto-approve"

echo ""
echo -e "${BLUE}🏁 Debug completed!${NC}"
