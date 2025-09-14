# Installation

KloneKit provides several installation methods to suit different environments and preferences.

## Homebrew (macOS/Linux - Recommended)

Install KloneKit using Homebrew for the easiest setup and automatic updates:

```bash
brew install salman-frs/klonekit/klonekit
```

### Verify Homebrew Installation

```bash
klonekit --help
```

### Upgrading via Homebrew

```bash
brew upgrade salman-frs/klonekit/klonekit
```

## GitHub Releases

The easiest way to install KloneKit is by downloading pre-compiled binaries from the GitHub Releases page.

### Download and Install

1. Visit the [KloneKit Releases page](https://github.com/salman-frs/klonekit/releases)
2. Download the appropriate binary for your operating system and architecture:

**For macOS:**
```bash
# Intel Macs
curl -L -o klonekit https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-darwin-amd64
chmod +x klonekit
sudo mv klonekit /usr/local/bin/

# Apple Silicon Macs (M1/M2/M3)
curl -L -o klonekit https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-darwin-arm64
chmod +x klonekit
sudo mv klonekit /usr/local/bin/
```

**For Linux:**
```bash
# AMD64/x86_64 systems
curl -L -o klonekit https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-linux-amd64
chmod +x klonekit
sudo mv klonekit /usr/local/bin/

# ARM64 systems
curl -L -o klonekit https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-linux-arm64
chmod +x klonekit
sudo mv klonekit /usr/local/bin/
```

**For Windows:**
```powershell
# Download using PowerShell
Invoke-WebRequest -Uri "https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-windows-amd64.exe" -OutFile "klonekit.exe"

# Move to a directory in your PATH, for example:
Move-Item klonekit.exe C:\Windows\System32\
```

Or manually:
1. Download `klonekit-windows-amd64.exe` from the releases page
2. Rename it to `klonekit.exe`
3. Move it to a directory in your PATH (e.g., `C:\Windows\System32\`)

### Verify Installation

After installation, verify KloneKit is working:

```bash
klonekit --help
```

You should see the KloneKit help output with available commands.

## Build from Source

If you prefer to build KloneKit from source or contribute to development:

### Prerequisites

- **Go 1.22.x or higher** - [Download Go](https://golang.org/dl/)
- **Git** - For cloning the repository
- **Make** (optional) - For using the Makefile

### Build Steps

1. **Clone the repository:**
```bash
git clone https://github.com/salman-frs/klonekit.git
cd klonekit
```

2. **Build using Go directly:**
```bash
go build -o klonekit ./cmd/klonekit
```

3. **Or build using Make:**
```bash
make build
```

4. **Install to system PATH:**
```bash
sudo cp klonekit /usr/local/bin/  # macOS/Linux
# or
copy klonekit.exe C:\Windows\System32\  # Windows
```

### Development Build

For development with all tools:
```bash
# Install development dependencies
make install-deps

# Run tests
make test

# Run linting
make lint

# Build and test
make all
```

## System Requirements

### Operating Systems
- **macOS**: 10.15 (Catalina) or later
- **Linux**: Any modern distribution (Ubuntu 18.04+, CentOS 7+, etc.)
- **Windows**: Windows 10 or later

### Architecture Support
- **AMD64/x86_64** (Intel/AMD 64-bit processors)
- **ARM64** (Apple Silicon, ARM64 processors)

### Dependencies

KloneKit is distributed as a single statically-linked binary with no external dependencies for basic functionality. However, you'll need:

- **Docker** - Required for Terraform execution (KloneKit runs Terraform in containers)
- **Internet connectivity** - For GitLab API calls and Docker image downloads

## Post-Installation Setup

### 1. Verify Docker Installation

KloneKit requires Docker for Terraform execution:

```bash
docker --version
docker info
```

If Docker isn't installed, visit [Docker's installation guide](https://docs.docker.com/get-docker/).

### 2. Set up GitLab Authentication

You'll need a GitLab Personal Access Token:

1. Go to GitLab.com → User Settings → Access Tokens
2. Create a token with these scopes:
   - `api` (for repository creation)
   - `read_repository` and `write_repository` (for pushing code)

3. Set the token as an environment variable:
```bash
export GITLAB_PRIVATE_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
```

Add this to your shell profile (`.bashrc`, `.zshrc`, etc.) to make it permanent.

### 3. Configure AWS Credentials

For infrastructure provisioning, configure AWS credentials:

```bash
# Option 1: AWS CLI
aws configure

# Option 2: Environment variables
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_DEFAULT_REGION="us-west-2"
```

### 4. Test Installation

Run a quick test to ensure everything works:

```bash
klonekit --help
```

## Upgrading KloneKit

### From GitHub Releases

To upgrade to the latest version:

1. Check your current version by trying different commands to see what's available
2. Download the latest release using the same method as installation
3. Replace the existing binary

### From Source

If you built from source:

```bash
cd klonekit
git pull origin main
make build
sudo cp klonekit /usr/local/bin/
```

## Troubleshooting Installation

### Command Not Found

If you get a "command not found" error after installation:

**macOS/Linux:**
1. Verify the binary is in your PATH: `echo $PATH`
2. Check if `/usr/local/bin` is included
3. Try the full path: `/usr/local/bin/klonekit --help`
4. Add to PATH if needed: `export PATH="/usr/local/bin:$PATH"`

**Windows:**
1. Ensure the directory containing `klonekit.exe` is in your PATH
2. Try running from the full path: `C:\path\to\klonekit.exe --help`

### Permission Denied (macOS)

On macOS, you may get a security warning for unsigned binaries:

1. Go to **System Preferences** → **Security & Privacy**
2. In the **General** tab, click **"Allow Anyway"** next to the KloneKit warning
3. Or run: `sudo spctl --add /usr/local/bin/klonekit`

### Permission Denied (Linux)

If you get permission errors:

```bash
chmod +x klonekit
sudo mv klonekit /usr/local/bin/
```

### Binary Won't Run

**Check architecture compatibility:**
```bash
# On macOS/Linux
file /usr/local/bin/klonekit
uname -m

# Ensure the binary matches your system architecture
```

**Check for missing dependencies (if building from source):**
```bash
ldd klonekit  # Linux
otool -L klonekit  # macOS
```

### Docker Issues

If KloneKit can't connect to Docker:

```bash
# Check Docker is running
docker info

# Check Docker permissions (Linux)
sudo usermod -aG docker $USER
# Then log out and back in

# Test Docker access
docker run hello-world
```

## Alternative Installation Methods

### Using wget instead of curl

If you don't have curl installed:

```bash
# Linux/macOS
wget -O klonekit https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-linux-amd64
chmod +x klonekit
sudo mv klonekit /usr/local/bin/
```

### Installing to User Directory

If you don't have sudo access:

```bash
# Create local bin directory
mkdir -p ~/.local/bin

# Download and install
curl -L -o ~/.local/bin/klonekit https://github.com/salman-frs/klonekit/releases/latest/download/klonekit-linux-amd64
chmod +x ~/.local/bin/klonekit

# Add to PATH (add to ~/.bashrc or ~/.zshrc)
export PATH="$HOME/.local/bin:$PATH"
```

## Uninstalling KloneKit

To remove KloneKit:

```bash
# Remove the binary
sudo rm /usr/local/bin/klonekit

# Remove configuration (if any)
rm -rf ~/.klonekit  # If any config files exist

# Remove environment variables from your shell profile
# Edit ~/.bashrc, ~/.zshrc, etc. and remove GITLAB_PRIVATE_TOKEN
```

## Getting Help

If you encounter installation issues:

1. Check the [troubleshooting section](#troubleshooting-installation) above
2. Visit the [GitHub Issues](https://github.com/salman-frs/klonekit/issues) page
3. Join the community discussions on GitHub

## Next Steps

Once KloneKit is installed, continue to the [Getting Started Guide](getting-started.md) to learn how to:
- Set up your first blueprint
- Configure GitLab and AWS credentials
- Deploy your first infrastructure project