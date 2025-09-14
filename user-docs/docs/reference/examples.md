# Blueprint Examples

A collection of real-world blueprint examples for common infrastructure patterns.

## Basic Examples

### Simple Web Server

A basic web server with load balancer and auto-scaling.

```yaml title="simple-web-server.yaml"
apiVersion: v1
kind: Blueprint
metadata:
  name: simple-web-server
  description: "Basic web server with auto-scaling"
  labels:
    pattern: web-server
    complexity: basic

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: simple-web-infrastructure
      namespace: my-username
      description: "Simple web server infrastructure"
      visibility: private

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ./templates/web-server
    destination: ./infrastructure

  variables:
    environment: "development"
    instance_type: "t3.micro"
    min_instances: 1
    max_instances: 3
    health_check_path: "/health"
```

### Static Website

S3-hosted static website with CloudFront CDN.

```yaml title="static-website.yaml"
apiVersion: v1
kind: Blueprint
metadata:
  name: static-website
  description: "S3 static website with CloudFront"

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: static-website-infrastructure
      namespace: my-username

  cloud:
    provider: aws
    region: us-east-1  # CloudFront requires us-east-1 for certificates

  scaffold:
    source: ./templates/static-site
    destination: ./infrastructure

  variables:
    domain_name: "example.com"
    subdomain: "www"
    index_document: "index.html"
    error_document: "error.html"
    enable_gzip: true
```

## Database Examples

### RDS PostgreSQL

Managed PostgreSQL database with read replicas.

```yaml title="postgresql-database.yaml"
apiVersion: v1
kind: Blueprint
metadata:
  name: postgresql-database
  description: "RDS PostgreSQL with read replicas"
  labels:
    pattern: database
    engine: postgresql

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: postgresql-infrastructure
      namespace: my-org

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ./templates/rds-postgresql
    destination: ./infrastructure

  variables:
    db_name: "myapp"
    db_username: "appuser"
    db_password: "${DATABASE_PASSWORD}"

    # Instance configuration
    instance_class: "db.r5.large"
    allocated_storage: 100
    max_allocated_storage: 1000
    storage_type: "gp2"
    storage_encrypted: true

    # High availability
    multi_az: true
    backup_retention_period: 7
    backup_window: "03:00-04:00"
    maintenance_window: "sun:04:00-sun:05:00"

    # Read replicas
    read_replica_count: 2
    read_replica_instance_class: "db.r5.large"

    # Monitoring
    monitoring_interval: 60
    performance_insights_enabled: true

    # Network
    vpc_cidr: "10.0.0.0/16"
    db_subnet_cidrs:
      - "10.0.1.0/24"
      - "10.0.2.0/24"
      - "10.0.3.0/24"
```

### DynamoDB Application

NoSQL application with DynamoDB and Lambda.

```yaml title="dynamodb-app.yaml"
apiVersion: v1
kind: Blueprint
metadata:
  name: dynamodb-serverless-app
  description: "Serverless application with DynamoDB"

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: serverless-app-infrastructure
      namespace: my-org

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ./templates/serverless-dynamodb
    destination: ./infrastructure

  variables:
    app_name: "my-serverless-app"
    environment: "production"

    # DynamoDB tables
    tables:
      users:
        hash_key: "user_id"
        billing_mode: "PAY_PER_REQUEST"
        point_in_time_recovery: true
      sessions:
        hash_key: "session_id"
        ttl_attribute: "expires_at"
        billing_mode: "PAY_PER_REQUEST"

    # Lambda functions
    lambda_functions:
      api:
        runtime: "python3.9"
        memory_size: 256
        timeout: 30
      worker:
        runtime: "python3.9"
        memory_size: 512
        timeout: 300

    # API Gateway
    api_gateway:
      stage_name: "prod"
      throttle_rate_limit: 10000
      throttle_burst_limit: 5000
```

## Multi-Tier Applications

### Three-Tier Web Application

Complete web application with presentation, application, and data tiers.

