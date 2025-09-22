variable "aws_region" {
    description = "The AWS region to deploy the TGW Spokes to"
    type        = string
    default     = "us-east-1"
}

variable "super_cidr_block" {
  type    = string
  default = "10.0.0.0/8"
}

