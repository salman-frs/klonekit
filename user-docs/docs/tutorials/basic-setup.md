# Basic Setup Tutorial

This comprehensive tutorial walks you through setting up your first KloneKit project from scratch.

## Prerequisites

Before starting this tutorial, ensure you have:

- [ ] KloneKit installed ([Installation Guide](../tasks/installation.md))
- [ ] Docker running
- [ ] GitLab account with Personal Access Token
- [ ] AWS account with configured credentials
- [ ] Basic understanding of Terraform and YAML

## Tutorial Overview

In this tutorial, you'll:

1. Set up authentication
2. Create a simple web application infrastructure
3. Write your first blueprint
4. Execute the complete KloneKit workflow
5. Verify the results

## Step 1: Environment Setup

### Create Project Directory

```bash
mkdir klonekit-web-app
cd klonekit-web-app
```

### Configure Authentication

Set up your GitLab token:

```bash
export GITLAB_PRIVATE_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
```

Verify AWS credentials:

```bash
aws sts get-caller-identity
```

You should see your AWS account information.

## Step 2: Create Terraform Infrastructure

### Directory Structure

Create the source template directory:

```bash
mkdir terraform-src
cd terraform-src
```

### Provider Configuration

Create the Terraform provider configuration:

```hcl title="terraform-src/providers.tf"
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}
```

### Variables Definition

Define the variables your infrastructure will use:

```hcl title="terraform-src/variables.tf"
variable "project_name" {
  description = "Name of the project"
  type        = string
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.micro"
}

variable "key_name" {
  description = "AWS Key Pair name"
  type        = string
  default     = ""
}
```

### VPC and Networking

Create a VPC with public and private subnets:

```hcl title="terraform-src/vpc.tf"
# VPC
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "${var.project_name}-${var.environment}-vpc"
    Environment = var.environment
    Project     = var.project_name
  }
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name        = "${var.project_name}-${var.environment}-igw"
    Environment = var.environment
    Project     = var.project_name
  }
}

# Public Subnet
resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = {
    Name        = "${var.project_name}-${var.environment}-public-subnet"
    Environment = var.environment
    Project     = var.project_name
  }
}

# Route Table for Public Subnet
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name        = "${var.project_name}-${var.environment}-public-rt"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}
```

### Security Group

Create a security group for the web server:

```hcl title="terraform-src/security.tf"
resource "aws_security_group" "web" {
  name_prefix = "${var.project_name}-${var.environment}-web-"
  vpc_id      = aws_vpc.main.id
  description = "Security group for web servers"

  # HTTP
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTPS
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # SSH (optional, only if key_name is provided)
  dynamic "ingress" {
    for_each = var.key_name != "" ? [1] : []
    content {
      from_port   = 22
      to_port     = 22
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]  # Restrict this in production
    }
  }

  # Outbound internet access
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.project_name}-${var.environment}-web-sg"
    Environment = var.environment
    Project     = var.project_name
  }
}
```

### EC2 Instance

Create the web server instance:

```hcl title="terraform-src/compute.tf"
# Data source for availability zones
data "aws_availability_zones" "available" {
  state = "available"
}

# Data source for latest Amazon Linux AMI
data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# Web Server Instance
resource "aws_instance" "web" {
  ami                    = data.aws_ami.amazon_linux.id
  instance_type          = var.instance_type
  key_name              = var.key_name != "" ? var.key_name : null
  vpc_security_group_ids = [aws_security_group.web.id]
  subnet_id             = aws_subnet.public.id

  user_data = <<-EOF
              #!/bin/bash
              yum update -y
              yum install -y httpd
              systemctl start httpd
              systemctl enable httpd
              echo "<h1>Hello from ${var.project_name} - ${var.environment}</h1>" > /var/www/html/index.html
              echo "<p>Instance ID: $(curl -s http://169.254.169.254/latest/meta-data/instance-id)</p>" >> /var/www/html/index.html
              EOF

  tags = {
    Name        = "${var.project_name}-${var.environment}-web"
    Environment = var.environment
    Project     = var.project_name
  }
}
```

### Outputs

Define outputs to display important information:

