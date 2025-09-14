# Getting Started with KloneKit

This tutorial walks you through creating your first KloneKit blueprint and deploying infrastructure. By the end, you'll understand how to use KloneKit to automate GitLab repository creation and AWS infrastructure provisioning.

## Prerequisites

Before starting, ensure you have:

- **KloneKit installed** (see [Installation Guide](installation.md))
- **GitLab Personal Access Token** with API and repository permissions
- **AWS credentials configured** (via AWS CLI, environment variables, or IAM roles)
- **Docker installed and running** (KloneKit uses Docker to run Terraform)
- **Basic understanding of Terraform** and infrastructure as code concepts

## Understanding KloneKit

KloneKit automates a common DevOps workflow:
1. **Scaffold** - Copy your Terraform files and generate variable files from the blueprint
2. **SCM** - Create a GitLab repository and push your infrastructure code
3. **Provision** - Execute Terraform via Docker container to create AWS resources

You can run these steps individually or all together with the `apply` command.

## Step 1: Set Up Authentication

### GitLab Authentication

Create a GitLab Personal Access Token with the following scopes:
- `api` (for repository creation)
- `read_repository` and `write_repository`

Set the token as an environment variable:

```bash
export GITLAB_PRIVATE_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
```

### AWS Authentication

Configure AWS credentials using one of these methods:

**Option 1: AWS CLI (Recommended)**
```bash
aws configure
```

**Option 2: Environment Variables**
```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_DEFAULT_REGION="us-west-2"
```

**Option 3: IAM Roles** (if running on EC2)
No additional configuration needed if your EC2 instance has appropriate IAM roles.

## Step 2: Create Your Project Structure

Create a directory for your first KloneKit project:

```bash
mkdir my-first-klonekit-project
cd my-first-klonekit-project
```

Create a Terraform source directory with a simple infrastructure configuration:

```bash
mkdir terraform
```

Create `terraform/main.tf` with example AWS infrastructure:

```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.region
}

variable "region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-west-2"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "development"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.micro"
}

# Example VPC
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "${var.environment}-vpc"
    Environment = var.environment
    ManagedBy   = "KloneKit"
  }
}

# Example subnet
resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = {
    Name        = "${var.environment}-public-subnet"
    Environment = var.environment
    ManagedBy   = "KloneKit"
  }
}

# Data source for AZs
data "aws_availability_zones" "available" {
  state = "available"
}

# Outputs
output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.main.id
}

output "subnet_id" {
  description = "ID of the public subnet"
  value       = aws_subnet.public.id
}
```

## Step 3: Create Your Blueprint

Create a `klonekit.yaml` blueprint file in your project root:

```yaml
apiVersion: v1
kind: Blueprint
metadata:
  name: my-first-infrastructure
  description: "My first KloneKit deployment - VPC and subnet"
  labels:
    project: tutorial
    owner: myname

spec:
  # GitLab configuration
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: my-first-klonekit-project
      namespace: your-gitlab-username-here  # Replace with your GitLab username
      description: "Infrastructure project created with KloneKit"
      visibility: private

  # AWS configuration
  cloud:
    provider: aws
    region: us-west-2

  # File scaffolding configuration
  scaffold:
    source: ./terraform
    destination: ./output

  # Variables passed to Terraform
  variables:
    region: us-west-2
    environment: tutorial
    instance_type: t3.micro
```

**Important:** Replace `your-gitlab-username-here` with your actual GitLab username or group name.

## Step 4: Validate Your Setup

Before running KloneKit, verify your configuration:

### Test Docker
```bash
docker info
```

### Test GitLab Token
```bash
curl -H "Authorization: Bearer $GITLAB_PRIVATE_TOKEN" https://gitlab.com/api/v4/user
```

### Test AWS Credentials
```bash
aws sts get-caller-identity
```

## Step 5: Run KloneKit

### Option 1: Complete Workflow (Recommended for first time)

Run the dry-run first to see what KloneKit will do:

```bash
klonekit apply --file klonekit.yaml --dry-run
```

If the dry-run looks good, execute the full workflow:

```bash
klonekit apply --file klonekit.yaml
```

This will:
1. Copy Terraform files from `./terraform` to `./output`
2. Generate `terraform.tfvars.json` with your variables
3. Create a GitLab repository
4. Push the scaffolded files to GitLab
5. Run `terraform init` and `terraform apply` via Docker

### Option 2: Step-by-Step Workflow

For better understanding, run each step individually:

**Step 1: Scaffold Files**
```bash
klonekit scaffold --file klonekit.yaml --dry-run  # Preview
klonekit scaffold --file klonekit.yaml           # Execute
```

Check the output directory:
```bash
ls -la ./output/
cat ./output/terraform.tfvars.json
```

**Step 2: Create GitLab Repository**
```bash
klonekit scm --file klonekit.yaml
```

**Step 3: Provision Infrastructure**
```bash
klonekit provision --file klonekit.yaml
```

## Step 6: Verify Results

### Check GitLab Repository

1. Visit GitLab.com and navigate to your new repository
2. Verify that your Terraform files were pushed
3. Check that `terraform.tfvars.json` contains your variables

### Check AWS Resources

