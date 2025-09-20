#!/usr/bin/env python3
"""
AWS Transit Gateway Route Viewer

This script displays all VPC attachments and routes for a given Transit Gateway ID.
It provides detailed information about attachment configurations and routing tables.

Usage:
    python tgw_route.py <tgw-id>
    
Or with uv:
    uv run tgw_route.py <tgw-id>
"""

import sys
import os
import yaml
from typing import List, Dict, Any, Optional
import boto3
import click
from rich.console import Console
from rich.table import Table
from rich.panel import Panel
from rich import print as rprint
from botocore.exceptions import ClientError, BotoCoreError

console = Console()


class TGWRouteViewer:
    """Handles Transit Gateway route and attachment viewing."""
    
    def __init__(self, tgw_id: str):
        """Initialize the TGW route viewer.
        
        Args:
            tgw_id: The Transit Gateway ID to analyze
            
        Note:
            AWS region and profile are determined from environment variables:
            - AWS_DEFAULT_REGION or AWS_REGION for region
            - AWS_PROFILE for profile (optional)
        """
        self.tgw_id = tgw_id
        
        # Validate required environment variables
        self._validate_aws_config()
        
        # Initialize AWS session and clients using environment variables
        session = boto3.Session()
        self.ec2_client = session.client('ec2')
        
        # Verify the Transit Gateway exists
        self._verify_tgw_exists()
    
    def _validate_aws_config(self) -> None:
        """Validate that AWS configuration is available."""
        required_vars = ['AWS_ACCESS_KEY_ID', 'AWS_SECRET_ACCESS_KEY']
        region_vars = ['AWS_DEFAULT_REGION', 'AWS_REGION']
        
        # Check for access credentials
        if not any(os.getenv(var) for var in required_vars):
            console.print("[red]Error: AWS credentials not found.[/red]")
            console.print("Please set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables.")
            console.print("Or configure AWS CLI with 'aws configure'.")
            sys.exit(1)
        
        # Check for region
        if not any(os.getenv(var) for var in region_vars):
            console.print("[yellow]Warning: No AWS region specified.[/yellow]")
            console.print("Please set AWS_DEFAULT_REGION or AWS_REGION environment variable.")
            console.print("Using 'us-east-1' as default.")
    
    def _verify_tgw_exists(self) -> None:
        """Verify that the specified Transit Gateway exists."""
        try:
            response = self.ec2_client.describe_transit_gateways(TransitGatewayIds=[self.tgw_id])
            if not response['TransitGateways']:
                console.print(f"[red]Error: Transit Gateway '{self.tgw_id}' not found.[/red]")
                sys.exit(1)
            
            tgw = response['TransitGateways'][0]
            self.tgw_state = tgw['State']
            self.tgw_description = tgw.get('Description', 'No description')
            self.tgw_owner_id = tgw['OwnerId']
            self.tgw_arn = tgw['TransitGatewayArn']
            
            console.print(f"[green]Found Transit Gateway: {self.tgw_id}[/green]")
            console.print(f"  State: {self.tgw_state}")
            console.print(f"  Owner: {self.tgw_owner_id}")
            console.print(f"  Description: {self.tgw_description}")
            console.print()
            
        except ClientError as e:
            error_code = e.response['Error']['Code']
            if error_code == 'InvalidTransitGatewayID.NotFound':
                console.print(f"[red]Error: Transit Gateway '{self.tgw_id}' not found.[/red]")
            else:
                console.print(f"[red]Error describing Transit Gateway: {e}[/red]")
            sys.exit(1)
    
    def get_vpc_attachments(self) -> List[Dict[str, Any]]:
        """Get all VPC attachments for the Transit Gateway."""
        try:
            response = self.ec2_client.describe_transit_gateway_vpc_attachments(
                Filters=[{'Name': 'transit-gateway-id', 'Values': [self.tgw_id]}]
            )
            return response['TransitGatewayVpcAttachments']
        except ClientError as e:
            console.print(f"[red]Error describing VPC attachments: {e}[/red]")
            return []
    
    def get_route_tables(self) -> List[Dict[str, Any]]:
        """Get all route tables associated with the Transit Gateway."""
        try:
            response = self.ec2_client.describe_transit_gateway_route_tables(
                Filters=[{'Name': 'transit-gateway-id', 'Values': [self.tgw_id]}]
            )
            return response['TransitGatewayRouteTables']
        except ClientError as e:
            console.print(f"[red]Error describing route tables: {e}[/red]")
            return []
    
    def get_routes(self, route_table_id: str) -> List[Dict[str, Any]]:
        """Get all routes for a specific Transit Gateway route table."""
        try:
            response = self.ec2_client.search_transit_gateway_routes(
                TransitGatewayRouteTableId=route_table_id,
                Filters=[{'Name': 'state', 'Values': ['active', 'blackhole']}]
            )
            return response['Routes']
        except ClientError as e:
            console.print(f"[red]Error searching routes for route table {route_table_id}: {e}[/red]")
            return []
    
    def get_attachment_associations(self, route_table_id: str) -> List[Dict[str, Any]]:
        """Get attachment associations for a route table."""
        try:
            response = self.ec2_client.get_transit_gateway_route_table_associations(
                TransitGatewayRouteTableId=route_table_id
            )
            return response['Associations']
        except ClientError as e:
            console.print(f"[red]Error getting associations for route table {route_table_id}: {e}[/red]")
            return []
    
    def get_attachment_propagations(self, route_table_id: str) -> List[Dict[str, Any]]:
        """Get attachment propagations for a route table."""
        try:
            response = self.ec2_client.get_transit_gateway_route_table_propagations(
                TransitGatewayRouteTableId=route_table_id
            )
            return response['TransitGatewayRouteTablePropagations']
        except ClientError as e:
            console.print(f"[red]Error getting propagations for route table {route_table_id}: {e}[/red]")
            return []
    
    def display_summary(self, attachments: List[Dict[str, Any]], route_tables: List[Dict[str, Any]]) -> None:
        """Display a summary of the Transit Gateway configuration."""
        # Count attachments by state
        attachment_states = {}
        for attachment in attachments:
            state = attachment['State']
            attachment_states[state] = attachment_states.get(state, 0) + 1
        
        # Count route tables
        route_table_count = len(route_tables)
        
        summary_data = [
            f"Transit Gateway ID: {self.tgw_id}",
            f"State: {self.tgw_state}",
            f"Owner: {self.tgw_owner_id}",
            f"Total VPC Attachments: {len(attachments)}",
            f"Route Tables: {route_table_count}"
        ]
        
        if attachment_states:
            summary_data.append("Attachment States:")
            for state, count in attachment_states.items():
                summary_data.append(f"  {state}: {count}")
        
        console.print(Panel("\n".join(summary_data), title="Transit Gateway Summary", border_style="green"))
    
    def display_attachments(self, attachments: List[Dict[str, Any]]) -> None:
        """Display VPC attachments in a formatted table."""
        if not attachments:
            console.print("[yellow]No VPC attachments found for this Transit Gateway.[/yellow]")
            return
        
        table = Table(title=f"VPC Attachments for Transit Gateway {self.tgw_id}")
        table.add_column("Attachment ID", style="cyan", width=20)
        table.add_column("VPC ID", style="green", width=15)
        table.add_column("Subnet IDs", style="blue", width=25)
        table.add_column("State", style="yellow", width=12)
        table.add_column("DNS Support", style="magenta", width=12)
        table.add_column("IPv6 Support", style="bright_blue", width=12)
        
        for attachment in attachments:
            subnet_ids = attachment.get('SubnetIds', [])
            dns_support = "Enabled" if attachment.get('Options', {}).get('DnsSupport') == 'enable' else "Disabled"
            ipv6_support = "Enabled" if attachment.get('Options', {}).get('Ipv6Support') == 'enable' else "Disabled"
            
            if subnet_ids:
                # Add first row with attachment info
                table.add_row(
                    attachment['TransitGatewayAttachmentId'],
                    attachment['VpcId'],
                    subnet_ids[0],
                    attachment['State'],
                    dns_support,
                    ipv6_support
                )
                
                # Add additional rows for remaining subnets
                for subnet_id in subnet_ids[1:]:
                    table.add_row(
                        "",  # Empty attachment ID
                        "",  # Empty VPC ID
                        subnet_id,
                        "",  # Empty state
                        "",  # Empty DNS support
                        ""   # Empty IPv6 support
                    )
            else:
                # No subnets case
                table.add_row(
                    attachment['TransitGatewayAttachmentId'],
                    attachment['VpcId'],
                    "No subnets",
                    attachment['State'],
                    dns_support,
                    ipv6_support
                )
        
        console.print(table)
        console.print()
    
    def display_route_tables(self, route_tables: List[Dict[str, Any]]) -> None:
        """Display route tables and their routes in formatted tables."""
        if not route_tables:
            console.print("[yellow]No route tables found for this Transit Gateway.[/yellow]")
            return
        
        for rt in route_tables:
            rt_id = rt['TransitGatewayRouteTableId']
            
            # Get route table name from tags
            rt_name = rt_id
            if rt.get('Tags'):
                for tag in rt['Tags']:
                    if tag['Key'] == 'Name':
                        rt_name = tag['Value']
                        break
            
            # Get associations and propagations
            associations = self.get_attachment_associations(rt_id)
            propagations = self.get_attachment_propagations(rt_id)
            
            # Get routes
            routes = self.get_routes(rt_id)
            
            # Display route table info
            info_text = f"Route Table: {rt_id}\nName: {rt_name}\nState: {rt['State']}"
            console.print(Panel(info_text, title="Route Table Information", border_style="blue"))
            
            # Display associations
            if associations:
                assoc_table = Table(title=f"Associations for {rt_name}")
                assoc_table.add_column("Resource Type", style="cyan", width=15)
                assoc_table.add_column("Resource ID", style="green", width=20)
                assoc_table.add_column("State", style="yellow", width=10)
                
                for assoc in associations:
                    assoc_table.add_row(
                        assoc['ResourceType'],
                        assoc['ResourceId'],
                        assoc['State']
                    )
                console.print(assoc_table)
                console.print()
            
            # Display propagations
            if propagations:
                prop_table = Table(title=f"Propagations for {rt_name}")
                prop_table.add_column("Resource Type", style="cyan", width=15)
                prop_table.add_column("Resource ID", style="green", width=20)
                prop_table.add_column("State", style="yellow", width=10)
                
                for prop in propagations:
                    prop_table.add_row(
                        prop['ResourceType'],
                        prop['ResourceId'],
                        prop['State']
                    )
                console.print(prop_table)
                console.print()
            
            # Display routes
            if routes:
                routes_table = Table(title=f"Routes in {rt_name}")
                routes_table.add_column("Destination", style="cyan", width=20)
                routes_table.add_column("Type", style="green", width=10)
                routes_table.add_column("State", style="yellow", width=10)
                routes_table.add_column("Attachment", style="blue", width=20)
                routes_table.add_column("Resource", style="magenta", width=20)
                
                for route in routes:
                    attachment_id = route.get('TransitGatewayAttachments', [{}])[0].get('TransitGatewayAttachmentId', 'N/A')
                    resource_id = route.get('TransitGatewayAttachments', [{}])[0].get('ResourceId', 'N/A')
                    
                    routes_table.add_row(
                        route['DestinationCidrBlock'],
                        route['Type'],
                        route['State'],
                        attachment_id,
                        resource_id
                    )
                console.print(routes_table)
                console.print()
            else:
                console.print(f"[yellow]No routes found in {rt_name}[/yellow]")
                console.print()
    
    def generate_yaml_output(self, attachments: List[Dict[str, Any]], route_tables: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Generate structured YAML output for the Transit Gateway data."""
        
        # Process attachments
        processed_attachments = []
        for attachment in attachments:
            processed_attachments.append({
                'attachment_id': attachment['TransitGatewayAttachmentId'],
                'vpc_id': attachment['VpcId'],
                'subnet_ids': attachment.get('SubnetIds', []),
                'state': attachment['State'],
                'dns_support': attachment.get('Options', {}).get('DnsSupport') == 'enable',
                'ipv6_support': attachment.get('Options', {}).get('Ipv6Support') == 'enable',
                'tags': {tag['Key']: tag['Value'] for tag in attachment.get('Tags', [])}
            })
        
        # Process route tables with routes
        processed_route_tables = []
        for rt in route_tables:
            rt_id = rt['TransitGatewayRouteTableId']
            
            # Get route table name
            rt_name = rt_id
            if rt.get('Tags'):
                for tag in rt['Tags']:
                    if tag['Key'] == 'Name':
                        rt_name = tag['Value']
                        break
            
            # Get associations, propagations, and routes
            associations = self.get_attachment_associations(rt_id)
            propagations = self.get_attachment_propagations(rt_id)
            routes = self.get_routes(rt_id)
            
            # Process routes
            processed_routes = []
            for route in routes:
                attachment_info = route.get('TransitGatewayAttachments', [{}])[0]
                processed_routes.append({
                    'destination': route['DestinationCidrBlock'],
                    'type': route['Type'],
                    'state': route['State'],
                    'attachment_id': attachment_info.get('TransitGatewayAttachmentId', 'N/A'),
                    'resource_id': attachment_info.get('ResourceId', 'N/A')
                })
            
            processed_route_tables.append({
                'route_table_id': rt_id,
                'name': rt_name,
                'state': rt['State'],
                'associations': [
                    {
                        'resource_type': assoc['ResourceType'],
                        'resource_id': assoc['ResourceId'],
                        'state': assoc['State']
                    } for assoc in associations
                ],
                'propagations': [
                    {
                        'resource_type': prop['ResourceType'],
                        'resource_id': prop['ResourceId'],
                        'state': prop['State']
                    } for prop in propagations
                ],
                'routes': processed_routes,
                'tags': {tag['Key']: tag['Value'] for tag in rt.get('Tags', [])}
            })
        
        # Calculate summary statistics
        attachment_states = {}
        for attachment in attachments:
            state = attachment['State']
            attachment_states[state] = attachment_states.get(state, 0) + 1
        
        return {
            'tgw_info': {
                'tgw_id': self.tgw_id,
                'state': self.tgw_state,
                'owner_id': self.tgw_owner_id,
                'description': self.tgw_description,
                'arn': self.tgw_arn
            },
            'summary': {
                'total_attachments': len(attachments),
                'attachment_states': attachment_states,
                'route_tables': len(route_tables),
                'total_routes': sum(len(self.get_routes(rt['TransitGatewayRouteTableId'])) for rt in route_tables)
            },
            'attachments': processed_attachments,
            'route_tables': processed_route_tables
        }


@click.command()
@click.argument('tgw_id')
@click.option('--output-format', 'output_format',
              type=click.Choice(['table', 'yaml'], case_sensitive=False),
              default='table',
              help='Output format: table (default) or yaml')
def main(tgw_id: str, output_format: str) -> None:
    """Display all VPC attachments and routes for the specified Transit Gateway.
    
    TGW_ID: The ID of the Transit Gateway to analyze
    """
    try:
        # Initialize the TGW route viewer
        viewer = TGWRouteViewer(tgw_id)
        
        # Get all resources
        console.print("[blue]Fetching Transit Gateway resources...[/blue]")
        attachments = viewer.get_vpc_attachments()
        route_tables = viewer.get_route_tables()
        
        # Display results based on output format
        if output_format == 'yaml':
            yaml_data = viewer.generate_yaml_output(attachments, route_tables)
            print(yaml.dump(yaml_data, default_flow_style=False, sort_keys=False))
        else:
            viewer.display_summary(attachments, route_tables)
            viewer.display_attachments(attachments)
            viewer.display_route_tables(route_tables)
            console.print("[green]Transit Gateway analysis complete![/green]")
        
    except KeyboardInterrupt:
        console.print("\n[yellow]Operation cancelled by user.[/yellow]")
        sys.exit(1)
    except Exception as e:
        console.print(f"[red]Unexpected error: {e}[/red]")
        sys.exit(1)


if __name__ == "__main__":
    main()
