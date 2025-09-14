# Blueprint Schema

Complete reference for KloneKit blueprint YAML structure and validation rules.

## Schema Overview

```yaml
apiVersion: v1                    # string, required
kind: Blueprint                   # string, required, must be "Blueprint"
metadata:                         # object, required
  name: string                    # required
  description: string             # optional
  labels:                         # optional
    key: value                    # string key-value pairs
spec:                            # object, required
  scm:                           # object, required
    provider: string             # required, must be "gitlab"
    url: string                  # required, GitLab instance URL
    token: string                # required, Personal Access Token
    project:                     # object, required
      name: string               # required, repository name
      namespace: string          # required, GitLab namespace/username
      description: string        # optional, repository description
      visibility: string         # optional, visibility level
  cloud:                         # object, required
    provider: string             # required, must be "aws"
    region: string               # required, AWS region
  scaffold:                      # object, required
    source: string               # required, source directory path
    destination: string          # required, destination directory path
  variables:                     # object, optional
    key: any                     # any type values for Terraform variables
```

## Field Reference

### `apiVersion`

**Type**: `string`
**Required**: Yes
**Valid Values**: `v1`

Specifies the API version of the blueprint schema.

```yaml
apiVersion: v1
```

### `kind`

**Type**: `string`
**Required**: Yes
**Valid Values**: `Blueprint`

Identifies the resource type. Must always be "Blueprint".

```yaml
kind: Blueprint
```

### `metadata`

**Type**: `object`
**Required**: Yes

Contains metadata about the blueprint.

#### `metadata.name`

**Type**: `string`
**Required**: Yes
**Validation**: Must be a valid identifier (alphanumeric, hyphens, underscores)

Unique identifier for the blueprint.

```yaml
metadata:
  name: my-web-application
```

#### `metadata.description`

**Type**: `string`
**Required**: No

Human-readable description of the blueprint.

```yaml
metadata:
  description: "Production web application infrastructure"
```

#### `metadata.labels`

**Type**: `object`
**Required**: No
**Values**: String key-value pairs

Arbitrary labels for organizing and categorizing blueprints.

```yaml
metadata:
  labels:
    environment: production
    team: platform
    version: "1.2.0"
```

### `spec`

**Type**: `object`
**Required**: Yes

Contains the blueprint specifications.

### `spec.scm`

**Type**: `object`
**Required**: Yes

Source control management configuration.

#### `spec.scm.provider`

**Type**: `string`
**Required**: Yes
**Valid Values**: `gitlab`

SCM provider type. Currently only GitLab is supported.

```yaml
spec:
  scm:
    provider: gitlab
```

#### `spec.scm.url`

**Type**: `string`
**Required**: Yes
**Format**: Valid URL

GitLab instance URL.

```yaml
spec:
  scm:
    url: https://gitlab.com           # GitLab.com
    # OR
    url: https://gitlab.company.com   # Self-hosted GitLab
```

#### `spec.scm.token`

**Type**: `string`
**Required**: Yes
**Format**: GitLab Personal Access Token or environment variable

GitLab Personal Access Token with API and repository permissions.

```yaml
spec:
  scm:
    token: glpat-xxxxxxxxxxxxxxxxxxxx        # Direct token (not recommended)
    # OR
    token: ${GITLAB_PRIVATE_TOKEN}           # Environment variable (recommended)
```

#### `spec.scm.project`

**Type**: `object`
**Required**: Yes

GitLab project configuration.

##### `spec.scm.project.name`

**Type**: `string`
**Required**: Yes
**Validation**: Valid GitLab project name

Repository name.

```yaml
spec:
  scm:
    project:
      name: my-infrastructure-project
```

##### `spec.scm.project.namespace`

**Type**: `string`
**Required**: Yes
**Validation**: Valid GitLab namespace (username or group)

GitLab namespace (username or group name) where the repository will be created.

```yaml
spec:
  scm:
    project:
      namespace: my-username
      # OR
      namespace: my-organization
```

