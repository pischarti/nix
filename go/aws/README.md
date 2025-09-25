# AWS CLI Tool

A Go-based command-line tool for interacting with AWS services using the gofr framework.

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

Manage AWS Network Load Balancers with functionality for listing NLBs in a VPC.

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
                "elasticloadbalancing:DescribeTags"
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
- **Comprehensive info**: Shows DNS names, subnets, availability zones, and more

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

**"Failed to load AWS config"**
- Configure AWS credentials using `aws configure`
- Set environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
- Use IAM roles if running on EC2

### Dependency Types

The tool checks for these dependency types:
- **EC2 Instances**: Running, pending, or stopping instances
- **Network Interfaces**: Attached Elastic Network Interfaces (ENIs)
- **VPC Endpoints**: Active VPC endpoints
- **Load Balancers**: Detected via ENI descriptions
