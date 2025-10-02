terraform {
  required_version = ">= 1.4.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.95.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.9, < 3.0"
    }
    kubectl = {
      source  = "alekc/kubectl"
      version = ">= 2.1"
    }
  }
  backend "s3" {
    bucket = "terraform-state-aws-poc-us-east-1"
    key    = "eks/automode-demo/terraform.tfstate"
    region = "us-east-1"
  }
}

locals {
  name   = "automode-poc"
  region = "us-east-1"

  cluster_version = "1.33"

  vpc_cidr = "10.0.0.0/16"
  azs      = slice(data.aws_availability_zones.available.names, 0, 3)

  tags = merge(var.tags, {
    Blueprint = local.name
  })
}

provider "aws" {
  region = local.region
}

data "aws_availability_zones" "available" {
  # only include zones in us-east-1 a,b,c
  filter {
    name   = "region-name"
    values = ["us-east-1"]
  }
  filter {
    name   = "zone-name"
    values = ["us-east-1a", "us-east-1b", "us-east-1c"]
  }
  filter {
    name   = "zone-type"
    values = ["availability-zone"]
  }
}

# Data source to get the latest EKS-optimized AMI
data "aws_ami" "eks_optimized" {
  count = var.enable_managed_node_groups && var.node_group_ami_id == null ? 1 : 0

  most_recent = true
  owners      = var.ami_owners

  filter {
    name   = "name"
    values = var.ami_name_filters
  }
}

###############################################################
# EKS Cluster
###############################################################

# Example usage for AMI specification:
# 
# 1. Use default EKS-optimized AMI (AL2_x86_64):
#    terraform apply -var="enable_managed_node_groups=true"
#
# 2. Use custom AMI ID:
#    terraform apply -var="enable_managed_node_groups=true" -var="node_group_ami_id=ami-1234567890abcdef0"
#
# 3. Use different AMI type (ARM64):
#    terraform apply -var="enable_managed_node_groups=true" -var="node_group_ami_type=AL2_ARM_64"
#
# 4. Use Bottlerocket OS:
#    terraform apply -var="enable_managed_node_groups=true" -var="node_group_ami_type=BOTTLEROCKET_x86_64"

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.34"

  cluster_name    = local.name
  cluster_version = local.cluster_version

  # if true, Your cluster API server is accessible from the internet. You can, optionally, limit the CIDR blocks that can access the public endpoint.
  #WARNING: Avoid using this option (cluster_endpoint_public_access = true) in preprod or prod accounts. This feature is designed for sandbox accounts, simplifying cluster deployment and testing.
  # Alternatively, create a bastion host in the same VPC as the cluster to access the cluster API server over a private connection
  cluster_endpoint_public_access = true

  vpc_id = module.vpc.vpc_id

  subnet_ids = module.vpc.private_subnets

  enable_cluster_creator_admin_permissions = true

  # Enable EKS AutoMode
  cluster_compute_config = {
    enabled    = true
    node_pools = []
  }

  # Node groups configuration
  eks_managed_node_groups = var.enable_managed_node_groups ? {
    general = {
      name = "general"

      # Use custom AMI ID if provided, otherwise use latest EKS-optimized AMI
      ami_id = var.node_group_ami_id != null ? var.node_group_ami_id : (var.enable_managed_node_groups && var.node_group_ami_id == null ? data.aws_ami.eks_optimized[0].id : null)
      ami_type = var.node_group_ami_type

      instance_types = ["t3.medium"]

      min_size     = 1
      max_size     = 3
      desired_size = 2

      vpc_security_group_ids = []
      subnet_ids             = module.vpc.private_subnets

      # Launch template configuration
      launch_template_name        = "${local.name}-general"
      launch_template_description = "EKS managed node group launch template for general purpose nodes"
      launch_template_version     = "$Latest"

      # IAM role for node group
      iam_role_name = "${local.name}-eks-managed-node-group-general"

      # Node group labels
      labels = {
        role = "general"
      }

      # Taints
      taints = []

      # Update configuration
      update_config = {
        max_unavailable_percentage = 50
      }

      # Tags
      tags = merge(var.tags, local.tags, {
        Name = "${local.name}-general"
      })
    }
  } : {}

  access_entries = {
    # One access entry with a policy associated
    custom_nodeclass_access = {
      principal_arn = aws_iam_role.custom_nodeclass_role.arn
      type          = "EC2"

      policy_associations = {
        auto = {
          policy_arn = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSAutoNodePolicy"
          access_scope = {
            type = "cluster"
          }
        }
      }
    }
  }

  tags = merge(var.tags, local.tags)
}

###############################################################
# Creating IAM Role for custom nodeclass nodes
###############################################################

# Create nodeclass role and associate with IAM policies
resource "aws_iam_role" "custom_nodeclass_role" {
  name = "${local.name}-AmazonEKSAutoNodeRole"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })

  tags = merge(var.tags, local.tags)
}

# Attach AmazonEKSWorkerNodeMinimalPolicy
resource "aws_iam_role_policy_attachment" "eks_worker_node_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodeMinimalPolicy"
  role       = aws_iam_role.custom_nodeclass_role.name
}

