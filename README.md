# KloneKit

**A blueprint-driven DevOps automation tool for GitLab and AWS infrastructure workflows.**

KloneKit simplifies the process of creating GitLab repositories and provisioning AWS infrastructure by orchestrating Terraform deployments through a single blueprint configuration file. It automates the typical DevOps workflow: scaffold Terraform files, create GitLab projects, and provision cloud infrastructure.

## Features

- **Declarative Setup** - Define your entire infrastructure and GitLab project in a single YAML blueprint
- **GitLab Integration** - Automatically create repositories, push code, and manage GitLab projects
- **Terraform Orchestration** - Run Terraform in Docker containers for consistent, isolated deployments
- **Resilient Execution** - Resume interrupted workflows with stateful execution and built-in error recovery
- **Workflow Orchestration** - Execute scaffolding, SCM, and provisioning steps individually or all together
- **Dry-Run Support** - Preview changes before execution across all commands

## Installation

### Homebrew (Recommended)

```bash
brew install salman-frs/klonekit/klonekit
```

### From GitHub Releases

Download the latest pre-built binary for your platform from the [GitHub Releases page](https://github.com/salman-frs/klonekit/releases):

```bash
# For macOS Intel
curl -L -o klonekit https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-darwin-amd64
chmod +x klonekit
sudo mv klonekit /usr/local/bin/

# For macOS Apple Silicon
curl -L -o klonekit https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-darwin-arm64
chmod +x klonekit
sudo mv klonekit /usr/local/bin/

# For Linux
curl -L -o klonekit https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-linux-amd64
chmod +x klonekit
sudo mv klonekit /usr/local/bin/
```

### From Source

Requirements:
- Go 1.22.x or higher
- Make (optional)

```bash
git clone https://github.com/salman-frs/klonekit.git
cd klonekit
go build -o klonekit ./cmd/klonekit
```

Or using Make:
```bash
make build
```

## Quick Start

### 1. Install KloneKit

```bash
# Using Homebrew (recommended)
brew install salman-frs/klonekit/klonekit

# Or download from GitHub Releases
curl -L -o klonekit https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-darwin-amd64
chmod +x klonekit
sudo mv klonekit /usr/local/bin/
```

### 2. Set up Authentication

KloneKit requires a GitLab Personal Access Token:

```bash
export GITLAB_PRIVATE_TOKEN="your-gitlab-token-here"
```

### 3. Create a Blueprint

Create a `klonekit.yaml` file:

```yaml
apiVersion: v1
kind: Blueprint
metadata:
  name: my-infrastructure
  description: "My first KloneKit deployment"

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: my-infrastructure-project
      namespace: your-gitlab-username
      description: "Infrastructure managed by KloneKit"
      visibility: private

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ./terraform
    destination: ./output

  variables:
    environment: development
    instance_type: t3.micro
```

### 4. Create Terraform Source Files

Create a `terraform/` directory with your infrastructure code:

```bash
mkdir terraform
```

`terraform/main.tf`:
```hcl
variable "environment" {
  description = "Environment name"
  type        = string
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.micro"
}

resource "aws_instance" "example" {
  ami           = "ami-0c55b159cbfafe1d0"  # Amazon Linux 2
  instance_type = var.instance_type

  tags = {
    Name        = "${var.environment}-instance"
    Environment = var.environment
  }
}
```

### 5. Run KloneKit

Execute the complete workflow:

```bash
klonekit apply --file klonekit.yaml
```

Or run individual steps:

```bash
# 1. Generate Terraform files with variables
klonekit scaffold --file klonekit.yaml

# 2. Create GitLab repository and push files
klonekit scm --file klonekit.yaml

# 3. Provision infrastructure via Docker
klonekit provision --file klonekit.yaml
```

## Available Commands

- **`apply`** - Execute the complete workflow (scaffold + scm + provision)
- **`scaffold`** - Generate Terraform files from blueprint and copy source files
- **`scm`** - Create GitLab repository and push the scaffolded files
- **`provision`** - Run Terraform in Docker container to provision infrastructure

### Command Options

All commands support:
- `--file, -f` - Path to blueprint YAML file (required)
- `--dry-run` - Simulate operations without making changes

The `apply` command additionally supports:
- `--retain-state` - Keep state files after completion for auditing

## Blueprint Reference

### Complete Blueprint Structure

```yaml
apiVersion: v1                    # Required: API version
kind: Blueprint                   # Required: Must be "Blueprint"

metadata:                         # Required: Project metadata
  name: string                    # Required: Unique identifier
  description: string             # Optional: Project description
  labels:                         # Optional: Key-value labels
    key: value

spec:                            # Required: Specifications
  scm:                           # Required: Source control config
    provider: gitlab             # Required: Only "gitlab" supported
    url: string                  # Required: GitLab instance URL
    token: string                # Required: GitLab token
    project:                     # Required: Project configuration
      name: string               # Required: Repository name
      namespace: string          # Required: GitLab namespace/username
      description: string        # Optional: Repository description
      visibility: string         # Optional: private|public|internal

  cloud:                         # Required: Cloud provider config
    provider: aws                # Required: Only "aws" supported
    region: string               # Required: AWS region

  scaffold:                      # Required: File scaffolding config
    source: string               # Required: Source directory path
    destination: string          # Required: Output directory path

  variables:                     # Optional: Terraform variables
    key: value                   # Variables passed to terraform.tfvars.json
```

## Prerequisites

- **GitLab Personal Access Token** with API and repository permissions
- **Docker** installed and running (for Terraform execution)
- **AWS Credentials** configured (for infrastructure provisioning)
- **Source Terraform files** in the directory specified by `scaffold.source`

## Authentication & Configuration

### GitLab Authentication

Set your GitLab Personal Access Token:
```bash
export GITLAB_PRIVATE_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
```

### AWS Authentication

Configure AWS credentials using any standard method:
```bash
# Option 1: AWS CLI
aws configure

# Option 2: Environment variables
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_DEFAULT_REGION="us-west-2"

# Option 3: IAM roles (if running on EC2)
```

## Development

### Building

```bash
make build          # Build the binary
make test           # Run tests
make lint           # Run linting
make clean          # Clean artifacts
make help           # Show available commands
```

### Requirements

- Go 1.22.x or higher
- Docker (for Terraform execution)
- golangci-lint (for linting)

### Project Structure

```
klonekit/
├── cmd/klonekit/           # CLI entry point
├── internal/               # Private application code
│   ├── app/               # Main workflow orchestrator
│   ├── parser/            # Blueprint parsing
│   ├── provisioner/       # Terraform Docker execution
│   ├── scaffolder/        # File generation
│   └── scm/               # GitLab integration
├── pkg/blueprint/         # Blueprint data structures
└── user-guide/           # Documentation
```

## Troubleshooting

### Common Issues

**"GITLAB_PRIVATE_TOKEN environment variable is required"**
- Solution: Set your GitLab token: `export GITLAB_PRIVATE_TOKEN="your-token"`

**"failed to connect to Docker daemon"**
- Solution: Ensure Docker is installed and running
- Check: `docker info`

**Blueprint validation errors**
- Solution: Verify YAML syntax and required fields
- Use: `--dry-run` to validate configuration

**GitLab repository creation fails**
- Check token permissions (API, repository access)
- Verify namespace exists and you have access
- Ensure repository name doesn't already exist

### Getting Help

Run any command with `--help` for detailed usage:
```bash
klonekit --help
klonekit apply --help
klonekit scaffold --help
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Run linting: `make lint`
6. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.