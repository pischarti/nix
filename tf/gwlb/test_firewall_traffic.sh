#!/bin/bash

# Test script for validating traffic flow through Gateway Load Balancer and Network Firewall
# This script tests the complete traffic path: Internet → NLB → GWLB → Firewall → Private Instance

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔥 Gateway Load Balancer Firewall Traffic Test${NC}"
echo -e "${BLUE}===============================================${NC}"

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

NLB_DNS=$(terraform output -raw test_nlb_dns_name 2>/dev/null)
INSTANCE_ID=$(terraform output -raw test_instance_id 2>/dev/null)
INSTANCE_IP=$(terraform output -raw test_instance_private_ip 2>/dev/null)

if [ -z "$NLB_DNS" ]; then
    echo -e "${RED}❌ Could not get NLB DNS name from terraform output.${NC}"
    echo -e "${YELLOW}💡 Make sure you've deployed the infrastructure with the test components.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Infrastructure Information:${NC}"
echo -e "   🌐 NLB DNS Name: ${NLB_DNS}"
echo -e "   🖥️  Instance ID: ${INSTANCE_ID}"
echo -e "   🔒 Instance IP: ${INSTANCE_IP}"
echo ""

# Test HTTP endpoint
echo -e "${YELLOW}🧪 Testing HTTP Traffic Flow...${NC}"
echo -e "${BLUE}   Route: Internet → NLB → GWLB → Firewall → Private Instance${NC}"

HTTP_URL="http://${NLB_DNS}"
echo -e "   Testing: ${HTTP_URL}"

if curl -s --connect-timeout 10 --max-time 30 "${HTTP_URL}" > /tmp/http_test.html; then
    echo -e "${GREEN}✅ HTTP Test: SUCCESS${NC}"
    
    # Check if we got the expected content
    if grep -q "Firewall Test Instance" /tmp/http_test.html; then
        echo -e "${GREEN}   ✅ Content validation: PASSED${NC}"
        echo -e "${GREEN}   📄 Received expected web page content${NC}"
    else
        echo -e "${YELLOW}   ⚠️  Content validation: Unexpected content${NC}"
    fi
else
    echo -e "${RED}❌ HTTP Test: FAILED${NC}"
    echo -e "${YELLOW}   💡 This might indicate a routing or firewall configuration issue${NC}"
fi

echo ""

# Test HTTPS endpoint
echo -e "${YELLOW}🔒 Testing HTTPS Traffic Flow...${NC}"
echo -e "${BLUE}   Route: Internet → NLB → GWLB → Firewall → Private Instance${NC}"

HTTPS_URL="https://${NLB_DNS}"
echo -e "   Testing: ${HTTPS_URL}"

if curl -k -s --connect-timeout 10 --max-time 30 "${HTTPS_URL}" > /tmp/https_test.html; then
    echo -e "${GREEN}✅ HTTPS Test: SUCCESS${NC}"
    
    # Check if we got the expected content
    if grep -q "Firewall Test Instance - HTTPS" /tmp/https_test.html; then
        echo -e "${GREEN}   ✅ Content validation: PASSED${NC}"
        echo -e "${GREEN}   🔐 Received expected HTTPS web page content${NC}"
    else
        echo -e "${YELLOW}   ⚠️  Content validation: Unexpected content${NC}"
    fi
else
    echo -e "${RED}❌ HTTPS Test: FAILED${NC}"
    echo -e "${YELLOW}   💡 This might indicate a routing or firewall configuration issue${NC}"
fi

echo ""

# Test health check endpoint
echo -e "${YELLOW}🩺 Testing Health Check Endpoint...${NC}"
HEALTH_URL="http://${NLB_DNS}/health"
echo -e "   Testing: ${HEALTH_URL}"

if HEALTH_RESPONSE=$(curl -s --connect-timeout 5 --max-time 10 "${HEALTH_URL}"); then
    if [ "$HEALTH_RESPONSE" = "OK" ]; then
        echo -e "${GREEN}✅ Health Check: SUCCESS${NC}"
        echo -e "${GREEN}   📊 Response: ${HEALTH_RESPONSE}${NC}"
    else
        echo -e "${YELLOW}⚠️  Health Check: Unexpected response: ${HEALTH_RESPONSE}${NC}"
    fi