# Attach AmazonEC2ContainerRegistryPullOnly
resource "aws_iam_role_policy_attachment" "ecr_pull_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPullOnly"
  role       = aws_iam_role.custom_nodeclass_role.name
}

###############################################################
# Supporting Resources
###############################################################
locals {
  # EKS subnets: 80% allocation using /18 (16,384 addresses per subnet)
  # Total for 3 AZs: 49,152 addresses (75% of VPC)
  eks_subnets = [for k, v in local.azs : cidrsubnet(local.vpc_cidr, 2, k)]
  
  # RDS subnets: 10% allocation using /22 (1,024 addresses per subnet) 
  # Total for 3 AZs: 3,072 addresses (4.7% of VPC)
  # Use explicit ranges to avoid overlaps: 10.0.192.0/22, 10.0.196.0/22, 10.0.200.0/22
  rds_subnets = [
    "10.0.192.0/22",   # us-east-1a: 10.0.192.0 - 10.0.195.255 (1,024 addresses)
    "10.0.196.0/22",   # us-east-1b: 10.0.196.0 - 10.0.199.255 (1,024 addresses)
    "10.0.200.0/22"    # us-east-1c: 10.0.200.0 - 10.0.203.255 (1,024 addresses)
  ]
  
  # Firewall subnets: Minimal allocation using /28 (16 addresses per subnet)
  # Total for 3 AZs: 48 addresses (0.07% of VPC)
  # Use explicit ranges to avoid overlaps: 10.0.216.0/28, 10.0.216.16/28, 10.0.216.32/28
  firewall_subnets = [
    "10.0.216.0/28",   # us-east-1a: 10.0.216.0 - 10.0.216.15 (16 addresses)
    "10.0.216.16/28",  # us-east-1b: 10.0.216.16 - 10.0.216.31 (16 addresses)
    "10.0.216.32/28"   # us-east-1c: 10.0.216.32 - 10.0.216.47 (16 addresses)
  ]
  
  # Public subnets: 10% allocation using /22 (1,024 addresses per subnet)
  # Total for 3 AZs: 3,072 addresses (4.7% of VPC)  
  # Use explicit ranges to avoid overlaps: 10.0.204.0/22, 10.0.208.0/22, 10.0.212.0/22
  public_subnets = [
    "10.0.204.0/22",   # us-east-1a: 10.0.204.0 - 10.0.207.255 (1,024 addresses)
    "10.0.208.0/22",   # us-east-1b: 10.0.208.0 - 10.0.211.255 (1,024 addresses)
    "10.0.212.0/22"    # us-east-1c: 10.0.212.0 - 10.0.215.255 (1,024 addresses)
  ]
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.19"

  name = local.name
  cidr = local.vpc_cidr

  azs             = local.azs
  private_subnets = local.eks_subnets
  public_subnets  = local.public_subnets

  enable_nat_gateway = true
  single_nat_gateway = true

  public_subnet_tags = { "kubernetes.io/role/elb" = 1 }

  private_subnet_tags = { 
    "kubernetes.io/role/internal-elb" = 1
    "kubernetes.io/cluster/${local.name}" = "shared"
  }

  tags = merge(var.tags, local.tags)
}

###############################################################
# Additional Subnet Groups
###############################################################

# RDS Subnets
resource "aws_subnet" "rds" {
  count = length(local.rds_subnets)

  vpc_id            = module.vpc.vpc_id
  cidr_block        = local.rds_subnets[count.index]
  availability_zone = local.azs[count.index]

  tags = merge(var.tags, local.tags, {
    Name = "${local.name}-rds-${local.azs[count.index]}"
    Type = "RDS"
  })
}

# Firewall Subnets
resource "aws_subnet" "firewall" {
  count = length(local.firewall_subnets)

  vpc_id            = module.vpc.vpc_id
  cidr_block        = local.firewall_subnets[count.index]
  availability_zone = local.azs[count.index]

  tags = merge(var.tags, local.tags, {
    Name = "${local.name}-firewall-${local.azs[count.index]}"
    Type = "Firewall"
  })
}


###############################################################
# Outputs
###############################################################

output "configure_kubectl" {
  description = "Configure kubectl: make sure you're logged in with the correct AWS profile and run the following command to update your kubeconfig"
  value       = "aws eks --region ${local.region} update-kubeconfig --name ${module.eks.cluster_name}"
}

output "node_group_ami_id" {
  description = "AMI ID used for the managed node group"
  value       = var.enable_managed_node_groups ? (var.node_group_ami_id != null ? var.node_group_ami_id : (var.node_group_ami_id == null ? data.aws_ami.eks_optimized[0].id : null)) : null
}

output "node_group_ami_type" {
  description = "AMI type used for the managed node group"
  value       = var.enable_managed_node_groups ? var.node_group_ami_type : null
}

output "managed_node_groups" {
  description = "EKS managed node groups"
  value       = var.enable_managed_node_groups ? module.eks.eks_managed_node_groups : {}
}