```yaml title="three-tier-webapp.yaml"
apiVersion: v1
kind: Blueprint
metadata:
  name: three-tier-webapp
  description: "Complete three-tier web application"
  labels:
    pattern: three-tier
    complexity: advanced

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: three-tier-webapp-infrastructure
      namespace: platform-team

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ./templates/three-tier-app
    destination: ./infrastructure

  variables:
    app_name: "ecommerce-platform"
    environment: "production"

    # Network configuration
    vpc_cidr: "10.0.0.0/16"
    availability_zones:
      - "us-west-2a"
      - "us-west-2b"
      - "us-west-2c"

    # Presentation tier (ALB + CloudFront)
    presentation:
      enable_waf: true
      ssl_certificate_arn: "${SSL_CERTIFICATE_ARN}"
      cloudfront_price_class: "PriceClass_100"

    # Application tier (ECS Fargate)
    application:
      container_image: "myorg/ecommerce-app:latest"
      cpu: 1024
      memory: 2048
      desired_count: 6
      min_capacity: 3
      max_capacity: 20
      target_cpu_utilization: 70

    # Data tier (RDS + ElastiCache)
    database:
      engine: "postgres"
      engine_version: "13.7"
      instance_class: "db.r5.2xlarge"
      allocated_storage: 500
      multi_az: true
      backup_retention_period: 30

    cache:
      node_type: "cache.r6g.large"
      num_cache_nodes: 3
      engine_version: "6.2"

    # Monitoring and logging
    monitoring:
      enable_container_insights: true
      log_retention_days: 30
      create_dashboard: true
```

### Microservices Platform

Kubernetes-based microservices platform on EKS.

```yaml title="microservices-platform.yaml"
apiVersion: v1
kind: Blueprint
metadata:
  name: microservices-platform
  description: "EKS-based microservices platform"
  labels:
    pattern: microservices
    orchestrator: kubernetes

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: microservices-platform-infrastructure
      namespace: platform-team

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ./templates/eks-microservices
    destination: ./infrastructure

  variables:
    cluster_name: "production-microservices"
    kubernetes_version: "1.28"

    # Network
    vpc_cidr: "10.0.0.0/16"
    availability_zones:
      - "us-west-2a"
      - "us-west-2b"
      - "us-west-2c"

    # EKS configuration
    eks:
      endpoint_private_access: true
      endpoint_public_access: true
      public_access_cidrs: ["0.0.0.0/0"]

    # Node groups
    node_groups:
      system:
        instance_types: ["t3.large"]
        capacity_type: "ON_DEMAND"
        desired_size: 3
        min_size: 3
        max_size: 6
        labels:
          role: "system"
        taints:
          - key: "dedicated"
            value: "system"
            effect: "NO_SCHEDULE"

      application:
        instance_types: ["c5.2xlarge", "c5.4xlarge"]
        capacity_type: "SPOT"
        desired_size: 6
        min_size: 3
        max_size: 20
        labels:
          role: "application"

    # Add-ons
    addons:
      enable_aws_load_balancer_controller: true
      enable_cluster_autoscaler: true
      enable_metrics_server: true
      enable_cert_manager: true

    # Monitoring
    monitoring:
      enable_prometheus: true
      enable_grafana: true
      enable_jaeger: true
      enable_fluentd: true
```

## Data Platform Examples

### Data Lake

S3-based data lake with Glue catalog and Athena.

```yaml title="data-lake.yaml"
apiVersion: v1
kind: Blueprint
metadata:
  name: analytics-data-lake
  description: "S3 data lake with AWS Glue and Athena"
  labels:
    pattern: data-lake
    use-case: analytics

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: data-lake-infrastructure
      namespace: data-team

  cloud:
    provider: aws
    region: us-east-1

  scaffold:
    source: ./templates/data-lake
    destination: ./infrastructure

  variables:
    data_lake_name: "company-data-lake"
    environment: "production"

    # S3 buckets
    buckets:
      raw: "company-data-lake-raw-prod"
      processed: "company-data-lake-processed-prod"
      curated: "company-data-lake-curated-prod"
      athena_results: "company-athena-results-prod"

    # Data lifecycle
    lifecycle_rules:
      raw_data:
        transition_to_ia_days: 30
        transition_to_glacier_days: 90
        expiration_days: 2555  # 7 years
      processed_data:
        transition_to_ia_days: 90
        transition_to_glacier_days: 365

    # Glue catalog
    glue:
      database_name: "company_data_lake"
      crawler_schedule: "cron(0 2 * * ? *)"  # Daily at 2 AM

    # Athena
    athena:
      workgroup_name: "data-analytics"
      query_result_encryption: true

    # Access control
    data_access_roles:
      - name: "data-analysts"
        permissions: ["SELECT"]
        databases: ["company_data_lake"]
      - name: "data-engineers"
        permissions: ["ALL"]
        databases: ["company_data_lake"]
```

### Real-time Analytics

Kinesis-based real-time data processing pipeline.

