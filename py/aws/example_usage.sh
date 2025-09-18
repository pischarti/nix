#!/bin/bash

# Example usage script for AWS VPC Deletion Tool
# This script demonstrates how to use the delete_vpc.py tool

set -e

echo "AWS VPC Deletion Tool - Example Usage"
echo "===================================="

# Check if VPC ID is provided
if [ $# -eq 0 ]; then
    echo "Usage: $0 <vpc-id>"
    echo ""
    echo "Set AWS region and profile via environment variables:"
    echo "  export AWS_DEFAULT_REGION=us-west-2"
    echo "  export AWS_PROFILE=myprofile"
    echo "  $0 vpc-12345678"
    exit 1
fi

VPC_ID=$1

echo "VPC ID: $VPC_ID"
echo "Region: ${AWS_DEFAULT_REGION:-${AWS_REGION:-default}}"
echo "Profile: ${AWS_PROFILE:-default}"
echo ""

# First, run in dry-run mode to see what would be deleted
echo "Step 1: Running dry-run to preview deletions..."
uv run delete_vpc.py "$VPC_ID" --dry-run

echo ""
echo "Step 2: If the dry-run looks correct, run the actual deletion:"
echo "uv run delete_vpc.py $VPC_ID"
echo ""
echo "Or to skip confirmation prompt:"
echo "uv run delete_vpc.py $VPC_ID --force"
