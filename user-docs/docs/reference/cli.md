# CLI Reference

Complete command-line interface reference for KloneKit.

## Global Options

These options are available for all commands:

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--help` | `-h` | Show help information | |
| `--version` | `-v` | Show version information | |
| `--verbose` | | Enable verbose logging | `false` |

## Commands

### `klonekit apply`

Execute the complete KloneKit workflow: scaffold, SCM, and provision.

```bash
klonekit apply --file <blueprint-file> [options]
```

**Options:**

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--file` | `-f` | Path to blueprint YAML file | **Required** |
| `--dry-run` | | Simulate operations without making changes | `false` |
| `--retain-state` | | Keep state files after completion | `false` |

**Examples:**

```bash
# Basic usage
klonekit apply --file klonekit.yaml

# Dry run to preview changes
klonekit apply --file klonekit.yaml --dry-run

# Keep state files for debugging
klonekit apply --file klonekit.yaml --retain-state
```

### `klonekit scaffold`

Generate Terraform files from blueprint and copy source files.

```bash
klonekit scaffold --file <blueprint-file> [options]
```

**Options:**

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--file` | `-f` | Path to blueprint YAML file | **Required** |
| `--dry-run` | | Show what would be generated | `false` |

**Examples:**

```bash
# Generate Terraform files
klonekit scaffold --file klonekit.yaml

# Preview file generation
klonekit scaffold --file klonekit.yaml --dry-run
```

**What it does:**
1. Copies files from `spec.scaffold.source` to `spec.scaffold.destination`
2. Creates `terraform.tfvars.json` with blueprint variables
3. Preserves file permissions and structure

### `klonekit scm`

Create GitLab repository and push the scaffolded files.

```bash
klonekit scm --file <blueprint-file> [options]
```

**Options:**

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--file` | `-f` | Path to blueprint YAML file | **Required** |
| `--dry-run` | | Show what would be done | `false` |

**Examples:**

```bash
# Create repository and push files
klonekit scm --file klonekit.yaml

# Preview SCM operations
klonekit scm --file klonekit.yaml --dry-run
```

**What it does:**
1. Creates GitLab repository using `spec.scm.project` configuration
2. Initializes local git repository in destination directory
3. Commits and pushes all files to GitLab

### `klonekit provision`

Run Terraform in Docker container to provision infrastructure.

```bash
klonekit provision --file <blueprint-file> [options]
```

**Options:**

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--file` | `-f` | Path to blueprint YAML file | **Required** |
| `--dry-run` | | Run terraform plan only | `false` |

**Examples:**

```bash
# Provision infrastructure
klonekit provision --file klonekit.yaml

# Plan only (no changes)
klonekit provision --file klonekit.yaml --dry-run
```

**What it does:**
1. Runs `terraform init` in destination directory
2. Executes `terraform plan` to preview changes
3. Applies configuration with `terraform apply` (unless dry-run)

## Environment Variables

KloneKit responds to these environment variables:

### Authentication

| Variable | Description | Required |
|----------|-------------|----------|
| `GITLAB_PRIVATE_TOKEN` | GitLab Personal Access Token | **Yes** |
| `AWS_ACCESS_KEY_ID` | AWS Access Key ID | **Yes** |
| `AWS_SECRET_ACCESS_KEY` | AWS Secret Access Key | **Yes** |
| `AWS_DEFAULT_REGION` | Default AWS region | No |

### Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `KLONEKIT_LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` |
| `KLONEKIT_LOG_FORMAT` | Log format (text, json) | `text` |
| `TERRAFORM_VERSION` | Terraform Docker image version | `1.8` |

### Terraform Variables

| Variable | Description | Usage |
|----------|-------------|-------|
| `TF_LOG` | Terraform log level | For debugging Terraform |
| `TF_CLI_ARGS` | Additional Terraform arguments | Performance tuning |

**Examples:**

```bash
# Enable debug logging
export KLONEKIT_LOG_LEVEL=debug

# Use specific Terraform version
export TERRAFORM_VERSION=1.7.5

# Increase Terraform parallelism
export TF_CLI_ARGS="-parallelism=20"

# Enable Terraform debugging
export TF_LOG=DEBUG
```

## Exit Codes

KloneKit uses standard exit codes:

| Code | Description |
|------|-------------|
| `0` | Success |
| `1` | General error |
| `2` | Misuse of shell command |
| `64` | Usage error (invalid arguments) |
| `65` | Data format error (invalid blueprint) |
| `69` | Service unavailable (GitLab/AWS API errors) |
| `70` | Internal software error |
| `71` | System error (file I/O, permissions) |
| `77` | Permission denied |
| `78` | Configuration error |

## Blueprint File Location

KloneKit searches for blueprint files in this order:

1. Specified with `--file` option
2. `klonekit.yaml` in current directory
3. `klonekit.yml` in current directory
4. `.klonekit.yaml` in current directory
5. `.klonekit.yml` in current directory

## Docker Integration

KloneKit runs Terraform in Docker containers with these characteristics:

### Volume Mounts
- Blueprint destination directory → `/workspace`
- AWS credentials → Container environment
- User's home directory → For SSH keys and git config

### Network Access
- Container has full internet access
- Can reach AWS APIs
- Can access GitLab APIs

### Resource Limits
Uses Docker's default resource limits. Adjust in Docker Desktop if needed for large Terraform plans.

## Debugging

### Verbose Logging

```bash
klonekit apply --file klonekit.yaml --verbose
```

### Environment Debug

```bash
export KLONEKIT_LOG_LEVEL=debug
export TF_LOG=DEBUG
klonekit apply --file klonekit.yaml
```

### State Inspection

```bash
# Retain files for inspection
klonekit apply --file klonekit.yaml --retain-state

# Examine generated files
ls -la infrastructure/
cat infrastructure/terraform.tfvars.json
```

## Common Patterns

### CI/CD Integration

```bash
# Validate in CI
klonekit apply --file klonekit.yaml --dry-run

# Deploy in CD
klonekit apply --file klonekit.yaml
```

### Multi-Environment

```bash
# Environment-specific blueprints
klonekit apply --file environments/dev.yaml
klonekit apply --file environments/prod.yaml
```

### State Management

```bash
# Keep state for debugging
klonekit apply --file klonekit.yaml --retain-state

# Clean deployment
klonekit apply --file klonekit.yaml
# State files are automatically cleaned up
```

## Troubleshooting Commands

### Test Authentication

```bash
# Test GitLab token
curl -H "PRIVATE-TOKEN: $GITLAB_PRIVATE_TOKEN" https://gitlab.com/api/v4/user

# Test AWS credentials
aws sts get-caller-identity

# Test Docker
docker info
```

### Validate Configuration

```bash
# Check blueprint syntax
python -c "import yaml; yaml.safe_load(open('klonekit.yaml'))"

# Validate Terraform
klonekit scaffold --file klonekit.yaml --dry-run
cd infrastructure && terraform validate
```

### Debug Execution

```bash
# Step-by-step execution
klonekit scaffold --file klonekit.yaml
klonekit scm --file klonekit.yaml
klonekit provision --file klonekit.yaml --dry-run
klonekit provision --file klonekit.yaml
```

## Next Steps

- Review [Blueprint Schema](blueprint-schema.md)
- Explore [Examples](examples.md)
- Check [Troubleshooting Guide](../tasks/troubleshooting.md)