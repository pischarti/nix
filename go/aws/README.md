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

#### Options

- `--vpc VPC_ID` (required): VPC ID to list subnets for
- `--zone AZ` (optional): Filter by availability zone (e.g., us-east-1a)
- `--sort SORT_BY` (optional): Sort by one of:
  - `cidr` (default): Sort by CIDR block in network order
  - `az`: Sort by availability zone
  - `name`: Sort by subnet name (from Name tag)
  - `type`: Sort by subnet type (from Type tag)

#### Output

The command displays a formatted table with the following columns:
- Subnet ID
- VPC ID
- CIDR Block
- AZ (Availability Zone)
- Name (from Name tag)
- State
- Type (from Type tag, defaults to "subnet")

#### Examples

```bash
# Show help
./aws subnets --help

# List all subnets in VPC, sorted by CIDR
./aws subnets --vpc vpc-0a1b2c3d4e5f6789

# List subnets in specific AZ, sorted by name
./aws subnets --vpc vpc-0a1b2c3d4e5f6789 --zone us-west-2a --sort name
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
                "ec2:DescribeSubnets"
            ],
            "Resource": "*"
        }
    ]
}
```

## Features

- **CIDR-based sorting**: Intelligently sorts CIDR blocks by network address and prefix length
- **Zone filtering**: Filter subnets by availability zone
- **Flexible sorting**: Sort by CIDR, availability zone, name, or type
- **Tag support**: Displays subnet names and types from AWS tags
- **Formatted output**: Uses go-pretty for clean, colored table output
- **Error handling**: Comprehensive error handling with helpful messages