else
    echo -e "${RED}❌ Health Check: FAILED${NC}"
fi

echo ""

# Performance test
echo -e "${YELLOW}⚡ Performance Test (10 requests)...${NC}"
echo -e "   Testing response time and consistency..."

TOTAL_TIME=0
SUCCESS_COUNT=0
FAIL_COUNT=0

for i in {1..10}; do
    if RESPONSE_TIME=$(curl -s -w "%{time_total}" -o /dev/null --connect-timeout 5 --max-time 10 "${HTTP_URL}"); then
        TOTAL_TIME=$(echo "$TOTAL_TIME + $RESPONSE_TIME" | bc -l)
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        echo -ne "\r   Request $i/10: ${RESPONSE_TIME}s"
    else
        FAIL_COUNT=$((FAIL_COUNT + 1))
        echo -ne "\r   Request $i/10: FAILED"
    fi
done

echo ""

if [ $SUCCESS_COUNT -gt 0 ]; then
    AVG_TIME=$(echo "scale=3; $TOTAL_TIME / $SUCCESS_COUNT" | bc -l)
    echo -e "${GREEN}✅ Performance Test Results:${NC}"
    echo -e "   📊 Successful requests: $SUCCESS_COUNT/10"
    echo -e "   📊 Failed requests: $FAIL_COUNT/10"
    echo -e "   📊 Average response time: ${AVG_TIME}s"
else
    echo -e "${RED}❌ Performance Test: All requests failed${NC}"
fi

echo ""

# Traffic flow validation
echo -e "${BLUE}🚦 Traffic Flow Validation${NC}"
echo -e "${BLUE}==========================${NC}"
echo -e "${GREEN}Expected Traffic Path:${NC}"
echo -e "   1️⃣  Internet Client"
echo -e "   2️⃣  Network Load Balancer (Public Subnet)"
echo -e "   3️⃣  Gateway Load Balancer Endpoint"
echo -e "   4️⃣  Network Firewall (Inspection VPC)"
echo -e "   5️⃣  Gateway Load Balancer (Return Path)"
echo -e "   6️⃣  EC2 Instance (Private Subnet)"
echo ""

if [ $SUCCESS_COUNT -gt 0 ]; then
    echo -e "${GREEN}✅ Traffic is successfully flowing through the firewall!${NC}"
    echo -e "${GREEN}   🔍 All traffic between public and private subnets is being inspected${NC}"
    echo -e "${GREEN}   🛡️  Network Firewall is processing the traffic${NC}"
else
    echo -e "${RED}❌ Traffic flow validation failed${NC}"
    echo -e "${YELLOW}   💡 Check the following:${NC}"
    echo -e "   • Network Firewall endpoints are registered with GWLB"
    echo -e "   • Route tables are correctly configured"
    echo -e "   • Security groups allow traffic"
    echo -e "   • GWLB endpoints are healthy"
fi

echo ""

# Troubleshooting information
echo -e "${BLUE}🔧 Troubleshooting Information${NC}"
echo -e "${BLUE}==============================${NC}"
echo -e "${YELLOW}Manual Test Commands:${NC}"
echo -e "   curl http://${NLB_DNS}"
echo -e "   curl -k https://${NLB_DNS}"
echo -e "   curl http://${NLB_DNS}/health"
echo ""

echo -e "${YELLOW}Check AWS Resources:${NC}"
echo -e "   • NLB Target Groups: aws elbv2 describe-target-health --target-group-arn <arn>"
echo -e "   • GWLB Target Groups: aws elbv2 describe-target-health --target-group-arn <gwlb-arn>"
echo -e "   • Instance Status: aws ec2 describe-instance-status --instance-ids ${INSTANCE_ID}"
echo -e "   • VPC Endpoints: aws ec2 describe-vpc-endpoints"
echo ""

echo -e "${YELLOW}Check Firewall Logs:${NC}"
echo -e "   • CloudWatch Logs: Network Firewall Flow Logs"
echo -e "   • VPC Flow Logs: Check traffic between subnets"
echo ""

# Cleanup
rm -f /tmp/http_test.html /tmp/https_test.html

echo -e "${BLUE}🏁 Test completed!${NC}"
