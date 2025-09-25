# AWS CLI Tool

A Go-based command-line tool for interacting with AWS services using the gofr framework.

## 🆕 Recent Updates

- **ECR Image Management**: Added `ecr` command to list ECR image versions and tags with filtering and sorting
- **NLB Subnet Management**: Added `remove-subnet` command to remove subnets from Network Load Balancers by availability zone
- **Enhanced NLB Listing**: Improved table display with AZ/subnet pairs and smart name fallbacks
- **Comprehensive Documentation**: Updated with troubleshooting guides and best practices

## Prerequisites

- Go 1.25.1 or later
- AWS credentials configured (via AWS CLI, environment variables, or IAM roles)

## Building

```bash
go build -o aws .
```

## Usage

### Subnets Command

Manage AWS subnets with comprehensive functionality for listing, deleting, and checking dependencies.

### NLB Command

Manage AWS Network Load Balancers with functionality for listing NLBs and managing subnets.

### ECR Command

Manage AWS ECR repositories with functionality for listing image versions and tags.

#### List Subnets

List all subnets in a VPC with optional filtering and sorting capabilities.

```bash
# Basic usage - list all subnets in a VPC (sorted by CIDR by default)
./aws subnets --vpc vpc-12345678

# Filter by availability zone
./aws subnets --vpc vpc-12345678 --zone us-east-1a

# Sort by different criteria
./aws subnets --vpc vpc-12345678 --sort az
./aws subnets --vpc vpc-12345678 --sort name
./aws subnets --vpc vpc-12345678 --sort type

# Combine filtering and sorting
./aws subnets --vpc vpc-12345678 --zone us-east-1a --sort name
```

#### Delete Subnet

Delete a subnet by ID with dependency checking and safety features.

```bash
# Delete subnet with confirmation prompt
./aws subnets delete --subnet-id subnet-12345678

# Delete subnet without confirmation (force mode)
./aws subnets delete --subnet-id subnet-12345678 --force
```

#### Check Dependencies

Check what resources are preventing a subnet from being deleted.

```bash
# Check dependencies for a subnet
./aws subnets check-dependencies --subnet-id subnet-12345678
```

#### List NLBs

List all Network Load Balancers in a VPC with optional filtering and sorting capabilities.

```bash
# Basic usage - list all NLBs in a VPC
./aws nlb --vpc vpc-12345678

# Filter by availability zone
./aws nlb --vpc vpc-12345678 --zone us-east-1a

# Sort by different criteria
./aws nlb --vpc vpc-12345678 --sort state
./aws nlb --vpc vpc-12345678 --sort name
./aws nlb --vpc vpc-12345678 --sort scheme

# Combine filtering and sorting
./aws nlb --vpc vpc-12345678 --zone us-east-1a --sort state
```

#### Add Subnet to NLB

Add subnets from a specific zone to NLBs in a VPC. This is useful when you need to add subnets before removing others.

```bash
# Add subnets from a zone to all NLBs in a VPC
./aws nlb add-subnet --vpc vpc-12345678 --zone us-east-1b

# Add subnets to a specific NLB only
./aws nlb add-subnet --vpc vpc-12345678 --zone us-east-1b --nlb-name my-nlb

# Add subnets without confirmation prompt
./aws nlb add-subnet --vpc vpc-12345678 --zone us-east-1b --force
```

#### Check NLB Associations

Check for service associations that might prevent subnet removal from NLBs.

```bash
# Check all NLBs in a VPC for associations
./aws nlb check-associations --vpc vpc-12345678

# Check a specific NLB for associations
./aws nlb check-associations --vpc vpc-12345678 --nlb-name my-nlb
```

#### Remove Subnet from NLB

Remove subnets from Network Load Balancers in a specific VPC and availability zone.

```bash
# Remove subnets in a specific zone from all NLBs in a VPC
./aws nlb remove-subnet --vpc vpc-12345678 --zone us-east-1a

# Remove subnets from a specific NLB only
./aws nlb remove-subnet --vpc vpc-12345678 --zone us-east-1a --nlb-name my-nlb

# Remove subnets without confirmation prompt
./aws nlb remove-subnet --vpc vpc-12345678 --zone us-east-1a --force
```

