# Getting Started

This guide walks you through creating your first KloneKit project from start to finish.

## Quick Start Checklist

- [ ] KloneKit installed and verified
- [ ] Docker running
- [ ] GitLab token configured
- [ ] AWS credentials configured

If you haven't completed these steps, see the [Installation Guide](installation.md).

## Create Your First Project

### 1. Create Project Directory

```bash
mkdir my-first-klonekit-project
cd my-first-klonekit-project
```

### 2. Create Terraform Templates

Create a directory for your Terraform source files:

```bash
mkdir terraform-templates
```

Create a simple EC2 instance template:

```bash title="terraform-templates/main.tf"
variable "environment" {
  description = "Environment name"
  type        = string
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.micro"
}

data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_instance" "example" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = var.instance_type

  tags = {
    Name        = "${var.environment}-instance"
    Environment = var.environment
    ManagedBy   = "KloneKit"
  }
}

output "instance_ip" {
  description = "Public IP of the instance"
  value       = aws_instance.example.public_ip
}
```

### 3. Create Blueprint Configuration

Create your blueprint file:

```yaml title="klonekit.yaml"
apiVersion: v1
kind: Blueprint
metadata:
  name: my-first-project
  description: "My first KloneKit deployment"

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: my-first-klonekit-project
      namespace: your-gitlab-username  # Replace with your username
      description: "Infrastructure managed by KloneKit"
      visibility: private

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ./terraform-templates
    destination: ./infrastructure

  variables:
    environment: development
    instance_type: t3.micro
```

!!! warning "Update Configuration"
    Replace `your-gitlab-username` with your actual GitLab username or organization.

### 4. Validate Your Setup

Test the configuration with a dry run:

```bash
klonekit apply --file klonekit.yaml --dry-run
```

This will show you what KloneKit would do without actually executing the operations.

### 5. Execute the Workflow

Run the complete workflow:

```bash
klonekit apply --file klonekit.yaml
```

This will:
1. Generate Terraform files in `./infrastructure/`
2. Create a GitLab repository
3. Push the files to GitLab
4. Provision AWS infrastructure

## Step-by-Step Execution

Alternatively, you can run individual steps:

### Step 1: Scaffold Files

```bash
klonekit scaffold --file klonekit.yaml
```

This generates the Terraform configuration with variables substituted.

### Step 2: Create Repository

```bash
klonekit scm --file klonekit.yaml
```

This creates the GitLab repository and pushes your files.

### Step 3: Provision Infrastructure

```bash
klonekit provision --file klonekit.yaml
```

This runs Terraform to create your AWS resources.

## Verify Results

### Check GitLab Repository

1. Visit your GitLab account
2. Find the newly created repository
3. Verify the Terraform files are present

### Check AWS Resources

1. Log into AWS Console
2. Navigate to EC2 service
3. Find your newly created instance

### Local Files

Check the generated files:

```bash
ls -la infrastructure/
cat infrastructure/terraform.tfvars.json
```

## Clean Up

To destroy the created infrastructure:

```bash
cd infrastructure
docker run --rm -v $(pwd):/workspace -w /workspace \
  -e AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY \
  -e AWS_DEFAULT_REGION \
  hashicorp/terraform:1.8 destroy -auto-approve
```

## What You've Learned

In this guide, you've:

- Created Terraform templates
- Written a KloneKit blueprint
- Executed a complete DevOps workflow
- Provisioned AWS infrastructure
- Integrated with GitLab

## Next Steps

- Learn about [Advanced Configuration](configuration.md)
- Try the [Advanced Workflows Tutorial](../tutorials/advanced-workflows.md)
- Explore [Blueprint Examples](../reference/examples.md)