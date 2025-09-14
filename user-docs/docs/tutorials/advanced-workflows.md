# Advanced Workflows

Learn advanced KloneKit patterns and workflows for complex infrastructure scenarios.

## Multi-Environment Deployments

Deploy the same infrastructure across multiple environments using environment-specific blueprints.

### Environment-Specific Blueprints

Create separate blueprints for each environment:

```yaml title="environments/dev.yaml"
apiVersion: v1
kind: Blueprint
metadata:
  name: web-app-dev
  description: "Development environment"

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: web-app-dev
      namespace: your-org
      visibility: private

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ../terraform-templates
    destination: ./dev-infrastructure

  variables:
    environment: development
    instance_type: t3.micro
    min_instances: 1
    max_instances: 2
```

```yaml title="environments/prod.yaml"
apiVersion: v1
kind: Blueprint
metadata:
  name: web-app-prod
  description: "Production environment"

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: web-app-prod
      namespace: your-org
      visibility: private

  cloud:
    provider: aws
    region: us-east-1

  scaffold:
    source: ../terraform-templates
    destination: ./prod-infrastructure

  variables:
    environment: production
    instance_type: t3.large
    min_instances: 3
    max_instances: 10
    enable_monitoring: true
```

### Deployment Script

```bash title="deploy.sh"
#!/bin/bash
set -e

ENVIRONMENT=${1:-dev}
BLUEPRINT_FILE="environments/${ENVIRONMENT}.yaml"

if [[ ! -f "$BLUEPRINT_FILE" ]]; then
    echo "Blueprint file $BLUEPRINT_FILE not found"
    exit 1
fi

echo "Deploying to $ENVIRONMENT environment..."
klonekit apply --file "$BLUEPRINT_FILE"

echo "Deployment to $ENVIRONMENT completed successfully"
```

Usage:
```bash
./deploy.sh dev   # Deploy to development
./deploy.sh prod  # Deploy to production
```

## Template-Based Infrastructure

Create reusable infrastructure templates with parameterized configurations.

### Base Template Structure

```
templates/
├── web-app/
│   ├── main.tf
│   ├── variables.tf
│   └── outputs.tf
├── database/
│   ├── rds.tf
│   ├── variables.tf
│   └── outputs.tf
└── monitoring/
    ├── cloudwatch.tf
    ├── variables.tf
    └── outputs.tf
```

### Composite Blueprint

Combine multiple templates in a single deployment:

```yaml title="composite-app.yaml"
apiVersion: v1
kind: Blueprint
metadata:
  name: full-stack-app
  description: "Complete application stack with database and monitoring"

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: full-stack-infrastructure
      namespace: your-org
      visibility: private

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ./templates
    destination: ./infrastructure

  variables:
    # Application variables
    app_name: "my-full-stack-app"
    environment: "production"

    # Web tier
    web_instance_type: "t3.large"
    web_min_instances: 2
    web_max_instances: 10

    # Database tier
    db_instance_class: "db.r5.large"
    db_allocated_storage: 100
    db_backup_retention: 7

    # Monitoring
    enable_detailed_monitoring: true
    notification_email: "alerts@company.com"
```

## GitOps Integration

Implement GitOps workflows where infrastructure changes are managed through Git commits.

### Repository Structure

```
gitops-infrastructure/
├── applications/
│   ├── web-app/
│   │   └── klonekit.yaml
│   ├── api-service/
│   │   └── klonekit.yaml
│   └── database/
│       └── klonekit.yaml
├── environments/
│   ├── dev/
│   ├── staging/
│   └── prod/
└── templates/
    └── terraform/
```

### GitLab CI Pipeline for GitOps

```yaml title=".gitlab-ci.yml"
stages:
  - validate
  - plan
  - deploy

variables:
  TERRAFORM_VERSION: "1.8"

.klonekit-base:
  image: alpine:latest
  before_script:
    - apk add --no-cache curl
    - curl -L -o klonekit https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-linux-amd64
    - chmod +x klonekit
    - mv klonekit /usr/local/bin/

validate:
  extends: .klonekit-base
  stage: validate
  script:
    - for app in applications/*/klonekit.yaml; do
        echo "Validating $app"
        klonekit apply --file "$app" --dry-run
      done

plan-dev:
  extends: .klonekit-base
  stage: plan
  script:
    - klonekit apply --file applications/web-app/klonekit.yaml --dry-run
  environment:
    name: development
  only:
    - merge_requests

deploy-dev:
  extends: .klonekit-base
  stage: deploy
  script:
    - klonekit apply --file applications/web-app/klonekit.yaml
  environment:
    name: development
  only:
    - develop

deploy-prod:
  extends: .klonekit-base
  stage: deploy
  script:
    - klonekit apply --file applications/web-app/klonekit.yaml
  environment:
    name: production
  only:
    - main
  when: manual
```

## State Management Patterns

Advanced patterns for managing Terraform state across environments.

### Remote State with S3 Backend

```hcl title="terraform-templates/backend.tf"
terraform {
  backend "s3" {
    bucket  = var.state_bucket
    key     = "${var.environment}/${var.project_name}/terraform.tfstate"
    region  = var.aws_region
    encrypt = true
  }
}
```

```yaml title="blueprint-with-remote-state.yaml"
variables:
  state_bucket: "my-terraform-state-bucket"
  project_name: "web-app"
  environment: "production"
  aws_region: "us-west-2"
```

### State Locking with DynamoDB

```hcl title="terraform-templates/backend.tf"
terraform {
  backend "s3" {
    bucket         = var.state_bucket
    key            = "${var.environment}/${var.project_name}/terraform.tfstate"
    region         = var.aws_region
    encrypt        = true
    dynamodb_table = var.state_lock_table
  }
}
```

