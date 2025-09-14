# Troubleshooting

Common issues and solutions when using KloneKit.

## Installation Issues

### Command Not Found

**Problem**: `command not found: klonekit`

**Solutions**:

1. Verify installation path:
   ```bash
   which klonekit
   echo $PATH
   ```

2. Check file permissions:
   ```bash
   ls -la /usr/local/bin/klonekit
   chmod +x /usr/local/bin/klonekit
   ```

3. Try full path:
   ```bash
   /usr/local/bin/klonekit --version
   ```

### Docker Issues

**Problem**: `failed to connect to Docker daemon`

**Solutions**:

1. Check Docker status:
   ```bash
   docker info
   docker ps
   ```

2. Start Docker service:
   ```bash
   # macOS/Windows: Start Docker Desktop
   # Linux:
   sudo systemctl start docker
   ```

3. Check Docker permissions:
   ```bash
   sudo usermod -aG docker $USER
   # Logout and login again
   ```

## Authentication Issues

### GitLab Token Issues

**Problem**: `401 Unauthorized` or `GITLAB_PRIVATE_TOKEN environment variable is required`

**Solutions**:

1. Verify token is set:
   ```bash
   echo $GITLAB_PRIVATE_TOKEN
   ```

2. Test token manually:
   ```bash
   curl -H "PRIVATE-TOKEN: $GITLAB_PRIVATE_TOKEN" \
        https://gitlab.com/api/v4/user
   ```

3. Check token permissions:
   - Ensure token has `api`, `read_repository`, and `write_repository` scopes
   - Verify token hasn't expired

4. For self-hosted GitLab, verify URL:
   ```yaml
   spec:
     scm:
       url: https://gitlab.company.com  # Not gitlab.com
   ```

### AWS Credential Issues

**Problem**: `AWS credentials not found` or `AccessDenied`

**Solutions**:

1. Verify credentials:
   ```bash
   aws sts get-caller-identity
   ```

2. Check environment variables:
   ```bash
   echo $AWS_ACCESS_KEY_ID
   echo $AWS_SECRET_ACCESS_KEY
   echo $AWS_DEFAULT_REGION
   ```

3. Verify IAM permissions:
   - EC2 permissions for instance creation
   - VPC permissions for networking resources
   - IAM permissions if creating roles

## Blueprint Issues

### YAML Syntax Errors

**Problem**: `yaml: unmarshal errors` or parsing failures

**Solutions**:

1. Validate YAML syntax:
   ```bash
   # Online YAML validator or
   python -c "import yaml; yaml.safe_load(open('klonekit.yaml'))"
   ```

2. Check indentation (use spaces, not tabs):
   ```yaml
   spec:
     scm:  # 2 spaces
       provider: gitlab  # 4 spaces
   ```

3. Verify required fields:
   ```yaml
   apiVersion: v1           # Required
   kind: Blueprint          # Required
   metadata:               # Required
     name: project-name    # Required
   spec:                   # Required
     scm: {}              # Required
     cloud: {}            # Required
     scaffold: {}         # Required
   ```

### Variable Substitution Issues

**Problem**: Variables not substituted or `${VAR_NAME}` appears literally

**Solutions**:

1. Verify environment variables are exported:
   ```bash
   export GITLAB_PRIVATE_TOKEN="your-token"
   env | grep GITLAB
   ```

2. Use correct syntax:
   ```yaml
   # Correct
   token: ${GITLAB_PRIVATE_TOKEN}

   # Incorrect
   token: $GITLAB_PRIVATE_TOKEN
   token: "${GITLAB_PRIVATE_TOKEN}"
   ```

## Execution Issues

### Scaffold Failures

**Problem**: `source directory not found` or file copy errors

**Solutions**:

1. Verify source directory exists:
   ```bash
   ls -la ./terraform-templates/
   ```

2. Check source path in blueprint:
   ```yaml
   spec:
     scaffold:
       source: ./terraform-templates  # Relative to blueprint file
   ```

3. Ensure source contains Terraform files:
   ```bash
   find ./terraform-templates -name "*.tf"
   ```

### SCM Failures

**Problem**: Repository creation or push failures

**Solutions**:

1. Check repository doesn't already exist:
   - Visit GitLab web interface
   - Try a different repository name

2. Verify namespace permissions:
   ```yaml
   spec:
     scm:
       project:
         namespace: correct-username-or-org
   ```

3. Check GitLab instance connectivity:
   ```bash
   curl https://gitlab.com/api/v4/version
   ```

### Provisioning Failures

**Problem**: Terraform execution errors in Docker

**Solutions**:

1. Check Terraform syntax locally:
   ```bash
   cd infrastructure/
   terraform validate
   ```

2. Verify AWS permissions for resources:
   ```bash
   # Test EC2 permissions
   aws ec2 describe-instances --region us-west-2
   ```

3. Check Docker can pull Terraform image:
   ```bash
   docker pull hashicorp/terraform:1.8
   ```

4. Enable detailed logging:
   ```bash
   export TF_LOG=DEBUG
   klonekit provision --file klonekit.yaml
   ```

## Common Error Messages

### "Repository already exists"

**Solution**: Either delete the existing repository or use a different name:

```yaml
spec:
  scm:
    project:
      name: my-project-v2  # Different name
```

### "Terraform state lock"

**Problem**: Previous Terraform run left a state lock

**Solutions**:

1. Wait for lock to expire (usually 20 minutes)
2. Force unlock (dangerous):
   ```bash
   cd infrastructure/
   terraform force-unlock LOCK_ID
   ```

### "Docker volume mount failed"

**Problem**: Permission issues with Docker volume mounts

**Solutions**:

1. Check directory permissions:
   ```bash
   chmod -R 755 ./infrastructure/
   ```

2. Ensure destination directory is writable:
   ```bash
   ls -la infrastructure/
   ```

## Getting Help

### Debug Mode

Enable verbose logging:

```bash
export KLONEKIT_LOG_LEVEL=debug
klonekit apply --file klonekit.yaml --verbose
```

### Dry Run

Test without making changes:

```bash
klonekit apply --file klonekit.yaml --dry-run
```

### State Inspection

Retain files for debugging:

```bash
klonekit apply --file klonekit.yaml --retain-state
ls -la infrastructure/
```

### Manual Verification

Test individual components:

```bash
# Test GitLab API
curl -H "PRIVATE-TOKEN: $GITLAB_PRIVATE_TOKEN" \
     https://gitlab.com/api/v4/projects

# Test AWS
aws ec2 describe-regions

# Test Docker
docker run --rm hashicorp/terraform:1.8 version
```

## Performance Issues

### Slow Terraform Execution

**Solutions**:

1. Increase parallelism:
   ```bash
   export TF_CLI_ARGS="-parallelism=20"
   ```

2. Use local state instead of remote state for testing
3. Optimize Terraform configuration

### Large File Transfers

**Solutions**:

1. Use `.gitignore` to exclude unnecessary files
2. Compress large binary files
3. Consider using Git LFS for large files

## Still Need Help?

If you're still experiencing issues:

1. Check the [GitHub Issues](https://github.com/salman-frs/klonekit/issues)
2. Review [Configuration Guide](configuration.md)
3. Create a new issue with:
   - KloneKit version (`klonekit --version`)
   - Operating system
   - Blueprint configuration (sanitized)
   - Complete error message
   - Steps to reproduce