#### List ECR Images

List all image versions in an ECR repository with optional filtering and sorting capabilities.

```bash
# Basic usage - list all images in a repository
./aws ecr --repository my-repo

# Filter by specific tag
./aws ecr --repository my-repo --tag latest

# Sort by different criteria
./aws ecr --repository my-repo --sort tag
./aws ecr --repository my-repo --sort size
# Default is sorted by push date (newest first)

# Combine filtering and sorting
./aws ecr --repository my-repo --tag v1.0 --sort pushed
```

#### Options

**List Subnets:**
- `--vpc VPC_ID` (required): VPC ID to list subnets for
- `--zone AZ` (optional): Filter by availability zone (e.g., us-east-1a)
- `--sort SORT_BY` (optional): Sort by one of:
  - `cidr` (default): Sort by CIDR block in network order
  - `az`: Sort by availability zone
  - `name`: Sort by subnet name (from Name tag)
  - `type`: Sort by subnet type (from Type tag)

**Delete Subnet:**
- `--subnet-id SUBNET_ID` (required): Subnet ID to delete
- `--force` (optional): Skip confirmation prompt

**Check Dependencies:**
- `--subnet-id SUBNET_ID` (required): Subnet ID to check dependencies for

**List NLBs:**
- `--vpc VPC_ID` (required): VPC ID to list NLBs for
- `--zone AZ` (optional): Filter by availability zone (e.g., us-east-1a)
- `--sort SORT_BY` (optional): Sort by one of:
  - `name` (default): Sort by NLB name
  - `state`: Sort by NLB state
  - `type`: Sort by NLB type
  - `scheme`: Sort by NLB scheme (internal/external)
  - `created`: Sort by creation time

**Add Subnet to NLB:**
- `--vpc VPC_ID` (required): VPC ID containing the NLB
- `--zone AZ` (required): Availability zone to add subnets from
- `--nlb-name NAME` (optional): Specific NLB name to target (adds to all NLBs if not specified)
- `--force` (optional): Skip confirmation prompt

**Check NLB Associations:**
- `--vpc VPC_ID` (required): VPC ID containing the NLB
- `--nlb-name NAME` (optional): Specific NLB name to check (checks all NLBs if not specified)

**Remove Subnet from NLB:**
- `--vpc VPC_ID` (required): VPC ID containing the NLB
- `--zone AZ` (required): Availability zone of the subnet to remove
- `--nlb-name NAME` (optional): Specific NLB name to target (removes from all NLBs if not specified)
- `--force` (optional): Skip confirmation prompt

**List ECR Images:**
- `--repository REPO_NAME` (required): ECR repository name
- `--tag TAG` (optional): Filter by specific image tag
- `--sort SORT_BY` (optional): Sort by one of:
  - `pushed` (default): Sort by push date (newest first)
  - `tag`: Sort by image tag
  - `size`: Sort by image size (largest first)

#### Output

**List Subnets:** Displays a formatted table with the following columns:
- Subnet ID
- CIDR Block
- AZ (Availability Zone)
- Name (from Name tag)
- State
- Type (from Type tag, defaults to "subnet")
- Tags (relevant tags like kubernetes.io/role/elb, each on a separate line)

**Delete Subnet:** Shows confirmation prompts and success/error messages.

**Check Dependencies:** Displays subnet information and dependency analysis:
- Subnet details (VPC, CIDR, AZ, State)
- List of dependencies preventing deletion (if any)
- Success message if no dependencies found

**List NLBs:** Displays a formatted table with the following columns:
- Name (from Name tag)
- State
- Scheme (internal/external)
- AZ / Subnet (availability zone and subnet pairs, each on a separate line)
- Created Time
- Tags (relevant tags like kubernetes.io/role/elb, each on a separate line)

