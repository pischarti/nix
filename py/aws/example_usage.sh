#!/bin/bash

# Example usage script for AWS VPC Deletion Tool
# This script demonstrates how to use the delete_vpc.py tool

set -e

echo "AWS VPC Deletion Tool - Example Usage"
echo "===================================="

# Check if VPC ID is provided
if [ $# -eq 0 ]; then
    echo "Usage: $0 <vpc-id> [region] [profile]"
    echo ""
    echo "Examples:"
    echo "  $0 vpc-12345678                    # Use default region and profile"
    echo "  $0 vpc-12345678 us-west-2          # Specify region"
    echo "  $0 vpc-12345678 us-west-2 myprofile # Specify region and profile"
    exit 1
fi

VPC_ID=$1
REGION=${2:-"us-east-1"}
PROFILE=${3:-"default"}

echo "VPC ID: $VPC_ID"
echo "Region: $REGION"
echo "Profile: $PROFILE"
echo ""

# First, run in dry-run mode to see what would be deleted
echo "Step 1: Running dry-run to preview deletions..."
uv run delete_vpc.py "$VPC_ID" --region "$REGION" --profile "$PROFILE" --dry-run

echo ""
echo "Step 2: If the dry-run looks correct, run the actual deletion:"
echo "uv run delete_vpc.py $VPC_ID --region $REGION --profile $PROFILE"
echo ""
echo "Or to skip confirmation prompt:"
echo "uv run delete_vpc.py $VPC_ID --region $REGION --profile $PROFILE --force"
