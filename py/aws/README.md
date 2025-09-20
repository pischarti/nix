# AWS VPC Tools

A comprehensive Python toolkit for managing AWS VPCs using Nix flakes and uv for environment management.

## Tools

### 1. VPC Deletion Tool (`delete_vpc.py`)
- **Complete VPC deletion**: Handles all AWS resources that would block VPC deletion from the console
- **Intelligent dependency resolution**: Automatically deletes resources in the correct order and handles complex dependencies like Gateway Load Balancers
- **Rich CLI interface**: Beautiful terminal output with progress indicators
- **Safety features**: Confirmation prompts and dry-run mode
- **Comprehensive logging**: Detailed tracking of all deleted resources
- **Error handling**: Graceful error handling with detailed error messages

### 2. VPC Route Viewer (`vpc_route.py`)
- **Subnet analysis**: View all subnets in a VPC with detailed information
- **Route table inspection**: Examine routing configurations and associations
- **Smart subnet classification**: Automatically detects subnet types (Public, Private, Firewall, Inspection, etc.)
- **Flexible sorting**: Sort subnets by type or CIDR block
- **Rich formatting**: Beautiful table output with color-coded information
- **Route target resolution**: Shows friendly names for NAT gateways, internet gateways, and other routing targets

### 3. Transit Gateway Route Viewer (`tgw_route.py`)
- **VPC attachment analysis**: View all VPC attachments to the Transit Gateway
- **Route table inspection**: Examine Transit Gateway route tables and their configurations
- **Route analysis**: Display all routes with destinations, attachments, and states
- **Association tracking**: Shows which resources are associated with each route table
- **Propagation monitoring**: Displays route propagation configurations
- **Rich formatting**: Beautiful table output with comprehensive attachment details

## Supported AWS Resources

The tool automatically detects and deletes the following AWS resources:

- **Compute**: EC2 instances, Lambda functions
- **Networking**: Internet Gateways, NAT Gateways, VPC Endpoints, VPC Endpoint Service Configurations, Peering Connections
- **Load Balancing**: Application Load Balancers, Network Load Balancers, Gateway Load Balancers, Classic Load Balancers
- **VPN**: VPN Connections, VPN Gateways, Customer Gateways
- **Storage**: RDS Subnet Groups
- **Security**: Security Groups (custom), Network ACLs (custom)
- **Infrastructure**: Subnets, Route Tables (custom), Network Interfaces, Elastic IPs

## Prerequisites

1. **Nix with flakes enabled**
2. **AWS CLI configured** with appropriate permissions
3. **Valid AWS credentials** (via AWS CLI, environment variables, or IAM roles)

## Installation & Setup

1. **Enter the Nix development environment:**
   ```bash
   cd /Users/steve/dev/nix
   nix develop
   ```

2. **Navigate to the AWS directory:**
   ```bash
   cd py/aws
   ```

3. **Install Python dependencies with uv:**
   ```bash
   uv sync
   ```

4. **Configure AWS credentials (if not already done):**
   ```bash
   aws configure
   ```
   Or set environment variables:
   ```bash
   export AWS_ACCESS_KEY_ID="your-access-key"
   export AWS_SECRET_ACCESS_KEY="your-secret-key"
   export AWS_DEFAULT_REGION="us-east-1"
   ```

5. **Test your AWS setup:**
   ```bash
   uv run test_aws_setup.py
   # Or with specific region/profile via environment variables:
   export AWS_DEFAULT_REGION=us-west-2
   export AWS_PROFILE=myprofile
   uv run test_aws_setup.py
   ```

### Quick Start with Just

