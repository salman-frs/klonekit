package runtime

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

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

// NewDockerRuntime creates a new DockerRuntime instance with dynamic socket detection.
func NewDockerRuntime() (*DockerRuntime, error) {
	dockerClient, err := createDockerClientWithDynamicSocket()
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

// createDockerClientWithDynamicSocket creates a Docker client with dynamic socket detection.
// It tries multiple socket locations in order of preference for different Docker setups.
func createDockerClientWithDynamicSocket() (*client.Client, error) {
	// Define potential Docker socket locations in order of preference
	socketPaths := getDockerSocketPaths()

	var lastErr error

	// Try each socket path
	for _, socketPath := range socketPaths {
		slog.Debug("Attempting to connect to Docker socket", "path", socketPath)

		// Check if socket exists
		if _, err := os.Stat(socketPath); os.IsNotExist(err) {
			slog.Debug("Docker socket not found", "path", socketPath)
			continue
		}

		// Try to create client with this socket
		dockerClient, err := client.NewClientWithOpts(
			client.WithHost("unix://"+socketPath),
			client.WithAPIVersionNegotiation(),
		)
		if err != nil {
			lastErr = err
			slog.Debug("Failed to create Docker client", "path", socketPath, "error", err)
			continue
		}

		// Test the connection
		ctx := context.Background()
		_, err = dockerClient.Ping(ctx)
		if err != nil {
			lastErr = err
			slog.Debug("Failed to ping Docker daemon", "path", socketPath, "error", err)
			if cerr := dockerClient.Close(); cerr != nil {
				slog.Debug("Error closing Docker client", "error", cerr)
			}
			continue
		}

		slog.Info("Successfully connected to Docker daemon", "socketPath", socketPath)
		return dockerClient, nil
	}

	// If all socket paths failed, try the default FromEnv approach
	slog.Debug("All socket paths failed, trying environment-based configuration")
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client with all methods: last error was %w", lastErr)
	}

	// Test the environment-based client
	ctx := context.Background()
	_, err = dockerClient.Ping(ctx)
	if err != nil {
		if cerr := dockerClient.Close(); cerr != nil {
			slog.Debug("Error closing Docker client", "error", cerr)
		}
		return nil, fmt.Errorf("failed to connect to Docker daemon with all methods: last error was %w", err)
	}

	slog.Info("Successfully connected to Docker daemon using environment configuration")
	return dockerClient, nil
}

// getDockerSocketPaths returns a list of potential Docker socket paths in order of preference.
func getDockerSocketPaths() []string {
	var socketPaths []string

	// Get home directory for user-specific paths
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		if currentUser := os.Getenv("USER"); currentUser != "" {
			homeDir = filepath.Join("/Users", currentUser)
		}
	}

	// Colima (macOS alternative to Docker Desktop)
	if homeDir != "" {
		socketPaths = append(socketPaths, filepath.Join(homeDir, ".colima", "docker.sock"))
		socketPaths = append(socketPaths, filepath.Join(homeDir, ".colima", "default", "docker.sock"))
	}

	// Docker Desktop (macOS/Windows)
	if homeDir != "" {
		socketPaths = append(socketPaths, filepath.Join(homeDir, ".docker", "run", "docker.sock"))
		socketPaths = append(socketPaths, filepath.Join(homeDir, ".docker", "desktop", "docker.sock"))
	}

	// Podman Desktop compatibility
	if homeDir != "" {
		socketPaths = append(socketPaths, filepath.Join(homeDir, ".local", "share", "containers", "podman", "machine", "podman.sock"))
	}

	// Standard Docker daemon socket (Linux/Docker CE)
	socketPaths = append(socketPaths, "/var/run/docker.sock")

	// Lima (another Docker alternative)
	if homeDir != "" {
		socketPaths = append(socketPaths, filepath.Join(homeDir, ".lima", "docker", "sock", "docker.sock"))
	}

	return socketPaths
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
		Mounts:      mounts,
		NetworkMode: "default", // Use default Docker network for internet access
		DNS:         []string{"8.8.8.8", "8.8.4.4"}, // Add public DNS servers
		DNSOptions:  []string{"ndots:0"}, // Improve DNS resolution performance
	}

	// Set container user if specified to avoid permission issues
	if opts.User != "" {
		containerConfig.User = opts.User
	}

	// Create container with optional name
	containerName := opts.ContainerName
	resp, err := d.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
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
		client:         d.client,
		containerID:    containerID,
		ctx:            ctx,
		retainContainer: opts.RetainContainer,
	}, nil
}

// containerReader wraps container output and handles cleanup.
type containerReader struct {
	client          *client.Client
	containerID     string
	ctx             context.Context
	reader          io.ReadCloser
	closed          bool
	exitCode        int64
	exitError       error
	retainContainer bool // If true, don't remove container on close
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

	// Wait for container to finish (with timeout to avoid hanging)
	waitCtx, cancel := context.WithTimeout(cr.ctx, 30*time.Second)
	defer cancel()

	statusCh, errCh := cr.client.ContainerWait(waitCtx, cr.containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			cr.exitError = err
			slog.Debug("Container wait completed with warning", "containerID", cr.containerID, "warning", err.Error())
		}
	case status := <-statusCh:
		// Container finished, capture exit code
		cr.exitCode = status.StatusCode
		if status.StatusCode != 0 {
			cr.exitError = fmt.Errorf("container exited with non-zero status: %d", status.StatusCode)
			slog.Debug("Container failed", "containerID", cr.containerID, "exitCode", status.StatusCode)
		} else {
			slog.Debug("Container finished successfully", "containerID", cr.containerID)
		}
	case <-waitCtx.Done():
		// Timeout reached
		cr.exitError = fmt.Errorf("container wait timeout")
		slog.Debug("Container wait timeout", "containerID", cr.containerID)
	}

	// Remove container only if not retaining - use a fresh context in case the original was cancelled
	if !cr.retainContainer {
		removeCtx, removeCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer removeCancel()

		if err := cr.client.ContainerRemove(removeCtx, cr.containerID, container.RemoveOptions{Force: true}); err != nil {
			// Log as debug instead of error - container cleanup is best effort
			slog.Debug("Container cleanup completed with warning", "containerID", cr.containerID, "warning", err.Error())
			// Don't return the error - container cleanup is best effort
		} else {
			slog.Debug("Container removed successfully", "containerID", cr.containerID)
		}
	} else {
		slog.Info("Container retained for state persistence", "containerID", cr.containerID)
	}

	// Return the exit error if the container failed
	return cr.exitError
}
