// Located in pkg/runtime/runtime.go
package runtime

import (
	"context"
	"io"
)

// RunOptions defines the parameters for running a container.
type RunOptions struct {
	Image            string
	Command          []string
	VolumeMounts     map[string]string
	EnvVars          map[string]string
	WorkingDirectory string
}

// ContainerRuntime defines the contract for container operations.
type ContainerRuntime interface {
	PullImage(ctx context.Context, image string) error
	RunContainer(ctx context.Context, opts RunOptions) (io.ReadCloser, error)
}
