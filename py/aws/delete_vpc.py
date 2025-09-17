#!/usr/bin/env python3
"""
AWS VPC Deletion Tool

This script deletes an AWS VPC and all its dependencies that would block
deletion from the AWS Console. It handles the proper order of resource
deletion to avoid dependency conflicts.

Usage:
    python delete_vpc.py <vpc-id>
    
Or with uv:
    uv run delete_vpc.py <vpc-id>
"""

import sys
import time
from typing import List, Dict, Any, Optional
import boto3
import click
from rich.console import Console
from rich.progress import Progress, SpinnerColumn, TextColumn
from rich.table import Table
from rich.panel import Panel
from botocore.exceptions import ClientError, BotoCoreError

console = Console()


class VPCDeleter:
    """Handles comprehensive VPC deletion with all dependencies."""
    
    def __init__(self, vpc_id: str, region: Optional[str] = None, profile: Optional[str] = None):
        """Initialize the VPC deleter.
        
        Args:
            vpc_id: The VPC ID to delete
            region: AWS region (defaults to session default)
            profile: AWS profile to use (defaults to default profile)
        """
        self.vpc_id = vpc_id
        self.region = region
        self.profile = profile
        
        # Initialize AWS session and clients
        session = boto3.Session(profile_name=profile, region_name=region)
        self.ec2 = session.client('ec2')
        self.elbv2 = session.client('elbv2')
        self.elb = session.client('elb')
        self.rds = session.client('rds')
        self.lambda_client = session.client('lambda')
        
        # Track deleted resources for reporting
        self.deleted_resources = {
            'instances': [],
            'load_balancers': [],
            'nat_gateways': [],
            'internet_gateways': [],
            'subnets': [],
            'route_tables': [],
            'security_groups': [],
            'network_acls': [],
            'vpc_endpoints': [],
            'peering_connections': [],
            'vpn_gateways': [],
            'customer_gateways': [],
            'vpn_connections': [],
            'db_subnet_groups': [],
            'lambda_functions': [],
            'elastic_ips': []
        }

    def verify_vpc_exists(self) -> bool:
        """Verify the VPC exists and get basic info."""
        try:
            response = self.ec2.describe_vpcs(VpcIds=[self.vpc_id])
            if not response['Vpcs']:
                console.print(f"[red]VPC {self.vpc_id} not found![/red]")
                return False
            
            vpc = response['Vpcs'][0]
            console.print(f"[green]Found VPC: {self.vpc_id}[/green]")
            console.print(f"  CIDR: {vpc.get('CidrBlock', 'N/A')}")
            console.print(f"  State: {vpc.get('State', 'N/A')}")
            console.print(f"  Default: {vpc.get('IsDefault', False)}")
            
            if vpc.get('IsDefault'):
                console.print("[yellow]⚠️  This is a default VPC. Deletion will affect default networking.[/yellow]")
            
            return True
            
        except ClientError as e:
            console.print(f"[red]Error verifying VPC: {e}[/red]")
            return False

    def delete_ec2_instances(self) -> None:
        """Delete all EC2 instances in the VPC."""
        try:
            response = self.ec2.describe_instances(
                Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
            )
            
            instance_ids = []
            for reservation in response['Reservations']:
                for instance in reservation['Instances']:
                    if instance['State']['Name'] not in ['terminated', 'terminating']:
                        instance_ids.append(instance['InstanceId'])
            
            if instance_ids:
                console.print(f"[yellow]Terminating {len(instance_ids)} EC2 instances...[/yellow]")
                self.ec2.terminate_instances(InstanceIds=instance_ids)
                
                # Wait for instances to terminate
                with Progress(SpinnerColumn(), TextColumn("[progress.description]{task.description}")) as progress:
                    task = progress.add_task("Waiting for instances to terminate...", total=None)
                    
                    waiter = self.ec2.get_waiter('instance_terminated')
                    waiter.wait(InstanceIds=instance_ids, WaiterConfig={'Delay': 15, 'MaxAttempts': 40})
                
                self.deleted_resources['instances'].extend(instance_ids)
                console.print(f"[green]✓ Terminated {len(instance_ids)} instances[/green]")
                
        except ClientError as e:
            console.print(f"[red]Error deleting instances: {e}[/red]")

    def delete_load_balancers(self) -> None:
        """Delete Application Load Balancers, Network Load Balancers, and Classic Load Balancers."""
        # Delete ALBs and NLBs
        try:
            response = self.elbv2.describe_load_balancers()
            vpc_lbs = [lb for lb in response['LoadBalancers'] if lb.get('VpcId') == self.vpc_id]
            
            for lb in vpc_lbs:
                console.print(f"[yellow]Deleting load balancer: {lb['LoadBalancerName']}[/yellow]")
                self.elbv2.delete_load_balancer(LoadBalancerArn=lb['LoadBalancerArn'])
                self.deleted_resources['load_balancers'].append(lb['LoadBalancerName'])
                
            if vpc_lbs:
                console.print(f"[green]✓ Deleted {len(vpc_lbs)} ALB/NLB load balancers[/green]")
                
        except ClientError as e:
            console.print(f"[red]Error deleting ALB/NLB load balancers: {e}[/red]")
        
        # Delete Classic Load Balancers
        try:
            response = self.elb.describe_load_balancers()
            vpc_clbs = [lb for lb in response['LoadBalancerDescriptions'] if lb.get('VPCId') == self.vpc_id]
            
            for lb in vpc_clbs:
                console.print(f"[yellow]Deleting classic load balancer: {lb['LoadBalancerName']}[/yellow]")
                self.elb.delete_load_balancer(LoadBalancerName=lb['LoadBalancerName'])
                self.deleted_resources['load_balancers'].append(lb['LoadBalancerName'])
                
            if vpc_clbs:
                console.print(f"[green]✓ Deleted {len(vpc_clbs)} classic load balancers[/green]")
                
        except ClientError as e:
            console.print(f"[red]Error deleting classic load balancers: {e}[/red]")

    def delete_rds_subnet_groups(self) -> None:
        """Delete RDS subnet groups in the VPC."""
        try:
            # Get all subnets in the VPC first
            subnets_response = self.ec2.describe_subnets(
                Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
            )
            vpc_subnet_ids = {subnet['SubnetId'] for subnet in subnets_response['Subnets']}
            
            # Get all DB subnet groups
            response = self.rds.describe_db_subnet_groups()
            
            for subnet_group in response['DBSubnetGroups']:
                # Check if any subnet in the group belongs to our VPC
                group_subnet_ids = {subnet['SubnetIdentifier'] for subnet in subnet_group['Subnets']}
                if vpc_subnet_ids.intersection(group_subnet_ids):
                    console.print(f"[yellow]Deleting DB subnet group: {subnet_group['DBSubnetGroupName']}[/yellow]")
                    self.rds.delete_db_subnet_group(DBSubnetGroupName=subnet_group['DBSubnetGroupName'])
                    self.deleted_resources['db_subnet_groups'].append(subnet_group['DBSubnetGroupName'])
                    
        except ClientError as e:
            if 'DBSubnetGroupNotFoundFault' not in str(e):
                console.print(f"[red]Error deleting DB subnet groups: {e}[/red]")

    def delete_lambda_functions(self) -> None:
        """Delete Lambda functions connected to VPC subnets."""
        try:
            # Get all subnets in the VPC
            subnets_response = self.ec2.describe_subnets(
                Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
            )
            vpc_subnet_ids = {subnet['SubnetId'] for subnet in subnets_response['Subnets']}
            
            # Get all Lambda functions
            response = self.lambda_client.list_functions()
            
            for function in response['Functions']:
                try:
                    # Get function configuration to check VPC config
                    config = self.lambda_client.get_function_configuration(
                        FunctionName=function['FunctionName']
                    )
                    
                    vpc_config = config.get('VpcConfig', {})
                    if vpc_config and vpc_config.get('SubnetIds'):
                        function_subnet_ids = set(vpc_config['SubnetIds'])
                        if vpc_subnet_ids.intersection(function_subnet_ids):
                            console.print(f"[yellow]Deleting Lambda function: {function['FunctionName']}[/yellow]")
                            self.lambda_client.delete_function(FunctionName=function['FunctionName'])
                            self.deleted_resources['lambda_functions'].append(function['FunctionName'])
                            
                except ClientError as e:
                    if 'ResourceNotFoundException' not in str(e):
                        console.print(f"[red]Error processing Lambda function {function['FunctionName']}: {e}[/red]")
                        
        except ClientError as e:
            console.print(f"[red]Error deleting Lambda functions: {e}[/red]")

    def delete_nat_gateways(self) -> None:
        """Delete NAT Gateways in the VPC."""
        try:
            response = self.ec2.describe_nat_gateways(
                Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
            )
            
            nat_gateways = [ng for ng in response['NatGateways'] if ng['State'] not in ['deleted', 'deleting']]
            
            for nat_gateway in nat_gateways:
                console.print(f"[yellow]Deleting NAT Gateway: {nat_gateway['NatGatewayId']}[/yellow]")
                self.ec2.delete_nat_gateway(NatGatewayId=nat_gateway['NatGatewayId'])
                self.deleted_resources['nat_gateways'].append(nat_gateway['NatGatewayId'])
            
            if nat_gateways:
                # Wait for NAT gateways to be deleted
                with Progress(SpinnerColumn(), TextColumn("[progress.description]{task.description}")) as progress:
                    task = progress.add_task("Waiting for NAT gateways to be deleted...", total=None)
                    
                    while True:
                        response = self.ec2.describe_nat_gateways(
                            NatGatewayIds=[ng['NatGatewayId'] for ng in nat_gateways]
                        )
                        if all(ng['State'] == 'deleted' for ng in response['NatGateways']):
                            break
                        time.sleep(10)
                
                console.print(f"[green]✓ Deleted {len(nat_gateways)} NAT gateways[/green]")
                
        except ClientError as e:
            console.print(f"[red]Error deleting NAT gateways: {e}[/red]")

    def delete_vpc_endpoints(self) -> None:
        """Delete VPC endpoints."""
        try:
            response = self.ec2.describe_vpc_endpoints(
                Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
            )
            
            for endpoint in response['VpcEndpoints']:
                if endpoint['State'] not in ['deleted', 'deleting']:
                    console.print(f"[yellow]Deleting VPC endpoint: {endpoint['VpcEndpointId']}[/yellow]")
                    self.ec2.delete_vpc_endpoint(VpcEndpointId=endpoint['VpcEndpointId'])
                    self.deleted_resources['vpc_endpoints'].append(endpoint['VpcEndpointId'])
                    
            if response['VpcEndpoints']:
                console.print(f"[green]✓ Deleted {len(response['VpcEndpoints'])} VPC endpoints[/green]")
                
        except ClientError as e:
            console.print(f"[red]Error deleting VPC endpoints: {e}[/red]")

    def delete_peering_connections(self) -> None:
        """Delete VPC peering connections."""
        try:
            response = self.ec2.describe_vpc_peering_connections(
                Filters=[
                    {'Name': 'requester-vpc-info.vpc-id', 'Values': [self.vpc_id]},
                    {'Name': 'accepter-vpc-info.vpc-id', 'Values': [self.vpc_id]}
                ]
            )
            
            for connection in response['VpcPeeringConnections']:
                if connection['Status']['Code'] not in ['deleted', 'deleting']:
                    console.print(f"[yellow]Deleting peering connection: {connection['VpcPeeringConnectionId']}[/yellow]")
                    self.ec2.delete_vpc_peering_connection(VpcPeeringConnectionId=connection['VpcPeeringConnectionId'])
                    self.deleted_resources['peering_connections'].append(connection['VpcPeeringConnectionId'])
                    
            if response['VpcPeeringConnections']:
                console.print(f"[green]✓ Deleted {len(response['VpcPeeringConnections'])} peering connections[/green]")
                
        except ClientError as e:
            console.print(f"[red]Error deleting peering connections: {e}[/red]")

    def delete_vpn_connections(self) -> None:
        """Delete VPN connections and gateways."""
        try:
            # Delete VPN connections
            response = self.ec2.describe_vpn_connections(
                Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
            )
            
            for vpn_conn in response['VpnConnections']:
                if vpn_conn['State'] not in ['deleted', 'deleting']:
                    console.print(f"[yellow]Deleting VPN connection: {vpn_conn['VpnConnectionId']}[/yellow]")
                    self.ec2.delete_vpn_connection(VpnConnectionId=vpn_conn['VpnConnectionId'])
                    self.deleted_resources['vpn_connections'].append(vpn_conn['VpnConnectionId'])
            
            # Delete VPN gateways
            response = self.ec2.describe_vpn_gateways(
                Filters=[{'Name': 'attachment.vpc-id', 'Values': [self.vpc_id]}]
            )
            
            for vpn_gw in response['VpnGateways']:
                if vpn_gw['State'] not in ['deleted', 'deleting']:
                    # Detach from VPC first
                    for attachment in vpn_gw.get('VpcAttachments', []):
                        if attachment['VpcId'] == self.vpc_id and attachment['State'] == 'attached':
                            self.ec2.detach_vpn_gateway(
                                VpnGatewayId=vpn_gw['VpnGatewayId'],
                                VpcId=self.vpc_id
                            )
                    
                    console.print(f"[yellow]Deleting VPN gateway: {vpn_gw['VpnGatewayId']}[/yellow]")
                    self.ec2.delete_vpn_gateway(VpnGatewayId=vpn_gw['VpnGatewayId'])
                    self.deleted_resources['vpn_gateways'].append(vpn_gw['VpnGatewayId'])
                    
        except ClientError as e:
            console.print(f"[red]Error deleting VPN connections/gateways: {e}[/red]")

    def delete_internet_gateways(self) -> None:
        """Delete and detach Internet Gateways."""
        try:
            response = self.ec2.describe_internet_gateways(
                Filters=[{'Name': 'attachment.vpc-id', 'Values': [self.vpc_id]}]
            )
            
            for igw in response['InternetGateways']:
                console.print(f"[yellow]Detaching Internet Gateway: {igw['InternetGatewayId']}[/yellow]")
                self.ec2.detach_internet_gateway(
                    InternetGatewayId=igw['InternetGatewayId'],
                    VpcId=self.vpc_id
                )
                
                console.print(f"[yellow]Deleting Internet Gateway: {igw['InternetGatewayId']}[/yellow]")
                self.ec2.delete_internet_gateway(InternetGatewayId=igw['InternetGatewayId'])
                self.deleted_resources['internet_gateways'].append(igw['InternetGatewayId'])
                
            if response['InternetGateways']:
                console.print(f"[green]✓ Deleted {len(response['InternetGateways'])} internet gateways[/green]")
                
        except ClientError as e:
            console.print(f"[red]Error deleting internet gateways: {e}[/red]")

    def delete_subnets(self) -> None:
        """Delete all subnets in the VPC."""
        try:
            response = self.ec2.describe_subnets(
                Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
            )
            
            for subnet in response['Subnets']:
                console.print(f"[yellow]Deleting subnet: {subnet['SubnetId']}[/yellow]")
                self.ec2.delete_subnet(SubnetId=subnet['SubnetId'])
                self.deleted_resources['subnets'].append(subnet['SubnetId'])
                
            if response['Subnets']:
                console.print(f"[green]✓ Deleted {len(response['Subnets'])} subnets[/green]")
                
        except ClientError as e:
            console.print(f"[red]Error deleting subnets: {e}[/red]")

    def delete_route_tables(self) -> None:
        """Delete custom route tables (not the main route table)."""
        try:
            response = self.ec2.describe_route_tables(
                Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
            )
            
            for route_table in response['RouteTables']:
                # Skip the main route table
                is_main = any(assoc.get('Main', False) for assoc in route_table.get('Associations', []))
                if not is_main:
                    console.print(f"[yellow]Deleting route table: {route_table['RouteTableId']}[/yellow]")
                    self.ec2.delete_route_table(RouteTableId=route_table['RouteTableId'])
                    self.deleted_resources['route_tables'].append(route_table['RouteTableId'])
                    
            custom_route_tables = [rt for rt in response['RouteTables'] 
                                 if not any(assoc.get('Main', False) for assoc in rt.get('Associations', []))]
            if custom_route_tables:
                console.print(f"[green]✓ Deleted {len(custom_route_tables)} custom route tables[/green]")
                
        except ClientError as e:
            console.print(f"[red]Error deleting route tables: {e}[/red]")

    def delete_security_groups(self) -> None:
        """Delete custom security groups (not the default security group)."""
        try:
            response = self.ec2.describe_security_groups(
                Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
            )
            
            # First pass: remove all rules to break circular dependencies
            custom_sgs = [sg for sg in response['SecurityGroups'] if sg['GroupName'] != 'default']
            for sg in custom_sgs:
                if sg['IpPermissions']:
                    console.print(f"[yellow]Removing ingress rules from: {sg['GroupId']}[/yellow]")
                    self.ec2.revoke_security_group_ingress(
                        GroupId=sg['GroupId'],
                        IpPermissions=sg['IpPermissions']
                    )
                
                if sg['IpPermissionsEgress']:
                    console.print(f"[yellow]Removing egress rules from: {sg['GroupId']}[/yellow]")
                    self.ec2.revoke_security_group_egress(
                        GroupId=sg['GroupId'],
                        IpPermissions=sg['IpPermissionsEgress']
                    )
            
            # Second pass: delete the security groups
            for sg in custom_sgs:
                console.print(f"[yellow]Deleting security group: {sg['GroupId']} ({sg['GroupName']})[/yellow]")
                self.ec2.delete_security_group(GroupId=sg['GroupId'])
                self.deleted_resources['security_groups'].append(sg['GroupId'])
                
            if custom_sgs:
                console.print(f"[green]✓ Deleted {len(custom_sgs)} custom security groups[/green]")
                
        except ClientError as e:
            console.print(f"[red]Error deleting security groups: {e}[/red]")

    def delete_network_acls(self) -> None:
        """Delete custom Network ACLs (not the default ACL)."""
        try:
            response = self.ec2.describe_network_acls(
                Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
            )
            
            custom_acls = [acl for acl in response['NetworkAcls'] if not acl['IsDefault']]
            for acl in custom_acls:
                console.print(f"[yellow]Deleting Network ACL: {acl['NetworkAclId']}[/yellow]")
                self.ec2.delete_network_acl(NetworkAclId=acl['NetworkAclId'])
                self.deleted_resources['network_acls'].append(acl['NetworkAclId'])
                
            if custom_acls:
                console.print(f"[green]✓ Deleted {len(custom_acls)} custom network ACLs[/green]")
                
        except ClientError as e:
            console.print(f"[red]Error deleting network ACLs: {e}[/red]")

    def release_elastic_ips(self) -> None:
        """Release Elastic IPs associated with the VPC."""
        try:
            response = self.ec2.describe_addresses()
            
            vpc_eips = []
            for address in response['Addresses']:
                # Check if EIP is associated with instances in our VPC
                if 'InstanceId' in address:
                    instance_response = self.ec2.describe_instances(InstanceIds=[address['InstanceId']])
                    for reservation in instance_response['Reservations']:
                        for instance in reservation['Instances']:
                            if instance.get('VpcId') == self.vpc_id:
                                vpc_eips.append(address)
                                break
                
                # Check if EIP is associated with NAT gateways in our VPC
                elif 'NetworkInterfaceId' in address:
                    try:
                        ni_response = self.ec2.describe_network_interfaces(
                            NetworkInterfaceIds=[address['NetworkInterfaceId']]
                        )
                        for ni in ni_response['NetworkInterfaces']:
                            if ni.get('VpcId') == self.vpc_id:
                                vpc_eips.append(address)
                                break
                    except ClientError:
                        pass  # Network interface might be deleted already
            
            for eip in vpc_eips:
                console.print(f"[yellow]Releasing Elastic IP: {eip.get('PublicIp', eip.get('AllocationId'))}[/yellow]")
                if 'AllocationId' in eip:
                    self.ec2.release_address(AllocationId=eip['AllocationId'])
                else:
                    self.ec2.release_address(PublicIp=eip['PublicIp'])
                self.deleted_resources['elastic_ips'].append(eip.get('AllocationId', eip.get('PublicIp')))
                
            if vpc_eips:
                console.print(f"[green]✓ Released {len(vpc_eips)} Elastic IPs[/green]")
                
        except ClientError as e:
            console.print(f"[red]Error releasing Elastic IPs: {e}[/red]")

    def delete_vpc(self) -> bool:
        """Delete the VPC itself."""
        try:
            console.print(f"[yellow]Deleting VPC: {self.vpc_id}[/yellow]")
            self.ec2.delete_vpc(VpcId=self.vpc_id)
            console.print(f"[green]✓ Successfully deleted VPC: {self.vpc_id}[/green]")
            return True
            
        except ClientError as e:
            console.print(f"[red]Error deleting VPC: {e}[/red]")
            return False

    def print_summary(self) -> None:
        """Print a summary of all deleted resources."""
        table = Table(title="Deletion Summary", show_header=True, header_style="bold magenta")
        table.add_column("Resource Type", style="cyan", no_wrap=True)
        table.add_column("Count", style="green", justify="right")
        table.add_column("IDs", style="yellow")
        
        for resource_type, resources in self.deleted_resources.items():
            if resources:
                count = len(resources)
                ids_str = ", ".join(resources[:3])  # Show first 3 IDs
                if len(resources) > 3:
                    ids_str += f" ... and {len(resources) - 3} more"
                table.add_row(resource_type.replace('_', ' ').title(), str(count), ids_str)
        
        console.print(table)

    def run_deletion(self, dry_run: bool = False) -> bool:
        """Run the complete VPC deletion process.
        
        Args:
            dry_run: If True, only show what would be deleted without actually deleting
            
        Returns:
            True if successful, False otherwise
        """
        if dry_run:
            console.print("[yellow]DRY RUN MODE - No resources will be deleted[/yellow]")
            return True
            
        console.print(Panel.fit(
            f"[bold red]DELETING VPC: {self.vpc_id}[/bold red]\n"
            "[yellow]This will delete ALL resources in the VPC and cannot be undone![/yellow]",
            border_style="red"
        ))
        
        if not self.verify_vpc_exists():
            return False
        
        # Execute deletion steps in the correct order
        steps = [
            ("EC2 Instances", self.delete_ec2_instances),
            ("Load Balancers", self.delete_load_balancers),
            ("Lambda Functions", self.delete_lambda_functions),
            ("RDS Subnet Groups", self.delete_rds_subnet_groups),
            ("NAT Gateways", self.delete_nat_gateways),
            ("VPC Endpoints", self.delete_vpc_endpoints),
            ("Peering Connections", self.delete_peering_connections),
            ("VPN Connections & Gateways", self.delete_vpn_connections),
            ("Elastic IPs", self.release_elastic_ips),
            ("Internet Gateways", self.delete_internet_gateways),
            ("Subnets", self.delete_subnets),
            ("Route Tables", self.delete_route_tables),
            ("Security Groups", self.delete_security_groups),
            ("Network ACLs", self.delete_network_acls),
        ]
        
        console.print("\n[bold blue]Starting VPC deletion process...[/bold blue]")
        
        for step_name, step_func in steps:
            console.print(f"\n[bold cyan]Step: {step_name}[/bold cyan]")
            try:
                step_func()
            except Exception as e:
                console.print(f"[red]Error in {step_name}: {e}[/red]")
                continue
        
        # Finally, delete the VPC
        console.print(f"\n[bold cyan]Final Step: VPC Deletion[/bold cyan]")
        success = self.delete_vpc()
        
        # Print summary
        console.print("\n" + "="*60)
        self.print_summary()
        
        return success


