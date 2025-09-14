# KloneKit User Guide

Welcome to the KloneKit User Guide! KloneKit is a blueprint-driven DevOps automation tool that simplifies GitLab and AWS infrastructure workflows by orchestrating Terraform deployments through single configuration files.

## Quick Navigation

### Getting Started
- **[Installation](installation.md)** - Install KloneKit from GitHub releases or build from source
- **[Getting Started Tutorial](getting-started.md)** - Complete walkthrough of your first deployment with working examples

## What is KloneKit?

KloneKit automates a common DevOps workflow that many teams do manually:
1. **Scaffold** - Copy Terraform files and generate variable configurations
2. **SCM** - Create GitLab repositories and push infrastructure code
3. **Provision** - Execute Terraform via Docker containers to create AWS resources

Instead of running these steps manually, KloneKit orchestrates them through a single blueprint YAML file.

## Core Features

✅ **Blueprint-Driven Configuration** - Define everything in one YAML file
✅ **GitLab Integration** - Automated repository creation and code pushing
✅ **Containerized Terraform** - Consistent execution via Docker
✅ **Workflow Orchestration** - Run individual steps or complete workflow
✅ **Dry-Run Support** - Preview all changes before execution
✅ **Stateful Execution** - Resume interrupted workflows automatically

## Available Commands

KloneKit provides 4 main commands:

- **`klonekit apply`** - Execute complete workflow (scaffold → scm → provision)
- **`klonekit scaffold`** - Generate Terraform files and variables from blueprint
- **`klonekit scm`** - Create GitLab repository and push scaffolded files
- **`klonekit provision`** - Run Terraform via Docker to provision infrastructure

All commands support `--dry-run` for safe previewing and `--file` to specify your blueprint.

## How KloneKit Works

### 1. Blueprint Configuration
You define your infrastructure and project settings in a `klonekit.yaml` blueprint:

```yaml
apiVersion: v1
kind: Blueprint
metadata:
  name: my-project

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: my-infrastructure
      namespace: username

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ./terraform
    destination: ./output

  variables:
    environment: production
    instance_type: t3.medium
```

### 2. Source Terraform Files
You create standard Terraform files in your source directory that define your infrastructure.

### 3. Automated Execution
KloneKit processes your blueprint and:
- Copies Terraform files to the destination
- Generates `terraform.tfvars.json` with your variables
- Creates GitLab repositories via API
- Pushes code using Git
- Runs Terraform in Docker containers

## Prerequisites

To use KloneKit effectively, you need:

- **GitLab Personal Access Token** with API and repository permissions
- **AWS Credentials** configured (CLI, environment variables, or IAM roles)
- **Docker** installed and running (for Terraform execution)
- **Source Terraform files** that define your infrastructure

## Getting Started Path

New to KloneKit? Follow this path:

1. **[Install KloneKit](installation.md)** - Get the binary on your system
2. **[Complete the tutorial](getting-started.md)** - Build your first project step-by-step
3. **Practice with examples** - Try the blueprint examples in the tutorial
4. **Scale up** - Apply KloneKit to your real infrastructure projects

## Real-World Use Cases

### Development Teams
- Standardize infrastructure deployment across projects
- Ensure consistent GitLab repository setup and naming
- Simplify onboarding for new team members

### DevOps Engineers
- Automate repetitive infrastructure provisioning tasks
- Reduce manual errors in multi-step deployments
- Maintain audit trails through GitLab version control

### Organizations
- Enforce infrastructure standards through blueprint templates
- Enable self-service infrastructure for development teams
- Integrate with existing GitLab and AWS workflows

## Blueprint Structure Overview

KloneKit blueprints have four main sections:

**Metadata**: Project identification and labels
```yaml
metadata:
  name: project-identifier
  description: "What this project does"
```

**SCM**: GitLab repository configuration
```yaml
spec:
  scm:
    provider: gitlab
    project:
      name: repository-name
      namespace: gitlab-username
```

**Cloud**: AWS provider settings
```yaml
  cloud:
    provider: aws
    region: us-west-2
```

**Scaffold**: File processing configuration
```yaml
  scaffold:
    source: ./terraform      # Your Terraform files
    destination: ./output    # Where to process them
```

**Variables**: Terraform variable values
```yaml
  variables:
    environment: production  # Passed to terraform.tfvars.json
    instance_type: t3.large
```

## Workflow Patterns

### Complete Automation
```bash
klonekit apply --file project.yaml
```
Runs all steps: scaffold → create GitLab repo → provision infrastructure

### Step-by-Step Control
```bash
klonekit scaffold --file project.yaml    # Process files
klonekit scm --file project.yaml        # Create repository
klonekit provision --file project.yaml  # Deploy infrastructure
```

### Safe Testing
```bash
klonekit apply --file project.yaml --dry-run
```
Preview all changes without making them

## Authentication

KloneKit uses standard authentication methods:

**GitLab**: Personal Access Token via `GITLAB_PRIVATE_TOKEN` environment variable
**AWS**: Standard credential chain (CLI, environment variables, IAM roles)
**Docker**: Local Docker daemon access

## Error Handling & Recovery

KloneKit includes built-in resilience:
- **State tracking** - Resume workflows from interruption points
- **Validation** - Check configuration before execution
- **Dry-run mode** - Preview changes safely
- **Detailed logging** - Understand what went wrong

## Current Limitations

Be aware of these current limitations:
- **GitLab only** - Other Git providers not yet supported
- **AWS only** - Other cloud providers not yet supported
- **Docker required** - Terraform execution requires Docker
- **No built-in state backends** - Use standard Terraform state management

## Getting Help

- **Installation issues**: See [Installation troubleshooting](installation.md#troubleshooting-installation)
- **Blueprint problems**: Check [Getting Started examples](getting-started.md#troubleshooting)
- **Command questions**: Run `klonekit [command] --help`
- **Project issues**: Visit [GitHub Issues](https://github.com/salman-frs/klonekit/issues)

## Contributing

KloneKit is open source and welcomes contributions:
- Report bugs and request features via GitHub Issues
- Submit pull requests with improvements
- Share blueprint examples and use cases
- Help improve documentation

Ready to automate your infrastructure workflows? Start with the [Installation Guide](installation.md)!