**Add Subnet to NLB:** Shows confirmation prompts and operation results:
- List of NLBs that will be modified
- List of subnets that will be added
- Confirmation prompt (unless --force is used)
- Success/failure messages for each NLB
- Summary of completed operations

**Check NLB Associations:** Shows detailed association analysis:
- NLB name, ARN, and current state
- Detection of listeners and target groups (indicators of service usage)
- Specific commands to check for Kubernetes and ECS associations
- Guidance on resolving common association issues

**Remove Subnet from NLB:** Shows confirmation prompts and operation results:
- List of NLBs that will be modified
- Confirmation prompt (unless --force is used)
- Success/failure messages for each NLB
- Summary of completed operations

**List ECR Images:** Displays a formatted table with the following columns:
- Repository (repository name)
- Tag (image tag, shows "<untagged>" for images without tags)
- Digest (image digest, truncated for display)
- Pushed At (push timestamp)
- Size (human-readable image size)
- Manifest (image manifest media type)

#### Examples

```bash
# Show help for all subnets commands
./aws subnets --help

# List all subnets in VPC, sorted by CIDR
./aws subnets --vpc vpc-0a1b2c3d4e5f6789

# List subnets in specific AZ, sorted by name
./aws subnets --vpc vpc-0a1b2c3d4e5f6789 --zone us-west-2a --sort name

# Check dependencies before deleting a subnet
./aws subnets check-dependencies --subnet-id subnet-0a87931be8d84c3df

# Delete subnet with confirmation
./aws subnets delete --subnet-id subnet-0a87931be8d84c3df

# Delete subnet without confirmation
./aws subnets delete --subnet-id subnet-0a87931be8d84c3df --force

# Show help for nlb commands
./aws nlb --help

# List all NLBs in VPC
./aws nlb --vpc vpc-0a1b2c3d4e5f6789

# List NLBs in specific AZ, sorted by state
./aws nlb --vpc vpc-0a1b2c3d4e5f6789 --zone us-west-2a --sort state

# List NLBs sorted by creation time
./aws nlb --vpc vpc-0a1b2c3d4e5f6789 --sort created

# Add subnets from another zone before removing subnets
./aws nlb add-subnet --vpc vpc-0a1b2c3d4e5f6789 --zone us-west-2b

# Add subnets to a specific NLB
./aws nlb add-subnet --vpc vpc-0a1b2c3d4e5f6789 --zone us-west-2b --nlb-name my-nlb

# Check for NLB associations before removing subnets
./aws nlb check-associations --vpc vpc-0a1b2c3d4e5f6789

# Check specific NLB for associations
./aws nlb check-associations --vpc vpc-0a1b2c3d4e5f6789 --nlb-name my-nlb

# Remove subnets from all NLBs in a zone
./aws nlb remove-subnet --vpc vpc-0a1b2c3d4e5f6789 --zone us-west-2a

# Remove subnets from a specific NLB
./aws nlb remove-subnet --vpc vpc-0a1b2c3d4e5f6789 --zone us-west-2a --nlb-name my-nlb

# Remove subnets without confirmation
./aws nlb remove-subnet --vpc vpc-0a1b2c3d4e5f6789 --zone us-west-2a --force

# Show help for ecr commands
./aws ecr --help

# List all images in a repository
./aws ecr --repository my-app

# List images with specific tag
./aws ecr --repository my-app --tag latest

# List images sorted by tag
./aws ecr --repository my-app --sort tag

# List images sorted by size
./aws ecr --repository my-app --sort size

# Default is sorted by push date (newest first)
```

## AWS Permissions

