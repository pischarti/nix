
variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

variable "vpc_cidr_edge" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.101.0.0/16"
}

variable "vpc_cidr_inspection" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "11.101.0.0/16"
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

variable "tgw_instance_type" {
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

variable "inspection_vpc_name" {
  description = "Name tag for the InspectionVPC"
  type        = string
  default     = "inspection-vpc"
}

variable "inspection_public_subnet_cidr" {
  description = "CIDR block for the public subnet in the inspection VPC"
  type        = string
  default     = "11.101.102.0/24"
}

variable "inspection_key_name" {
  description = "Name of an existing EC2 Key Pair to enable SSH access"
  type        = string
  default     = null
}

variable "vpc_cidr_app" {
  description = "CIDR block for the App VPC"
  type        = string
  default     = "12.101.0.0/16"
}

variable "app_vpc_name" {
  description = "Name tag for the AppVPC"
  type        = string
  default     = "app-vpc"
}

variable "app_public_subnet_cidr" {
  description = "CIDR block for the public subnet in the app VPC"
  type        = string
  default     = "12.101.102.0/24"
}

variable "app_key_name" {
  description = "Name of an existing EC2 Key Pair to enable SSH access"
  type        = string
  default     = null
}

# Network Firewall Variables
variable "firewall_subnet_1_cidr" {
  description = "CIDR block for the first firewall subnet in the inspection VPC"
  type        = string
  default     = "11.101.1.0/24"
}

variable "firewall_subnet_2_cidr" {
  description = "CIDR block for the second firewall subnet in the inspection VPC"
  type        = string
  default     = "11.101.2.0/24"
}

variable "firewall_subnet_3_cidr" {
  description = "CIDR block for the third firewall subnet in the inspection VPC"
  type        = string
  default     = "11.101.3.0/24"
}
