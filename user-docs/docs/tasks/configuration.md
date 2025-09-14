# Configuration

Learn how to configure KloneKit for your specific environment and requirements.

## Authentication Configuration

### GitLab Authentication

KloneKit requires a GitLab Personal Access Token for repository operations.

#### Creating a Token

1. Navigate to GitLab → Settings → Access Tokens
2. Create a token with these scopes:
   - `api` - Full API access
   - `read_repository` - Read repository content
   - `write_repository` - Write repository content

#### Setting the Token

```bash
export GITLAB_PRIVATE_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
```

#### Token Security Best Practices

- Use environment-specific tokens
- Regularly rotate tokens
- Limit token scope to necessary permissions
- Never commit tokens to version control

### AWS Credentials

Configure AWS credentials for infrastructure provisioning:

#### Method 1: AWS CLI

```bash
aws configure
```

#### Method 2: Environment Variables

```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_DEFAULT_REGION="us-west-2"
```

#### Method 3: IAM Roles (Recommended for EC2)

When running on EC2 instances, use IAM roles for automatic credential management.

## Blueprint Configuration

### Environment Variables in Blueprints

Use environment variable substitution for sensitive or environment-specific values:

```yaml
spec:
  scm:
    token: ${GITLAB_PRIVATE_TOKEN}
  variables:
    environment: ${ENVIRONMENT_NAME}
    aws_region: ${AWS_DEFAULT_REGION}
```

### Advanced Blueprint Options

#### Custom GitLab Instance

For self-hosted GitLab:

```yaml
spec:
  scm:
    provider: gitlab
    url: https://gitlab.company.com
    token: ${GITLAB_PRIVATE_TOKEN}
```

#### Multi-Region Deployments

```yaml
spec:
  cloud:
    provider: aws
    region: ${AWS_REGION}
  variables:
    backup_region: ${AWS_BACKUP_REGION}
```

## Docker Configuration

### Custom Terraform Version

Override the default Terraform version:

```bash
# KloneKit uses environment variables to customize Docker execution
export TERRAFORM_VERSION="1.8.0"
klonekit provision --file blueprint.yaml
```

### Docker Resource Limits

For large Terraform plans, you may need to adjust Docker resource limits in Docker Desktop or your container runtime.

## Logging Configuration

### Structured Logging

KloneKit uses structured logging with configurable levels:

```bash
# Set log level (debug, info, warn, error)
export KLONEKIT_LOG_LEVEL="debug"

# Enable JSON output
export KLONEKIT_LOG_FORMAT="json"
```

### Debug Mode

Enable detailed logging for troubleshooting:

```bash
klonekit apply --file blueprint.yaml --verbose
```

## Workspace Configuration

### Custom Working Directory

Specify a different working directory:

```yaml
spec:
  scaffold:
    source: ./terraform-source
    destination: /tmp/klonekit-workspace
```

### State Management

Configure Terraform state handling:

```bash
# Keep state files after completion
klonekit apply --file blueprint.yaml --retain-state
```

## CI/CD Integration

### GitLab CI Integration

```yaml title=".gitlab-ci.yml"
stages:
  - validate
  - deploy

variables:
  GITLAB_PRIVATE_TOKEN: $GITLAB_TOKEN
  AWS_ACCESS_KEY_ID: $AWS_ACCESS_KEY_ID
  AWS_SECRET_ACCESS_KEY: $AWS_SECRET_ACCESS_KEY

validate:
  stage: validate
  script:
    - klonekit apply --file klonekit.yaml --dry-run

deploy:
  stage: deploy
  script:
    - klonekit apply --file klonekit.yaml
  only:
    - main
```

### GitHub Actions Integration

```yaml title=".github/workflows/klonekit.yml"
name: KloneKit Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup KloneKit
        run: |
          curl -L -o klonekit https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-linux-amd64
          chmod +x klonekit
          sudo mv klonekit /usr/local/bin/
      - name: Deploy Infrastructure
        env:
          GITLAB_PRIVATE_TOKEN: ${{ secrets.GITLAB_TOKEN }}
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        run: klonekit apply --file klonekit.yaml
```

## Performance Tuning

### Parallel Execution

For large deployments, Terraform supports parallel resource creation:

```bash
# Increase parallelism (default is 10)
export TF_CLI_ARGS="-parallelism=20"
```

### Resource Caching

Use consistent working directories to leverage Terraform's provider caching:

```yaml
spec:
  scaffold:
    destination: ./.klonekit-cache
```

## Troubleshooting Configuration

### Common Configuration Issues

**Token Permission Errors**
- Verify token has required scopes
- Check token expiration
- Ensure correct GitLab instance URL

**AWS Credential Errors**
- Validate credentials with `aws sts get-caller-identity`
- Check IAM permissions for required services
- Verify region configuration

**Docker Connection Issues**
- Ensure Docker daemon is running
- Check Docker permissions for your user
- Verify Docker can pull Terraform images

### Configuration Validation

Test your configuration:

```bash
# Validate blueprint structure
klonekit apply --file blueprint.yaml --dry-run

# Test GitLab connectivity
curl -H "PRIVATE-TOKEN: $GITLAB_PRIVATE_TOKEN" https://gitlab.com/api/v4/user

# Test AWS connectivity
aws sts get-caller-identity
```

## Next Steps

- Review [Troubleshooting Guide](troubleshooting.md)
- Try [Advanced Workflows](../tutorials/advanced-workflows.md)
- Explore [CLI Reference](../reference/cli.md)