# Installation

This guide covers installing KloneKit on different platforms and verifying the installation.

## Prerequisites

Before installing KloneKit, ensure you have:

- **Docker** installed and running (required for Terraform execution)
- **GitLab Personal Access Token** with API and repository permissions
- **AWS Credentials** configured for infrastructure provisioning

## Installation Methods

### Homebrew (Recommended for macOS)

The easiest way to install KloneKit on macOS:

```bash
brew install salman-frs/klonekit/klonekit
```

### GitHub Releases (All Platforms)

Download pre-built binaries for your platform:

=== "macOS Intel"

    ```bash
    curl -L -o klonekit https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-darwin-amd64
    chmod +x klonekit
    sudo mv klonekit /usr/local/bin/
    ```

=== "macOS Apple Silicon"

    ```bash
    curl -L -o klonekit https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-darwin-arm64
    chmod +x klonekit
    sudo mv klonekit /usr/local/bin/
    ```

=== "Linux"

    ```bash
    curl -L -o klonekit https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-linux-amd64
    chmod +x klonekit
    sudo mv klonekit /usr/local/bin/
    ```

=== "Windows"

    ```powershell
    # Download the Windows executable
    Invoke-WebRequest -Uri "https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-windows-amd64.exe" -OutFile "klonekit.exe"

    # Add to PATH or move to a directory in your PATH
    ```

### From Source

If you have Go 1.22.x or higher installed:

```bash
git clone https://github.com/salman-frs/klonekit.git
cd klonekit
go build -o klonekit ./cmd/klonekit
```

Or using Make:

```bash
make build
```

## Verify Installation

Check that KloneKit is installed correctly:

```bash
klonekit --version
```

You should see version information printed.

Get help with available commands:

```bash
klonekit --help
```

## Docker Setup

KloneKit requires Docker to run Terraform in isolated containers. Verify Docker is working:

```bash
docker info
```

If Docker is not running, start the Docker service:

=== "macOS"

    Start Docker Desktop application.

=== "Linux"

    ```bash
    sudo systemctl start docker
    sudo systemctl enable docker
    ```

=== "Windows"

    Start Docker Desktop application.

## Authentication Setup

### GitLab Personal Access Token

1. Go to GitLab → Settings → Access Tokens
2. Create a new token with these scopes:
   - `api` (full API access)
   - `read_repository`
   - `write_repository`
3. Export the token:

```bash
export GITLAB_PRIVATE_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
```

!!! tip "Persistent Token"
    Add the export command to your shell profile (`.bashrc`, `.zshrc`) to make it persistent.

### AWS Credentials

Configure AWS credentials using one of these methods:

=== "AWS CLI"

    ```bash
    aws configure
    ```

=== "Environment Variables"

    ```bash
    export AWS_ACCESS_KEY_ID="your-access-key"
    export AWS_SECRET_ACCESS_KEY="your-secret-key"
    export AWS_DEFAULT_REGION="us-west-2"
    ```

=== "IAM Roles"

    If running on EC2, use IAM roles for automatic credential management.

## Verification

Test that everything is working:

```bash
# Verify KloneKit can access Docker
klonekit provision --help

# Create a test blueprint to validate setup
klonekit scaffold --help
```

## Troubleshooting

### Common Issues

**"command not found: klonekit"**
- Verify the binary is in your PATH
- Try the full path to the binary
- Check file permissions (should be executable)

**"GITLAB_PRIVATE_TOKEN environment variable is required"**
- Export your GitLab token as shown above
- Verify the token has correct permissions

**"failed to connect to Docker daemon"**
- Ensure Docker is installed and running
- Check Docker permissions for your user
- Try `docker info` to verify Docker access

**"AWS credentials not found"**
- Configure AWS credentials as shown above
- Verify credentials with `aws sts get-caller-identity`

## Next Steps

- Complete the [Basic Setup Tutorial](../tutorials/basic-setup.md)
- Learn about [Configuration Options](configuration.md)
- Review [Blueprint Structure](../concepts/blueprints.md)