##### `spec.scm.project.description`

**Type**: `string`
**Required**: No

Repository description.

```yaml
spec:
  scm:
    project:
      description: "Infrastructure managed by KloneKit"
```

##### `spec.scm.project.visibility`

**Type**: `string`
**Required**: No
**Valid Values**: `private`, `public`, `internal`
**Default**: `private`

Repository visibility level.

```yaml
spec:
  scm:
    project:
      visibility: private    # Only accessible to namespace members
      # OR
      visibility: public     # Publicly accessible
      # OR
      visibility: internal   # Accessible to all logged-in users (GitLab instance)
```

### `spec.cloud`

**Type**: `object`
**Required**: Yes

Cloud provider configuration.

#### `spec.cloud.provider`

**Type**: `string`
**Required**: Yes
**Valid Values**: `aws`

Cloud provider type. Currently only AWS is supported.

```yaml
spec:
  cloud:
    provider: aws
```

#### `spec.cloud.region`

**Type**: `string`
**Required**: Yes
**Validation**: Valid AWS region

AWS region where resources will be provisioned.

```yaml
spec:
  cloud:
    region: us-west-2
    # OR
    region: ${AWS_DEFAULT_REGION}    # Environment variable
```

### `spec.scaffold`

**Type**: `object`
**Required**: Yes

File scaffolding configuration.

#### `spec.scaffold.source`

**Type**: `string`
**Required**: Yes
**Format**: Directory path (relative or absolute)

Source directory containing Terraform templates and other files to be scaffolded.

```yaml
spec:
  scaffold:
    source: ./terraform-templates    # Relative path
    # OR
    source: /path/to/templates       # Absolute path
```

#### `spec.scaffold.destination`

**Type**: `string`
**Required**: Yes
**Format**: Directory path (relative or absolute)

Destination directory where scaffolded files will be generated.

```yaml
spec:
  scaffold:
    destination: ./infrastructure    # Relative path
    # OR
    destination: /tmp/output        # Absolute path
```

### `spec.variables`

**Type**: `object`
**Required**: No
**Values**: Any YAML-compatible type

Variables that will be passed to Terraform as `terraform.tfvars.json`.

```yaml
spec:
  variables:
    # String variables
    environment: production
    project_name: my-app
    aws_region: ${AWS_DEFAULT_REGION}

    # Number variables
    instance_count: 3
    disk_size: 100

    # Boolean variables
    enable_monitoring: true
    encrypt_storage: true

    # List variables
    availability_zones:
      - us-west-2a
      - us-west-2b
      - us-west-2c

    # Object variables
    tags:
      Environment: production
      Owner: platform-team
      Project: my-app

    # Complex nested structures
    database_config:
      engine: postgres
      version: "13.7"
      instance_class: db.r5.large
      storage:
        type: gp2
        size: 100
        encrypted: true
```

## Environment Variable Substitution

Blueprint values support environment variable substitution using `${VAR_NAME}` syntax.

### Supported Locations

Environment variables can be used in:
- `spec.scm.token`
- `spec.scm.url`
- `spec.cloud.region`
- `spec.scaffold.source`
- `spec.scaffold.destination`
- Any value in `spec.variables`

### Syntax

```yaml
# Basic substitution
token: ${GITLAB_PRIVATE_TOKEN}

# With default values (not currently supported, use shell scripts)
# token: ${GITLAB_PRIVATE_TOKEN:-default-value}  # Future feature

# In complex structures
variables:
  aws_region: ${AWS_DEFAULT_REGION}
  database_url: "postgres://user:${DB_PASSWORD}@${DB_HOST}:5432/myapp"
```

### Resolution Rules

1. Environment variables are resolved when the blueprint is loaded
2. Missing environment variables cause validation errors
3. Empty environment variables result in empty strings
4. Variable names are case-sensitive

## Validation Rules

### Blueprint Structure

