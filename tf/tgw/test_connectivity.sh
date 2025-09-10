#!/bin/bash

# Test connectivity script for TGW setup
echo "=== Testing TGW Connectivity ==="

# Get IPs
EDGE_PUBLIC_IP=$(terraform output -raw edge_public_instance_public_ip)
APP_PUBLIC_IP=$(terraform output -raw app_public_instance_public_ip)
APP_PRIVATE_IP=$(terraform output -raw app_public_instance_private_ip)
INSPECTION_PUBLIC_IP=$(terraform output -raw inspection_public_instance_public_ip)

echo "Edge Public IP: $EDGE_PUBLIC_IP"
echo "App Public IP: $APP_PUBLIC_IP"
echo "App Private IP: $APP_PRIVATE_IP"
echo "Inspection Public IP: $INSPECTION_PUBLIC_IP"

echo ""
echo "=== Testing from Edge to App (Public IP) ==="
ssh -i edge_generated.pem -o StrictHostKeyChecking=no ec2-user@$EDGE_PUBLIC_IP "curl -v http://$APP_PUBLIC_IP" || echo "Failed to connect to app public IP"

echo ""
echo "=== Testing from Edge to App (Private IP) ==="
ssh -i edge_generated.pem -o StrictHostKeyChecking=no ec2-user@$EDGE_PUBLIC_IP "curl -v http://$APP_PRIVATE_IP" || echo "Failed to connect to app private IP"

echo ""
echo "=== Testing from Edge to Inspection (Public IP) ==="
ssh -i edge_generated.pem -o StrictHostKeyChecking=no ec2-user@$EDGE_PUBLIC_IP "curl -v http://$INSPECTION_PUBLIC_IP" || echo "Failed to connect to inspection public IP"

echo ""
echo "=== Testing from App to Edge (Public IP) ==="
ssh -i edge_generated.pem -o StrictHostKeyChecking=no ec2-user@$APP_PUBLIC_IP "curl -v http://$EDGE_PUBLIC_IP" || echo "Failed to connect to edge public IP"

echo ""
echo "=== Testing from Inspection to App (Public IP) ==="
ssh -i edge_generated.pem -o StrictHostKeyChecking=no ec2-user@$INSPECTION_PUBLIC_IP "curl -v http://$APP_PUBLIC_IP" || echo "Failed to connect to app public IP"

echo ""
echo "=== Testing from Inspection to App (Private IP) ==="
ssh -i edge_generated.pem -o StrictHostKeyChecking=no ec2-user@$INSPECTION_PUBLIC_IP "curl -v http://$APP_PRIVATE_IP" || echo "Failed to connect to app private IP"