This project includes a `Justfile` for common tasks. Install [just](https://just.systems/) and run:

```bash
# Show available commands
just

# Test AWS setup
just test-setup

# Preview what would be deleted (dry run)
just dry-run vpc-12345678

# Delete VPC with confirmation
just delete vpc-12345678

# Delete VPC without confirmation
just force-delete vpc-12345678

# Show examples
just examples
```

## Usage

### VPC Route Viewer (`vpc_route.py`)

#### Basic Usage

```bash
# View all subnets and routes for a VPC (sorted by type by default)
uv run vpc_route.py vpc-12345678

# Sort by CIDR block instead of type
uv run vpc_route.py vpc-12345678 --sort cidr

# Sort by type explicitly
uv run vpc_route.py vpc-12345678 --sort type

# Output in YAML format for programmatic processing
uv run vpc_route.py vpc-12345678 --output-format yaml

# Combine sorting and output format options
uv run vpc_route.py vpc-12345678 --sort cidr --output-format yaml
```

#### Command Line Options

- `VPC_ID` (required): The ID of the VPC to analyze (e.g., vpc-12345678)
- `--sort`: Sort subnets by 'type' (default) or 'cidr' block
- `--output-format`: Output format - 'table' (default) or 'yaml'
- `--help`: Show help message

#### Example Output

**Table Format (default):**
- **VPC Summary**: Overview with CIDR block, subnet counts, and route table count
- **Subnets Table**: All subnets with ID, CIDR block, AZ, type, state, and associated route table
- **Route Tables**: Detailed routing information with destinations, targets, and target types

**YAML Format:**
Structured data including:
- VPC information (ID, CIDR block, state)
- Summary statistics (subnet counts by type, route table count)
- Complete subnet details with tags
- Route table configurations with all routes
- NAT gateway and internet gateway information

### Transit Gateway Route Viewer (`tgw_route.py`)

#### Basic Usage

```bash
# View all VPC attachments and routes for a Transit Gateway
uv run tgw_route.py tgw-12345678

# Output in YAML format for programmatic processing
uv run tgw_route.py tgw-12345678 --output-format yaml
```

#### Command Line Options

- `TGW_ID` (required): The ID of the Transit Gateway to analyze (e.g., tgw-12345678)
- `--output-format`: Output format - 'table' (default) or 'yaml'
- `--help`: Show help message

#### Example Output

**Table Format (default):**
- **TGW Summary**: Overview with state, owner, attachment counts, and route table count
- **VPC Attachments**: All attachments with VPC ID, subnet IDs, state, and DNS/IPv6 support
- **Route Tables**: Each route table with associations, propagations, and routes
- **Routes**: Detailed routing information with destinations, types, states, and attachments

**YAML Format:**
Structured data including:
- Transit Gateway information (ID, state, owner, description, ARN)
- Summary statistics (attachment counts by state, route table count, total routes)
- Complete attachment details with configuration options
- Route table configurations with associations, propagations, and all routes

### VPC Deletion Tool (`delete_vpc.py`)

#### Basic Usage

```bash
# Delete a VPC with confirmation prompt
uv run delete_vpc.py vpc-12345678

# Or run directly with Python
python delete_vpc.py vpc-12345678
```

### Advanced Usage

```bash
# Set AWS region and profile via environment variables
export AWS_DEFAULT_REGION=us-west-2
export AWS_PROFILE=my-profile
uv run delete_vpc.py vpc-12345678

# Dry run mode (shows what would be deleted without actually deleting)
uv run delete_vpc.py vpc-12345678 --dry-run

# Force deletion without confirmation prompt
uv run delete_vpc.py vpc-12345678 --force

# Combine options with environment variables
export AWS_DEFAULT_REGION=us-east-1
export AWS_PROFILE=prod
uv run delete_vpc.py vpc-12345678 --dry-run
```

### Command Line Options

- `VPC_ID` (required): The ID of the VPC to delete (e.g., vpc-12345678)
- `--dry-run`: Show what would be deleted without actually deleting anything
- `--force`: Skip the confirmation prompt
- `--help`: Show help message

### Environment Variables

The tool uses the following environment variables for AWS configuration:

- `AWS_DEFAULT_REGION` or `AWS_REGION`: AWS region to use
- `AWS_PROFILE`: AWS profile to use (optional, defaults to default profile)
- `AWS_ACCESS_KEY_ID`: AWS access key (if not using profiles)
- `AWS_SECRET_ACCESS_KEY`: AWS secret key (if not using profiles)

## AWS Permissions Required

The tool requires the following AWS permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:*",
                "elasticloadbalancing:*",
                "rds:DescribeDBSubnetGroups",
                "rds:DeleteDBSubnetGroup",
                "lambda:ListFunctions",
                "lambda:GetFunctionConfiguration",
                "lambda:DeleteFunction"
            ],
            "Resource": "*"
        }
    ]
}
```

## Safety Features

1. **Configuration Validation**: Validates AWS credentials and region before starting
2. **VPC Verification**: Confirms the VPC exists before starting deletion
3. **Confirmation Prompt**: Asks for explicit confirmation before deletion (unless `--force` is used)
4. **Dry Run Mode**: Use `--dry-run` to see what would be deleted without making changes
5. **Default VPC Warning**: Shows a warning if you're trying to delete a default VPC
6. **Error Handling**: Continues with other resources even if some deletions fail
7. **Helpful Error Messages**: Provides clear guidance when configuration is missing
8. **Batch Processing**: Efficiently handles bulk operations like VPC endpoint deletion
9. **Automated Dependency Resolution**: Automatically handles complex dependencies like Gateway Load Balancer cleanup

## Example Output

```
AWS VPC Deletion Tool
VPC ID: vpc-12345678
Region: us-west-2
Profile: default

