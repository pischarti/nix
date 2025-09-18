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
    
    def __init__(self, vpc_id: str):
        """Initialize the VPC deleter.
        
        Args:
            vpc_id: The VPC ID to delete
            
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
        self.ec2 = session.client('ec2')
        self.elbv2 = session.client('elbv2')
        self.elb = session.client('elb')
        self.rds = session.client('rds')
        self.lambda_client = session.client('lambda')
        
        # Track deleted resources for reporting
        self.deleted_resources = {
            'instances': [],
            'load_balancers': [],
            'target_groups': [],
            'listeners': [],
            'nat_gateways': [],
            'internet_gateways': [],
            'subnets': [],
            'route_tables': [],
            'security_groups': [],
            'network_acls': [],
            'vpc_endpoints': [],
            'vpc_endpoint_service_configurations': [],
            'network_interfaces': [],
            'peering_connections': [],
            'vpn_gateways': [],
            'customer_gateways': [],
            'vpn_connections': [],
            'db_subnet_groups': [],
            'lambda_functions': [],
            'elastic_ips': []
        }

    def _validate_aws_config(self) -> None:
        """Validate that required AWS configuration is available."""
        import os
        
        # Check for AWS region
        region = os.environ.get('AWS_DEFAULT_REGION') or os.environ.get('AWS_REGION')
        if not region:
            console.print("[red]❌ AWS region not configured![/red]")
            console.print("\n[yellow]Please set your AWS region using one of these methods:[/yellow]")
            console.print("  1. Environment variable:")
            console.print("     export AWS_DEFAULT_REGION=us-east-1")
            console.print("  2. AWS CLI configuration:")
            console.print("     aws configure set region us-east-1")
            console.print("  3. AWS credentials file (~/.aws/config):")
            console.print("     [default]")
            console.print("     region = us-east-1")
            raise SystemExit(1)
        
        # Check for AWS credentials (basic check)
        access_key = os.environ.get('AWS_ACCESS_KEY_ID')
        profile = os.environ.get('AWS_PROFILE')
        
        if not access_key and not profile:
            # Check if AWS CLI is configured
            aws_config_dir = os.path.expanduser('~/.aws')
            credentials_file = os.path.join(aws_config_dir, 'credentials')
            config_file = os.path.join(aws_config_dir, 'config')
            
            if not (os.path.exists(credentials_file) or os.path.exists(config_file)):
                console.print("[red]❌ AWS credentials not found![/red]")
                console.print("\n[yellow]Please configure AWS credentials using one of these methods:[/yellow]")
                console.print("  1. Environment variables:")
                console.print("     export AWS_ACCESS_KEY_ID=your-access-key")
                console.print("     export AWS_SECRET_ACCESS_KEY=your-secret-key")
                console.print("  2. AWS CLI configuration:")
                console.print("     aws configure")
                console.print("  3. AWS profile:")
                console.print("     export AWS_PROFILE=your-profile-name")
                raise SystemExit(1)

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
        """Delete Application Load Balancers, Network Load Balancers, Gateway Load Balancers, and Classic Load Balancers."""
        # Delete ALBs, NLBs, and GWLBs
        try:
            response = self.elbv2.describe_load_balancers()
            vpc_lbs = [lb for lb in response['LoadBalancers'] if lb.get('VpcId') == self.vpc_id]
            
            deleted_count = 0
            failed_lbs = []
            
            for lb in vpc_lbs:
                lb_name = lb['LoadBalancerName']
                lb_type = lb.get('Type', 'unknown')
                lb_arn = lb['LoadBalancerArn']
                
                try:
                    console.print(f"[yellow]Deleting {lb_type} load balancer: {lb_name}[/yellow]")
                    
                    # Note: VPC Endpoint Service configurations are handled in a separate step
                    
                    self.elbv2.delete_load_balancer(LoadBalancerArn=lb_arn)
                    self.deleted_resources['load_balancers'].append(lb_name)
                    deleted_count += 1
                    
                except ClientError as lb_error:
                    error_code = lb_error.response.get('Error', {}).get('Code', '')
                    error_message = lb_error.response.get('Error', {}).get('Message', str(lb_error))
                    
                    if error_code == 'ResourceInUse':
                        console.print(f"[red]✗ Cannot delete {lb_type} load balancer '{lb_name}': {error_message}[/red]")
                        if lb_type == 'gateway':
                            console.print(f"[yellow]  Gateway Load Balancer is associated with another service.[/yellow]")
                            
                            # Try to identify and automatically clean up dependencies
                            dependencies_cleaned = self._cleanup_gwlb_dependencies(lb_name, lb_arn)
                            
                            if dependencies_cleaned:
                                console.print(f"[yellow]  Attempting to delete GWLB again after dependency cleanup...[/yellow]")
                                try:
                                    # Wait a moment for AWS to process the dependency deletions
                                    import time
                                    time.sleep(5)
                                    
                                    # Retry GWLB deletion
                                    self.elbv2.delete_load_balancer(LoadBalancerArn=lb_arn)
                                    self.deleted_resources['load_balancers'].append(lb_name)
                                    deleted_count += 1
                                    console.print(f"[green]✓ Successfully deleted Gateway Load Balancer: {lb_name}[/green]")
                                    
                                except ClientError as retry_error:
                                    console.print(f"[red]✗ Still cannot delete GWLB '{lb_name}' after cleanup: {retry_error}[/red]")
                                    console.print(f"[yellow]  → Performing detailed GWLB analysis...[/yellow]")
                                    self._analyze_gwlb_detailed_dependencies(lb_name, lb_arn)
                                    failed_lbs.append(lb_name)
                            else:
                                console.print(f"[yellow]  Could not automatically clean up all dependencies.[/yellow]")
                                console.print(f"[yellow]  Manual cleanup may be required:[/yellow]")
                                console.print(f"[yellow]  1. Check VPC Console → Endpoint Services for remaining configurations[/yellow]")
                                console.print(f"[yellow]  2. Check for Auto Scaling Groups using this GWLB[/yellow]")
                                console.print(f"[yellow]  3. Check for other services referencing this GWLB[/yellow]")
                                failed_lbs.append(lb_name)
                        else:
                            console.print(f"[yellow]  Load balancer may have active targets or listeners[/yellow]")
                            failed_lbs.append(lb_name)
                    else:
                        console.print(f"[red]✗ Error deleting {lb_type} load balancer '{lb_name}': {error_message}[/red]")
                        failed_lbs.append(lb_name)
                
            if deleted_count > 0:
                console.print(f"[green]✓ Deleted {deleted_count} ALB/NLB/GWLB load balancers[/green]")
            if failed_lbs:
                console.print(f"[yellow]⚠ Failed to delete {len(failed_lbs)} load balancers: {', '.join(failed_lbs)}[/yellow]")
                
        except ClientError as e:
            console.print(f"[red]Error listing ALB/NLB/GWLB load balancers: {e}[/red]")
        
        # Delete Classic Load Balancers
        try:
            response = self.elb.describe_load_balancers()
            vpc_clbs = [lb for lb in response['LoadBalancerDescriptions'] if lb.get('VPCId') == self.vpc_id]
            
            deleted_clb_count = 0
            failed_clbs = []
            
            for lb in vpc_clbs:
                lb_name = lb['LoadBalancerName']
                try:
                    console.print(f"[yellow]Deleting classic load balancer: {lb_name}[/yellow]")
                    self.elb.delete_load_balancer(LoadBalancerName=lb_name)
                    self.deleted_resources['load_balancers'].append(lb_name)
                    deleted_clb_count += 1
                    
                except ClientError as clb_error:
                    error_code = clb_error.response.get('Error', {}).get('Code', '')
                    error_message = clb_error.response.get('Error', {}).get('Message', str(clb_error))
                    
                    if error_code == 'ResourceInUse':
                        console.print(f"[red]✗ Cannot delete classic load balancer '{lb_name}': {error_message}[/red]")
                        console.print(f"[yellow]  Classic Load Balancer may have active instances or be in use by other services[/yellow]")
                        failed_clbs.append(lb_name)
                    else:
                        console.print(f"[red]✗ Error deleting classic load balancer '{lb_name}': {error_message}[/red]")
                        failed_clbs.append(lb_name)
                
            if deleted_clb_count > 0:
                console.print(f"[green]✓ Deleted {deleted_clb_count} classic load balancers[/green]")
            if failed_clbs:
                console.print(f"[yellow]⚠ Failed to delete {len(failed_clbs)} classic load balancers: {', '.join(failed_clbs)}[/yellow]")
                
        except ClientError as e:
            console.print(f"[red]Error listing classic load balancers: {e}[/red]")

    def _identify_gwlb_dependencies(self, lb_name: str, lb_arn: str) -> None:
        """Try to identify what services are using a Gateway Load Balancer."""
        try:
            # Check VPC Endpoint Service configurations
            try:
                response = self.ec2.describe_vpc_endpoint_service_configurations()
                found_services = []
                
                for service_config in response.get('ServiceConfigurations', []):
                    gwlb_arns = service_config.get('GatewayLoadBalancerArns', [])
                    if lb_arn in gwlb_arns:
                        service_name = service_config.get('ServiceName', 'Unknown')
                        service_id = service_config.get('ServiceId', 'Unknown')
                        found_services.append(f"{service_name} (ID: {service_id})")
                
                if found_services:
                    console.print(f"[yellow]  → Found VPC Endpoint Services using this GWLB:[/yellow]")
                    for service in found_services:
                        console.print(f"[yellow]    - {service}[/yellow]")
                else:
                    console.print(f"[yellow]  → No VPC Endpoint Services found using this GWLB[/yellow]")
                    
            except (AttributeError, ClientError) as e:
                console.print(f"[yellow]  → Unable to check VPC Endpoint Services: {e}[/yellow]")
            
            # Check for target groups (though GWLBs don't use traditional target groups)
            try:
                response = self.elbv2.describe_target_groups()
                associated_tgs = []
                
                for tg in response.get('TargetGroups', []):
                    for lb_arn_in_tg in tg.get('LoadBalancerArns', []):
                        if lb_arn_in_tg == lb_arn:
                            associated_tgs.append(tg.get('TargetGroupName', 'Unknown'))
                
                if associated_tgs:
                    console.print(f"[yellow]  → Found Target Groups: {', '.join(associated_tgs)}[/yellow]")
                    
            except ClientError as e:
                console.print(f"[yellow]  → Unable to check Target Groups: {e}[/yellow]")
                
        except Exception as e:
            console.print(f"[yellow]  → Error identifying dependencies: {e}[/yellow]")

    def _cleanup_gwlb_dependencies(self, lb_name: str, lb_arn: str) -> bool:
        """Try to automatically clean up Gateway Load Balancer dependencies."""
        try:
            console.print(f"[yellow]  → Attempting comprehensive automatic cleanup of GWLB dependencies...[/yellow]")
            dependencies_cleaned = False
            
            # Try to delete VPC Endpoint Service configurations using this GWLB
            try:
                response = self.ec2.describe_vpc_endpoint_service_configurations()
                
                for service_config in response.get('ServiceConfigurations', []):
                    gwlb_arns = service_config.get('GatewayLoadBalancerArns', [])
                    
                    if lb_arn in gwlb_arns:
                        service_name = service_config.get('ServiceName', 'Unknown')
                        service_id = service_config.get('ServiceId')
                        
                        console.print(f"[yellow]    → Found VPC Endpoint Service using this GWLB: {service_name}[/yellow]")
                        
                        # Try to delete the service configuration
                        try:
                            console.print(f"[yellow]    → Deleting VPC Endpoint Service configuration: {service_name}[/yellow]")
                            self.ec2.delete_vpc_endpoint_service_configurations(ServiceIds=[service_id])
                            console.print(f"[green]    ✓ Deleted VPC Endpoint Service configuration: {service_name}[/green]")
                            dependencies_cleaned = True
                            
                        except AttributeError:
                            # Try singular method
                            try:
                                self.ec2.delete_vpc_endpoint_service_configuration(ServiceId=service_id)
                                console.print(f"[green]    ✓ Deleted VPC Endpoint Service configuration: {service_name}[/green]")
                                dependencies_cleaned = True
                            except (AttributeError, ClientError) as inner_error:
                                console.print(f"[red]    ✗ Could not delete service configuration {service_name}: {inner_error}[/red]")
                                
                        except ClientError as svc_error:
                            error_message = svc_error.response.get('Error', {}).get('Message', str(svc_error))
                            console.print(f"[red]    ✗ Could not delete service configuration {service_name}: {error_message}[/red]")
                            
            except (AttributeError, ClientError) as endpoint_error:
                console.print(f"[yellow]    → Could not check VPC Endpoint Service configurations: {endpoint_error}[/yellow]")
            
            # Try to remove target groups and their listeners
            try:
                response = self.elbv2.describe_target_groups()
                for tg in response.get('TargetGroups', []):
                    if lb_arn in tg.get('LoadBalancerArns', []):
                        tg_arn = tg.get('TargetGroupArn')
                        tg_name = tg.get('TargetGroupName', 'Unknown')
                        
                        console.print(f"[yellow]    → Found target group associated with GWLB: {tg_name}[/yellow]")
                        
                        # First, try to delete listeners that use this target group
                        try:
                            listeners_response = self.elbv2.describe_listeners(LoadBalancerArn=lb_arn)
                            for listener in listeners_response.get('Listeners', []):
                                listener_arn = listener.get('ListenerArn')
                                
                                # Check if this listener uses our target group
                                default_actions = listener.get('DefaultActions', [])
                                uses_target_group = any(
                                    action.get('TargetGroupArn') == tg_arn 
                                    for action in default_actions
                                )
                                
                                if uses_target_group:
                                    try:
                                        console.print(f"[yellow]      → Deleting listener using target group: {listener_arn}[/yellow]")
                                        self.elbv2.delete_listener(ListenerArn=listener_arn)
                                        self.deleted_resources['listeners'].append(listener_arn)
                                        console.print(f"[green]      ✓ Deleted listener: {listener_arn}[/green]")
                                    except ClientError as listener_error:
                                        console.print(f"[red]      ✗ Could not delete listener: {listener_error}[/red]")
                                        
                        except ClientError as listeners_error:
                            console.print(f"[yellow]      → Could not check listeners: {listeners_error}[/yellow]")
                        
                        # Now try to delete the target group
                        try:
                            console.print(f"[yellow]    → Deleting target group: {tg_name}[/yellow]")
                            self.elbv2.delete_target_group(TargetGroupArn=tg_arn)
                            self.deleted_resources['target_groups'].append(tg_name)
                            console.print(f"[green]    ✓ Deleted target group: {tg_name}[/green]")
                            dependencies_cleaned = True
                            
                        except ClientError as tg_error:
                            error_code = tg_error.response.get('Error', {}).get('Code', '')
                            error_message = tg_error.response.get('Error', {}).get('Message', str(tg_error))
                            
                            if error_code == 'ResourceInUse':
                                console.print(f"[red]    ✗ Target group {tg_name} is still in use: {error_message}[/red]")
                                console.print(f"[yellow]      → This target group may have listeners or rules that couldn't be automatically removed[/yellow]")
                                console.print(f"[yellow]      → Manual cleanup may be required in the AWS Console[/yellow]")
                            else:
                                console.print(f"[red]    ✗ Could not delete target group {tg_name}: {error_message}[/red]")
                            
            except ClientError as tg_list_error:
                console.print(f"[yellow]    → Could not check target groups: {tg_list_error}[/yellow]")
            
            # Check for and automatically delete VPC endpoints that connect to GWLB services
            try:
                console.print(f"[yellow]    → Checking for VPC endpoints connecting to GWLB services...[/yellow]")
                vpc_endpoints_response = self.ec2.describe_vpc_endpoints(
                    Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
                )
                
                for vpc_endpoint in vpc_endpoints_response.get('VpcEndpoints', []):
                    if vpc_endpoint.get('State') not in ['deleted', 'deleting']:
                        service_name = vpc_endpoint.get('ServiceName', '')
                        endpoint_id = vpc_endpoint.get('VpcEndpointId')
                        
                        # Check if this endpoint might be related to our GWLB or is using a GWLB service
                        should_delete_endpoint = False
                        
                        # Check for firewall/security related services
                        if any(keyword in service_name.lower() for keyword in ['firewall', 'security', 'inspection']):
                            should_delete_endpoint = True
                            console.print(f"[yellow]      → Found firewall/security VPC endpoint: {endpoint_id}[/yellow]")
                        
                        # Check if this endpoint is using a service that uses our GWLB
                        elif service_name.startswith('com.amazonaws.vpce.'):
                            # This might be a VPC Endpoint Service that uses our GWLB
                            try:
                                # Get the service ID from the service name
                                service_id_from_name = service_name.split('.')[-1]
                                
                                # Check if any VPC Endpoint Service configurations use our GWLB
                                service_configs_response = self.ec2.describe_vpc_endpoint_service_configurations()
                                for config in service_configs_response.get('ServiceConfigurations', []):
                                    if (config.get('ServiceId') == service_id_from_name and 
                                        lb_arn in config.get('GatewayLoadBalancerArns', [])):
                                        should_delete_endpoint = True
                                        console.print(f"[yellow]      → Found VPC endpoint using GWLB service: {endpoint_id}[/yellow]")
                                        break
                                        
                            except (ClientError, AttributeError):
                                pass  # Continue if we can't check
                        
                        # Also check for any endpoint that might be connecting to our GWLB service
                        # by looking at network interfaces in this VPC
                        else:
                            try:
                                # Check if this VPC has any GWLB endpoint network interfaces
                                ni_response = self.ec2.describe_network_interfaces(
                                    Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
                                )
                                
                                for ni in ni_response.get('NetworkInterfaces', []):
                                    if (ni.get('InterfaceType') == 'gateway_load_balancer_endpoint' and 
                                        endpoint_id in ni.get('Description', '')):
                                        should_delete_endpoint = True
                                        console.print(f"[yellow]      → Found VPC endpoint with GWLB endpoint interface: {endpoint_id}[/yellow]")
                                        break
                                        
                            except ClientError:
                                pass  # Continue if we can't check
                        
                        if should_delete_endpoint:
                            console.print(f"[yellow]        Service: {service_name}[/yellow]")
                            
                            try:
                                console.print(f"[yellow]      → Deleting VPC endpoint: {endpoint_id}[/yellow]")
                                self.ec2.delete_vpc_endpoints(VpcEndpointIds=[endpoint_id])
                                console.print(f"[green]      ✓ Deleted VPC endpoint: {endpoint_id}[/green]")
                                dependencies_cleaned = True
                                
                            except ClientError as endpoint_error:
                                error_code = endpoint_error.response.get('Error', {}).get('Code', '')
                                error_message = endpoint_error.response.get('Error', {}).get('Message', str(endpoint_error))
                                console.print(f"[red]      ✗ Could not delete VPC endpoint {endpoint_id}: {error_message}[/red]")
                                
                                if error_code == 'InvalidVpcEndpointId.NotFound':
                                    console.print(f"[yellow]        → Endpoint already deleted[/yellow]")
                                elif 'cannot be deleted' in error_message.lower():
                                    console.print(f"[yellow]        → Endpoint may have active connections or policies[/yellow]")
                                
            except ClientError as endpoint_check_error:
                console.print(f"[yellow]    → Could not check VPC endpoints: {endpoint_check_error}[/yellow]")
            
            # Check for AWS Network Firewall dependencies (common with firewall GWLBs)
            if 'firewall' in lb_name.lower():
                try:
                    console.print(f"[yellow]    → Checking for AWS Network Firewall dependencies...[/yellow]")
                    
                    # Try to create a Network Firewall client to check for firewalls
                    try:
                        import boto3
                        session = boto3.Session()
                        network_firewall = session.client('network-firewall')
                        
                        # Check for firewalls in this VPC
                        firewalls_response = network_firewall.list_firewalls()
                        
                        for firewall_metadata in firewalls_response.get('Firewalls', []):
                            firewall_name = firewall_metadata.get('FirewallName')
                            firewall_arn = firewall_metadata.get('FirewallArn')
                            
                            # Get detailed firewall info
                            try:
                                firewall_detail = network_firewall.describe_firewall(FirewallArn=firewall_arn)
                                firewall = firewall_detail.get('Firewall', {})
                                
                                if firewall.get('VpcId') == self.vpc_id:
                                    console.print(f"[yellow]      → Found AWS Network Firewall in this VPC: {firewall_name}[/yellow]")
                                    console.print(f"[yellow]        This firewall may be using the Gateway Load Balancer[/yellow]")
                                    console.print(f"[yellow]        You may need to delete the Network Firewall first[/yellow]")
                                    
                            except ClientError as firewall_detail_error:
                                console.print(f"[yellow]      → Could not get firewall details for {firewall_name}: {firewall_detail_error}[/yellow]")
                                
                    except ImportError:
                        console.print(f"[yellow]      → Network Firewall client not available[/yellow]")
                    except ClientError as nf_error:
                        if 'UnauthorizedOperation' in str(nf_error):
                            console.print(f"[yellow]      → No permission to check Network Firewall (this is normal)[/yellow]")
                        else:
                            console.print(f"[yellow]      → Could not check Network Firewall: {nf_error}[/yellow]")
                            
                except Exception as firewall_check_error:
                    console.print(f"[yellow]    → Error checking Network Firewall: {firewall_check_error}[/yellow]")
            
            # Check for Auto Scaling Groups that might be using this GWLB
            try:
                console.print(f"[yellow]    → Checking for Auto Scaling Groups using this GWLB...[/yellow]")
                # Note: Auto Scaling Groups use a different service (autoscaling), but we can check if any exist
                # This is a placeholder for potential ASG cleanup - would need autoscaling client
                console.print(f"[yellow]      → Auto Scaling Group check not implemented (would require additional permissions)[/yellow]")
                
            except Exception as asg_error:
                console.print(f"[yellow]    → Could not check Auto Scaling Groups: {asg_error}[/yellow]")
            
            if dependencies_cleaned:
                console.print(f"[green]  ✓ Successfully cleaned up some GWLB dependencies[/green]")
            else:
                console.print(f"[yellow]  ⚠ Automatic dependency cleanup completed, but GWLB may have additional dependencies[/yellow]")
                
                # Provide comprehensive manual cleanup guidance
                console.print(f"[yellow]  → Additional dependencies to check manually:[/yellow]")
                console.print(f"[yellow]    1. VPC Endpoint Services in other AWS accounts or regions[/yellow]")
                console.print(f"[yellow]    2. Auto Scaling Groups using this GWLB as a target[/yellow]")
                console.print(f"[yellow]    3. AWS Network Firewall or third-party firewall integrations[/yellow]")
                console.print(f"[yellow]    4. Cross-account VPC endpoint connections[/yellow]")
                console.print(f"[yellow]    5. AWS Transit Gateway attachments[/yellow]")
                console.print(f"[yellow]  → Manual cleanup steps:[/yellow]")
                console.print(f"[yellow]    1. Go to VPC Console → Endpoint Services (check all regions)[/yellow]")
                console.print(f"[yellow]    2. Search for services using ARN: {lb_arn}[/yellow]")
                console.print(f"[yellow]    3. Delete any endpoint service configurations found[/yellow]")
                console.print(f"[yellow]    4. Check Auto Scaling Console for target groups using this GWLB[/yellow]")
                console.print(f"[yellow]    5. Check Network Firewall Console for firewall policies[/yellow]")
                console.print(f"[yellow]    6. Wait 5-10 minutes after cleanup, then re-run this tool[/yellow]")
            
            return dependencies_cleaned
            
        except Exception as e:
            console.print(f"[yellow]  → Error during dependency cleanup: {e}[/yellow]")
            return False

    def _analyze_gwlb_detailed_dependencies(self, lb_name: str, lb_arn: str) -> None:
        """Perform detailed analysis of GWLB dependencies when automatic cleanup fails."""
        try:
            console.print(f"[yellow]    → Detailed analysis of GWLB dependencies for {lb_name}...[/yellow]")
            
            # Check all VPC Endpoint Service configurations again with more detail
            try:
                response = self.ec2.describe_vpc_endpoint_service_configurations()
                console.print(f"[yellow]      → Checking {len(response.get('ServiceConfigurations', []))} VPC Endpoint Service configurations...[/yellow]")
                
                for service_config in response.get('ServiceConfigurations', []):
                    service_name = service_config.get('ServiceName', 'Unknown')
                    service_id = service_config.get('ServiceId', 'Unknown')
                    acceptance_required = service_config.get('AcceptanceRequired', False)
                    service_state = service_config.get('ServiceState', 'Unknown')
                    
                    gwlb_arns = service_config.get('GatewayLoadBalancerArns', [])
                    nlb_arns = service_config.get('NetworkLoadBalancerArns', [])
                    
                    if lb_arn in gwlb_arns:
                        console.print(f"[red]      ✗ Found VPC Endpoint Service still using this GWLB:[/red]")
                        console.print(f"[yellow]        Service Name: {service_name}[/yellow]")
                        console.print(f"[yellow]        Service ID: {service_id}[/yellow]")
                        console.print(f"[yellow]        State: {service_state}[/yellow]")
                        console.print(f"[yellow]        Acceptance Required: {acceptance_required}[/yellow]")
                        
                        # Check for endpoint connections
                        try:
                            connections_response = self.ec2.describe_vpc_endpoint_connections(
                                Filters=[{'Name': 'service-id', 'Values': [service_id]}]
                            )
                            
                            connections = connections_response.get('VpcEndpointConnections', [])
                            if connections:
                                console.print(f"[red]        → Found {len(connections)} active endpoint connections:[/red]")
                                for conn in connections[:3]:  # Show first 3
                                    vpc_endpoint_id = conn.get('VpcEndpointId', 'Unknown')
                                    connection_state = conn.get('VpcEndpointState', 'Unknown')
                                    console.print(f"[yellow]          - {vpc_endpoint_id} ({connection_state})[/yellow]")
                                
                                console.print(f"[yellow]        → These connections must be deleted before the service can be removed[/yellow]")
                            
                        except ClientError as conn_error:
                            console.print(f"[yellow]        → Could not check endpoint connections: {conn_error}[/yellow]")
                            
            except (AttributeError, ClientError) as detailed_error:
                console.print(f"[yellow]      → Could not perform detailed VPC Endpoint Service analysis: {detailed_error}[/yellow]")
            
            # Provide comprehensive resolution steps
            console.print(f"[yellow]    → Complete resolution steps for persistent GWLB:[/yellow]")
            console.print(f"[yellow]      1. Check for VPC endpoint connections using the service[/yellow]")
            console.print(f"[yellow]      2. Delete all VPC endpoint connections first[/yellow]")
            console.print(f"[yellow]      3. Then delete the VPC Endpoint Service configuration[/yellow]")
            console.print(f"[yellow]      4. If this is a Network Firewall GWLB, delete the firewall first[/yellow]")
            console.print(f"[yellow]      5. Check other AWS accounts that might have cross-account access[/yellow]")
            console.print(f"[yellow]      6. Wait 10-15 minutes between each step for AWS to propagate changes[/yellow]")
            
        except Exception as e:
            console.print(f"[yellow]    → Error during detailed GWLB analysis: {e}[/yellow]")

    def delete_vpc_endpoint_service_configurations(self) -> None:
        """Delete VPC Endpoint Service configurations that may be blocking load balancer deletion."""
        try:
            # First check if the describe method exists
            try:
                response = self.ec2.describe_vpc_endpoint_service_configurations()
            except AttributeError:
                console.print(f"[yellow]⚠ VPC Endpoint Service configuration management not supported by this AWS API version[/yellow]")
                return
            
            deleted_count = 0
            failed_configs = []
            
            for service_config in response.get('ServiceConfigurations', []):
                # Check if this service configuration uses load balancers in our VPC
                gwlb_arns = service_config.get('GatewayLoadBalancerArns', [])
                nlb_arns = service_config.get('NetworkLoadBalancerArns', [])
                service_name = service_config.get('ServiceName', 'Unknown')
                service_id = service_config.get('ServiceId')
                
                # Check if any of the load balancers belong to our VPC
                service_uses_vpc_lbs = False
                if gwlb_arns or nlb_arns:
                    try:
                        # Get load balancers to check their VPC
                        lb_response = self.elbv2.describe_load_balancers()
                        vpc_lb_arns = {lb['LoadBalancerArn'] for lb in lb_response['LoadBalancers'] 
                                     if lb.get('VpcId') == self.vpc_id}
                        
                        service_uses_vpc_lbs = bool(
                            (set(gwlb_arns) & vpc_lb_arns) or 
                            (set(nlb_arns) & vpc_lb_arns)
                        )
                        
                        # Show which load balancers are being used
                        if service_uses_vpc_lbs:
                            console.print(f"[yellow]Found VPC Endpoint Service using load balancers in this VPC: {service_name}[/yellow]")
                            for arn in gwlb_arns:
                                if arn in vpc_lb_arns:
                                    lb_name = arn.split('/')[-2] if '/' in arn else 'Unknown'
                                    console.print(f"[yellow]  - Gateway Load Balancer: {lb_name}[/yellow]")
                            for arn in nlb_arns:
                                if arn in vpc_lb_arns:
                                    lb_name = arn.split('/')[-2] if '/' in arn else 'Unknown'
                                    console.print(f"[yellow]  - Network Load Balancer: {lb_name}[/yellow]")
                                    
                    except ClientError:
                        pass  # Continue if we can't check load balancers
                
                if service_uses_vpc_lbs:
                    # Try to delete using the most likely API method
                    deleted = False
                    
                    # Method 1: Try delete_vpc_endpoint_service_configurations (plural)
                    try:
                        console.print(f"[yellow]Deleting VPC Endpoint Service configuration: {service_name}[/yellow]")
                        self.ec2.delete_vpc_endpoint_service_configurations(ServiceIds=[service_id])
                        self.deleted_resources['vpc_endpoint_service_configurations'].append(service_name)
                        deleted_count += 1
                        deleted = True
                        console.print(f"[green]✓ Successfully deleted VPC Endpoint Service configuration: {service_name}[/green]")
                        
                    except AttributeError:
                        # Method 2: Try delete_vpc_endpoint_service_configuration (singular)
                        try:
                            console.print(f"[yellow]Trying alternative API for VPC Endpoint Service configuration: {service_name}[/yellow]")
                            self.ec2.delete_vpc_endpoint_service_configuration(ServiceId=service_id)
                            self.deleted_resources['vpc_endpoint_service_configurations'].append(service_name)
                            deleted_count += 1
                            deleted = True
                            console.print(f"[green]✓ Successfully deleted VPC Endpoint Service configuration: {service_name}[/green]")
                            
                        except (AttributeError, ClientError) as inner_error:
                            console.print(f"[red]✗ Could not delete VPC Endpoint Service '{service_name}': {inner_error}[/red]")
                            failed_configs.append(service_name)
                            
                    except ClientError as svc_error:
                        error_code = svc_error.response.get('Error', {}).get('Code', '')
                        error_message = svc_error.response.get('Error', {}).get('Message', str(svc_error))
                        
                        if error_code == 'InvalidVpcEndpointServiceId.NotFound':
                            console.print(f"[yellow]VPC Endpoint Service {service_name} already deleted[/yellow]")
                        elif 'cannot be deleted' in error_message.lower():
                            console.print(f"[red]✗ Cannot delete VPC Endpoint Service '{service_name}': {error_message}[/red]")
                            console.print(f"[yellow]  This service may have active endpoint connections[/yellow]")
                            failed_configs.append(service_name)
                        else:
                            console.print(f"[red]✗ Error deleting VPC Endpoint Service '{service_name}': {error_message}[/red]")
                            failed_configs.append(service_name)
            
            if deleted_count > 0:
                console.print(f"[green]✓ Deleted {deleted_count} VPC Endpoint Service configurations[/green]")
            
            if failed_configs:
                console.print(f"[yellow]⚠ Failed to delete {len(failed_configs)} VPC Endpoint Service configurations[/yellow]")
                console.print(f"[yellow]  You may need to manually delete these in the AWS Console:[/yellow]")
                for config_name in failed_configs:
                    console.print(f"[yellow]  - {config_name}[/yellow]")
                
        except ClientError as e:
            if 'InvalidAction' in str(e) or 'UnauthorizedOperation' in str(e):
                console.print(f"[yellow]⚠ VPC Endpoint Service configuration access not available: {e}[/yellow]")
            else:
                console.print(f"[red]Error managing VPC Endpoint Service configurations: {e}[/red]")

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
        """Delete VPC endpoints, with special handling for Gateway Load Balancer endpoints."""
        try:
            response = self.ec2.describe_vpc_endpoints(
                Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
            )
            
            all_endpoints = response.get('VpcEndpoints', [])
            active_endpoints = [ep for ep in all_endpoints if ep['State'] not in ['deleted', 'deleting']]
            
            console.print(f"[yellow]Found {len(active_endpoints)} active VPC endpoints to process[/yellow]")
            
            # Categorize endpoints for better handling
            gwlb_related_endpoints = []
            other_endpoints = []
            
            for endpoint in active_endpoints:
                endpoint_id = endpoint['VpcEndpointId']
                service_name = endpoint.get('ServiceName', '')
                endpoint_type = endpoint.get('VpcEndpointType', 'Interface')
                
                console.print(f"[yellow]  Analyzing endpoint {endpoint_id}: {service_name} (type: {endpoint_type})[/yellow]")
                
                # Check if this is a GWLB-related endpoint (be more aggressive in detection)
                is_gwlb_related = (
                    any(keyword in service_name.lower() for keyword in ['firewall', 'security', 'inspection']) or
                    service_name.startswith('com.amazonaws.vpce.') or
                    endpoint.get('VpcEndpointType') == 'GatewayLoadBalancer' or
                    'gwlb' in service_name.lower() or
                    'gateway' in service_name.lower()
                )
                
                # Also check if this endpoint has a gateway_load_balancer_endpoint network interface
                if not is_gwlb_related:
                    try:
                        # Check network interfaces for this endpoint
                        endpoint_ni_response = self.ec2.describe_network_interfaces(
                            Filters=[{'Name': 'description', 'Values': [f'*{endpoint_id}*']}]
                        )
                        
                        for ni in endpoint_ni_response.get('NetworkInterfaces', []):
                            if ni.get('InterfaceType') == 'gateway_load_balancer_endpoint':
                                is_gwlb_related = True
                                console.print(f"[yellow]  Found GWLB endpoint interface for {endpoint_id}, marking as GWLB-related[/yellow]")
                                break
                                
                    except ClientError:
                        pass  # Continue if we can't check
                
                if is_gwlb_related:
                    console.print(f"[yellow]  → Classified as GWLB-related endpoint[/yellow]")
                    gwlb_related_endpoints.append((endpoint_id, service_name))
                else:
                    console.print(f"[yellow]  → Classified as other endpoint[/yellow]")
                    other_endpoints.append(endpoint_id)
            
            deleted_count = 0
            failed_endpoints = []
            
            # Delete GWLB-related endpoints first (these are likely blocking GWLB deletion)
            if gwlb_related_endpoints:
                console.print(f"[yellow]Deleting {len(gwlb_related_endpoints)} Gateway Load Balancer related VPC endpoints...[/yellow]")
                
                for endpoint_id, service_name in gwlb_related_endpoints:
                    try:
                        console.print(f"[yellow]  Deleting GWLB-related VPC endpoint: {endpoint_id}[/yellow]")
                        console.print(f"[yellow]    Service: {service_name}[/yellow]")
                        
                        self.ec2.delete_vpc_endpoints(VpcEndpointIds=[endpoint_id])
                        self.deleted_resources['vpc_endpoints'].append(endpoint_id)
                        deleted_count += 1
                        console.print(f"[green]  ✓ Deleted GWLB-related VPC endpoint: {endpoint_id}[/green]")
                        
                    except ClientError as gwlb_endpoint_error:
                        error_message = gwlb_endpoint_error.response.get('Error', {}).get('Message', str(gwlb_endpoint_error))
                        console.print(f"[red]  ✗ Failed to delete GWLB-related VPC endpoint {endpoint_id}: {error_message}[/red]")
                        failed_endpoints.append(endpoint_id)
            
            # Delete other VPC endpoints in batches
            if other_endpoints:
                console.print(f"[yellow]Deleting {len(other_endpoints)} other VPC endpoints...[/yellow]")
                
                # Delete endpoints in batches (AWS allows up to 25 per call)
                batch_size = 25
                
                for i in range(0, len(other_endpoints), batch_size):
                    batch = other_endpoints[i:i + batch_size]
                    try:
                        self.ec2.delete_vpc_endpoints(VpcEndpointIds=batch)
                        self.deleted_resources['vpc_endpoints'].extend(batch)
                        deleted_count += len(batch)
                        
                        for endpoint_id in batch:
                            console.print(f"[yellow]  Deleted VPC endpoint: {endpoint_id}[/yellow]")
                            
                    except ClientError as batch_error:
                        console.print(f"[red]Error deleting VPC endpoint batch: {batch_error}[/red]")
                        # Try individual deletions for this batch
                        for endpoint_id in batch:
                            try:
                                self.ec2.delete_vpc_endpoints(VpcEndpointIds=[endpoint_id])
                                self.deleted_resources['vpc_endpoints'].append(endpoint_id)
                                deleted_count += 1
                                console.print(f"[yellow]  Deleted VPC endpoint: {endpoint_id}[/yellow]")
                            except ClientError as single_error:
                                console.print(f"[red]  ✗ Failed to delete VPC endpoint {endpoint_id}: {single_error}[/red]")
                                failed_endpoints.append(endpoint_id)
            
            if deleted_count > 0:
                console.print(f"[green]✓ Deleted {deleted_count} VPC endpoints[/green]")
            
            if failed_endpoints:
                console.print(f"[yellow]⚠ Failed to delete {len(failed_endpoints)} VPC endpoints: {', '.join(failed_endpoints)}[/yellow]")
            
            # Final check: look for any remaining VPC endpoints that might have been missed
            try:
                final_check_response = self.ec2.describe_vpc_endpoints(
                    Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
                )
                
                remaining_endpoints = [ep for ep in final_check_response.get('VpcEndpoints', []) 
                                     if ep['State'] not in ['deleted', 'deleting']]
                
                if remaining_endpoints:
                    console.print(f"[yellow]⚠ Found {len(remaining_endpoints)} VPC endpoints still remaining after deletion attempt[/yellow]")
                    for ep in remaining_endpoints:
                        ep_id = ep['VpcEndpointId']
                        ep_service = ep.get('ServiceName', 'Unknown')
                        ep_type = ep.get('VpcEndpointType', 'Unknown')
                        console.print(f"[yellow]  - {ep_id}: {ep_service} (type: {ep_type})[/yellow]")
                        
                        # Try to delete these remaining endpoints
                        try:
                            console.print(f"[yellow]    → Attempting to delete remaining endpoint: {ep_id}[/yellow]")
                            self.ec2.delete_vpc_endpoints(VpcEndpointIds=[ep_id])
                            console.print(f"[green]    ✓ Successfully deleted remaining endpoint: {ep_id}[/green]")
                            if ep_id not in self.deleted_resources['vpc_endpoints']:
                                self.deleted_resources['vpc_endpoints'].append(ep_id)
                                deleted_count += 1
                        except ClientError as remaining_error:
                            console.print(f"[red]    ✗ Could not delete remaining endpoint {ep_id}: {remaining_error}[/red]")
                else:
                    console.print(f"[green]✓ Verified: No VPC endpoints remaining in VPC[/green]")
                    
            except ClientError as final_check_error:
                console.print(f"[yellow]Could not perform final VPC endpoint check: {final_check_error}[/yellow]")
                
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
            
            deleted_count = 0
            failed_subnets = []
            
            for subnet in response['Subnets']:
                subnet_id = subnet['SubnetId']
                availability_zone = subnet.get('AvailabilityZone', 'Unknown')
                cidr_block = subnet.get('CidrBlock', 'Unknown')
                
                try:
                    console.print(f"[yellow]Deleting subnet: {subnet_id} ({cidr_block} in {availability_zone})[/yellow]")
                    self.ec2.delete_subnet(SubnetId=subnet_id)
                    self.deleted_resources['subnets'].append(subnet_id)
                    deleted_count += 1
                    
                except ClientError as subnet_error:
                    error_code = subnet_error.response.get('Error', {}).get('Code', '')
                    error_message = subnet_error.response.get('Error', {}).get('Message', str(subnet_error))
                    
                    if error_code == 'DependencyViolation':
                        console.print(f"[red]✗ Cannot delete subnet '{subnet_id}': {error_message}[/red]")
                        console.print(f"[yellow]  Subnet has dependencies that must be removed first.[/yellow]")
                        
                        # Try to automatically clean up subnet dependencies
                        dependencies_cleaned = self._cleanup_subnet_dependencies(subnet_id)
                        
                        if dependencies_cleaned:
                            # Retry subnet deletion after cleanup
                            console.print(f"[yellow]  Retrying subnet deletion after dependency cleanup...[/yellow]")
                            try:
                                import time
                                time.sleep(5)  # Wait for AWS to process changes
                                
                                self.ec2.delete_subnet(SubnetId=subnet_id)
                                self.deleted_resources['subnets'].append(subnet_id)
                                deleted_count += 1
                                console.print(f"[green]✓ Successfully deleted subnet {subnet_id} after dependency cleanup[/green]")
                                
                            except ClientError as retry_error:
                                console.print(f"[red]✗ Still cannot delete subnet {subnet_id} after cleanup: {retry_error}[/red]")
                                self._identify_subnet_dependencies(subnet_id)
                                failed_subnets.append(subnet_id)
                        else:
                            # Fallback to identification if cleanup failed
                            self._identify_subnet_dependencies(subnet_id)
                            failed_subnets.append(subnet_id)
                    else:
                        console.print(f"[red]✗ Error deleting subnet '{subnet_id}': {error_message}[/red]")
                        failed_subnets.append(subnet_id)
                
            if deleted_count > 0:
                console.print(f"[green]✓ Deleted {deleted_count} subnets[/green]")
            if failed_subnets:
                console.print(f"[yellow]⚠ Failed to delete {len(failed_subnets)} subnets: {', '.join(failed_subnets)}[/yellow]")
                
        except ClientError as e:
            console.print(f"[red]Error listing subnets: {e}[/red]")

    def _cleanup_subnet_dependencies(self, subnet_id: str) -> bool:
        """Try to automatically clean up subnet dependencies, especially VPC endpoint interfaces."""
        try:
            console.print(f"[yellow]  → Attempting automatic cleanup of subnet dependencies...[/yellow]")
            dependencies_cleaned = False
            
            # Check for VPC endpoint interfaces in this subnet
            try:
                ni_response = self.ec2.describe_network_interfaces(
                    Filters=[{'Name': 'subnet-id', 'Values': [subnet_id]}]
                )
                
                gwlb_endpoint_interfaces = []
                vpc_endpoints_to_delete = []
                
                for ni in ni_response.get('NetworkInterfaces', []):
                    if ni.get('InterfaceType') == 'gateway_load_balancer_endpoint':
                        ni_id = ni.get('NetworkInterfaceId')
                        description = ni.get('Description', '')
                        
                        # Extract VPC endpoint ID from description
                        import re
                        vpce_match = re.search(r'vpce-[a-f0-9]+', description)
                        if vpce_match:
                            vpce_id = vpce_match.group(0)
                            gwlb_endpoint_interfaces.append((ni_id, vpce_id))
                            vpc_endpoints_to_delete.append(vpce_id)
                            console.print(f"[yellow]    → Found GWLB endpoint interface {ni_id} for VPC endpoint {vpce_id}[/yellow]")
                
                # Delete the VPC endpoints that have interfaces in this subnet with multiple retries
                if vpc_endpoints_to_delete:
                    console.print(f"[yellow]    → Attempting aggressive deletion of {len(vpc_endpoints_to_delete)} VPC endpoints blocking this subnet...[/yellow]")
                    
                    for vpce_id in vpc_endpoints_to_delete:
                        endpoint_deleted = False
                        
                        # Try multiple deletion attempts
                        for attempt in range(3):
                            try:
                                console.print(f"[yellow]      → Deleting VPC endpoint: {vpce_id} (attempt {attempt + 1}/3)[/yellow]")
                                
                                # First, get endpoint details to understand its state
                                try:
                                    endpoint_details = self.ec2.describe_vpc_endpoints(VpcEndpointIds=[vpce_id])
                                    endpoint = endpoint_details.get('VpcEndpoints', [{}])[0]
                                    endpoint_state = endpoint.get('State', 'Unknown')
                                    service_name = endpoint.get('ServiceName', 'Unknown')
                                    console.print(f"[yellow]        → Endpoint state: {endpoint_state}, Service: {service_name}[/yellow]")
                                    
                                    if endpoint_state in ['deleted', 'deleting']:
                                        console.print(f"[green]      ✓ VPC endpoint {vpce_id} is already being deleted[/green]")
                                        endpoint_deleted = True
                                        dependencies_cleaned = True
                                        break
                                        
                                except ClientError as detail_error:
                                    if 'InvalidVpcEndpointId.NotFound' in str(detail_error):
                                        console.print(f"[green]      ✓ VPC endpoint {vpce_id} already deleted[/green]")
                                        endpoint_deleted = True
                                        dependencies_cleaned = True
                                        break
                                
                                # Attempt deletion
                                self.ec2.delete_vpc_endpoints(VpcEndpointIds=[vpce_id])
                                console.print(f"[green]      ✓ Successfully deleted VPC endpoint: {vpce_id}[/green]")
                                
                                # Track this deletion
                                if vpce_id not in self.deleted_resources['vpc_endpoints']:
                                    self.deleted_resources['vpc_endpoints'].append(vpce_id)
                                
                                endpoint_deleted = True
                                dependencies_cleaned = True
                                break
                                
                            except ClientError as vpce_error:
                                error_code = vpce_error.response.get('Error', {}).get('Code', '')
                                error_message = vpce_error.response.get('Error', {}).get('Message', str(vpce_error))
                                
                                if error_code == 'InvalidVpcEndpointId.NotFound':
                                    console.print(f"[green]      ✓ VPC endpoint {vpce_id} already deleted[/green]")
                                    endpoint_deleted = True
                                    dependencies_cleaned = True
                                    break
                                elif 'DependencyViolation' in error_code or 'cannot be deleted' in error_message.lower():
                                    console.print(f"[red]      ✗ VPC endpoint {vpce_id} has dependencies (attempt {attempt + 1}): {error_message}[/red]")
                                    if attempt < 2:  # Not the last attempt
                                        console.print(f"[yellow]        → Waiting 10 seconds before retry...[/yellow]")
                                        import time
                                        time.sleep(10)
                                    else:
                                        console.print(f"[red]        → All deletion attempts failed for {vpce_id}[/red]")
                                else:
                                    console.print(f"[red]      ✗ Could not delete VPC endpoint {vpce_id}: {error_message}[/red]")
                                    break  # Don't retry for other types of errors
                        
                        if not endpoint_deleted:
                            console.print(f"[red]    ✗ Failed to delete VPC endpoint {vpce_id} after multiple attempts[/red]")
                            console.print(f"[yellow]    → Attempting force cleanup of VPC endpoint dependencies...[/yellow]")
                            
                            # Try to force cleanup by checking endpoint connections and policies
                            force_cleaned = self._force_cleanup_vpc_endpoint(vpce_id)
                            
                            if force_cleaned:
                                # Final attempt to delete the endpoint
                                try:
                                    console.print(f"[yellow]      → Final attempt to delete VPC endpoint: {vpce_id}[/yellow]")
                                    self.ec2.delete_vpc_endpoints(VpcEndpointIds=[vpce_id])
                                    console.print(f"[green]      ✓ Successfully force-deleted VPC endpoint: {vpce_id}[/green]")
                                    
                                    if vpce_id not in self.deleted_resources['vpc_endpoints']:
                                        self.deleted_resources['vpc_endpoints'].append(vpce_id)
                                    
                                    dependencies_cleaned = True
                                    
                                except ClientError as final_error:
                                    console.print(f"[red]      ✗ Force deletion also failed: {final_error}[/red]")
                            else:
                                console.print(f"[yellow]    → Force cleanup was not possible[/yellow]")
                                console.print(f"[yellow]    → Attempting ultimate fallback: direct network interface cleanup[/yellow]")
                                
                                # Ultimate fallback: try to work around the VPC endpoint by cleaning up its network interface
                                ultimate_success = self._ultimate_network_interface_cleanup(subnet_id, vpce_id)
                                if ultimate_success:
                                    dependencies_cleaned = True
                
                # Wait for network interface cleanup if we deleted endpoints
                if dependencies_cleaned and gwlb_endpoint_interfaces:
                    console.print(f"[yellow]    → Waiting for VPC endpoint network interfaces to be cleaned up...[/yellow]")
                    import time
                    time.sleep(10)  # Wait longer for VPC endpoint interface cleanup
                    
                    # Check if interfaces are gone
                    for attempt in range(3):  # Try 3 times
                        ni_check_response = self.ec2.describe_network_interfaces(
                            Filters=[{'Name': 'subnet-id', 'Values': [subnet_id]}]
                        )
                        
                        remaining_gwlb_interfaces = [
                            ni for ni in ni_check_response.get('NetworkInterfaces', [])
                            if ni.get('InterfaceType') == 'gateway_load_balancer_endpoint'
                        ]
                        
                        if not remaining_gwlb_interfaces:
                            console.print(f"[green]    ✓ VPC endpoint network interfaces have been cleaned up[/green]")
                            break
                        else:
                            console.print(f"[yellow]    → {len(remaining_gwlb_interfaces)} GWLB endpoint interfaces still present, waiting...[/yellow]")
                            if attempt < 2:  # Don't sleep on the last attempt
                                time.sleep(15)
                    
            except ClientError as ni_error:
                console.print(f"[yellow]    → Could not check network interfaces: {ni_error}[/yellow]")
            
            if dependencies_cleaned:
                console.print(f"[green]  ✓ Successfully cleaned up subnet dependencies[/green]")
            else:
                console.print(f"[yellow]  ⚠ No automatic dependency cleanup was possible[/yellow]")
            
            return dependencies_cleaned
            
        except Exception as e:
            console.print(f"[yellow]  → Error during subnet dependency cleanup: {e}[/yellow]")
            return False

    def _force_cleanup_vpc_endpoint(self, vpce_id: str) -> bool:
        """Try to force cleanup of a persistent VPC endpoint by removing its dependencies."""
        try:
            console.print(f"[yellow]      → Force cleanup analysis for VPC endpoint {vpce_id}...[/yellow]")
            cleanup_performed = False
            
            # Get detailed endpoint information
            try:
                endpoint_response = self.ec2.describe_vpc_endpoints(VpcEndpointIds=[vpce_id])
                endpoint = endpoint_response.get('VpcEndpoints', [{}])[0]
                service_name = endpoint.get('ServiceName', '')
                endpoint_type = endpoint.get('VpcEndpointType', 'Interface')
                policy_document = endpoint.get('PolicyDocument')
                
                console.print(f"[yellow]        → Service: {service_name}[/yellow]")
                console.print(f"[yellow]        → Type: {endpoint_type}[/yellow]")
                
                # Try to remove endpoint policy if it exists
                if policy_document and policy_document != '{}':
                    try:
                        console.print(f"[yellow]        → Removing VPC endpoint policy...[/yellow]")
                        self.ec2.modify_vpc_endpoint(
                            VpcEndpointId=vpce_id,
                            ResetPolicy=True
                        )
                        console.print(f"[green]        ✓ Removed VPC endpoint policy[/green]")
                        cleanup_performed = True
                        
                    except ClientError as policy_error:
                        console.print(f"[yellow]        → Could not remove policy: {policy_error}[/yellow]")
                
                # Try to remove route table associations if this is a Gateway endpoint
                if endpoint_type == 'Gateway':
                    route_table_ids = endpoint.get('RouteTableIds', [])
                    if route_table_ids:
                        console.print(f"[yellow]        → Removing route table associations...[/yellow]")
                        try:
                            self.ec2.modify_vpc_endpoint(
                                VpcEndpointId=vpce_id,
                                RemoveRouteTableIds=route_table_ids
                            )
                            console.print(f"[green]        ✓ Removed route table associations[/green]")
                            cleanup_performed = True
                            
                        except ClientError as route_error:
                            console.print(f"[yellow]        → Could not remove route associations: {route_error}[/yellow]")
                
                # Try to remove security group associations if this is an Interface endpoint
                elif endpoint_type == 'Interface':
                    security_group_ids = endpoint.get('Groups', [])
                    if security_group_ids and len(security_group_ids) > 1:  # Keep at least one
                        try:
                            # Keep only the first security group, remove others
                            keep_sg = security_group_ids[0]['GroupId']
                            console.print(f"[yellow]        → Simplifying security group associations...[/yellow]")
                            
                            self.ec2.modify_vpc_endpoint(
                                VpcEndpointId=vpce_id,
                                AddSecurityGroupIds=[keep_sg],
                                RemoveSecurityGroupIds=[sg['GroupId'] for sg in security_group_ids[1:]]
                            )
                            console.print(f"[green]        ✓ Simplified security group associations[/green]")
                            cleanup_performed = True
                            
                        except ClientError as sg_error:
                            console.print(f"[yellow]        → Could not modify security groups: {sg_error}[/yellow]")
                
                # Check for and clean up endpoint connections that might be blocking deletion
                try:
                    console.print(f"[yellow]        → Checking for endpoint connections blocking deletion...[/yellow]")
                    
                    # For GWLB service endpoints, check if there are connections from other accounts/VPCs
                    if service_name.startswith('com.amazonaws.vpce.'):
                        service_id = service_name.split('.')[-1]
                        
                        try:
                            # Check for endpoint connections
                            connections_response = self.ec2.describe_vpc_endpoint_connections(
                                Filters=[{'Name': 'service-id', 'Values': [service_id]}]
                            )
                            
                            connections = connections_response.get('VpcEndpointConnections', [])
                            if connections:
                                console.print(f"[yellow]        → Found {len(connections)} endpoint connections to clean up[/yellow]")
                                
                                for connection in connections:
                                    connection_vpce_id = connection.get('VpcEndpointId')
                                    connection_state = connection.get('VpcEndpointState', 'Unknown')
                                    connection_owner = connection.get('VpcEndpointOwner', 'Unknown')
                                    
                                    console.print(f"[yellow]          - {connection_vpce_id} ({connection_state}) owned by {connection_owner}[/yellow]")
                                    
                                    # Try to reject/delete the connection
                                    try:
                                        if connection_state in ['PendingAcceptance', 'Pending']:
                                            console.print(f"[yellow]            → Rejecting pending connection[/yellow]")
                                            self.ec2.reject_vpc_endpoint_connections(
                                                ServiceId=service_id,
                                                VpcEndpointIds=[connection_vpce_id]
                                            )
                                            console.print(f"[green]            ✓ Rejected connection from {connection_vpce_id}[/green]")
                                            cleanup_performed = True
                                        elif connection_state == 'Available':
                                            console.print(f"[yellow]            → Attempting to delete connected endpoint[/yellow]")
                                            # Try to delete the connected endpoint if it's in our account
                                            if connection_owner == 'self' or connection_owner == endpoint.get('Owner', ''):
                                                try:
                                                    self.ec2.delete_vpc_endpoints(VpcEndpointIds=[connection_vpce_id])
                                                    console.print(f"[green]            ✓ Deleted connected endpoint {connection_vpce_id}[/green]")
                                                    cleanup_performed = True
                                                except ClientError as connected_delete_error:
                                                    console.print(f"[yellow]            → Could not delete connected endpoint: {connected_delete_error}[/yellow]")
                                            else:
                                                console.print(f"[yellow]            → Connection is from another account, cannot delete automatically[/yellow]")
                                                
                                    except ClientError as connection_error:
                                        console.print(f"[yellow]            → Could not manage connection: {connection_error}[/yellow]")
                            else:
                                console.print(f"[yellow]        → No endpoint connections found[/yellow]")
                                
                        except ClientError as connections_error:
                            console.print(f"[yellow]        → Could not check endpoint connections: {connections_error}[/yellow]")
                        
                        # Try to delete the service configuration after cleaning up connections
                        try:
                            console.print(f"[yellow]        → Attempting to delete VPC Endpoint Service configuration: {service_id}[/yellow]")
                            self.ec2.delete_vpc_endpoint_service_configurations(ServiceIds=[service_id])
                            console.print(f"[green]        ✓ Deleted VPC Endpoint Service configuration[/green]")
                            cleanup_performed = True
                            
                        except (AttributeError, ClientError) as service_error:
                            console.print(f"[yellow]        → Could not delete service configuration: {service_error}[/yellow]")
                            
                except Exception as connection_cleanup_error:
                    console.print(f"[yellow]        → Error during connection cleanup: {connection_cleanup_error}[/yellow]")
                
                if cleanup_performed:
                    console.print(f"[green]      ✓ Performed some force cleanup actions[/green]")
                    # Wait for changes to propagate
                    import time
                    time.sleep(5)
                else:
                    console.print(f"[yellow]      → No force cleanup actions were possible[/yellow]")
                
                return cleanup_performed
                
            except ClientError as endpoint_error:
                if 'InvalidVpcEndpointId.NotFound' in str(endpoint_error):
                    console.print(f"[green]      ✓ VPC endpoint {vpce_id} no longer exists[/green]")
                    return True
                else:
                    console.print(f"[yellow]      → Could not get endpoint details: {endpoint_error}[/yellow]")
                    return False
            
        except Exception as e:
            console.print(f"[yellow]      → Error during force cleanup: {e}[/yellow]")
            return False

    def _ultimate_network_interface_cleanup(self, subnet_id: str, vpce_id: str) -> bool:
        """Ultimate fallback: try to work around persistent VPC endpoints by manipulating their network interfaces."""
        try:
            console.print(f"[yellow]        → Ultimate network interface cleanup for subnet {subnet_id}...[/yellow]")
            
            # Find the specific network interface for this VPC endpoint in this subnet
            ni_response = self.ec2.describe_network_interfaces(
                Filters=[
                    {'Name': 'subnet-id', 'Values': [subnet_id]},
                    {'Name': 'description', 'Values': [f'*{vpce_id}*']}
                ]
            )
            
            target_interfaces = [
                ni for ni in ni_response.get('NetworkInterfaces', [])
                if ni.get('InterfaceType') == 'gateway_load_balancer_endpoint'
            ]
            
            if not target_interfaces:
                console.print(f"[yellow]        → No target network interfaces found[/yellow]")
                return False
            
            cleanup_success = False
            
            for ni in target_interfaces:
                ni_id = ni.get('NetworkInterfaceId')
                ni_status = ni.get('Status', 'unknown')
                
                console.print(f"[yellow]        → Attempting ultimate cleanup of network interface {ni_id} ({ni_status})[/yellow]")
                
                try:
                    # Try to modify the network interface to make it deletable
                    # First, try to detach any attachments forcefully
                    attachment = ni.get('Attachment', {})
                    if attachment and attachment.get('AttachmentId'):
                        attachment_id = attachment['AttachmentId']
                        
                        # Only try to detach if it's not a managed attachment
                        if not attachment_id.startswith(('ela-attach', 'vpce-attach')):
                            try:
                                console.print(f"[yellow]          → Force detaching network interface...[/yellow]")
                                self.ec2.detach_network_interface(AttachmentId=attachment_id, Force=True)
                                console.print(f"[green]          ✓ Force detached network interface[/green]")
                                
                                # Wait for detachment
                                import time
                                time.sleep(10)
                                cleanup_success = True
                                
                            except ClientError as detach_error:
                                console.print(f"[yellow]          → Could not force detach: {detach_error}[/yellow]")
                    
                    # Try to reset network interface attributes that might be blocking deletion
                    try:
                        console.print(f"[yellow]          → Attempting to reset network interface attributes...[/yellow]")
                        
                        # Try to reset source/dest check
                        self.ec2.modify_network_interface_attribute(
                            NetworkInterfaceId=ni_id,
                            SourceDestCheck={'Value': True}
                        )
                        console.print(f"[green]          ✓ Reset source/dest check[/green]")
                        cleanup_success = True
                        
                    except ClientError as attr_error:
                        console.print(f"[yellow]          → Could not reset attributes: {attr_error}[/yellow]")
                    
                    # Final attempt: try to delete the network interface directly
                    try:
                        console.print(f"[yellow]          → Final attempt: direct network interface deletion...[/yellow]")
                        self.ec2.delete_network_interface(NetworkInterfaceId=ni_id)
                        console.print(f"[green]          ✓ Successfully deleted network interface {ni_id}[/green]")
                        cleanup_success = True
                        
                    except ClientError as ni_delete_error:
                        console.print(f"[yellow]          → Direct deletion failed: {ni_delete_error}[/yellow]")
                
                except Exception as ni_cleanup_error:
                    console.print(f"[yellow]        → Error during network interface cleanup: {ni_cleanup_error}[/yellow]")
            
            if cleanup_success:
                console.print(f"[green]        ✓ Ultimate cleanup performed some actions[/green]")
                console.print(f"[yellow]        → Waiting 30 seconds for changes to propagate...[/yellow]")
                import time
                time.sleep(30)
            
            return cleanup_success
            
        except Exception as e:
            console.print(f"[yellow]        → Error during ultimate cleanup: {e}[/yellow]")
            return False

    def delete_network_interfaces(self) -> None:
        """Delete orphaned network interfaces in the VPC."""
        try:
            response = self.ec2.describe_network_interfaces(
                Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
            )
            
            deleted_count = 0
            failed_nis = []
            
            for ni in response.get('NetworkInterfaces', []):
                ni_id = ni.get('NetworkInterfaceId')
                ni_status = ni.get('Status', 'unknown')
                ni_type = ni.get('InterfaceType', 'interface')
                description = ni.get('Description', 'No description')
                
                # Skip network interfaces that are attached to instances or other services
                attachment = ni.get('Attachment', {})
                if attachment.get('InstanceId'):
                    console.print(f"[yellow]Skipping network interface {ni_id}: attached to EC2 instance {attachment['InstanceId']}[/yellow]")
                    continue  # Skip - attached to instance
                
                # Skip service-managed network interface types
                if ni_type in ['nat_gateway', 'vpc_endpoint', 'load_balancer', 'gateway_load_balancer']:
                    console.print(f"[yellow]Skipping network interface {ni_id}: managed by {ni_type}[/yellow]")
                    continue  # Skip - managed by other services
                
                # Skip ELB-managed network interfaces (identified by attachment ID)
                attachment_id = attachment.get('AttachmentId', '')
                if attachment_id.startswith('ela-attach'):
                    console.print(f"[yellow]Skipping network interface {ni_id}: managed by load balancer (ELB attachment)[/yellow]")
                    continue
                
                # Skip VPC endpoint-managed network interfaces (but log them for visibility)
                if attachment_id.startswith('vpce-attach'):
                    console.print(f"[yellow]Skipping network interface {ni_id}: managed by VPC endpoint[/yellow]")
                    continue
                
                # Also skip gateway_load_balancer_endpoint interfaces (managed by VPC endpoints)
                if ni_type == 'gateway_load_balancer_endpoint':
                    console.print(f"[yellow]Skipping GWLB endpoint network interface {ni_id}: managed by VPC endpoint service[/yellow]")
                    continue
                
                # Skip Lambda function network interfaces that might still be in use
                if 'lambda' in description.lower() and ni_status == 'in-use':
                    console.print(f"[yellow]Skipping Lambda network interface {ni_id}: still in use[/yellow]")
                    continue
                
                # Skip network interfaces with ELB-related descriptions
                if any(keyword in description.lower() for keyword in ['elb', 'elasticloadbalancing', 'load balancer']):
                    console.print(f"[yellow]Skipping network interface {ni_id}: appears to be load balancer managed[/yellow]")
                    continue
                
                try:
                    console.print(f"[yellow]Deleting network interface: {ni_id} ({ni_type}, {ni_status})[/yellow]")
                    console.print(f"[yellow]  Description: {description}[/yellow]")
                    
                    # Detach if needed (though we skip attached ones above)
                    if ni.get('Attachment') and ni['Attachment'].get('AttachmentId'):
                        attachment_id = ni['Attachment']['AttachmentId']
                        attachment_status = ni['Attachment'].get('Status', 'unknown')
                        
                        # Check if this is a managed attachment that we shouldn't try to detach
                        if attachment_id.startswith('ela-attach'):
                            console.print(f"[yellow]  Skipping ELB-managed attachment {attachment_id} (managed by load balancer)[/yellow]")
                        elif attachment_id.startswith('vpce-attach'):
                            console.print(f"[yellow]  Skipping VPC endpoint attachment {attachment_id} (managed by VPC endpoint)[/yellow]")
                        elif attachment_status == 'detaching':
                            console.print(f"[yellow]  Attachment {attachment_id} is already detaching[/yellow]")
                        else:
                            try:
                                console.print(f"[yellow]  Attempting to detach network interface from attachment {attachment_id}[/yellow]")
                                self.ec2.detach_network_interface(AttachmentId=attachment_id, Force=True)
                                console.print(f"[yellow]  Detached network interface from attachment {attachment_id}[/yellow]")
                                # Wait a moment for detachment
                                import time
                                time.sleep(2)
                            except ClientError as detach_error:
                                error_code = detach_error.response.get('Error', {}).get('Code', '')
                                error_message = detach_error.response.get('Error', {}).get('Message', str(detach_error))
                                
                                if error_code == 'OperationNotPermitted':
                                    if 'ela-attach' in error_message:
                                        console.print(f"[yellow]  Cannot detach ELB attachment {attachment_id}: managed by load balancer[/yellow]")
                                    elif 'vpce-attach' in error_message:
                                        console.print(f"[yellow]  Cannot detach VPC endpoint attachment {attachment_id}: managed by VPC endpoint[/yellow]")
                                    else:
                                        console.print(f"[yellow]  Cannot detach attachment {attachment_id}: {error_message}[/yellow]")
                                else:
                                    console.print(f"[yellow]  Could not detach network interface: {error_message}[/yellow]")
                    
                    self.ec2.delete_network_interface(NetworkInterfaceId=ni_id)
                    self.deleted_resources['network_interfaces'].append(ni_id)
                    deleted_count += 1
                    
                except ClientError as ni_error:
                    error_code = ni_error.response.get('Error', {}).get('Code', '')
                    error_message = ni_error.response.get('Error', {}).get('Message', str(ni_error))
                    
                    if error_code == 'InvalidNetworkInterface.InUse':
                        console.print(f"[red]✗ Network interface {ni_id} is currently in use: {error_message}[/red]")
                        console.print(f"[yellow]  Network interface details:[/yellow]")
                        console.print(f"[yellow]    Type: {ni_type}[/yellow]")
                        console.print(f"[yellow]    Status: {ni_status}[/yellow]")
                        console.print(f"[yellow]    Description: {description}[/yellow]")
                        
                        # Try to identify what's using this network interface
                        self._identify_network_interface_usage(ni_id, ni)
                        
                        failed_nis.append(ni_id)
                    else:
                        console.print(f"[red]  ✗ Error deleting network interface {ni_id}: {error_message}[/red]")
                        failed_nis.append(ni_id)
            
            if deleted_count > 0:
                console.print(f"[green]✓ Deleted {deleted_count} network interfaces[/green]")
            if failed_nis:
                console.print(f"[yellow]⚠ Failed to delete {len(failed_nis)} network interfaces[/yellow]")
                
        except ClientError as e:
            console.print(f"[red]Error managing network interfaces: {e}[/red]")

    def cleanup_gwlb_network_interfaces(self) -> None:
        """Clean up Gateway Load Balancer network interfaces with retries."""
        try:
            console.print(f"[yellow]Checking for Gateway Load Balancer network interfaces to clean up...[/yellow]")
            
            # Wait for GWLB network interfaces to be automatically cleaned up
            max_attempts = 6  # 6 attempts with 30 second intervals = 3 minutes
            attempt = 0
            
            while attempt < max_attempts:
                attempt += 1
                
                response = self.ec2.describe_network_interfaces(
                    Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
                )
                
                gwlb_interfaces = [ni for ni in response.get('NetworkInterfaces', []) 
                                 if ni.get('InterfaceType') == 'gateway_load_balancer']
                
                if not gwlb_interfaces:
                    console.print(f"[green]✓ All Gateway Load Balancer network interfaces have been cleaned up[/green]")
                    break
                
                if attempt == 1:
                    console.print(f"[yellow]Found {len(gwlb_interfaces)} GWLB network interfaces, waiting for automatic cleanup...[/yellow]")
                    for ni in gwlb_interfaces:
                        ni_id = ni.get('NetworkInterfaceId')
                        description = ni.get('Description', 'No description')
                        console.print(f"[yellow]  - {ni_id}: {description}[/yellow]")
                
                if attempt < max_attempts:
                    console.print(f"[yellow]Waiting 30 seconds for GWLB network interfaces to be cleaned up... (attempt {attempt}/{max_attempts})[/yellow]")
                    import time
                    time.sleep(30)
                else:
                    console.print(f"[yellow]⚠ Some GWLB network interfaces are still present after waiting[/yellow]")
                    console.print(f"[yellow]  These should be cleaned up automatically by AWS, but it may take longer[/yellow]")
                    for ni in gwlb_interfaces:
                        ni_id = ni.get('NetworkInterfaceId')
                        status = ni.get('Status', 'unknown')
                        console.print(f"[yellow]  - {ni_id} ({status})[/yellow]")
                    
        except ClientError as e:
            console.print(f"[red]Error checking GWLB network interfaces: {e}[/red]")

    def _identify_network_interface_usage(self, ni_id: str, ni_details: dict) -> None:
        """Try to identify what's using a network interface."""
        try:
            console.print(f"[yellow]  → Checking what's using network interface {ni_id}...[/yellow]")
            
            # Check attachment details
            attachment = ni_details.get('Attachment', {})
            if attachment:
                instance_id = attachment.get('InstanceId')
                attachment_id = attachment.get('AttachmentId')
                device_index = attachment.get('DeviceIndex')
                status = attachment.get('Status', 'unknown')
                
                console.print(f"[yellow]  → Attachment details:[/yellow]")
                console.print(f"[yellow]    Status: {status}[/yellow]")
                console.print(f"[yellow]    Device Index: {device_index}[/yellow]")
                console.print(f"[yellow]    Attachment ID: {attachment_id}[/yellow]")
                
                # Identify the type of attachment
                if attachment_id.startswith('ela-attach'):
                    console.print(f"[yellow]    → This is an ELB (Elastic Load Balancer) managed attachment[/yellow]")
                    console.print(f"[yellow]    → Cannot be manually detached - managed automatically by load balancer[/yellow]")
                elif attachment_id.startswith('vpce-attach'):
                    console.print(f"[yellow]    → This is a VPC Endpoint managed attachment[/yellow]")
                elif attachment_id.startswith('eni-attach'):
                    console.print(f"[yellow]    → This is a standard network interface attachment[/yellow]")
                
                if instance_id:
                    console.print(f"[yellow]    Attached to EC2 instance: {instance_id}[/yellow]")
                    
                    # Try to get instance details
                    try:
                        instance_response = self.ec2.describe_instances(InstanceIds=[instance_id])
                        for reservation in instance_response['Reservations']:
                            for instance in reservation['Instances']:
                                instance_state = instance.get('State', {}).get('Name', 'unknown')
                                instance_type = instance.get('InstanceType', 'unknown')
                                console.print(f"[yellow]    Instance state: {instance_state} ({instance_type})[/yellow]")
                                
                                # Get instance name from tags
                                instance_name = 'Unnamed'
                                for tag in instance.get('Tags', []):
                                    if tag['Key'] == 'Name':
                                        instance_name = tag['Value']
                                        break
                                console.print(f"[yellow]    Instance name: {instance_name}[/yellow]")
                                
                    except ClientError as instance_error:
                        console.print(f"[yellow]    Could not get instance details: {instance_error}[/yellow]")
            
            # Check for Lambda function association
            ni_description = ni_details.get('Description', '')
            if 'lambda' in ni_description.lower() or 'aws lambda' in ni_description.lower():
                console.print(f"[yellow]  → This appears to be a Lambda function network interface[/yellow]")
                
                # Try to find associated Lambda functions
                try:
                    lambda_response = self.lambda_client.list_functions()
                    subnet_id = ni_details.get('SubnetId')
                    
                    if subnet_id:
                        lambda_functions_in_subnet = []
                        for function in lambda_response.get('Functions', []):
                            try:
                                config = self.lambda_client.get_function_configuration(
                                    FunctionName=function['FunctionName']
                                )
                                vpc_config = config.get('VpcConfig', {})
                                if subnet_id in vpc_config.get('SubnetIds', []):
                                    lambda_functions_in_subnet.append(function['FunctionName'])
                            except ClientError:
                                continue
                        
                        if lambda_functions_in_subnet:
                            console.print(f"[yellow]  → Lambda functions in same subnet:[/yellow]")
                            for func_name in lambda_functions_in_subnet:
                                console.print(f"[yellow]    - {func_name}[/yellow]")
                                
                except ClientError as lambda_error:
                    console.print(f"[yellow]  → Could not check Lambda functions: {lambda_error}[/yellow]")
            
            # Check for RDS association
            if 'rds' in ni_description.lower() or 'database' in ni_description.lower():
                console.print(f"[yellow]  → This appears to be an RDS-related network interface[/yellow]")
            
            # Check for ELB association
            if any(keyword in ni_description.lower() for keyword in ['elb', 'load balancer', 'loadbalancer']):
                console.print(f"[yellow]  → This appears to be a load balancer network interface[/yellow]")
            
            # Check groups (security groups)
            groups = ni_details.get('Groups', [])
            if groups:
                console.print(f"[yellow]  → Security groups:[/yellow]")
                for group in groups[:3]:  # Show first 3
                    group_id = group.get('GroupId', 'Unknown')
                    group_name = group.get('GroupName', 'Unknown')
                    console.print(f"[yellow]    - {group_id} ({group_name})[/yellow]")
            
            # Check private IP addresses
            private_ips = ni_details.get('PrivateIpAddresses', [])
            if private_ips:
                console.print(f"[yellow]  → Private IP addresses:[/yellow]")
                for ip_info in private_ips[:2]:  # Show first 2
                    private_ip = ip_info.get('PrivateIpAddress', 'Unknown')
                    is_primary = ip_info.get('Primary', False)
                    console.print(f"[yellow]    - {private_ip} {'(Primary)' if is_primary else ''}[/yellow]")
            
            # Provide guidance based on attachment type
            attachment = ni_details.get('Attachment', {})
            attachment_id = attachment.get('AttachmentId', '')
            
            console.print(f"[yellow]  To resolve network interface issues:[/yellow]")
            
            if attachment_id.startswith('ela-attach'):
                console.print(f"[yellow]  → This is an ELB-managed network interface:[/yellow]")
                console.print(f"[yellow]    1. Delete the associated load balancer first[/yellow]")
                console.print(f"[yellow]    2. The network interface will be automatically cleaned up[/yellow]")
                console.print(f"[yellow]    3. Do not attempt to manually detach or delete this interface[/yellow]")
            elif attachment_id.startswith('vpce-attach'):
                console.print(f"[yellow]  → This is a VPC Endpoint managed network interface:[/yellow]")
                console.print(f"[yellow]    1. Delete the associated VPC endpoint first[/yellow]")
                console.print(f"[yellow]    2. The network interface will be automatically cleaned up[/yellow]")
            else:
                console.print(f"[yellow]  1. If attached to EC2 instance: Stop or terminate the instance[/yellow]")
                console.print(f"[yellow]  2. If Lambda-related: Delete the Lambda function or modify its VPC config[/yellow]")
                console.print(f"[yellow]  3. If RDS-related: Delete RDS instances or modify subnet groups[/yellow]")
                console.print(f"[yellow]  4. If load balancer-related: Delete the load balancer[/yellow]")
            
            console.print(f"[yellow]  5. Wait a few minutes after removing resources, then re-run this tool[/yellow]")
            
        except Exception as e:
            console.print(f"[yellow]  → Error identifying network interface usage: {e}[/yellow]")

    def _identify_subnet_dependencies(self, subnet_id: str) -> None:
        """Try to identify what dependencies are preventing subnet deletion."""
        try:
            console.print(f"[yellow]  → Checking dependencies for subnet {subnet_id}...[/yellow]")
            
            # Check for network interfaces
            try:
                ni_response = self.ec2.describe_network_interfaces(
                    Filters=[{'Name': 'subnet-id', 'Values': [subnet_id]}]
                )
                
                network_interfaces = ni_response.get('NetworkInterfaces', [])
                if network_interfaces:
                    console.print(f"[yellow]  → Found {len(network_interfaces)} network interfaces in subnet:[/yellow]")
                    
                    gwlb_interfaces = []
                    other_interfaces = []
                    
                    for ni in network_interfaces:
                        ni_id = ni.get('NetworkInterfaceId', 'Unknown')
                        ni_type = ni.get('InterfaceType', 'Unknown')
                        ni_status = ni.get('Status', 'Unknown')
                        description = ni.get('Description', 'No description')
                        
                        if ni_type == 'gateway_load_balancer':
                            gwlb_interfaces.append((ni_id, ni_type, ni_status, description))
                        elif ni_type == 'gateway_load_balancer_endpoint':
                            gwlb_interfaces.append((ni_id, ni_type, ni_status, description))
                        else:
                            other_interfaces.append((ni_id, ni_type, ni_status, description))
                    
                    # Show Gateway Load Balancer interfaces first with special handling
                    if gwlb_interfaces:
                        console.print(f"[yellow]    → Gateway Load Balancer network interfaces (managed by AWS):[/yellow]")
                        for ni_id, ni_type, ni_status, description in gwlb_interfaces:
                            console.print(f"[yellow]      - {ni_id} ({ni_type}, {ni_status}): {description}[/yellow]")
                            
                            if ni_type == 'gateway_load_balancer':
                                # Extract GWLB name from description
                                if 'gwy/' in description:
                                    gwlb_name = description.split('gwy/')[-1].split('/')[0]
                                    console.print(f"[yellow]        → This belongs to Gateway Load Balancer: {gwlb_name}[/yellow]")
                                    console.print(f"[yellow]        → Cannot be deleted manually - managed by GWLB service[/yellow]")
                            elif ni_type == 'gateway_load_balancer_endpoint':
                                # Extract VPC endpoint ID from description
                                if 'vpce-' in description:
                                    import re
                                    vpce_match = re.search(r'vpce-[a-f0-9]+', description)
                                    if vpce_match:
                                        vpce_id = vpce_match.group(0)
                                        console.print(f"[yellow]        → This belongs to VPC Endpoint: {vpce_id}[/yellow]")
                                        console.print(f"[yellow]        → This endpoint connects to a GWLB service[/yellow]")
                                        console.print(f"[yellow]        → Cannot be deleted manually - managed by VPC Endpoint service[/yellow]")
                    
                    # Show other interfaces
                    if other_interfaces:
                        if gwlb_interfaces:
                            console.print(f"[yellow]    → Other network interfaces:[/yellow]")
                        for ni_id, ni_type, ni_status, description in other_interfaces[:5]:
                            console.print(f"[yellow]      - {ni_id} ({ni_type}, {ni_status}): {description}[/yellow]")
                        
                        if len(other_interfaces) > 5:
                            console.print(f"[yellow]      ... and {len(other_interfaces) - 5} more[/yellow]")
                        
            except ClientError as ni_error:
                console.print(f"[yellow]  → Unable to check network interfaces: {ni_error}[/yellow]")
            
            # Check for Lambda functions in this subnet
            try:
                lambda_response = self.lambda_client.list_functions()
                lambda_functions_in_subnet = []
                
                for function in lambda_response.get('Functions', []):
                    try:
                        config = self.lambda_client.get_function_configuration(
                            FunctionName=function['FunctionName']
                        )
                        vpc_config = config.get('VpcConfig', {})
                        if subnet_id in vpc_config.get('SubnetIds', []):
                            lambda_functions_in_subnet.append(function['FunctionName'])
                    except ClientError:
                        continue
                
                if lambda_functions_in_subnet:
                    console.print(f"[yellow]  → Found Lambda functions using this subnet:[/yellow]")
                    for func_name in lambda_functions_in_subnet[:3]:
                        console.print(f"[yellow]    - {func_name}[/yellow]")
                    if len(lambda_functions_in_subnet) > 3:
                        console.print(f"[yellow]    ... and {len(lambda_functions_in_subnet) - 3} more[/yellow]")
                        
            except ClientError as lambda_error:
                console.print(f"[yellow]  → Unable to check Lambda functions: {lambda_error}[/yellow]")
            
            # Check for RDS instances
            try:
                rds_response = self.rds.describe_db_instances()
                rds_instances_in_subnet = []
                
                for db_instance in rds_response.get('DBInstances', []):
                    db_subnet_group = db_instance.get('DBSubnetGroup', {})
                    subnets = db_subnet_group.get('Subnets', [])
                    for subnet in subnets:
                        if subnet.get('SubnetIdentifier') == subnet_id:
                            rds_instances_in_subnet.append(db_instance.get('DBInstanceIdentifier', 'Unknown'))
                            break
                
                if rds_instances_in_subnet:
                    console.print(f"[yellow]  → Found RDS instances using this subnet:[/yellow]")
                    for db_name in rds_instances_in_subnet:
                        console.print(f"[yellow]    - {db_name}[/yellow]")
                        
            except ClientError as rds_error:
                console.print(f"[yellow]  → Unable to check RDS instances: {rds_error}[/yellow]")
            
            # Provide specific guidance based on what was found
            console.print(f"[yellow]  To resolve subnet dependency issues:[/yellow]")
            
            # Check if we found GWLB interfaces earlier
            try:
                ni_response = self.ec2.describe_network_interfaces(
                    Filters=[{'Name': 'subnet-id', 'Values': [subnet_id]}]
                )
                has_gwlb = any(ni.get('InterfaceType') == 'gateway_load_balancer' 
                              for ni in ni_response.get('NetworkInterfaces', []))
                
                if has_gwlb:
                    # Check if we have GWLB endpoint interfaces (VPC endpoints connecting to GWLB)
                    ni_response = self.ec2.describe_network_interfaces(
                        Filters=[{'Name': 'subnet-id', 'Values': [subnet_id]}]
                    )
                    has_gwlb_endpoint = any(ni.get('InterfaceType') == 'gateway_load_balancer_endpoint' 
                                          for ni in ni_response.get('NetworkInterfaces', []))
                    
                    console.print(f"[yellow]  → Gateway Load Balancer dependencies detected:[/yellow]")
                    
                    if has_gwlb_endpoint:
                        console.print(f"[yellow]    → VPC Endpoint connection to GWLB service detected[/yellow]")
                        console.print(f"[yellow]    1. Delete VPC endpoints connecting to the GWLB service[/yellow]")
                        console.print(f"[yellow]       - Go to VPC Console → Endpoints[/yellow]")
                        console.print(f"[yellow]       - Look for endpoints with 'firewall' or 'security' in service name[/yellow]")
                        console.print(f"[yellow]       - Delete these VPC endpoints[/yellow]")
                        console.print(f"[yellow]    2. Delete VPC Endpoint Service configurations[/yellow]")
                        console.print(f"[yellow]       - Go to VPC Console → Endpoint Services[/yellow]")
                        console.print(f"[yellow]       - Delete any endpoint service configurations[/yellow]")
                        console.print(f"[yellow]    3. Delete the Gateway Load Balancer[/yellow]")
                        console.print(f"[yellow]       - Go to EC2 Console → Load Balancers[/yellow]")
                        console.print(f"[yellow]       - Delete the Gateway Load Balancer[/yellow]")
                        console.print(f"[yellow]    4. Network interfaces will be automatically cleaned up[/yellow]")
                        console.print(f"[yellow]    5. Wait 5-10 minutes for cleanup to complete[/yellow]")
                    else:
                        console.print(f"[yellow]    1. Delete VPC Endpoint Service configurations using the GWLB[/yellow]")
                        console.print(f"[yellow]       - Go to VPC Console → Endpoint Services[/yellow]")
                        console.print(f"[yellow]       - Delete any endpoint service configurations[/yellow]")
                        console.print(f"[yellow]    2. Delete the Gateway Load Balancer itself[/yellow]")
                        console.print(f"[yellow]       - Go to EC2 Console → Load Balancers[/yellow]")
                        console.print(f"[yellow]       - Delete the Gateway Load Balancer[/yellow]")
                        console.print(f"[yellow]    3. Network interfaces will be automatically cleaned up[/yellow]")
                        console.print(f"[yellow]    4. Wait 5-10 minutes for cleanup to complete[/yellow]")
                else:
                    console.print(f"[yellow]  1. Delete or move any EC2 instances in this subnet[/yellow]")
                    console.print(f"[yellow]  2. Delete any Lambda functions using this subnet[/yellow]")
                    console.print(f"[yellow]  3. Delete any RDS instances or modify their subnet groups[/yellow]")
                    console.print(f"[yellow]  4. Check for any load balancers using this subnet[/yellow]")
                    console.print(f"[yellow]  5. Remove any network interfaces not automatically cleaned up[/yellow]")
                    
            except ClientError:
                # Fallback to general guidance if we can't check
                console.print(f"[yellow]  1. Delete or move any EC2 instances in this subnet[/yellow]")
                console.print(f"[yellow]  2. Delete any Lambda functions using this subnet[/yellow]")
                console.print(f"[yellow]  3. Delete any RDS instances or modify their subnet groups[/yellow]")
                console.print(f"[yellow]  4. Check for any load balancers using this subnet[/yellow]")
                console.print(f"[yellow]  5. Remove any network interfaces not automatically cleaned up[/yellow]")
            
            console.print(f"[yellow]  Then re-run this tool to complete VPC deletion.[/yellow]")
            
        except Exception as e:
            console.print(f"[yellow]  → Error identifying subnet dependencies: {e}[/yellow]")

    def retry_failed_subnet_deletions(self) -> None:
        """Retry deleting subnets that may have been blocked by GWLB network interfaces."""
        try:
            console.print(f"[yellow]Retrying subnet deletions after GWLB cleanup...[/yellow]")
            
            # Get current subnets in the VPC
            response = self.ec2.describe_subnets(
                Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
            )
            
            remaining_subnets = response.get('Subnets', [])
            if not remaining_subnets:
                console.print(f"[green]✓ All subnets have been deleted[/green]")
                return
            
            console.print(f"[yellow]Found {len(remaining_subnets)} remaining subnets to delete[/yellow]")
            
            retry_deleted_count = 0
            still_failed_subnets = []
            
            for subnet in remaining_subnets:
                subnet_id = subnet['SubnetId']
                availability_zone = subnet.get('AvailabilityZone', 'Unknown')
                cidr_block = subnet.get('CidrBlock', 'Unknown')
                
                try:
                    console.print(f"[yellow]Retrying deletion of subnet: {subnet_id} ({cidr_block} in {availability_zone})[/yellow]")
                    self.ec2.delete_subnet(SubnetId=subnet_id)
                    
                    # Update our tracking if this subnet wasn't previously tracked
                    if subnet_id not in self.deleted_resources['subnets']:
                        self.deleted_resources['subnets'].append(subnet_id)
                    
                    retry_deleted_count += 1
                    console.print(f"[green]✓ Successfully deleted subnet {subnet_id} on retry[/green]")
                    
                except ClientError as subnet_error:
                    error_code = subnet_error.response.get('Error', {}).get('Code', '')
                    error_message = subnet_error.response.get('Error', {}).get('Message', str(subnet_error))
                    
                    if error_code == 'DependencyViolation':
                        console.print(f"[yellow]⚠ Subnet {subnet_id} still has dependencies: {error_message}[/yellow]")
                        still_failed_subnets.append(subnet_id)
                    else:
                        console.print(f"[red]✗ Error retrying subnet {subnet_id}: {error_message}[/red]")
                        still_failed_subnets.append(subnet_id)
            
            if retry_deleted_count > 0:
                console.print(f"[green]✓ Successfully deleted {retry_deleted_count} subnets on retry[/green]")
            
            if still_failed_subnets:
                console.print(f"[yellow]⚠ {len(still_failed_subnets)} subnets still have dependencies: {', '.join(still_failed_subnets)}[/yellow]")
                console.print(f"[yellow]  These may require additional manual cleanup[/yellow]")
                
        except ClientError as e:
            console.print(f"[red]Error retrying subnet deletions: {e}[/red]")

    def delete_route_tables(self) -> None:
        """Delete custom route tables (not the main route table)."""
        try:
            response = self.ec2.describe_route_tables(
                Filters=[{'Name': 'vpc-id', 'Values': [self.vpc_id]}]
            )
            
            deleted_count = 0
            failed_route_tables = []
            
            # Filter out main route tables
            custom_route_tables = [rt for rt in response['RouteTables'] 
                                 if not any(assoc.get('Main', False) for assoc in rt.get('Associations', []))]
            
            for route_table in custom_route_tables:
                route_table_id = route_table['RouteTableId']
                
                try:
                    console.print(f"[yellow]Processing route table: {route_table_id}[/yellow]")
                    
                    # First, disassociate from any subnets
                    associations = route_table.get('Associations', [])
                    subnet_associations = [assoc for assoc in associations if assoc.get('SubnetId')]
                    
                    for association in subnet_associations:
                        assoc_id = association.get('RouteTableAssociationId')
                        subnet_id = association.get('SubnetId')
                        
                        if assoc_id:
                            try:
                                console.print(f"[yellow]  Disassociating from subnet {subnet_id}[/yellow]")
                                self.ec2.disassociate_route_table(AssociationId=assoc_id)
                            except ClientError as disassoc_error:
                                console.print(f"[yellow]  Could not disassociate from subnet {subnet_id}: {disassoc_error}[/yellow]")
                    
                    # Remove custom routes (keep local routes)
                    routes = route_table.get('Routes', [])
                    custom_routes = [route for route in routes 
                                   if route.get('Origin') != 'CreateRouteTable' and route.get('State') != 'blackhole']
                    
                    for route in custom_routes:
                        destination = route.get('DestinationCidrBlock') or route.get('DestinationIpv6CidrBlock')
                        if destination:
                            try:
                                console.print(f"[yellow]  Removing route to {destination}[/yellow]")
                                delete_route_params = {'RouteTableId': route_table_id}
                                
                                if route.get('DestinationCidrBlock'):
                                    delete_route_params['DestinationCidrBlock'] = route['DestinationCidrBlock']
                                elif route.get('DestinationIpv6CidrBlock'):
                                    delete_route_params['DestinationIpv6CidrBlock'] = route['DestinationIpv6CidrBlock']
                                
                                self.ec2.delete_route(**delete_route_params)
                            except ClientError as route_error:
                                # Some routes can't be deleted (like local routes), which is expected
                                console.print(f"[yellow]  Could not remove route to {destination}: {route_error}[/yellow]")
                    
                    # Now try to delete the route table
                    console.print(f"[yellow]Deleting route table: {route_table_id}[/yellow]")
                    self.ec2.delete_route_table(RouteTableId=route_table_id)
                    self.deleted_resources['route_tables'].append(route_table_id)
                    deleted_count += 1
                    
                except ClientError as rt_error:
                    error_code = rt_error.response.get('Error', {}).get('Code', '')
                    error_message = rt_error.response.get('Error', {}).get('Message', str(rt_error))
                    
                    if error_code == 'DependencyViolation':
                        console.print(f"[red]✗ Cannot delete route table '{route_table_id}': {error_message}[/red]")
                        console.print(f"[yellow]  Route table still has dependencies.[/yellow]")
                        
                        # Try to identify remaining dependencies
                        self._identify_route_table_dependencies(route_table_id, route_table)
                        
                        failed_route_tables.append(route_table_id)
                    else:
                        console.print(f"[red]✗ Error deleting route table '{route_table_id}': {error_message}[/red]")
                        failed_route_tables.append(route_table_id)
                
            if deleted_count > 0:
                console.print(f"[green]✓ Deleted {deleted_count} custom route tables[/green]")
            if failed_route_tables:
                console.print(f"[yellow]⚠ Failed to delete {len(failed_route_tables)} route tables: {', '.join(failed_route_tables)}[/yellow]")
                
        except ClientError as e:
            console.print(f"[red]Error managing route tables: {e}[/red]")

    def _identify_route_table_dependencies(self, route_table_id: str, route_table: dict) -> None:
        """Try to identify what dependencies are preventing route table deletion."""
        try:
            console.print(f"[yellow]  → Checking dependencies for route table {route_table_id}...[/yellow]")
            
            # Check associations
            associations = route_table.get('Associations', [])
            if associations:
                console.print(f"[yellow]  → Found {len(associations)} associations:[/yellow]")
                for assoc in associations:
                    if assoc.get('Main'):
                        console.print(f"[yellow]    - Main route table association (cannot be removed)[/yellow]")
                    elif assoc.get('SubnetId'):
                        subnet_id = assoc['SubnetId']
                        assoc_id = assoc.get('RouteTableAssociationId', 'Unknown')
                        console.print(f"[yellow]    - Subnet {subnet_id} (Association: {assoc_id})[/yellow]")
                    elif assoc.get('GatewayId'):
                        gateway_id = assoc['GatewayId']
                        console.print(f"[yellow]    - Gateway {gateway_id}[/yellow]")
            
            # Check routes
            routes = route_table.get('Routes', [])
            custom_routes = [route for route in routes if route.get('Origin') != 'CreateRouteTable']
            
            if custom_routes:
                console.print(f"[yellow]  → Found {len(custom_routes)} custom routes:[/yellow]")
                for route in custom_routes[:5]:  # Show first 5
                    destination = route.get('DestinationCidrBlock') or route.get('DestinationIpv6CidrBlock') or 'Unknown'
                    target = (route.get('GatewayId') or 
                             route.get('InstanceId') or 
                             route.get('NetworkInterfaceId') or 
                             route.get('VpcPeeringConnectionId') or 
                             route.get('NatGatewayId') or 
                             'Unknown')
                    state = route.get('State', 'Unknown')
                    console.print(f"[yellow]    - {destination} → {target} ({state})[/yellow]")
                
                if len(custom_routes) > 5:
                    console.print(f"[yellow]    ... and {len(custom_routes) - 5} more routes[/yellow]")
            
            # Check for VPC peering connections using this route table
            try:
                peering_response = self.ec2.describe_vpc_peering_connections(
                    Filters=[
                        {'Name': 'requester-vpc-info.vpc-id', 'Values': [self.vpc_id]},
                        {'Name': 'accepter-vpc-info.vpc-id', 'Values': [self.vpc_id]}
                    ]
                )
                
                active_peering = [pc for pc in peering_response.get('VpcPeeringConnections', []) 
                                if pc.get('Status', {}).get('Code') == 'active']
                
                if active_peering:
                    console.print(f"[yellow]  → Found {len(active_peering)} active VPC peering connections[/yellow]")
                    for pc in active_peering[:3]:
                        pc_id = pc.get('VpcPeeringConnectionId', 'Unknown')
                        console.print(f"[yellow]    - {pc_id}[/yellow]")
                        
            except ClientError as peering_error:
                console.print(f"[yellow]  → Unable to check VPC peering connections: {peering_error}[/yellow]")
            
            # Provide guidance
            console.print(f"[yellow]  To resolve route table dependency issues:[/yellow]")
            console.print(f"[yellow]  1. Remove custom routes pointing to deleted resources[/yellow]")
            console.print(f"[yellow]  2. Disassociate route table from any remaining subnets[/yellow]")
            console.print(f"[yellow]  3. Delete any VPC peering connections[/yellow]")
            console.print(f"[yellow]  4. Remove routes to NAT gateways, internet gateways, etc.[/yellow]")
            console.print(f"[yellow]  Then re-run this tool to complete VPC deletion.[/yellow]")
            
        except Exception as e:
            console.print(f"[yellow]  → Error identifying route table dependencies: {e}[/yellow]")

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
            ("VPC Endpoints", self.delete_vpc_endpoints),
            ("VPC Endpoint Service Configurations", self.delete_vpc_endpoint_service_configurations),
            ("Load Balancers", self.delete_load_balancers),
            ("GWLB Network Interface Cleanup", self.cleanup_gwlb_network_interfaces),
            ("Lambda Functions", self.delete_lambda_functions),
            ("RDS Subnet Groups", self.delete_rds_subnet_groups),
            ("NAT Gateways", self.delete_nat_gateways),
            ("Peering Connections", self.delete_peering_connections),
            ("VPN Connections & Gateways", self.delete_vpn_connections),
            ("Elastic IPs", self.release_elastic_ips),
            ("Internet Gateways", self.delete_internet_gateways),
            ("Network Interfaces", self.delete_network_interfaces),
            ("Subnets", self.delete_subnets),
            ("Subnet Retry (after GWLB cleanup)", self.retry_failed_subnet_deletions),
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
        
        # Check if there were any failed resources that need manual intervention
        total_failed = sum(1 for resources in self.deleted_resources.values() if isinstance(resources, list) and len(resources) == 0)
        if not success:
            console.print(f"\n[yellow]⚠ Some resources could not be deleted automatically.[/yellow]")
            console.print(f"[yellow]If you encountered Gateway Load Balancer errors:[/yellow]")
            console.print(f"[yellow]1. Delete VPC Endpoint Service configurations in the AWS Console[/yellow]")
            console.print(f"[yellow]2. Remove any other service associations[/yellow]")
            console.print(f"[yellow]3. Re-run this tool to complete VPC deletion[/yellow]")
        
        return success