## Infrastructure Testing

Integrate testing into your KloneKit workflows.

### Validation Pipeline

```bash title="validate-infrastructure.sh"
#!/bin/bash
set -e

BLUEPRINT_FILE=${1:-klonekit.yaml}

echo "Step 1: Blueprint validation"
klonekit apply --file "$BLUEPRINT_FILE" --dry-run

echo "Step 2: Terraform syntax validation"
klonekit scaffold --file "$BLUEPRINT_FILE"
cd infrastructure/
terraform init -backend=false
terraform validate
terraform fmt -check
cd ..

echo "Step 3: Security scanning"
# Example with tfsec
docker run --rm -v $(pwd)/infrastructure:/workspace aquasec/tfsec /workspace

echo "Step 4: Cost estimation"
# Example with Infracost
infracost breakdown --path infrastructure/

echo "All validations passed!"
```

### Post-Deployment Testing

```bash title="test-deployment.sh"
#!/bin/bash
set -e

# Extract outputs from Terraform state
cd infrastructure/
WEB_URL=$(terraform output -raw website_url)
INSTANCE_ID=$(terraform output -raw web_instance_id)
cd ..

echo "Testing deployment..."

# Test web application
if curl -f "$WEB_URL" > /dev/null; then
    echo "✓ Web application is responding"
else
    echo "✗ Web application is not responding"
    exit 1
fi

# Test AWS resources
if aws ec2 describe-instances --instance-ids "$INSTANCE_ID" > /dev/null; then
    echo "✓ EC2 instance is running"
else
    echo "✗ EC2 instance not found"
    exit 1
fi

echo "All tests passed!"
```

## Disaster Recovery

Implement disaster recovery patterns with KloneKit.

### Cross-Region Backup

```yaml title="disaster-recovery.yaml"
apiVersion: v1
kind: Blueprint
metadata:
  name: dr-infrastructure
  description: "Disaster recovery infrastructure"

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: dr-backup-infrastructure
      namespace: your-org
      visibility: private

  cloud:
    provider: aws
    region: us-east-1  # Different region for DR

  scaffold:
    source: ./terraform-templates
    destination: ./dr-infrastructure

  variables:
    environment: "dr"
    primary_region: "us-west-2"
    backup_region: "us-east-1"
    enable_cross_region_backup: true
    backup_retention_days: 30
```

### Recovery Script

```bash title="disaster-recovery.sh"
#!/bin/bash
set -e

ACTION=${1:-prepare}  # prepare, activate, failback

case $ACTION in
    prepare)
        echo "Preparing disaster recovery infrastructure..."
        klonekit apply --file disaster-recovery.yaml
        ;;
    activate)
        echo "Activating disaster recovery..."
        # Update DNS, load balancers, etc.
        klonekit apply --file disaster-recovery.yaml
        ./scripts/update-dns-to-dr.sh
        ;;
    failback)
        echo "Failing back to primary region..."
        ./scripts/update-dns-to-primary.sh
        ;;
    *)
        echo "Usage: $0 [prepare|activate|failback]"
        exit 1
        ;;
esac
```

## Security Hardening

Advanced security patterns for KloneKit deployments.

### Secret Management

```yaml title="secure-blueprint.yaml"
apiVersion: v1
kind: Blueprint
metadata:
  name: secure-app
  description: "Security-hardened application"

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: secure-infrastructure
      namespace: your-org
      visibility: private

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ./secure-templates
    destination: ./infrastructure

  variables:
    # Use AWS Secrets Manager
    db_password_secret_arn: ${DB_PASSWORD_SECRET_ARN}
    api_key_secret_arn: ${API_KEY_SECRET_ARN}

    # Security settings
    enable_encryption: true
    enable_cloudtrail: true
    enable_vpc_flow_logs: true
    restrict_ssh_access: true
    allowed_ssh_cidrs: ["10.0.0.0/8"]
```

### Security Scanning Integration

```yaml title=".gitlab-ci.yml"
security-scan:
  stage: validate
  script:
    - klonekit scaffold --file klonekit.yaml
    - docker run --rm -v $(pwd)/infrastructure:/workspace aquasec/tfsec /workspace
    - docker run --rm -v $(pwd)/infrastructure:/workspace bridgecrew/checkov -d /workspace
  artifacts:
    reports:
      junit: security-report.xml
```

## Performance Optimization

Optimize KloneKit workflows for large-scale deployments.

### Parallel Deployments

```bash title="parallel-deploy.sh"
#!/bin/bash

# Deploy multiple applications in parallel
declare -a APPS=("web-app" "api-service" "database" "monitoring")

for app in "${APPS[@]}"; do
    echo "Starting deployment of $app..."
    (
        cd "applications/$app"
        klonekit apply --file klonekit.yaml
        echo "$app deployment completed"
    ) &
done

# Wait for all deployments to complete
wait
echo "All deployments completed!"
```

### Resource Optimization

```yaml title="optimized-blueprint.yaml"
variables:
  # Terraform performance tuning
  terraform_parallelism: 20
  terraform_refresh: false

  # AWS optimization
  instance_types:
    web: "t3.large"
    api: "c5.xlarge"
    database: "r5.2xlarge"

  # Auto-scaling configuration
  auto_scaling:
    target_cpu_utilization: 70
    scale_up_cooldown: 300
    scale_down_cooldown: 600
```

## Next Steps

- Review [CLI Reference](../reference/cli.md)
- Explore [Blueprint Examples](../reference/examples.md)
- Check [Configuration Options](../tasks/configuration.md)
- Learn about [Troubleshooting](../tasks/troubleshooting.md)