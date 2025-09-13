aws_region  = "us-east-1"
# az_a        = "us-east-1a"
team        = "devsecops"
env         = "dev"

#Inspection VPC
vpc_cidr_inspection                 = "11.3.0.0/16"
inspection_vpc_tgw_subnet_cidr      = "11.3.144.0/20"
inspection_vpc_firewall_subnet_cidr = "11.3.128.0/20"

#App VPC
app_vpc_cidr                             = "12.101.0.0/16"
# app_vpc_cidr                             = "10.1.0.0/16"
app_vpc_tgw_subnet_cidr                  = "12.101.128.0/20"
# app_vpc_tgw_subnet_cidr                  = "10.1.128.0/20"
app_vpc_application_workload_subnet_cidr = "12.101.144.0/20"
# app_vpc_application_workload_subnet_cidr = "10.1.144.0/20"

#App VPC
egress_vpc_cidr            = "10.2.0.0/16"
egress_vpc_tgw_subnet_cidr = "10.2.128.0/20"
egress_vpc_igw_subnet_cidr = "10.2.144.0/20"

#SSH Key -
# ssh_key = ""
