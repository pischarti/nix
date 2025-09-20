#!/usr/bin/env python3
"""
AWS VPC Route Viewer

This script displays all subnets and routes for a given VPC ID.
It provides detailed information about subnet configuration and routing tables.

Usage:
    python vpc_route.py <vpc-id>
    
Or with uv:
    uv run vpc_route.py <vpc-id>
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


class VPCRouteViewer:
    """Handles VPC route and subnet viewing."""
    
    def __init__(self, vpc_id: str):
        """Initialize the VPC route viewer.
        
        Args:
            vpc_id: The VPC ID to analyze
            
        Note:
            AWS region and profile are determined from environment variables:
            - AWS_DEFAULT_REGION or AWS_REGION for region
            - AWS_PROFILE for profile (optional)
        """
        self.vpc_id = vpc_id
        
        # Validate required environment variables
        self._validate_aws_config()
        
        # Initialize AWS session and clients using environment variables
        session = boto3.Session()
        self.ec2_client = session.client('ec2')
        
        # Verify the VPC exists
        self._verify_vpc_exists()
    
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
    
    def _verify_vpc_exists(self) -> None:
        """Verify that the specified VPC exists."""
        try:
            response = self.ec2_client.describe_vpcs(VpcIds=[self.vpc_id])
            if not response['Vpcs']:
                console.print(f"[red]Error: VPC '{self.vpc_id}' not found.[/red]")
                sys.exit(1)
            
            vpc = response['Vpcs'][0]
            self.vpc_cidr = vpc['CidrBlock']
            self.vpc_state = vpc['State']
            
            console.print(f"[green]Found VPC: {self.vpc_id}[/green]")
            console.print(f"  CIDR Block: {self.vpc_cidr}")
            console.print(f"  State: {self.vpc_state}")
            console.print()
            
        except ClientError as e:
            error_code = e.response['Error']['Code']
            if error_code == 'InvalidVpcID.NotFound':
                console.print(f"[red]Error: VPC '{self.vpc_id}' not found.[/red]")
            else:
                console.print(f"[red]Error describing VPC: {e}[/red]")
            sys.exit(1)
    
    def get_subnets(self) -> List[Dict[str, Any]]:
        """Get all subnets in the VPC."""
        try:
            response = self.ec2_client.describe_subnets(
                Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
            )
            return response['Subnets']
        except ClientError as e:
            console.print(f"[red]Error describing subnets: {e}[/red]")
            return []
    
    def get_route_tables(self) -> List[Dict[str, Any]]:
        """Get all route tables associated with the VPC."""
        try:
            response = self.ec2_client.describe_route_tables(
                Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
            )
            return response['RouteTables']
        except ClientError as e:
            console.print(f"[red]Error describing route tables: {e}[/red]")
            return []
    
    def get_nat_gateways(self) -> Dict[str, Dict[str, Any]]:
        """Get NAT gateways in the VPC for route analysis."""
        try:
            response = self.ec2_client.describe_nat_gateways(
                Filter=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
            )
            return {ng['NatGatewayId']: ng for ng in response['NatGateways']}
        except ClientError as e:
            console.print(f"[red]Error describing NAT gateways: {e}[/red]")
            return {}
    
    def get_internet_gateways(self) -> Dict[str, Dict[str, Any]]:
        """Get internet gateways attached to the VPC."""
        try:
            response = self.ec2_client.describe_internet_gateways(
                Filters=[{'Name': 'attachment.vpc-id', 'Values': [self.vpc_id]}]
            )
            return {ig['InternetGatewayId']: ig for ig in response['InternetGateways']}
        except ClientError as e:
            console.print(f"[red]Error describing internet gateways: {e}[/red]")
            return {}
    
    def display_subnets(self, subnets: List[Dict[str, Any]], route_tables: List[Dict[str, Any]], 
                       sort_by: str = "type") -> None:
        """Display subnets in a formatted table.
        
        Args:
            subnets: List of subnet data from AWS
            route_tables: List of route table data from AWS
            sort_by: Sort order - 'type' or 'cidr' (default: 'type')
        """
        if not subnets:
            console.print("[yellow]No subnets found in this VPC.[/yellow]")
            return
        
        # Create a mapping of subnet ID to route table ID
        subnet_to_route_table = {}
        for rt in route_tables:
            for association in rt['Associations']:
                if 'SubnetId' in association:
                    subnet_id = association['SubnetId']
                    rt_id = rt['RouteTableId']
                    # Get route table name from tags
                    rt_name = rt_id
                    if rt.get('Tags'):
                        for tag in rt['Tags']:
                            if tag['Key'] == 'Name':
                                rt_name = tag['Value']
                                break
                    subnet_to_route_table[subnet_id] = rt_name
        
        # Prepare subnet data with type information
        subnet_data = []
        for subnet in subnets:
            # Determine subnet type based on tags or naming
            subnet_type = "Private"  # Default fallback
            
            # First check explicit type tags
            if subnet.get('Tags'):
                for tag in subnet['Tags']:
                    if tag['Key'].lower() in ['type', 'subnet-type', 'subnet_type']:
                        # Abbreviate the type name if it's long
                        tag_value = tag['Value'].lower()
                        # Check in order of priority (inspection before firewall)
                        if 'inspection' in tag_value or 'inspect' in tag_value:
                            subnet_type = "Inspect"
                        elif 'firewall' in tag_value or 'fw' in tag_value:
                            subnet_type = "FW"
                        elif 'public' in tag_value:
                            subnet_type = "Public"
                        elif 'private' in tag_value:
                            subnet_type = "Private"
                        elif 'database' in tag_value or 'db' in tag_value:
                            subnet_type = "DB"
                        elif 'application' in tag_value or 'app' in tag_value:
                            subnet_type = "App"
                        elif 'management' in tag_value or 'mgmt' in tag_value:
                            subnet_type = "Mgmt"
                        else:
                            # Keep original if no abbreviation matches
                            subnet_type = tag['Value']
                        break
                else:
                    # If no explicit type tag, infer from name tag
                    for tag in subnet['Tags']:
                        if tag['Key'].lower() == 'name':
                            name = tag['Value'].lower()
                            # Check in order of priority (inspection before firewall)
                            if 'inspection' in name or 'inspect' in name:
                                subnet_type = "Inspect"
                            elif 'firewall' in name or 'fw-' in name:
                                subnet_type = "FW"
                            elif 'public' in name:
                                subnet_type = "Public"
                            elif 'private' in name:
                                subnet_type = "Private"
                            elif 'database' in name or 'db-' in name:
                                subnet_type = "DB"
                            elif 'application' in name or 'app-' in name:
                                subnet_type = "App"
                            elif 'management' in name or 'mgmt-' in name:
                                subnet_type = "Mgmt"
                            break
            
            # Get associated route table
            route_table = subnet_to_route_table.get(subnet['SubnetId'], "Main Route Table")
            
            # Get subnet name from tags
            subnet_name = "No name"
            if subnet.get('Tags'):
                for tag in subnet['Tags']:
                    if tag['Key'] == 'Name':
                        subnet_name = tag['Value']
                        break
            
            subnet_data.append({
                'subnet': subnet,
                'type': subnet_type,
                'route_table': route_table,
                'name': subnet_name
            })
        
        # Sort subnets based on the specified criteria
        if sort_by == "cidr":
            # Sort by CIDR block (convert to IP address for proper sorting)
            import ipaddress
            subnet_data.sort(key=lambda x: ipaddress.IPv4Network(x['subnet']['CidrBlock']))
        else:  # default to type
            # Sort by type with logical ordering, then by CIDR block
            type_order = {
                "Public": 0,
                "FW": 1,
                "Inspect": 2,
                "App": 3,
                "DB": 4,
                "Mgmt": 5,
                "Private": 6
            }
            subnet_data.sort(key=lambda x: (type_order.get(x['type'], 999), x['subnet']['CidrBlock']))
        
        table = Table(title=f"Subnets in VPC {self.vpc_id} (sorted by {sort_by})")
        table.add_column("Subnet ID", style="cyan", width=20)
        table.add_column("Subnet Name", style="bright_cyan", width=20)
        table.add_column("CIDR Block", style="green", width=18)
        table.add_column("Availability Zone", style="blue", width=20)
        table.add_column("Type", style="magenta", width=10)
        table.add_column("State", style="yellow", width=10)
        table.add_column("Route Table", style="bright_blue", width=35)
        
        for data in subnet_data:
            subnet = data['subnet']
            table.add_row(
                subnet['SubnetId'],
                data['name'],
                subnet['CidrBlock'],
                subnet['AvailabilityZone'],
                data['type'],
                subnet['State'],
                data['route_table']
            )
        
        console.print(table)
        console.print()
    
    def display_route_tables(self, route_tables: List[Dict[str, Any]], 
                           nat_gateways: Dict[str, Dict[str, Any]], 
                           internet_gateways: Dict[str, Dict[str, Any]]) -> None:
        """Display route tables and their routes in formatted tables."""
        if not route_tables:
            console.print("[yellow]No route tables found in this VPC.[/yellow]")
            return
        
        for rt in route_tables:
            rt_id = rt['RouteTableId']
            
            # Get route table name from tags
            rt_name = "Main Route Table" if rt['Associations'] and rt['Associations'][0].get('Main') else rt_id
            if rt.get('Tags'):
                for tag in rt['Tags']:
                    if tag['Key'] == 'Name':
                        rt_name = tag['Value']
                        break
            
            # Display associated subnets
            associated_subnets = []
            for association in rt['Associations']:
                if 'SubnetId' in association:
                    associated_subnets.append(association['SubnetId'])
            
            # Create route table info panel
            info_text = f"Route Table: {rt_id}\nName: {rt_name}"
            console.print(Panel(info_text, title="Route Table Information", border_style="blue"))
            
            # Display associated subnets in a separate table if any exist
            if associated_subnets:
                subnet_table = Table(title=f"Associated Subnets for {rt_name}")
                subnet_table.add_column("Subnet ID", style="cyan", width=20)
                
                for subnet_id in associated_subnets:
                    subnet_table.add_row(subnet_id)
                
                console.print(subnet_table)
                console.print()
            else:
                console.print(f"[yellow]No associated subnets for {rt_name}[/yellow]")
                console.print()
            
            # Create routes table
            routes_table = Table(title=f"Routes in {rt_name}")
            routes_table.add_column("Destination", style="cyan")
            routes_table.add_column("Target", style="green")
            routes_table.add_column("Target Type", style="blue")
            routes_table.add_column("State", style="yellow")
            routes_table.add_column("Origin", style="magenta")
            
            for route in rt['Routes']:
                destination = route.get('DestinationCidrBlock', route.get('DestinationPrefixListId', 'N/A'))
                target = route.get('GatewayId', route.get('InstanceId', route.get('NatGatewayId', 
                        route.get('VpcPeeringConnectionId', route.get('TransitGatewayId', 'N/A')))))
                
                # Determine target type and get friendly name
                target_type = "Unknown"
                if route.get('GatewayId'):
                    if route['GatewayId'].startswith('igw-'):
                        target_type = "Internet Gateway"
                        if route['GatewayId'] in internet_gateways:
                            target = f"{route['GatewayId']} (IGW)"
                    elif route['GatewayId'].startswith('vgw-'):
                        target_type = "Virtual Private Gateway"
                    else:
                        target_type = "Gateway"
                elif route.get('NatGatewayId'):
                    target_type = "NAT Gateway"
                    if route['NatGatewayId'] in nat_gateways:
                        ng = nat_gateways[route['NatGatewayId']]
                        target = f"{route['NatGatewayId']} ({ng['State']})"
                elif route.get('InstanceId'):
                    target_type = "Instance"
                elif route.get('VpcPeeringConnectionId'):
                    target_type = "VPC Peering"
                elif route.get('TransitGatewayId'):
                    target_type = "Transit Gateway"
                
                routes_table.add_row(
                    destination,
                    target,
                    target_type,
                    route.get('State', 'active'),
                    route.get('Origin', 'N/A')
                )
            
            console.print(routes_table)
            console.print()
    
    def display_summary(self, subnets: List[Dict[str, Any]], route_tables: List[Dict[str, Any]]) -> None:
        """Display a summary of the VPC configuration."""
        public_subnets = 0
        private_subnets = 0
        
        for subnet in subnets:
            if subnet.get('Tags'):
                for tag in subnet['Tags']:
                    if tag['Key'].lower() in ['type', 'subnet-type'] and 'public' in tag['Value'].lower():
                        public_subnets += 1
                        break
                    elif tag['Key'].lower() == 'name' and 'public' in tag['Value'].lower():
                        public_subnets += 1
                        break
                else:
                    private_subnets += 1
            else:
                private_subnets += 1
        
        summary_data = [
            f"VPC ID: {self.vpc_id}",
            f"CIDR Block: {self.vpc_cidr}",
            f"Total Subnets: {len(subnets)}",
            f"Public Subnets: {public_subnets}",
            f"Private Subnets: {private_subnets}",
            f"Route Tables: {len(route_tables)}"
        ]
        
        console.print(Panel("\n".join(summary_data), title="VPC Summary", border_style="green"))
    
    def generate_yaml_output(self, subnets: List[Dict[str, Any]], route_tables: List[Dict[str, Any]], 
                           nat_gateways: Dict[str, Dict[str, Any]], 
                           internet_gateways: Dict[str, Dict[str, Any]], sort_by: str = "type") -> Dict[str, Any]:
        """Generate structured YAML output for the VPC data."""
        
        # Create a mapping of subnet ID to route table ID
        subnet_to_route_table = {}
        for rt in route_tables:
            for association in rt['Associations']:
                if 'SubnetId' in association:
                    subnet_id = association['SubnetId']
                    rt_id = rt['RouteTableId']
                    # Get route table name from tags
                    rt_name = rt_id
                    if rt.get('Tags'):
                        for tag in rt['Tags']:
                            if tag['Key'] == 'Name':
                                rt_name = tag['Value']
                                break
                    subnet_to_route_table[subnet_id] = rt_name
        
        # Process subnets with type detection
        processed_subnets = []
        for subnet in subnets:
            # Determine subnet type
            subnet_type = "Private"
            if subnet.get('Tags'):
                for tag in subnet['Tags']:
                    if tag['Key'].lower() in ['type', 'subnet-type', 'subnet_type']:
                        tag_value = tag['Value'].lower()
                        if 'inspection' in tag_value or 'inspect' in tag_value:
                            subnet_type = "Inspect"
                        elif 'firewall' in tag_value or 'fw' in tag_value:
                            subnet_type = "FW"
                        elif 'public' in tag_value:
                            subnet_type = "Public"
                        elif 'private' in tag_value:
                            subnet_type = "Private"
                        elif 'database' in tag_value or 'db' in tag_value:
                            subnet_type = "DB"
                        elif 'application' in tag_value or 'app' in tag_value:
                            subnet_type = "App"
                        elif 'management' in tag_value or 'mgmt' in tag_value:
                            subnet_type = "Mgmt"
                        else:
                            subnet_type = tag['Value']
                        break
                else:
                    for tag in subnet['Tags']:
                        if tag['Key'].lower() == 'name':
                            name = tag['Value'].lower()
                            if 'inspection' in name or 'inspect' in name:
                                subnet_type = "Inspect"
                            elif 'firewall' in name or 'fw-' in name:
                                subnet_type = "FW"
                            elif 'public' in name:
                                subnet_type = "Public"
                            elif 'private' in name:
                                subnet_type = "Private"
                            elif 'database' in name or 'db-' in name:
                                subnet_type = "DB"
                            elif 'application' in name or 'app-' in name:
                                subnet_type = "App"
                            elif 'management' in name or 'mgmt-' in name:
                                subnet_type = "Mgmt"
                            break
            
            route_table = subnet_to_route_table.get(subnet['SubnetId'], "Main Route Table")
            
            # Get subnet name from tags
            subnet_name = "No name"
            if subnet.get('Tags'):
                for tag in subnet['Tags']:
                    if tag['Key'] == 'Name':
                        subnet_name = tag['Value']
                        break
            
            processed_subnets.append({
                'subnet_id': subnet['SubnetId'],
                'name': subnet_name,
                'cidr_block': subnet['CidrBlock'],
                'availability_zone': subnet['AvailabilityZone'],
                'type': subnet_type,
                'state': subnet['State'],
                'route_table': route_table,
                'tags': {tag['Key']: tag['Value'] for tag in subnet.get('Tags', [])}
            })
        
        # Sort subnets
        if sort_by == "cidr":
            import ipaddress
            processed_subnets.sort(key=lambda x: ipaddress.IPv4Network(x['cidr_block']))
        else:
            type_order = {
                "Public": 0, "FW": 1, "Inspect": 2, "App": 3, 
                "DB": 4, "Mgmt": 5, "Private": 6
            }
            processed_subnets.sort(key=lambda x: (type_order.get(x['type'], 999), x['cidr_block']))
        
        # Process route tables
        processed_route_tables = []
        for rt in route_tables:
            # Get route table name
            rt_name = rt['RouteTableId']
            if rt.get('Tags'):
                for tag in rt['Tags']:
                    if tag['Key'] == 'Name':
                        rt_name = tag['Value']
                        break
            
            # Get associated subnets
            associated_subnets = []
            for association in rt['Associations']:
                if 'SubnetId' in association:
                    associated_subnets.append(association['SubnetId'])
            
            # Process routes
            routes = []
            for route in rt['Routes']:
                destination = route.get('DestinationCidrBlock', route.get('DestinationPrefixListId', 'N/A'))
                target = route.get('GatewayId', route.get('InstanceId', route.get('NatGatewayId', 
                        route.get('VpcPeeringConnectionId', route.get('TransitGatewayId', 'N/A')))))
                
                # Determine target type
                target_type = "Unknown"
                if route.get('GatewayId'):
                    if route['GatewayId'].startswith('igw-'):
                        target_type = "Internet Gateway"
                    elif route['GatewayId'].startswith('vgw-'):
                        target_type = "Virtual Private Gateway"
                    else:
                        target_type = "Gateway"
                elif route.get('NatGatewayId'):
                    target_type = "NAT Gateway"
                elif route.get('InstanceId'):
                    target_type = "Instance"
                elif route.get('VpcPeeringConnectionId'):
                    target_type = "VPC Peering"
                elif route.get('TransitGatewayId'):
                    target_type = "Transit Gateway"
                
                routes.append({
                    'destination': destination,
                    'target': target,
                    'target_type': target_type,
                    'state': route.get('State', 'active'),
                    'origin': route.get('Origin', 'N/A')
                })
            
            processed_route_tables.append({
                'route_table_id': rt['RouteTableId'],
                'name': rt_name,
                'associated_subnets': associated_subnets,
                'routes': routes,
                'tags': {tag['Key']: tag['Value'] for tag in rt.get('Tags', [])}
            })
        
        # Calculate summary statistics
        type_counts = {}
        for subnet in processed_subnets:
            subnet_type = subnet['type']
            type_counts[subnet_type] = type_counts.get(subnet_type, 0) + 1
        
        public_count = type_counts.get('Public', 0)
        private_count = sum(count for type_name, count in type_counts.items() if type_name != 'Public')
        
        return {
            'vpc_info': {
                'vpc_id': self.vpc_id,
                'cidr_block': self.vpc_cidr,
                'state': self.vpc_state
            },
            'summary': {
                'total_subnets': len(processed_subnets),
                'public_subnets': public_count,
                'private_subnets': private_count,
                'route_tables': len(processed_route_tables),
                'subnet_types': type_counts
            },
            'subnets': processed_subnets,
            'route_tables': processed_route_tables,
            'nat_gateways': {ng_id: {'state': ng['State'], 'public_ip': ng.get('NatGatewayAddresses', [{}])[0].get('PublicIp', 'N/A')} 
                           for ng_id, ng in nat_gateways.items()},
            'internet_gateways': list(internet_gateways.keys())
        }


@click.command()
@click.argument('vpc_id')
@click.option('--sort', 'sort_by', 
              type=click.Choice(['type', 'cidr'], case_sensitive=False),
              default='type',
              help='Sort subnets by type (default) or CIDR block')
@click.option('--output-format', 'output_format',
              type=click.Choice(['table', 'yaml'], case_sensitive=False),
              default='table',
              help='Output format: table (default) or yaml')
def main(vpc_id: str, sort_by: str, output_format: str) -> None:
    """Display all subnets and routes for the specified VPC.
    
    VPC_ID: The ID of the VPC to analyze
    """
    try:
        # Initialize the VPC route viewer
        viewer = VPCRouteViewer(vpc_id)
        
        # Get all resources
        console.print("[blue]Fetching VPC resources...[/blue]")
        subnets = viewer.get_subnets()
        route_tables = viewer.get_route_tables()
        nat_gateways = viewer.get_nat_gateways()
        internet_gateways = viewer.get_internet_gateways()
        
        # Display results based on output format
        if output_format == 'yaml':
            yaml_data = viewer.generate_yaml_output(subnets, route_tables, nat_gateways, internet_gateways, sort_by)
            print(yaml.dump(yaml_data, default_flow_style=False, sort_keys=False))
        else:
            viewer.display_summary(subnets, route_tables)
            viewer.display_subnets(subnets, route_tables, sort_by)
            viewer.display_route_tables(route_tables, nat_gateways, internet_gateways)
            console.print("[green]VPC analysis complete![/green]")
        
    except KeyboardInterrupt:
        console.print("\n[yellow]Operation cancelled by user.[/yellow]")
        sys.exit(1)
    except Exception as e:
        console.print(f"[red]Unexpected error: {e}[/red]")
        sys.exit(1)


if __name__ == "__main__":
    main()