The tool requires the following AWS IAM permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeSubnets",
                "ec2:DescribeInstances",
                "ec2:DescribeNetworkInterfaces",
                "ec2:DescribeVpcEndpoints",
                "ec2:DeleteSubnet",
                "elasticloadbalancing:DescribeLoadBalancers",
                "elasticloadbalancing:DescribeTags",
                "elasticloadbalancing:SetSubnets",
                "ecr:DescribeImages",
                "ecr:DescribeRepositories"
            ],
            "Resource": "*"
        }
    ]
}
```

**Permission Details:**
- `ec2:DescribeSubnets` - List subnets and their properties
- `ec2:DescribeInstances` - Check for EC2 instances in subnets
- `ec2:DescribeNetworkInterfaces` - Check for network interfaces
- `ec2:DescribeVpcEndpoints` - Check for VPC endpoints
- `ec2:DeleteSubnet` - Delete subnets (only needed for delete operations)
- `elasticloadbalancing:DescribeLoadBalancers` - List load balancers and their properties
- `elasticloadbalancing:DescribeTags` - Get tags for load balancers
- `elasticloadbalancing:SetSubnets` - Modify NLB subnet configuration (only needed for remove-subnet operations)
- `ecr:DescribeImages` - List ECR images and their properties
- `ecr:DescribeRepositories` - List ECR repositories (only needed for ECR operations)

## Features

### Subnet Listing
- **CIDR-based sorting**: Intelligently sorts CIDR blocks by network address and prefix length
- **Zone filtering**: Filter subnets by availability zone
- **Flexible sorting**: Sort by CIDR, availability zone, name, or type
- **Tag support**: Displays subnet names, types, and relevant tags from AWS
- **Formatted output**: Uses go-pretty for clean, colored table output

### Subnet Deletion
- **Dependency checking**: Automatically checks for resources that prevent deletion
- **Safety features**: Confirmation prompts and force mode options
- **Comprehensive checks**: Validates EC2 instances, network interfaces, VPC endpoints, and load balancers
- **Clear error messages**: Provides specific guidance on what needs to be removed first

### Dependency Analysis
- **Pre-deletion validation**: Check what resources are blocking subnet deletion
- **Detailed reporting**: Shows specific resource IDs and types preventing deletion
- **Resource identification**: Identifies EC2 instances, ENIs, VPC endpoints, and load balancers

### NLB Listing
- **VPC filtering**: List only Network Load Balancers in a specific VPC
- **Zone filtering**: Filter NLBs by availability zone
- **Flexible sorting**: Sort by name, state, type, scheme, or creation time
- **Tag support**: Displays NLB names and relevant tags from AWS
- **Formatted output**: Uses go-pretty for clean, colored table output
- **Comprehensive info**: Shows subnets, availability zones, and more
- **Smart naming**: Falls back to Load Balancer ARN when Name tag is missing

### NLB Subnet Management
- **Zone-based removal**: Remove subnets from NLBs by availability zone
- **Selective targeting**: Target specific NLBs by name or all NLBs in a VPC
- **Safety checks**: Prevents removing all subnets from an NLB
- **Confirmation prompts**: Interactive confirmation before making changes
- **Force mode**: Skip confirmation for automated operations
- **Detailed reporting**: Shows which NLBs will be modified and operation results

### ECR Image Listing
- **Repository filtering**: List images from specific ECR repositories
- **Tag filtering**: Filter images by specific tags
- **Flexible sorting**: Sort by tag, push date, or image size
- **Human-readable output**: Formatted table with digest truncation and size formatting
- **Untagged image support**: Shows untagged images with special indicator

### General
- **Error handling**: Comprehensive error handling with helpful messages
- **Nested commands**: Intuitive command structure with sub-commands
- **Help system**: Detailed help for all commands and options

## Troubleshooting

### Common Issues

**"Subnet has dependencies and cannot be deleted"**
```bash
# Check what's preventing deletion
./aws subnets check-dependencies --subnet-id subnet-12345678

