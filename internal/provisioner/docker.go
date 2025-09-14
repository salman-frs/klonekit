package provisioner

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"

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

// Provision executes Terraform init and apply commands within a Docker container.
func (p *TerraformDockerProvisioner) Provision(spec *blueprint.Spec) error {
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
	if err := p.runTerraformCommand(ctx, absScaffoldDir, awsCredsDir, "init"); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	// Execute Terraform apply with auto-approve
	if err := p.runTerraformCommand(ctx, absScaffoldDir, awsCredsDir, "apply", "-auto-approve"); err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	slog.Info("Infrastructure provisioning completed successfully")
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
func (p *TerraformDockerProvisioner) runTerraformCommand(ctx context.Context, scaffoldDir, awsCredsDir string, args ...string) error {
	// Build the terraform command
	cmd := append([]string{"terraform"}, args...)

	slog.Info("Executing Terraform command", "command", cmd)

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
		},
		WorkingDirectory: WorkingDirectory,
	}

	// Run the container
	reader, err := p.containerRuntime.RunContainer(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to run container: %w", err)
	}
	defer reader.Close()

	// Stream the output
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip Docker log headers for clean output
		if len(line) > 8 {
			line = line[8:] // Remove Docker log header
		}
		slog.Info("Terraform output", "line", line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading container output: %w", err)
	}

	slog.Info("Terraform command completed successfully", "command", cmd)
	return nil
}