```hcl title="terraform-src/outputs.tf"
output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.main.id
}

output "public_subnet_id" {
  description = "ID of the public subnet"
  value       = aws_subnet.public.id
}

output "web_instance_id" {
  description = "ID of the web server instance"
  value       = aws_instance.web.id
}

output "web_instance_public_ip" {
  description = "Public IP of the web server"
  value       = aws_instance.web.public_ip
}

output "web_instance_public_dns" {
  description = "Public DNS of the web server"
  value       = aws_instance.web.public_dns
}

output "website_url" {
  description = "URL of the website"
  value       = "http://${aws_instance.web.public_ip}"
}
```

Return to the project root:

```bash
cd ..
```

## Step 3: Create KloneKit Blueprint

Create your blueprint configuration:

```yaml title="klonekit.yaml"
apiVersion: v1
kind: Blueprint
metadata:
  name: web-app-tutorial
  description: "Tutorial web application infrastructure"
  labels:
    tutorial: basic-setup
    type: web-application

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: klonekit-web-app-tutorial
      namespace: YOUR_GITLAB_USERNAME  # Replace with your username
      description: "Web application infrastructure managed by KloneKit"
      visibility: private

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ./terraform-src
    destination: ./infrastructure

  variables:
    project_name: "klonekit-tutorial"
    environment: "development"
    aws_region: "us-west-2"
    instance_type: "t3.micro"
    # key_name: "my-key-pair"  # Uncomment if you have an EC2 key pair
```

!!! warning "Update Configuration"
    Replace `YOUR_GITLAB_USERNAME` with your actual GitLab username.

## Step 4: Validate Configuration

Before executing, validate your setup:

```bash
# Test the blueprint with a dry run
klonekit apply --file klonekit.yaml --dry-run
```

This command will:
- Parse and validate your blueprint
- Check that source files exist
- Verify authentication
- Show what would be executed

## Step 5: Execute the Workflow

### Option 1: Complete Workflow

Run the entire workflow:

```bash
klonekit apply --file klonekit.yaml
```

### Option 2: Step-by-Step Execution

For learning purposes, execute each step individually:

```bash
# Step 1: Scaffold files
klonekit scaffold --file klonekit.yaml
ls -la infrastructure/

# Step 2: Create repository and push
klonekit scm --file klonekit.yaml

# Step 3: Provision infrastructure
klonekit provision --file klonekit.yaml
```

## Step 6: Verify Results

### Check Generated Files

```bash
# View the generated Terraform configuration
ls -la infrastructure/
cat infrastructure/terraform.tfvars.json
```

### Verify GitLab Repository

1. Visit your GitLab account
2. Find the `klonekit-web-app-tutorial` repository
3. Verify all Terraform files are present

### Check AWS Infrastructure

1. Log into AWS Console
2. Navigate to EC2 → Instances
3. Find your new instance
4. Copy the public IP address
5. Visit `http://YOUR_INSTANCE_IP` in a browser

You should see your web application running!

### View Terraform Outputs

```bash
cd infrastructure/
cat terraform.tfstate | grep -A 10 '"outputs"'
```

## Step 7: Understanding What Happened

Let's review what KloneKit accomplished:

### 1. File Scaffolding
- Copied Terraform files from `terraform-src/` to `infrastructure/`
- Created `terraform.tfvars.json` with your blueprint variables
- Preserved file structure and permissions

### 2. SCM Integration
- Created a new GitLab repository
- Initialized git repository locally
- Pushed all generated files to GitLab
- Set up proper repository metadata

### 3. Infrastructure Provisioning
- Ran `terraform init` in a Docker container
- Executed `terraform plan` to preview changes
- Applied the configuration with `terraform apply`
- Created all AWS resources as defined

## Step 8: Clean Up (Optional)

To remove the created infrastructure:

```bash
cd infrastructure/
docker run --rm \
  -v $(pwd):/workspace \
  -w /workspace \
  -e AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY \
  -e AWS_DEFAULT_REGION \
  hashicorp/terraform:1.8 \
  destroy -auto-approve
```

## What You've Learned

In this tutorial, you've learned to:

- Structure Terraform code for KloneKit
- Write comprehensive blueprint configurations
- Execute complete DevOps workflows
- Integrate GitLab repository management
- Provision AWS infrastructure declaratively

## Next Steps

- Explore [Advanced Workflows](advanced-workflows.md)
- Learn about [Configuration Options](../tasks/configuration.md)
- Review [Blueprint Examples](../reference/examples.md)
- Try [Multi-Environment Deployments](advanced-workflows.md)