- Root level must contain `apiVersion`, `kind`, `metadata`, and `spec`
- All required fields must be present
- Field types must match schema definitions

### Name Validation

- `metadata.name` must match pattern: `^[a-zA-Z0-9-_]+$`
- `spec.scm.project.name` must be a valid GitLab project name
- `spec.scm.project.namespace` must be a valid GitLab namespace

### Path Validation

- `spec.scaffold.source` must exist and be readable
- `spec.scaffold.destination` parent directory must exist or be creatable
- Paths can be relative (to blueprint file) or absolute

### Token Validation

- `spec.scm.token` must be a valid GitLab Personal Access Token format
- Token must have required scopes: `api`, `read_repository`, `write_repository`

### Region Validation

- `spec.cloud.region` must be a valid AWS region identifier
- Region must be accessible with provided AWS credentials

## Example Schemas

### Minimal Blueprint

```yaml
apiVersion: v1
kind: Blueprint
metadata:
  name: minimal-example
spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: minimal-infrastructure
      namespace: my-username
  cloud:
    provider: aws
    region: us-west-2
  scaffold:
    source: ./terraform
    destination: ./output
```

### Complete Blueprint

```yaml
apiVersion: v1
kind: Blueprint
metadata:
  name: complete-web-application
  description: "Complete web application infrastructure with database and monitoring"
  labels:
    environment: production
    team: platform
    version: "2.1.0"
    cost-center: engineering

spec:
  scm:
    provider: gitlab
    url: https://gitlab.company.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: web-app-infrastructure
      namespace: platform-team
      description: "Production web application infrastructure managed by KloneKit"
      visibility: private

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ./terraform-templates
    destination: ./infrastructure-output

  variables:
    # Application configuration
    app_name: "web-application"
    environment: "production"
    aws_region: "us-west-2"

    # Networking
    vpc_cidr: "10.0.0.0/16"
    availability_zones:
      - "us-west-2a"
      - "us-west-2b"
      - "us-west-2c"

    # Compute
    web_instance_type: "t3.large"
    web_min_instances: 3
    web_max_instances: 12
    web_desired_capacity: 6

    # Database
    database_config:
      engine: "postgres"
      engine_version: "13.7"
      instance_class: "db.r5.xlarge"
      allocated_storage: 500
      storage_encrypted: true
      backup_retention_period: 30
      multi_az: true

    # Storage
    s3_buckets:
      assets: "web-app-assets-prod"
      backups: "web-app-backups-prod"
      logs: "web-app-logs-prod"

    # Monitoring
    monitoring:
      enable_cloudwatch: true
      enable_x_ray: true
      log_retention_days: 90
      alarm_email: "alerts@company.com"

    # Security
    ssl_certificate_arn: "${SSL_CERT_ARN}"
    allowed_cidr_blocks:
      - "10.0.0.0/8"
      - "172.16.0.0/12"

    # Tags
    common_tags:
      Environment: "production"
      Project: "web-application"
      Team: "platform"
      Owner: "platform-team@company.com"
      CostCenter: "engineering"
      Backup: "required"
      Compliance: "pci-dss"
```

## Schema Validation

KloneKit validates blueprints at runtime. Common validation errors:

### Missing Required Fields

```
Error: missing required field 'metadata.name'
Error: missing required field 'spec.scm.provider'
```

### Invalid Field Values

```
Error: invalid provider 'azure', must be 'gitlab'
Error: invalid region 'invalid-region'
```

### Environment Variable Errors

```
Error: environment variable 'GITLAB_PRIVATE_TOKEN' not set
Error: failed to resolve variable '${MISSING_VAR}'
```

### Path Errors

```
Error: source directory './nonexistent' does not exist
Error: cannot create destination directory './output': permission denied
```

## Next Steps

- Review [CLI Reference](cli.md)
- Explore [Blueprint Examples](examples.md)
- Learn about [Advanced Workflows](../tutorials/advanced-workflows.md)