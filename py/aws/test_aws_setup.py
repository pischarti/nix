#!/usr/bin/env python3
"""
Test AWS Setup Script

This script tests if your AWS credentials and region are properly configured
for use with the VPC deletion tool.
"""

import sys
from typing import Optional
import boto3
import click
from rich.console import Console
from rich.panel import Panel
from botocore.exceptions import ClientError, NoCredentialsError, NoRegionError

console = Console()


def test_aws_credentials(profile: Optional[str] = None, region: Optional[str] = None) -> bool:
    """Test AWS credentials and basic connectivity."""
    try:
        # Create session
        session = boto3.Session(profile_name=profile, region_name=region)
        
        # Test credentials with STS
        sts = session.client('sts')
        identity = sts.get_caller_identity()
        
        console.print("[green]✓ AWS Credentials: Valid[/green]")
        console.print(f"  Account ID: {identity.get('Account', 'N/A')}")
        console.print(f"  User ARN: {identity.get('Arn', 'N/A')}")
        
        return True
        
    except NoCredentialsError:
        console.print("[red]✗ AWS Credentials: Not found[/red]")
        console.print("  Please configure AWS credentials using:")
        console.print("  - aws configure")
        console.print("  - AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables")
        console.print("  - IAM roles (for EC2 instances)")
        return False
        
    except NoRegionError:
        console.print("[red]✗ AWS Region: Not configured[/red]")
        console.print("  Please set a default region using:")
        console.print("  - aws configure set region us-east-1")
        console.print("  - AWS_DEFAULT_REGION environment variable")
        console.print("  - --region parameter when running the script")
        return False
        
    except ClientError as e:
        console.print(f"[red]✗ AWS API Error: {e}[/red]")
        return False


def test_ec2_permissions(profile: Optional[str] = None, region: Optional[str] = None) -> bool:
    """Test EC2 permissions required for VPC operations."""
    try:
        session = boto3.Session(profile_name=profile, region_name=region)
        ec2 = session.client('ec2')
        
        # Test basic EC2 describe permissions
        ec2.describe_vpcs(MaxResults=1)
        console.print("[green]✓ EC2 Permissions: Basic describe operations work[/green]")
        
        return True
        
    except ClientError as e:
        error_code = e.response.get('Error', {}).get('Code', '')
        if error_code == 'UnauthorizedOperation':
            console.print("[red]✗ EC2 Permissions: Insufficient permissions[/red]")
            console.print("  The current credentials don't have EC2 describe permissions")
        else:
            console.print(f"[red]✗ EC2 API Error: {e}[/red]")
        return False


def list_vpcs(profile: Optional[str] = None, region: Optional[str] = None) -> None:
    """List available VPCs for testing."""
    try:
        session = boto3.Session(profile_name=profile, region_name=region)
        ec2 = session.client('ec2')
        
        response = ec2.describe_vpcs()
        vpcs = response.get('Vpcs', [])
        
        if not vpcs:
            console.print("[yellow]No VPCs found in the current region[/yellow]")
            return
        
        console.print(f"[blue]Found {len(vpcs)} VPC(s) in {session.region_name}:[/blue]")
        
        for vpc in vpcs:
            vpc_id = vpc['VpcId']
            cidr = vpc.get('CidrBlock', 'N/A')
            is_default = vpc.get('IsDefault', False)
            state = vpc.get('State', 'N/A')
            
            # Get VPC name from tags
            vpc_name = 'N/A'
            for tag in vpc.get('Tags', []):
                if tag['Key'] == 'Name':
                    vpc_name = tag['Value']
                    break
            
            default_indicator = " (DEFAULT)" if is_default else ""
            console.print(f"  • {vpc_id}: {vpc_name} - {cidr} - {state}{default_indicator}")
            
    except ClientError as e:
        console.print(f"[red]Error listing VPCs: {e}[/red]")


@click.command()
@click.option('--region', '-r', help='AWS region to test')
@click.option('--profile', '-p', help='AWS profile to test')
def main(region: Optional[str], profile: Optional[str]):
    """Test AWS setup for VPC deletion tool."""
    console.print(Panel.fit(
        "[bold blue]AWS Setup Test for VPC Deletion Tool[/bold blue]",
        border_style="blue"
    ))
    
    console.print(f"Testing AWS setup...")
    console.print(f"Region: {region or 'default'}")
    console.print(f"Profile: {profile or 'default'}")
    console.print()
    
    # Test credentials
    creds_ok = test_aws_credentials(profile, region)
    if not creds_ok:
        console.print("\n[red]❌ AWS setup test failed![/red]")
        sys.exit(1)
    
    console.print()
    
    # Test EC2 permissions
    perms_ok = test_ec2_permissions(profile, region)
    if not perms_ok:
        console.print("\n[yellow]⚠️  Limited permissions detected[/yellow]")
        console.print("The VPC deletion tool may not work properly with current permissions.")
    
    console.print()
    
    # List VPCs for reference
    console.print("[blue]Available VPCs:[/blue]")
    list_vpcs(profile, region)
    
    console.print()
    
    if creds_ok and perms_ok:
        console.print("[bold green]✅ AWS setup looks good![/bold green]")
        console.print("\nYou can now use the VPC deletion tool:")
        console.print("  uv run delete_vpc.py <vpc-id> --dry-run")
    else:
        console.print("[bold yellow]⚠️  AWS setup has issues[/bold yellow]")
        console.print("Please fix the issues above before using the VPC deletion tool.")


if __name__ == "__main__":
    main()
