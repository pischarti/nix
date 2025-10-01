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
