variable "region" {
    type = string
    default = "us-east-1"
}

variable "tags" {
    type = map(string)
    default = {
        Environment = "poc"
        Project = "EKS Automode PoC"
    }
}

variable "node_group_ami_type" {
  description = "Type of Amazon Machine Image (AMI) associated with the EKS Node Group. Valid values: AL2_x86_64, AL2_x86_64_GPU, AL2_ARM_64, CUSTOM, BOTTLEROCKET_ARM_64, BOTTLEROCKET_x86_64, BOTTLEROCKET_ARM_64_NVIDIA, BOTTLEROCKET_x86_64_NVIDIA"
  type        = string
  default     = "AL2_x86_64"
}

variable "node_group_ami_id" {
  description = "The AMI ID to use for the node group. If not specified, will use the latest EKS-optimized AMI for the specified AMI type"
  type        = string
  default     = null
}

variable "enable_managed_node_groups" {
  description = "Enable managed node groups alongside EKS AutoMode"
  type        = bool
  default     = false
}

variable "ami_owners" {
  description = "The AWS account IDs that own the AMIs"
  type        = list(string)
  default     = []
}

variable "ami_name_filters" {
  description = "The name filters for the AMIs"
  type        = list(string)
  default     = []
}
