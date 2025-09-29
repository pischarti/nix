# Network Firewall Single VPC Architecture

This directory contains AWS Diagram as Code (DAC) files for visualizing a single VPC architecture with Network Firewall.

## Files

- `nfw_single_vpc.yml` - YAML definition for the VPC architecture diagram
- `nfw-single-vpc.png` - Generated diagram output

## Architecture Overview

The diagram shows a multi-AZ VPC architecture with:

- **Two Availability Zones**: us-west-1a and us-west-2b
- **Database Tier**: RDS instances in private subnets
- **Application Tier**: EKS clusters in private subnets  
- **Network Firewall Tier**: AWS Network Firewall in private subnets
- **Public Tier**: Network Load Balancers in public subnets
- **Internet Gateway**: For external connectivity

## How to Generate PNG

### Prerequisites

Install the AWS Diagram as Code tool:

```bash
go install github.com/awslabs/diagram-as-code/cmd/awsdac@latest
```

### Generate Diagram

To generate the PNG diagram from the YAML file:

```bash
awsdac nfw_single_vpc.yml -t
```

This will create `nfw_single_vpc.png` in the current directory.

### Alternative Output Formats

You can also specify different output formats:

```bash
# Generate SVG with template
awsdac nfw_single_vpc.yml -t -o nfw_single_vpc.svg

# Generate with custom output name
awsdac nfw_single_vpc.yml -t -o my_diagram.png
```

## Architecture Flow

1. **User** → **Internet Gateway** → **Network Load Balancers**
2. **NLB** → **Network Firewall** (traffic inspection)
3. **Network Firewall** → **EKS Clusters** (filtered traffic)
4. **EKS Clusters** → **RDS Databases** (application data)

## Multi-AZ Design

Each Availability Zone contains a complete stack:
- Database subnet with RDS
- Application subnet with EKS
- Network Firewall subnet
- Public subnet with NLB

This provides high availability and fault tolerance across both AZs.
