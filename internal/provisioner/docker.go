package provisioner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"klonekit/pkg/blueprint"
	"klonekit/pkg/runtime"
)

const (
	// TerraformDockerImage is the official HashiCorp Terraform Docker image version
	TerraformDockerImage = "hashicorp/terraform:1.8.0"

	// WorkingDirectory is the container working directory
	WorkingDirectory = "/workspace"
)


// TerraformDockerProvisioner implements the Provisioner interface using container runtime.
type TerraformDockerProvisioner struct {
	containerRuntime runtime.ContainerRuntime
	containerName    string // Name for the persistent Terraform container
}

// NewTerraformDockerProvisioner creates a new TerraformDockerProvisioner.
func NewTerraformDockerProvisioner(containerRuntime runtime.ContainerRuntime) *TerraformDockerProvisioner {
	// Generate unique container name for this session
	containerName := fmt.Sprintf("klonekit-terraform-%d", os.Getpid())

	return &TerraformDockerProvisioner{
		containerRuntime: containerRuntime,
		containerName:    containerName,
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
	if err := p.runTerraformCommand(ctx, absScaffoldDir, awsCredsDir, spec.Cloud.Region, false, "init"); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	// Execute Terraform plan for validation
	if err := p.runTerraformCommand(ctx, absScaffoldDir, awsCredsDir, spec.Cloud.Region, false, "plan"); err != nil {
		return fmt.Errorf("terraform plan failed: %w", err)
	}

	// Only execute apply if auto-approve is enabled
	if autoApprove {
		// Backup state file before apply operation (critical for safety)
		if err := p.backupStateFile(absScaffoldDir); err != nil {
			slog.Warn("Failed to backup state file before apply", "error", err.Error())
			// Continue anyway - backup failure shouldn't block apply
		}

		if err := p.runTerraformCommand(ctx, absScaffoldDir, awsCredsDir, spec.Cloud.Region, true, "apply", "-auto-approve"); err != nil {
			return fmt.Errorf("terraform apply failed: %w", err)
		}
		slog.Info("Infrastructure provisioning completed successfully")
	} else {
		slog.Info("Infrastructure validation completed successfully - use --auto-approve to provision")
	}

	return nil
}

// backupStateFile creates a backup of terraform.tfstate before critical operations.
// This prevents permanent state loss in case of failures.
func (p *TerraformDockerProvisioner) backupStateFile(scaffoldDir string) error {
	stateFile := filepath.Join(scaffoldDir, "terraform.tfstate")

	// Check if state file exists
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		slog.Debug("No state file found to backup", "path", stateFile)
		return nil // Not an error - might be first run
	}

	// Create backup with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupFile := filepath.Join(scaffoldDir, fmt.Sprintf("terraform.tfstate.backup.%s", timestamp))

	// Copy state file to backup
	if err := copyFile(stateFile, backupFile); err != nil {
		return fmt.Errorf("failed to backup state file: %w", err)
	}

	slog.Info("State file backed up successfully", "backup", backupFile)
	return nil
}

// validatePath ensures the path is safe and doesn't contain directory traversal sequences
func validatePath(path string) error {
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path contains directory traversal: %s", path)
	}
	return nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	// Validate paths to prevent directory traversal
	if err := validatePath(src); err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}
	if err := validatePath(dst); err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// getCurrentUserID returns the current user ID in format "uid:gid" for Docker containers.
// This ensures files created by containers have the same ownership as the host user,
// preventing permission issues during cleanup.
func getCurrentUserID() string {
	// In most cases, we can use os.Getuid() and os.Getgid()
	uid := os.Getuid()
	gid := os.Getgid()
	return fmt.Sprintf("%d:%d", uid, gid)
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
func (p *TerraformDockerProvisioner) runTerraformCommand(ctx context.Context, scaffoldDir, awsCredsDir, region string, retainContainer bool, args ...string) error {
	// Use args directly since the container's ENTRYPOINT is already 'terraform'
	cmd := args

	slog.Info("Executing Terraform command", "command", append([]string{"terraform"}, cmd...))

	// Create RunOptions for the container
	opts := runtime.RunOptions{
		Image:   TerraformDockerImage,
		Command: cmd,
		VolumeMounts: map[string]string{
			scaffoldDir: WorkingDirectory,
			awsCredsDir: "/home/terraform/.aws", // Use non-root path for AWS credentials
		},
		EnvVars: map[string]string{
			"AWS_SHARED_CREDENTIALS_FILE": "/home/terraform/.aws/credentials",
			"AWS_CONFIG_FILE":             "/home/terraform/.aws/config",
			"AWS_DEFAULT_REGION":          region,
			"AWS_REGION":                  region,
		},
		WorkingDirectory: WorkingDirectory,
		User:             getCurrentUserID(),    // Run container as current user to avoid permission issues
		RetainContainer:  retainContainer,      // Retain container for state persistence
		ContainerName:    p.containerName,      // Use consistent container name
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
		if cerr := reader.Close(); cerr != nil {
			slog.Debug("Error closing container output reader", "error", cerr)
		}
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

// bracketRegex is a compiled regex for bracket-only color codes (Docker log format)
var bracketRegex = regexp.MustCompile(`\[[0-9;]*[a-zA-Z]`)

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

	// Remove bracket-only color codes (common in Docker logs)
	line = bracketRegex.ReplaceAllString(line, "")

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
