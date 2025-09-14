package runtime

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"

	"klonekit/pkg/runtime"
)

// DockerRuntime implements the ContainerRuntime interface using Docker client.
type DockerRuntime struct {
	client *client.Client
}

// NewDockerRuntime creates a new DockerRuntime instance using client.FromEnv.
func NewDockerRuntime() (*DockerRuntime, error) {
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

	return &DockerRuntime{
		client: dockerClient,
	}, nil
}

// PullImage pulls a Docker image.
func (d *DockerRuntime) PullImage(ctx context.Context, imageName string) error {
	slog.Info("Pulling Docker image", "image", imageName)

	reader, err := d.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	// Stream the pull output (but don't print it to avoid clutter)
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return fmt.Errorf("failed to stream image pull output: %w", err)
	}

	slog.Info("Successfully pulled Docker image", "image", imageName)
	return nil
}

// RunContainer runs a container and returns the output reader.
func (d *DockerRuntime) RunContainer(ctx context.Context, opts runtime.RunOptions) (io.ReadCloser, error) {
	slog.Info("Running container", "image", opts.Image, "command", opts.Command)

	// Create volume mounts
	var mounts []mount.Mount
	for hostPath, containerPath := range opts.VolumeMounts {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: hostPath,
			Target: containerPath,
		})
	}

	// Convert env vars to slice format
	var envVars []string
	for key, value := range opts.EnvVars {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}

	// Create container configuration
	containerConfig := &container.Config{
		Image:      opts.Image,
		Cmd:        opts.Command,
		Env:        envVars,
		WorkingDir: opts.WorkingDirectory,
	}

	hostConfig := &container.HostConfig{
		Mounts: mounts,
	}

	// Create container
	resp, err := d.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	containerID := resp.ID

	// Start container
	if err := d.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		// Clean up on start failure
		if removeErr := d.client.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); removeErr != nil {
			slog.Error("Failed to remove container after start failure", "containerID", containerID, "error", removeErr)
		}
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Create a reader that will automatically clean up the container when closed
	return &containerReader{
		client:      d.client,
		containerID: containerID,
		ctx:         ctx,
	}, nil
}

// containerReader wraps container output and handles cleanup.
type containerReader struct {
	client      *client.Client
	containerID string
	ctx         context.Context
	reader      io.ReadCloser
	closed      bool
}

// Read reads from the container output.
func (cr *containerReader) Read(p []byte) (n int, err error) {
	if cr.reader == nil {
		// Initialize the reader on first read
		logs, err := cr.client.ContainerLogs(cr.ctx, cr.containerID, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
		})
		if err != nil {
			return 0, fmt.Errorf("failed to get container logs: %w", err)
		}
		cr.reader = logs
	}

	return cr.reader.Read(p)
}

// Close closes the reader and cleans up the container.
func (cr *containerReader) Close() error {
	if cr.closed {
		return nil
	}
	cr.closed = true

	// Close the reader if it exists
	if cr.reader != nil {
		cr.reader.Close()
	}

	// Wait for container to finish
	if _, err := cr.client.ContainerWait(cr.ctx, cr.containerID, container.WaitConditionNotRunning); err != nil {
		slog.Error("Failed to wait for container", "containerID", cr.containerID, "error", err)
	}

	// Remove container
	if err := cr.client.ContainerRemove(cr.ctx, cr.containerID, container.RemoveOptions{Force: true}); err != nil {
		slog.Error("Failed to remove container", "containerID", cr.containerID, "error", err)
		return err
	}

	return nil
}
