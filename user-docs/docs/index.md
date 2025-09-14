# Welcome to KloneKit Documentation

**A blueprint-driven DevOps automation tool for GitLab and AWS infrastructure workflows.**

KloneKit simplifies the process of creating GitLab repositories and provisioning AWS infrastructure by orchestrating Terraform deployments through a single blueprint configuration file. It automates the typical DevOps workflow: scaffold Terraform files, create GitLab projects, and provision cloud infrastructure.

## Features

- **Declarative Setup** - Define your entire infrastructure and GitLab project in a single YAML blueprint
- **GitLab Integration** - Automatically create repositories, push code, and manage GitLab projects
- **Terraform Orchestration** - Run Terraform in Docker containers for consistent, isolated deployments
- **Resilient Execution** - Resume interrupted workflows with stateful execution and built-in error recovery
- **Workflow Orchestration** - Execute scaffolding, SCM, and provisioning steps individually or all together
- **Dry-Run Support** - Preview changes before execution across all commands

## Quick Navigation

<div class="grid cards" markdown>

-   :material-lightbulb-outline:{ .lg .middle } __Concepts__

    ---

    Understand the core ideas and architecture behind KloneKit

    [:octicons-arrow-right-24: Learn concepts](concepts/overview.md)

-   :material-format-list-checks:{ .lg .middle } __Tasks__

    ---

    Step-by-step guides for common operations and configurations

    [:octicons-arrow-right-24: View tasks](tasks/installation.md)

-   :material-school:{ .lg .middle } __Tutorials__

    ---

    Complete walkthroughs from basic setup to advanced workflows

    [:octicons-arrow-right-24: Start tutorials](tutorials/basic-setup.md)

-   :material-book-open-variant:{ .lg .middle } __Reference__

    ---

    Complete CLI documentation and configuration reference

    [:octicons-arrow-right-24: Browse reference](reference/cli.md)

</div>

## Getting Started

New to KloneKit? Start here:

1. **[Install KloneKit](tasks/installation.md)** - Get KloneKit installed on your system
2. **[Basic Setup](tutorials/basic-setup.md)** - Create your first blueprint and deploy infrastructure
3. **[Configuration Guide](tasks/configuration.md)** - Configure authentication and advanced options

## Need Help?

- Browse the [Tasks](tasks/installation.md) section for how-to guides
- Check [Troubleshooting](tasks/troubleshooting.md) for common issues
- View [Examples](reference/examples.md) for blueprint templates