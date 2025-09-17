

variable "aws_region" {
  description = "The AWS region to deploy the routing table"
  type        = string
  default     = "us-east-1"
}

variable "tags" {
  description = "The tags to deploy the routing table"
  type        = map(string)
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
  default     = ["10.0.10.0/24", "10.0.20.0/24"]
}

# Network Firewall VPC Variables
variable "firewall_vpc_cidr" {
  description = "CIDR block for the Network Firewall VPC"
  type        = string
  default     = "10.1.0.0/16"
}

variable "firewall_vpc_name" {
  description = "Name for the Network Firewall VPC"
  type        = string
  default     = "firewall-vpc"
}

variable "firewall_subnet_cidrs" {
  description = "CIDR blocks for Network Firewall subnets"
  type        = list(string)
  default     = ["10.1.1.0/24", "10.1.2.0/24"]
}

variable "firewall_tgw_subnet_cidrs" {
  description = "CIDR blocks for Transit Gateway subnets in firewall VPC"
  type        = list(string)
  default     = ["10.1.10.0/24", "10.1.20.0/24"]
}

# Transit Gateway Variables
variable "tgw_name" {
  description = "Name for the Transit Gateway"
  type        = string
  default     = "main-tgw"
}

variable "tgw_description" {
  description = "Description for the Transit Gateway"
  type        = string
  default     = "Transit Gateway for routing traffic through Network Firewall"
}

variable "main_vpc_tgw_subnet_cidrs" {
  description = "CIDR blocks for Transit Gateway subnets in main VPC"
  type        = list(string)
  default     = ["10.0.100.0/24", "10.0.200.0/24"]
} 
