# Nix Development Environment

This repository contains various development projects and tools organized by language and technology.

## Projects

### Go Projects

#### AWS CLI (`go/aws/`)
A command-line tool for managing AWS resources, particularly focused on VPC subnets and Network Load Balancers.

**Features:**
- List, delete, and check dependencies for AWS subnets
- Manage Network Load Balancers and their subnet associations
- Built with the GoFr framework for clean CLI structure

**Usage:**
```bash
# List subnets in a VPC
gaws subnets --vpc vpc-12345678

# Delete a subnet
gaws subnets delete --subnet-id subnet-12345678

# List Network Load Balancers
gaws nlb --vpc vpc-12345678

# Add subnets to NLBs
gaws nlb add-subnet --vpc vpc-12345678 --zone us-east-1b
```

#### Kubernetes CLI (`go/kube/`)
Kubernetes management tools and utilities.

#### Flappy Bird Game (`go/glappy/`)
A Go implementation of the classic Flappy Bird game using Ebiten.

### Python Projects

#### AWS Tools (`py/aws/`)
Python-based AWS automation scripts for VPC management.

#### FastAPI Demo (`py/fastapi/`)
Simple FastAPI application with testing.

#### Flappy Bird (`py/flappy/`)
Python implementation of Flappy Bird using Pygame.

#### Data Analysis (`py/panda/`, `py/sql/`, `py/sqlalchemy/`)
Various data analysis and SQL projects using pandas, SQLAlchemy, and other libraries.

### Terraform Infrastructure

#### Network Firewall (`tf/nfw/`)
Terraform configurations for AWS Network Firewall deployments.

#### Gateway Load Balancer (`tf/gwlb/`)
Gateway Load Balancer configurations for traffic inspection.

#### Transit Gateway (`tf/tgw/`)
Transit Gateway configurations for network connectivity.

## Development

### Prerequisites

- **Go**: Version 1.25.1 or later
- **Python**: 3.8+ with uv for dependency management
- **Terraform**: Latest version
- **Nix**: For reproducible development environments

### Getting Started

1. Clone the repository:
```bash
git clone <repository-url>
cd nix
```

2. Set up development environment:
```bash
# For Go projects
cd go/aws
go mod download

# For Python projects
cd py/aws
uv sync

# For Terraform projects
cd tf/nfw
terraform init
```

### Building and Releasing

#### Go AWS CLI (gaws)

The Go AWS CLI has automated build and release workflows:

**Manual Release:**
```bash
# Create a new release (patch, minor, or major)
./scripts/release.sh patch
```

**GitHub Actions:**
- **Version Management**: Use the "Version Management" workflow to create version tags
- **Build and Release**: Automatically triggered on version tags to build for:
  - Linux (amd64, arm64)
  - macOS (amd64, arm64)

**Download Releases:**
Visit the [Releases page](https://github.com/your-org/nix/releases) to download pre-built binaries.

#### Installation

```bash
# Download for your platform
wget https://github.com/your-org/nix/releases/download/v1.0.0/gaws-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m).tar.gz
tar -xzf gaws-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m).tar.gz
sudo mv gaws /usr/local/bin/
```

### Usage

```bash
gaws --help
gaws subnets --help
gaws nlb --help
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.