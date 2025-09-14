# Blueprints

Blueprints are the heart of KloneKit - YAML configuration files that define your entire DevOps workflow in a declarative manner.

## What is a Blueprint?

A blueprint is a structured YAML file that describes:
- Your GitLab repository configuration
- AWS infrastructure requirements
- File scaffolding instructions
- Variable definitions for Terraform

## Blueprint Structure

Every blueprint follows this high-level structure:

```yaml
apiVersion: v1                    # API version
kind: Blueprint                   # Resource type
metadata:                         # Project metadata
  name: string
  description: string
spec:                            # Specifications
  scm: {}                        # Source control config
  cloud: {}                      # Cloud provider config
  scaffold: {}                   # File scaffolding config
  variables: {}                  # Terraform variables
```

## Core Sections

### Metadata

The metadata section provides basic information about your project:

```yaml
metadata:
  name: my-infrastructure
  description: "Production infrastructure for web application"
  labels:
    environment: production
    team: platform
```

### SCM Configuration

Defines GitLab repository settings:

```yaml
spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: my-infrastructure-project
      namespace: my-organization
      description: "Infrastructure managed by KloneKit"
      visibility: private
```

### Cloud Configuration

Specifies cloud provider settings:

```yaml
spec:
  cloud:
    provider: aws
    region: us-west-2
```

### Scaffolding Configuration

Controls how files are processed and generated:

```yaml
spec:
  scaffold:
    source: ./terraform-templates
    destination: ./infrastructure-output
```

### Variables

Define values that will be substituted in your Terraform files:

```yaml
spec:
  variables:
    environment: production
    instance_type: t3.large
    availability_zones:
      - us-west-2a
      - us-west-2b
    enable_monitoring: true
```

## Variable Substitution

Variables defined in your blueprint are automatically:
1. Written to `terraform.tfvars.json` in the output directory
2. Available for use in your Terraform configurations
3. Accessible during the scaffolding process

## Environment Variable Support

Blueprint values support environment variable substitution using `${VAR_NAME}` syntax:

```yaml
spec:
  scm:
    token: ${GITLAB_PRIVATE_TOKEN}
  variables:
    aws_region: ${AWS_DEFAULT_REGION}
```

## Blueprint Validation

KloneKit validates blueprints for:
- Required fields and structure
- Valid provider configurations
- Accessible source directories
- Environment variable resolution

## Example: Complete Blueprint

```yaml
apiVersion: v1
kind: Blueprint
metadata:
  name: web-app-infrastructure
  description: "Complete web application infrastructure"
  labels:
    environment: production
    project: web-app

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: web-app-infrastructure
      namespace: my-org
      description: "Production infrastructure for web application"
      visibility: private

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ./terraform-templates
    destination: ./infrastructure

  variables:
    environment: production
    instance_type: t3.large
    min_instances: 2
    max_instances: 10
    enable_logging: true
    availability_zones:
      - us-west-2a
      - us-west-2b
      - us-west-2c
```

## Best Practices

### Organization
- Use descriptive names and descriptions
- Apply consistent labeling
- Group related variables logically

### Security
- Never commit sensitive values directly
- Use environment variables for secrets
- Leverage GitLab's CI/CD variables for automation

### Maintainability
- Keep blueprints version controlled
- Document variable purposes
- Use consistent naming conventions

## Next Steps

- Learn about [Architecture](architecture.md)
- Review the [Overview](overview.md)
- Start with [Basic Setup Tutorial](../tutorials/basic-setup.md)