Found VPC: vpc-12345678
  CIDR: 10.0.0.0/16
  State: available
  Default: False

Starting VPC deletion process...

Step: EC2 Instances
✓ Terminated 2 instances

Step: Load Balancers
✓ Deleted 1 ALB/NLB load balancers

Step: NAT Gateways
✓ Deleted 2 NAT gateways

...

✓ Successfully deleted VPC: vpc-12345678

============================================================
                        Deletion Summary                        
┏━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ Resource Type        ┃ Count ┃ IDs                             ┃
┡━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┩
│ Instances            │     2 │ i-1234567890abcdef0, i-fedcba   │
│ Load Balancers       │     1 │ my-load-balancer                │
│ Nat Gateways         │     2 │ nat-12345, nat-67890            │
│ Internet Gateways    │     1 │ igw-12345678                    │
│ Subnets              │     4 │ subnet-123, subnet-456, subnet │
│ Security Groups      │     3 │ sg-123, sg-456, sg-789          │
└──────────────────────┴───────┴─────────────────────────────────┘

✅ VPC deletion completed successfully!
```

## Troubleshooting

### Common Issues

1. **Permission Denied**: Ensure your AWS credentials have the required permissions
2. **VPC Not Found**: Verify the VPC ID is correct and exists in the specified region
3. **Dependencies Still Exist**: Some resources might take time to delete; retry after a few minutes
4. **Rate Limiting**: The tool includes built-in delays, but you might need to wait and retry for heavily loaded accounts

### Getting Help

```bash
uv run delete_vpc.py --help
```

## Development

### Project Structure

```
py/aws/
├── delete_vpc.py       # VPC deletion script
├── vpc_route.py        # VPC route and subnet viewer
├── tgw_route.py        # Transit Gateway route and attachment viewer
├── test_aws_setup.py   # AWS setup validation script
├── example_usage.sh    # Example usage script
├── Justfile           # Task runner with common commands
├── pyproject.toml      # Project configuration and dependencies
├── uv.lock            # Locked dependency versions
└── README.md          # This file
```

### Running Tests

```bash
# Install development dependencies
uv sync --group dev

# Run tests (when available)
uv run pytest
```

### Code Quality

```bash
# Format code
uv run black delete_vpc.py vpc_route.py tgw_route.py

# Lint code
uv run ruff check delete_vpc.py vpc_route.py tgw_route.py
```

## Security Considerations

- **Irreversible Operation**: VPC deletion cannot be undone
- **Credential Security**: Ensure AWS credentials are properly secured
- **Production Safety**: Always use `--dry-run` first in production environments
- **Access Logging**: Consider enabling CloudTrail for audit purposes

## License

This project is licensed under the same license as the parent repository.
