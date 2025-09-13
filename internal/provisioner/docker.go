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

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"

	"klonekit/pkg/blueprint"
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

// TerraformDockerProvisioner implements the Provisioner interface using Docker containers.
type TerraformDockerProvisioner struct {
	dockerClient *client.Client
}

// NewTerraformDockerProvisioner creates a new TerraformDockerProvisioner.
func NewTerraformDockerProvisioner() (*TerraformDockerProvisioner, error) {
	// Create Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Check if Docker daemon is accessible
	ctx := context.Background()
	_, err = dockerClient.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker daemon: %w", err)
	}

	return &TerraformDockerProvisioner{
		dockerClient: dockerClient,
	}, nil
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
	if err := p.pullTerraformImage(ctx); err != nil {
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

	// Create container configuration
	containerConfig, hostConfig, err := p.createContainerConfig(absScaffoldDir, awsCredsDir)
	if err != nil {
		return fmt.Errorf("failed to create container configuration: %w", err)
	}

	// Execute Terraform init
	if err := p.runTerraformCommand(ctx, containerConfig, hostConfig, "init"); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	// Execute Terraform apply with auto-approve
	if err := p.runTerraformCommand(ctx, containerConfig, hostConfig, "apply", "-auto-approve"); err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	slog.Info("Infrastructure provisioning completed successfully")
	return nil
}

// pullTerraformImage pulls the Terraform Docker image if not already present.
func (p *TerraformDockerProvisioner) pullTerraformImage(ctx context.Context) error {
	slog.Info("Pulling Terraform Docker image", "image", TerraformDockerImage)

	reader, err := p.dockerClient.ImagePull(ctx, TerraformDockerImage, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", TerraformDockerImage, err)
	}
	defer reader.Close()

	// Stream the pull output (but don't print it to avoid clutter)
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return fmt.Errorf("failed to stream image pull output: %w", err)
	}

	slog.Info("Successfully pulled Terraform Docker image")
	return nil
}

// getAWSCredentialsDir returns the path to the user's AWS credentials directory.
func (p *TerraformDockerProvisioner) getAWSCredentialsDir() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	awsDir := filepath.Join(currentUser.HomeDir, ".aws")

	// Check if AWS credentials directory exists
	if _, err := os.Stat(awsDir); os.IsNotExist(err) {
		return "", fmt.Errorf("AWS credentials directory not found: %s. Please configure AWS credentials", awsDir)
	}

	return awsDir, nil
}

// createContainerConfig creates Docker container and host configurations.
func (p *TerraformDockerProvisioner) createContainerConfig(scaffoldDir, awsCredsDir string) (*container.Config, *container.HostConfig, error) {
	containerConfig := &container.Config{
		Image:      TerraformDockerImage,
		WorkingDir: WorkingDirectory,
		Tty:        true,
		Env: []string{
			"AWS_SHARED_CREDENTIALS_FILE=/root/.aws/credentials",
			"AWS_CONFIG_FILE=/root/.aws/config",
		},
	}

	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: scaffoldDir,
				Target: WorkingDirectory,
			},
			{
				Type:     mount.TypeBind,
				Source:   awsCredsDir,
				Target:   "/root/.aws",
				ReadOnly: true, // Mount AWS credentials as read-only for security
			},
		},
	}

	return containerConfig, hostConfig, nil
}

// runTerraformCommand executes a Terraform command within a Docker container.
func (p *TerraformDockerProvisioner) runTerraformCommand(ctx context.Context, containerConfig *container.Config, hostConfig *container.HostConfig, args ...string) error {
	// Set the command to execute
	cmd := append([]string{"terraform"}, args...)
	containerConfig.Cmd = cmd

	slog.Info("Executing Terraform command", "command", cmd)

	// Create container
	resp, err := p.dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	containerID := resp.ID
	defer func() {
		// Clean up container
		p.dockerClient.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
	}()

	// Start container
	if err := p.dockerClient.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Attach to container to stream output
	if err := p.streamContainerOutput(ctx, containerID); err != nil {
		return fmt.Errorf("failed to stream container output: %w", err)
	}

	// Wait for container to finish
	statusCh, errCh := p.dockerClient.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return fmt.Errorf("terraform command failed with exit code: %d", status.StatusCode)
		}
	}

	slog.Info("Terraform command completed successfully", "command", cmd)
	return nil
}

// streamContainerOutput streams container stdout/stderr to the console in real-time.
func (p *TerraformDockerProvisioner) streamContainerOutput(ctx context.Context, containerID string) error {
	out, err := p.dockerClient.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	})
	if err != nil {
		return fmt.Errorf("failed to get container logs: %w", err)
	}
	defer out.Close()

	// Stream output to stdout in real-time
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := scanner.Text()
		// Docker logs include a header, skip it for clean output
		if len(line) > 8 {
			fmt.Println(line[8:]) // Skip the 8-byte Docker log header
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading container output: %w", err)
	}

	return nil
}