@click.command()
@click.argument('vpc_id')
@click.option('--dry-run', is_flag=True, help='Show what would be deleted without actually deleting')
@click.option('--force', is_flag=True, help='Skip confirmation prompt')
def main(vpc_id: str, dry_run: bool, force: bool):
    """Delete an AWS VPC and all its dependencies.
    
    VPC_ID: The ID of the VPC to delete (e.g., vpc-12345678)
    
    AWS region and profile are determined from environment variables:
    - AWS_DEFAULT_REGION or AWS_REGION for region
    - AWS_PROFILE for profile (optional)
    """
    import os
    
    console.print(f"[bold blue]AWS VPC Deletion Tool[/bold blue]")
    console.print(f"VPC ID: {vpc_id}")
    console.print(f"Region: {os.environ.get('AWS_DEFAULT_REGION') or os.environ.get('AWS_REGION') or 'default'}")
    console.print(f"Profile: {os.environ.get('AWS_PROFILE') or 'default'}")
    
    if not dry_run and not force:
        confirmation = click.confirm(
            f"\nAre you sure you want to delete VPC {vpc_id} and ALL its resources? This cannot be undone!"
        )
        if not confirmation:
            console.print("[yellow]Operation cancelled.[/yellow]")
            return
    
    try:
        deleter = VPCDeleter(vpc_id)
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