1. Log into the AWS Console
2. Navigate to VPC service
3. Verify that your VPC and subnet were created
4. Check the tags to confirm they're managed by KloneKit

### Local Output Files

Check the scaffolded files locally:
```bash
tree ./output/
cat ./output/terraform.tfvars.json
```

## Understanding the Blueprint Structure

Let's break down each section of your blueprint:

### Metadata Section
```yaml
metadata:
  name: my-first-infrastructure        # Unique identifier for this project
  description: "Project description"   # Human-readable description
  labels:                             # Optional key-value tags
    project: tutorial
```

### SCM Section
```yaml
spec:
  scm:
    provider: gitlab                   # Only GitLab is currently supported
    url: https://gitlab.com           # GitLab instance URL
    token: ${GITLAB_PRIVATE_TOKEN}    # Environment variable reference
    project:
      name: repository-name           # GitLab repository name
      namespace: username-or-group    # GitLab namespace (user/group)
      visibility: private             # private|public|internal
```

### Cloud Section
```yaml
  cloud:
    provider: aws                     # Only AWS is currently supported
    region: us-west-2                # AWS region for resources
```

### Scaffold Section
```yaml
  scaffold:
    source: ./terraform              # Directory containing your Terraform files
    destination: ./output            # Where to copy and process files
```

### Variables Section
```yaml
  variables:
    environment: tutorial            # Passed to terraform.tfvars.json
    instance_type: t3.micro         # Custom variables for your Terraform code
```

## Common Workflow Patterns

### Development Workflow
```bash
# 1. Create/modify Terraform files
vim terraform/main.tf

# 2. Update blueprint variables
vim klonekit.yaml

# 3. Test scaffolding
klonekit scaffold --file klonekit.yaml --dry-run

# 4. Apply changes
klonekit apply --file klonekit.yaml
```

### Testing Changes
```bash
# Always dry-run first
klonekit apply --file klonekit.yaml --dry-run

# Then apply
klonekit apply --file klonekit.yaml
```

### Iterating on Infrastructure
```bash
# Modify Terraform files
vim terraform/main.tf

# Re-scaffold and update GitLab repo
klonekit scaffold --file klonekit.yaml
klonekit scm --file klonekit.yaml

# Apply Terraform changes
klonekit provision --file klonekit.yaml
```

## Troubleshooting

### Common Issues

**"GITLAB_PRIVATE_TOKEN environment variable is required"**
- Ensure you set the environment variable: `export GITLAB_PRIVATE_TOKEN="your-token"`
- Verify the token is valid: `echo $GITLAB_PRIVATE_TOKEN`

**"failed to connect to Docker daemon"**
- Start Docker Desktop or Docker service
- Test with: `docker info`

**GitLab repository already exists**
- KloneKit will skip creation and push to existing repo
- Ensure you have push permissions to the repository

**Terraform errors in provision step**
- Check AWS credentials: `aws sts get-caller-identity`
- Verify Terraform syntax in your source files
- Check Docker can pull Terraform image: `docker pull hashicorp/terraform:1.8.0`

**Blueprint parsing errors**
- Verify YAML syntax with: `yamllint klonekit.yaml`
- Check all required fields are present
- Ensure proper indentation (spaces, not tabs)

### Getting Detailed Information

**Verbose logging:**
```bash
export KLONEKIT_LOG_LEVEL=debug
klonekit apply --file klonekit.yaml
```

**Docker troubleshooting:**
```bash
# Test Docker connectivity
docker run hello-world

# Check Terraform Docker image
docker run --rm hashicorp/terraform:1.8.0 version
```

## Next Steps

Now that you've completed your first KloneKit deployment:

1. **Explore Advanced Features:**
   - Use the `--retain-state` flag for auditing
   - Experiment with different AWS resources
   - Try different GitLab visibility settings

2. **Customize Your Workflow:**
   - Create more complex Terraform modules
   - Use Terraform workspaces for environments
   - Add more sophisticated variable configurations

3. **Production Considerations:**
   - Use IAM roles instead of access keys
   - Implement Terraform backends for state management
   - Set up GitLab CI/CD pipelines for automated deployments

4. **Learn More:**
   - Study the [Blueprint Reference](https://github.com/salman-frs/klonekit#blueprint-reference) for all available options
   - Explore the individual commands for more control
   - Check out example blueprints in the project repository

## Blueprint Examples

### Multi-Environment Setup
```yaml
apiVersion: v1
kind: Blueprint
metadata:
  name: production-infrastructure

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: production-infra
      namespace: mycompany
      visibility: private

  cloud:
    provider: aws
    region: us-east-1

  scaffold:
    source: ./terraform
    destination: ./output

  variables:
    environment: production
    instance_type: t3.medium
    min_size: 2
    max_size: 10
    enable_logging: true
```

### Development Environment
```yaml
apiVersion: v1
kind: Blueprint
metadata:
  name: dev-infrastructure

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: dev-infra
      namespace: mycompany
      visibility: internal

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ./terraform
    destination: ./output

  variables:
    environment: development
    instance_type: t3.micro
    enable_logging: false
    auto_scaling: false
```

KloneKit simplifies the DevOps workflow by providing a consistent, repeatable process for infrastructure management. Start with simple examples and gradually build more complex infrastructure as you become comfortable with the tool.