@click.command()
@click.argument('vpc_id')
@click.option('--region', '-r', help='AWS region')
@click.option('--profile', '-p', help='AWS profile to use')
@click.option('--dry-run', is_flag=True, help='Show what would be deleted without actually deleting')
@click.option('--force', is_flag=True, help='Skip confirmation prompt')
def main(vpc_id: str, region: Optional[str], profile: Optional[str], dry_run: bool, force: bool):
    """Delete an AWS VPC and all its dependencies.
    
    VPC_ID: The ID of the VPC to delete (e.g., vpc-12345678)
    """
    console.print(f"[bold blue]AWS VPC Deletion Tool[/bold blue]")
    console.print(f"VPC ID: {vpc_id}")
    console.print(f"Region: {region or 'default'}")
    console.print(f"Profile: {profile or 'default'}")
    
    if not dry_run and not force:
        confirmation = click.confirm(
            f"\nAre you sure you want to delete VPC {vpc_id} and ALL its resources? This cannot be undone!"
        )
        if not confirmation:
            console.print("[yellow]Operation cancelled.[/yellow]")
            return
    
    try:
        deleter = VPCDeleter(vpc_id, region, profile)
        success = deleter.run_deletion(dry_run)
        
        if success:
            console.print(f"\n[bold green]✅ VPC deletion completed successfully![/bold green]")
            sys.exit(0)
        else:
            console.print(f"\n[bold red]❌ VPC deletion failed![/bold red]")
            sys.exit(1)
            
    except KeyboardInterrupt:
        console.print(f"\n[yellow]Operation interrupted by user.[/yellow]")
        sys.exit(1)
    except Exception as e:
        console.print(f"\n[red]Unexpected error: {e}[/red]")
        sys.exit(1)


if __name__ == "__main__":
    main()
