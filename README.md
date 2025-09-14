# KloneKit

**A blueprint-driven DevOps automation tool for GitLab and AWS infrastructure workflows.**

KloneKit simplifies the process of creating GitLab repositories and provisioning AWS infrastructure by orchestrating Terraform deployments through a single blueprint configuration file. It automates the typical DevOps workflow: scaffold Terraform files, create GitLab projects, and provision cloud infrastructure.

## 📖 Documentation

**[View the complete documentation →](https://salman-frs.github.io/klonekit/)**

The full documentation includes:
- **[Installation Guide](https://salman-frs.github.io/klonekit/tasks/installation/)** - Get started with KloneKit
- **[Tutorials](https://salman-frs.github.io/klonekit/tutorials/basic-setup/)** - Step-by-step walkthroughs
- **[CLI Reference](https://salman-frs.github.io/klonekit/reference/cli/)** - Complete command documentation
- **[Examples](https://salman-frs.github.io/klonekit/reference/examples/)** - Blueprint templates and patterns

## Features

- **Declarative Setup** - Define your entire infrastructure and GitLab project in a single YAML blueprint
- **GitLab Integration** - Automatically create repositories, push code, and manage GitLab projects
- **Terraform Orchestration** - Run Terraform in Docker containers for consistent, isolated deployments
- **Resilient Execution** - Resume interrupted workflows with stateful execution and built-in error recovery
- **Workflow Orchestration** - Execute scaffolding, SCM, and provisioning steps individually or all together
- **Dry-Run Support** - Preview changes before execution across all commands

## Quick Install

```bash
# macOS (Homebrew)
brew install salman-frs/klonekit/klonekit

# Or download directly
curl -L -o klonekit https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64
chmod +x klonekit && sudo mv klonekit /usr/local/bin/
```

> **See the [Installation Guide](https://salman-frs.github.io/klonekit/tasks/installation/) for detailed instructions**

## Quick Start

1. **Install KloneKit** and set up authentication:
   ```bash
   export GITLAB_PRIVATE_TOKEN="your-gitlab-token"
   ```

2. **Create a blueprint** (`klonekit.yaml`):
   ```yaml
   apiVersion: v1
   kind: Blueprint
   metadata:
     name: my-project
   spec:
     scm:
       provider: gitlab
       token: ${GITLAB_PRIVATE_TOKEN}
       project:
         name: my-infrastructure
         namespace: your-username
     cloud:
       provider: aws
       region: us-west-2
     scaffold:
       source: ./terraform
       destination: ./infrastructure
     variables:
       environment: development
   ```

3. **Run KloneKit**:
   ```bash
   klonekit apply --file klonekit.yaml
   ```

> **Follow the [Basic Setup Tutorial](https://salman-frs.github.io/klonekit/tutorials/basic-setup/) for a complete walkthrough**

## Commands

- **`klonekit apply`** - Complete workflow (scaffold + scm + provision)
- **`klonekit scaffold`** - Generate files from blueprint
- **`klonekit scm`** - Create GitLab repository and push files
- **`klonekit provision`** - Run Terraform in Docker

> **See the [CLI Reference](https://salman-frs.github.io/klonekit/reference/cli/) for detailed command documentation**

## Prerequisites

- GitLab Personal Access Token with API permissions
- Docker installed and running
- AWS credentials configured

## Getting Help

```bash
klonekit --help                    # Show available commands
klonekit apply --help             # Command-specific help
```

**Need more help?** Check the [troubleshooting guide](https://salman-frs.github.io/klonekit/tasks/troubleshooting/) or browse [examples](https://salman-frs.github.io/klonekit/reference/examples/).

## Development

Requirements: Go 1.22.x, Docker, Make

```bash
make build          # Build binary
make test           # Run tests
make lint           # Run linting
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.