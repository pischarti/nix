## Edge VPC (tgw) Terraform summary

### What this deploys
- **Providers**: AWS, TLS
- **VPC**: `aws_vpc.edge` with CIDR `var.vpc_cidr_space`, DNS support/hostnames enabled
- **Internet Gateway**: `aws_internet_gateway.edge`
- **Public Subnet**: `aws_subnet.edge_public` with `map_public_ip_on_launch = true`
- **Routing**: `aws_route_table.edge_public` + default route `0.0.0.0/0` via IGW and association to the public subnet
- **Security Group**: `aws_security_group.edge_public_ssh` allowing SSH from `var.edge_ssh_ingress_cidr`, HTTP (80) from anywhere, and all egress
- **EC2 Instance**: `aws_instance.edge_public` (Amazon Linux 2) in the public subnet
- **User data**: Installs and starts Apache (httpd) and serves `index.html` showing hostname, instance ID, and local IP
- **Key Pair**: Generated with `tls_private_key.edge` and `aws_key_pair.edge_generated` when `var.edge_key_name` is not provided

### Inputs (variables)
- **aws_region**: AWS region (default: `us-east-1`)
- **vpc_cidr_space**: VPC CIDR (default: `10.101.0.0/16`)
- **edge_vpc_name**: Name tag prefix (default: `edge-vpc`)
- **edge_public_subnet_cidr**: Public subnet CIDR (default: `10.101.101.0/24`)
- **edge_instance_type**: EC2 instance type (default: `t3.micro`)
- **edge_key_name**: Existing EC2 key pair name. If `null`, a new key is generated
- **edge_ssh_ingress_cidr**: CIDR allowed for SSH (default: `0.0.0.0/0`) — restrict for security
- **enable_dns_hostnames / enable_dns_support**: Enable DNS features in VPC (default: `true`)
- **tags**: Map of tags (default includes `environment = "tgw-nfw-demo"`)

### Outputs
- **edge_vpc_id**, **edge_vpc_arn**, **edge_vpc_cidr_block**
- **edge_igw_id**
- **edge_public_subnet_id**, **edge_public_route_table_id**
- **edge_public_instance_id**, **edge_public_instance_public_ip**
- **edge_generated_key_pair_name**, **edge_generated_public_key**, **edge_generated_private_key_pem** (sensitive)

### Quickstart
```bash
cd tf/tgw
terraform init
terraform apply -auto-approve

# If a key was generated (edge_key_name = null):
terraform output -raw edge_generated_private_key_pem > edge_generated.pem
chmod 600 edge_generated.pem

IP=$(terraform output -raw edge_public_instance_public_ip)
ssh -i edge_generated.pem ec2-user@$IP
```

Verify HTTP (Apache):
```bash
curl -s http://$(terraform output -raw edge_public_instance_public_ip)
```

### Notes
- The generated private key is stored in Terraform state. Handle the state file securely and rotate keys as needed.
- To use an existing key pair, set `edge_key_name` and keep `edge_ssh_ingress_cidr` restricted to your IP.

### Cleanup
```bash
terraform destroy -auto-approve
```