# Remove the dependencies (e.g., terminate instances, delete endpoints)
# Then retry deletion
./aws subnets delete --subnet-id subnet-12345678
```

**"Subnet not found"**
- Verify the subnet ID is correct
- Check that the subnet exists in your current AWS region
- Ensure you have the necessary permissions

**"No NLBs found in VPC"**
- Verify the VPC ID is correct
- Check that the VPC contains Network Load Balancers
- Ensure you have the necessary permissions

**"No NLBs found with subnets in zone"**
- Verify the availability zone is correct
- Check that NLBs in the VPC have subnets in the specified zone
- Use `aws nlb list --vpc VPC_ID --zone AZ` to see which NLBs have subnets in that zone

**"Cannot remove all subnets from NLB"**
- NLBs must have at least one subnet
- Remove subnets from other zones first, or add subnets to other zones before removing the last subnet

**"ResourceInUse: Subnets cannot be removed from load balancer because the load balancer is currently associated with another service"**
- The NLB is being used by another AWS service (Kubernetes, ECS, etc.)
- Check for Kubernetes service associations: `kubectl get services -o wide`
- Check for ECS service associations: `aws ecs describe-services --cluster CLUSTER_NAME`
- Delete or modify the associated service first
- Wait a few minutes for the association to be removed, then retry

**"Failed to load AWS config"**
- Configure AWS credentials using `aws configure`
- Set environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
- Use IAM roles if running on EC2

**"Repository not found"**
- Verify the repository name is correct
- Check that the repository exists in your current AWS region
- Ensure you have the necessary permissions

**"No images found in the repository"**
- Verify the repository name is correct
- Check that the repository contains images
- Ensure you have the necessary permissions

### Dependency Types

The tool checks for these dependency types:
- **EC2 Instances**: Running, pending, or stopping instances
- **Network Interfaces**: Attached Elastic Network Interfaces (ENIs)
- **VPC Endpoints**: Active VPC endpoints
- **Load Balancers**: Detected via ENI descriptions

### NLB Subnet Management Best Practices

**Before removing subnets:**
1. List NLBs to understand current configuration: `./aws nlb list --vpc VPC_ID`
2. Check which NLBs have subnets in the target zone: `./aws nlb list --vpc VPC_ID --zone AZ`
3. Check for service associations that might prevent removal:
   - Kubernetes: `kubectl get services -o wide | grep LOADBALANCER`
   - ECS: `aws ecs describe-services --cluster CLUSTER_NAME --services SERVICE_NAME`
4. Ensure NLBs will have remaining subnets in other zones after removal

**Safe removal workflow:**
```bash
# 1. Check current NLB configuration
./aws nlb list --vpc vpc-12345678

# 2. See which NLBs will be affected
./aws nlb list --vpc vpc-12345678 --zone us-east-1a

# 3. Remove subnets (with confirmation)
./aws nlb remove-subnet --vpc vpc-12345678 --zone us-east-1a

# 4. Verify changes
./aws nlb list --vpc vpc-12345678
```

## Quick Reference

### Subnet Commands
```bash
# List subnets
./aws subnets --vpc vpc-12345678
./aws subnets list --vpc vpc-12345678 --zone us-east-1a --sort name

# Delete subnet
./aws subnets delete --subnet-id subnet-12345678
./aws subnets delete --subnet-id subnet-12345678 --force

# Check dependencies
./aws subnets check-dependencies --subnet-id subnet-12345678
```

### NLB Commands
```bash
# List NLBs
./aws nlb --vpc vpc-12345678
./aws nlb list --vpc vpc-12345678 --zone us-east-1a --sort state

# Add subnets to NLBs
./aws nlb add-subnet --vpc vpc-12345678 --zone us-east-1b
./aws nlb add-subnet --vpc vpc-12345678 --zone us-east-1b --nlb-name my-nlb

# Check for associations
./aws nlb check-associations --vpc vpc-12345678
./aws nlb check-associations --vpc vpc-12345678 --nlb-name my-nlb

# Remove subnets from NLBs
./aws nlb remove-subnet --vpc vpc-12345678 --zone us-east-1a
./aws nlb remove-subnet --vpc vpc-12345678 --zone us-east-1a --nlb-name my-nlb --force
```

### ECR Commands
```bash
# List ECR images
./aws ecr --repository my-repo
./aws ecr list --repository my-repo --tag latest --sort pushed
```

### Help Commands
```bash
# General help
./aws --help

# Subnet help
./aws subnets --help
./aws subnets delete --help

# NLB help
./aws nlb --help
./aws nlb remove-subnet --help

# ECR help
./aws ecr --help
```
