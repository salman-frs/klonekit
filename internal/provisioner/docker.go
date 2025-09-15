package provisioner

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"klonekit/pkg/blueprint"
	"klonekit/pkg/runtime"
)

const (
	// TerraformDockerImage is the official HashiCorp Terraform Docker image version
	TerraformDockerImage = "hashicorp/terraform:1.8.0"

	// WorkingDirectory is the container working directory
	WorkingDirectory = "/workspace"
)

// Provisioner defines the interface for infrastructure provisioning operations.
type Provisioner interface {
	Provision(spec *blueprint.Spec) error
}

// TerraformDockerProvisioner implements the Provisioner interface using container runtime.
type TerraformDockerProvisioner struct {
	containerRuntime runtime.ContainerRuntime
}

// NewTerraformDockerProvisioner creates a new TerraformDockerProvisioner.
func NewTerraformDockerProvisioner(containerRuntime runtime.ContainerRuntime) *TerraformDockerProvisioner {
	return &TerraformDockerProvisioner{
		containerRuntime: containerRuntime,
	}
}

// Provision executes Terraform init and optionally apply commands within a Docker container.
// If autoApprove is false, only terraform init and plan will be executed for validation.
func (p *TerraformDockerProvisioner) Provision(spec *blueprint.Spec, autoApprove bool) error {
	ctx := context.Background()

	// Validate that scaffold directory exists
	scaffoldDir := spec.Scaffold.Destination
	if _, err := os.Stat(scaffoldDir); os.IsNotExist(err) {
		return fmt.Errorf("scaffold directory does not exist: %s", scaffoldDir)
	}

	slog.Info("Starting infrastructure provisioning", "scaffoldDir", scaffoldDir)

	// Pull Terraform Docker image
	if err := p.containerRuntime.PullImage(ctx, TerraformDockerImage); err != nil {
		return fmt.Errorf("failed to pull Terraform image: %w", err)
	}

	// Get absolute path of scaffold directory
	absScaffoldDir, err := filepath.Abs(scaffoldDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for scaffold directory: %w", err)
	}

	// Get user's AWS credentials directory
	awsCredsDir, err := p.getAWSCredentialsDir()
	if err != nil {
		return fmt.Errorf("failed to locate AWS credentials directory: %w", err)
	}

	// Execute Terraform init
	if err := p.runTerraformCommand(ctx, absScaffoldDir, awsCredsDir, spec.Cloud.Region, "init"); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	// Execute Terraform plan for validation
	if err := p.runTerraformCommand(ctx, absScaffoldDir, awsCredsDir, spec.Cloud.Region, "plan"); err != nil {
		return fmt.Errorf("terraform plan failed: %w", err)
	}

	// Only execute apply if auto-approve is enabled
	if autoApprove {
		if err := p.runTerraformCommand(ctx, absScaffoldDir, awsCredsDir, spec.Cloud.Region, "apply", "-auto-approve"); err != nil {
			return fmt.Errorf("terraform apply failed: %w", err)
		}
		slog.Info("Infrastructure provisioning completed successfully")
	} else {
		slog.Info("Infrastructure validation completed successfully - use --auto-approve to provision")
	}

	return nil
}

// getAWSCredentialsDir returns the path to the user's AWS credentials directory.
func (p *TerraformDockerProvisioner) getAWSCredentialsDir() (string, error) {
	var homeDir string

	// First try to get HOME from environment variable (respects test overrides)
	if envHome := os.Getenv("HOME"); envHome != "" {
		homeDir = envHome
	} else {
		// Fallback to system user home directory
		currentUser, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("failed to get current user: %w", err)
		}
		homeDir = currentUser.HomeDir
	}

	awsDir := filepath.Join(homeDir, ".aws")

	// Check if AWS credentials directory exists
	if _, err := os.Stat(awsDir); os.IsNotExist(err) {
		return "", fmt.Errorf("AWS credentials directory not found: %s. Please configure AWS credentials", awsDir)
	}

	return awsDir, nil
}

// runTerraformCommand executes a Terraform command using the container runtime.
func (p *TerraformDockerProvisioner) runTerraformCommand(ctx context.Context, scaffoldDir, awsCredsDir, region string, args ...string) error {
	// Use args directly since the container's ENTRYPOINT is already 'terraform'
	cmd := args

	slog.Info("Executing Terraform command", "command", append([]string{"terraform"}, cmd...))

	// Create RunOptions for the container
	opts := runtime.RunOptions{
		Image:   TerraformDockerImage,
		Command: cmd,
		VolumeMounts: map[string]string{
			scaffoldDir: WorkingDirectory,
			awsCredsDir: "/root/.aws",
		},
		EnvVars: map[string]string{
			"AWS_SHARED_CREDENTIALS_FILE": "/root/.aws/credentials",
			"AWS_CONFIG_FILE":             "/root/.aws/config",
			"AWS_DEFAULT_REGION":          region,
			"AWS_REGION":                  region,
		},
		WorkingDirectory: WorkingDirectory,
	}

	// Run the container
	reader, err := p.containerRuntime.RunContainer(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to run container: %w", err)
	}

	// Stream the output
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		// Clean up Docker log output
		cleanLine := cleanDockerLogLine(line)
		if cleanLine != "" {
			slog.Info("Terraform output", "line", cleanLine)
		}
	}

	if err := scanner.Err(); err != nil {
		reader.Close() // Best effort cleanup
		return fmt.Errorf("error reading container output: %w", err)
	}

	// Check container exit status
	if err := reader.Close(); err != nil {
		return fmt.Errorf("terraform command failed: %w", err)
	}

	slog.Info("Terraform command completed successfully", "command", append([]string{"terraform"}, cmd...))
	return nil
}

// ansiRegex is a compiled regex for ANSI escape sequences
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// cleanDockerLogLine removes Docker log headers, ANSI escape sequences, and filters out binary/control characters.
func cleanDockerLogLine(line string) string {
	// Skip empty lines
	if len(line) == 0 {
		return ""
	}

	// Docker log format has 8-byte header: [STREAM_TYPE][0][0][0][SIZE]
	// Remove Docker log header if present
	if len(line) >= 8 {
		// Check if line starts with Docker log header pattern
		if line[0] == 1 || line[0] == 2 { // stdout or stderr stream type
			if len(line) > 8 {
				line = line[8:]
			} else {
				return "" // Header only, no content
			}
		}
	}

	// Remove ANSI escape sequences (colors, formatting, etc.)
	line = ansiRegex.ReplaceAllString(line, "")

	// Remove common control characters
	line = strings.ReplaceAll(line, "\x00", "")
	line = strings.ReplaceAll(line, "\x01", "")
	line = strings.ReplaceAll(line, "\x02", "")
	line = strings.ReplaceAll(line, "\x03", "")

	// Trim whitespace
	line = strings.TrimSpace(line)

	// Skip empty lines after cleaning
	if len(line) == 0 {
		return ""
	}

	// Filter out lines that are mostly binary/control characters
	printableChars := 0
	for _, r := range line {
		if r >= 32 && r <= 126 { // printable ASCII range
			printableChars++
		}
	}

	// If less than 50% printable characters, skip the line
	if len(line) > 0 && float64(printableChars)/float64(len(line)) < 0.5 {
		return ""
	}

	return line
}
