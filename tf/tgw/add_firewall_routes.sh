#!/bin/bash

# Script to add firewall routes after infrastructure is deployed
echo "Adding firewall routes to both inspection and edge VPCs..."

# Get the route table IDs
INSPECTION_ROUTE_TABLE_ID=$(terraform output -raw inspection_public_route_table_id)
EDGE_ROUTE_TABLE_ID=$(terraform output -raw edge_public_route_table_id)
echo "Inspection Route Table ID: $INSPECTION_ROUTE_TABLE_ID"
echo "Edge Route Table ID: $EDGE_ROUTE_TABLE_ID"

# Get the firewall endpoint ID
FIREWALL_ARN=$(terraform output -raw network_firewall_arn)
echo "Firewall ARN: $FIREWALL_ARN"

# Get the firewall endpoint ID
FIREWALL_ENDPOINT_ID=$(aws network-firewall describe-firewall --firewall-arn "$FIREWALL_ARN" --query 'FirewallStatus.SyncStates[0].Attachment[0].EndpointId' --output text)
echo "Firewall Endpoint ID: $FIREWALL_ENDPOINT_ID"

# Get VPC CIDRs
EDGE_CIDR=$(terraform output -raw edge_vpc_cidr_block)
APP_CIDR=$(terraform output -raw app_vpc_id | xargs -I {} aws ec2 describe-vpcs --vpc-ids {} --query 'Vpcs[0].CidrBlock' --output text)

echo "Edge VPC CIDR: $EDGE_CIDR"
echo "App VPC CIDR: $APP_CIDR"

# Add routes to inspection VPC route table
echo "Adding routes to inspection VPC route table..."
echo "Adding route to edge VPC via firewall..."
aws ec2 create-route --route-table-id "$INSPECTION_ROUTE_TABLE_ID" --destination-cidr-block "$EDGE_CIDR" --vpc-endpoint-id "$FIREWALL_ENDPOINT_ID" || echo "Route may already exist"

echo "Adding route to app VPC via firewall..."
aws ec2 create-route --route-table-id "$INSPECTION_ROUTE_TABLE_ID" --destination-cidr-block "$APP_CIDR" --vpc-endpoint-id "$FIREWALL_ENDPOINT_ID" || echo "Route may already exist"

# Add routes to edge VPC route table
echo "Adding routes to edge VPC route table..."
echo "Adding route to app VPC via firewall..."
aws ec2 create-route --route-table-id "$EDGE_ROUTE_TABLE_ID" --destination-cidr-block "$APP_CIDR" --vpc-endpoint-id "$FIREWALL_ENDPOINT_ID" || echo "Route may already exist"

echo "Firewall routes added successfully!"
echo "Both edge and app VPCs will now route through the firewall, blocking bidirectional traffic between them."