```yaml title="realtime-analytics.yaml"
apiVersion: v1
kind: Blueprint
metadata:
  name: realtime-analytics-pipeline
  description: "Kinesis real-time data processing"

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: realtime-analytics-infrastructure
      namespace: data-team

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ./templates/kinesis-analytics
    destination: ./infrastructure

  variables:
    pipeline_name: "user-behavior-analytics"

    # Kinesis Data Streams
    data_streams:
      clickstream:
        shard_count: 10
        retention_period: 168  # 7 days
      user_events:
        shard_count: 5
        retention_period: 168

    # Kinesis Data Firehose
    delivery_streams:
      s3_archive:
        destination: "s3"
        s3_bucket: "analytics-archive-prod"
        buffering_size: 128  # MB
        buffering_interval: 60  # seconds
        compression: "GZIP"
      redshift_load:
        destination: "redshift"
        cluster_jdbcurl: "${REDSHIFT_JDBC_URL}"
        copy_options: "JSON 'auto'"

    # Kinesis Analytics
    analytics_applications:
      real_time_dashboard:
        sql_query: |
          CREATE STREAM aggregated_metrics AS
          SELECT
            ROWTIME_RANGE_START,
            ROWTIME_RANGE_END,
            user_type,
            COUNT(*) as event_count,
            AVG(session_duration) as avg_session_duration
          FROM SOURCE_SQL_STREAM_001
          GROUP BY
            user_type,
            RANGE (INTERVAL '1' MINUTE);

    # Lambda processors
    lambda_functions:
      data_enrichment:
        runtime: "python3.9"
        memory_size: 512
        timeout: 300
        environment_variables:
          ENRICHMENT_API_URL: "${ENRICHMENT_API_URL}"
```

## Security-Hardened Examples

### Zero-Trust Network

Security-focused infrastructure with zero-trust networking.

```yaml title="zero-trust-network.yaml"
apiVersion: v1
kind: Blueprint
metadata:
  name: zero-trust-infrastructure
  description: "Security-hardened zero-trust network"
  labels:
    pattern: security
    compliance: soc2

spec:
  scm:
    provider: gitlab
    url: https://gitlab.company.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: zero-trust-infrastructure
      namespace: security-team
      visibility: private

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ./templates/zero-trust
    destination: ./infrastructure

  variables:
    environment: "production"
    compliance_standard: "SOC2"

    # Network segmentation
    vpc_cidr: "10.0.0.0/16"
    network_acls:
      dmz:
        ingress_rules:
          - rule_number: 100
            protocol: "tcp"
            rule_action: "allow"
            port_range: "443"
            cidr_block: "0.0.0.0/0"
      private:
        ingress_rules:
          - rule_number: 100
            protocol: "tcp"
            rule_action: "allow"
            port_range: "443"
            cidr_block: "10.0.0.0/16"

    # WAF rules
    waf:
      enable_aws_managed_rules: true
      enable_ip_reputation: true
      enable_known_bad_inputs: true
      rate_limit: 2000

    # Security monitoring
    security:
      enable_cloudtrail: true
      enable_config: true
      enable_guardduty: true
      enable_security_hub: true
      enable_vpc_flow_logs: true

      # Encryption
      kms_key_rotation: true
      s3_bucket_encryption: "AES256"
      ebs_encryption_by_default: true

      # Access control
      enforce_mfa: true
      password_policy:
        min_length: 14
        require_uppercase: true
        require_lowercase: true
        require_numbers: true
        require_symbols: true
        max_age: 90
```

## Development Environments

### Development Sandbox

Cost-optimized development environment.

```yaml title="dev-sandbox.yaml"
apiVersion: v1
kind: Blueprint
metadata:
  name: development-sandbox
  description: "Cost-optimized development environment"
  labels:
    environment: development
    cost-optimized: true

spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_PRIVATE_TOKEN}
    project:
      name: dev-sandbox-infrastructure
      namespace: dev-team

  cloud:
    provider: aws
    region: us-west-2

  scaffold:
    source: ./templates/dev-environment
    destination: ./infrastructure

  variables:
    environment: "development"

    # Cost optimization
    instance_types:
      web: "t3.micro"      # Burstable performance
      database: "db.t3.micro"  # Burstable database

    auto_shutdown:
      enable: true
      schedule: "cron(0 19 * * 1-5)"  # Shutdown at 7 PM weekdays
      startup: "cron(0 8 * * 1-5)"    # Start at 8 AM weekdays

    # Simplified networking
    single_az_deployment: true
    nat_gateway_count: 1  # Single NAT gateway

    # Development features
    enable_ssh_access: true
    ssh_key_name: "dev-team-key"
    allowed_ssh_cidrs: ["10.0.0.0/8"]

    # Monitoring (minimal)
    enable_detailed_monitoring: false
    log_retention_days: 7
```

## Next Steps

- Study the [Blueprint Schema](blueprint-schema.md) for detailed field reference
- Review [CLI Reference](cli.md) for command usage
- Try the [Basic Setup Tutorial](../tutorials/basic-setup.md) with these examples
- Explore [Advanced Workflows](../tutorials/advanced-workflows.md) for complex scenarios