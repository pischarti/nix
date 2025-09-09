
variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

variable "vpc_cidr_space" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.101.0.0/16"
}

variable "edge_vpc_name" {
  description = "Name tag for the EdgeVPC"
  type        = string
  default     = "edge-vpc"
}

variable "enable_dns_hostnames" {
  description = "Enable DNS hostnames in the VPC"
  type        = bool
  default     = true
}

variable "enable_dns_support" {
  description = "Enable DNS support in the VPC"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Additional tags to apply to resources"
  type        = map(string)
  default     = {
    "environment" = "tgw-nfw-demo"
  }
}

variable "edge_public_subnet_cidr" {
  description = "CIDR block for the public subnet in the edge VPC"
  type        = string
  default     = "10.101.101.0/24"
}

variable "edge_instance_type" {
  description = "Instance type for the edge EC2 instance"
  type        = string
  default     = "t3.micro"
}

variable "edge_key_name" {
  description = "Name of an existing EC2 Key Pair to enable SSH access"
  type        = string
  default     = null
}

variable "edge_ssh_ingress_cidr" {
  description = "CIDR block allowed to SSH into the edge instance"
  type        = string
  default     = "0.0.0.0/0"
}
