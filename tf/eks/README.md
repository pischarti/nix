# EKS AutoMode with Custom NodePools

This Terraform configuration deploys an Amazon EKS AutoMode cluster with custom NodeClass and NodePool objects, providing fine-grained control over compute resources for different workload types.

## Overview

EKS AutoMode is a simplified way to run Kubernetes on AWS where AWS manages the control plane, data plane, and Kubernetes add-ons. This setup extends the default AutoMode configuration with custom NodePools to support specific compute requirements like different architectures (amd64, arm64) and instance types.

## Architecture

- **EKS AutoMode Cluster**: Managed control plane with version 1.33
- **VPC**: Custom VPC with public/private subnets across 2 AZs (us-east-1a, us-east-1b)
- **Custom NodePools**: 
  - AMD64 instances (c, r, m families)
  - Graviton/ARM64 instances (c, r, m families)
- **NodeClasses**: 
  - Basic configuration (default EBS settings)
  - EBS-optimized configuration (enhanced IOPS and throughput)
- **Storage**: EBS CSI driver (AWS-managed)
- **Load Balancing**: AWS Load Balancer Controller (AWS-managed)

## Prerequisites

- AWS CLI configured with appropriate permissions
- Terraform >= 1.4.0
- kubectl
- AWS IAM permissions for EKS, VPC, and EC2 resources

## Configuration

### Variables

The configuration supports the following variables (defined in `variables.tf`):

- `region`: AWS region (default: "us-east-1")
- `tags`: Additional tags for resources (default: Environment="poc", Project="EKS Automode PoC")

### Key Features

1. **EKS AutoMode Cluster**
   - Kubernetes version: 1.33
   - Public endpoint access enabled (for demo purposes)
   - Custom IAM role for NodeClass instances

2. **Custom NodePools**
   - `nodepool-amd64.yaml`: x86_64 architecture instances
   - `nodepool-graviton.yaml`: ARM64/Graviton instances
   - Both support c, r, m instance families

3. **Custom NodeClasses**
   - `nodeclass-basic.yaml`: Standard EBS configuration
   - `nodeclass-ebs-optimized.yaml`: Enhanced EBS with optimized IOPS and throughput

4. **AWS-Managed Add-ons**
   - EBS CSI Driver (for persistent storage)
   - AWS Load Balancer Controller (for ingress)

## Deployment

1. **Initialize Terraform**
   ```bash
   terraform init
   ```

2. **Plan the deployment**
   ```bash
   terraform plan
   ```

3. **Apply the configuration**
   ```bash
   terraform apply
   ```

4. **Configure kubectl**
   ```bash
   aws eks --region us-east-1 update-kubeconfig --name <cluster-name>
   ```

## Post-Deployment Configuration

After deployment, the following Kubernetes resources are automatically created:

### Storage Classes
- `auto-ebs-sc`: Default EBS storage class for AutoMode

### Ingress Classes
- `alb`: AWS Load Balancer Controller ingress class
- `alb-params`: Ingress class parameters for ALB configuration

### NodeClasses
- `basic`: Standard node configuration
- `ebs-optimized`: Enhanced EBS configuration

### NodePools
- `amd64`: x86_64 architecture nodes
- `graviton`: ARM64 architecture nodes

## Testing the Setup

You can validate the deployment by checking the cluster status:

```bash
# Check cluster nodes
kubectl get nodes

# Check available node classes
kubectl get nodeclass

# Check available node pools
kubectl get nodepool

# Check storage classes
kubectl get storageclass

# Check ingress classes
kubectl get ingressclass
```

## Customizing NodePools

To add new NodePools or modify existing ones:

1. Create or modify YAML files in the `eks-automode-config` directory
2. Update the `local` variables in `eks-automode-config.tf` to include new file names
3. Apply changes with `terraform apply`

## Security Considerations

⚠️ **Important**: This configuration enables public endpoint access for the EKS cluster API server, which is suitable for development and testing but **NOT recommended for production environments**.

For production deployments:
- Set `cluster_endpoint_public_access = false`
- Use a bastion host or VPN for cluster access
- Implement proper network segmentation

## Cleanup

To destroy the infrastructure:

1. Remove any applications deployed to the cluster
2. Run terraform destroy:
   ```bash
   terraform destroy
   ```

## State Management

This configuration uses an S3 backend for Terraform state:
- Bucket: `terraform-state-aws-poc-us-east-1`
- Key: `eks/automode-demo/terraform.tfstate`
- Region: `us-east-1`

## Troubleshooting

### Common Issues

1. **NodePool not provisioning nodes**
   - Check NodeClass IAM role permissions
   - Verify NodePool configuration in YAML files
   - Ensure sufficient EC2 capacity in target AZs

2. **Storage issues**
   - Verify EBS CSI driver is running
   - Check storage class configuration
   - Ensure proper IAM permissions for EBS operations

3. **Ingress not working**
   - Verify AWS Load Balancer Controller is installed
   - Check ingress class configuration
   - Ensure proper security group rules

### Useful Commands

```bash
# Check cluster status
kubectl get nodes -o wide

# View NodePool status
kubectl describe nodepool <nodepool-name>

# Check pod scheduling
kubectl get pods -o wide

# View events
kubectl get events --sort-by=.metadata.creationTimestamp
```

## References

- [EKS AutoMode Documentation](https://docs.aws.amazon.com/eks/latest/userguide/eks-automode.html)
- [AWS EKS Blueprints](https://github.com/aws-ia/terraform-aws-eks-blueprints)
- [EKS NodePool Documentation](https://docs.aws.amazon.com/eks/latest/userguide/eks-node-pool.html)
