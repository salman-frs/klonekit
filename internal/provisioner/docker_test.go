package provisioner

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/mock"

	"klonekit/internal/runtime"
	"klonekit/pkg/blueprint"
	runtimePkg "klonekit/pkg/runtime"
)

// MockContainerRuntime is a mock implementation of the ContainerRuntime interface
type MockContainerRuntime struct {
	*mock.Mock
}

func NewMockContainerRuntime() *MockContainerRuntime {
	return &MockContainerRuntime{Mock: &mock.Mock{}}
}

func (m *MockContainerRuntime) PullImage(ctx context.Context, image string) error {
	args := m.Called(ctx, image)
	return args.Error(0)
}

func (m *MockContainerRuntime) RunContainer(ctx context.Context, opts runtimePkg.RunOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

// MockReadCloser for testing container output
type MockReadCloser struct {
	data []byte
	pos  int
}

func (m *MockReadCloser) Read(p []byte) (int, error) {
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	n := copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

func (m *MockReadCloser) Close() error {
	return nil
}

func TestTerraformDockerProvisioner_WithMock(t *testing.T) {
	tests := []struct {
		name          string
		spec          *blueprint.Spec
		setupMock     func(*MockContainerRuntime)
		expectError   bool
		errorContains string
	}{
		{
			name: "Successful provision with mock runtime",
			spec: &blueprint.Spec{
				Scaffold: blueprint.Scaffold{
					Destination: t.TempDir(),
				},
				Cloud: blueprint.CloudProvider{
					Region: "us-east-1",
				},
			},
			setupMock: func(m *MockContainerRuntime) {
				m.On("PullImage", mock.Anything, "hashicorp/terraform:1.8.0").Return(nil)
				m.On("RunContainer", mock.Anything, mock.MatchedBy(func(opts runtimePkg.RunOptions) bool { return true })).Return(&MockReadCloser{data: []byte("Terraform initialized successfully")}, nil)
			},
			expectError: false,
		},
		{
			name: "Pull image failure",
			spec: &blueprint.Spec{
				Scaffold: blueprint.Scaffold{
					Destination: t.TempDir(),
				},
				Cloud: blueprint.CloudProvider{
					Region: "us-east-1",
				},
			},
			setupMock: func(m *MockContainerRuntime) {
				m.On("PullImage", mock.Anything, "hashicorp/terraform:1.8.0").Return(errors.New("failed to pull image"))
			},
			expectError:   true,
			errorContains: "failed to pull image",
		},
		{
			name: "Container run failure",
			spec: &blueprint.Spec{
				Scaffold: blueprint.Scaffold{
					Destination: t.TempDir(),
				},
				Cloud: blueprint.CloudProvider{
					Region: "us-east-1",
				},
			},
			setupMock: func(m *MockContainerRuntime) {
				m.On("PullImage", mock.Anything, "hashicorp/terraform:1.8.0").Return(nil)
				m.On("RunContainer", mock.Anything, mock.MatchedBy(func(opts runtimePkg.RunOptions) bool { return true })).Return((*MockReadCloser)(nil), errors.New("container failed to run"))
			},
			expectError:   true,
			errorContains: "container failed to run",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock runtime
			mockRuntime := NewMockContainerRuntime()
			tt.setupMock(mockRuntime)

			// Create provisioner with mock
			provisioner := NewTerraformDockerProvisioner(mockRuntime)

			err := provisioner.Provision(tt.spec)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got: %s", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %s", err)
				}
			}

			// Verify all expectations were met
			mockRuntime.AssertExpectations(t)
		})
	}
}

func TestTerraformDockerProvisioner_Basic(t *testing.T) {
	tests := []struct {
		name        string
		spec        *blueprint.Spec
		expectError bool
		errorMsg    string
	}{
		{
			name: "Scaffold directory does not exist",
			spec: &blueprint.Spec{
				Scaffold: blueprint.Scaffold{
					Destination: "/nonexistent/path",
				},
				Cloud: blueprint.CloudProvider{
					Region: "us-east-1",
				},
			},
			expectError: true,
			errorMsg:    "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Docker runtime
			dockerRuntime, err := runtime.NewDockerRuntime()
			if err != nil {
				t.Skipf("Skipping test: Docker not available in test environment: %s", err)
				return
			}

			// Create provisioner
			provisioner := NewTerraformDockerProvisioner(dockerRuntime)

			err = provisioner.Provision(tt.spec)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}

			if tt.expectError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error containing '%s', got: %s", tt.errorMsg, err)
			}
		})
	}
}

func TestTerraformDockerProvisioner_getAWSCredentialsDir(t *testing.T) {
	// Create Docker runtime
	dockerRuntime, err := runtime.NewDockerRuntime()
	if err != nil {
		t.Skipf("Skipping test: Docker not available: %s", err)
		return
	}

	provisioner := NewTerraformDockerProvisioner(dockerRuntime)
	awsDir, err := provisioner.getAWSCredentialsDir()

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	if awsDir == "" {
		t.Error("Expected non-empty AWS credentials directory path")
	}

	// Verify the path structure is reasonable (should contain .aws)
	if !strings.Contains(awsDir, ".aws") {
		t.Errorf("Expected AWS credentials directory to contain '.aws', got: %s", awsDir)
	}
}

func TestTerraformDockerProvisioner_E2E_Local(t *testing.T) {
	// Check if Docker daemon is accessible
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		t.Skipf("Skipping E2E test: Docker daemon not accessible: %s", err)
		return
	}

	ctx := context.Background()
	_, err = dockerClient.Ping(ctx)
	if err != nil {
		t.Skipf("Skipping E2E test: Docker daemon not accessible: %s", err)
		return
	}

	// Create a temporary directory structure for testing
	tempDir := t.TempDir()
	scaffoldDir := filepath.Join(tempDir, "scaffold")
	if err := os.MkdirAll(scaffoldDir, 0755); err != nil {
		t.Fatalf("Failed to create temp scaffold directory: %s", err)
	}

	// Create a minimal Terraform file
	terraformFile := filepath.Join(scaffoldDir, "main.tf")
	terraformContent := `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}
`
	if err := os.WriteFile(terraformFile, []byte(terraformContent), 0644); err != nil {
		t.Fatalf("Failed to create test Terraform file: %s", err)
	}

	// Create Docker runtime
	dockerRuntime, err := runtime.NewDockerRuntime()
	if err != nil {
		t.Fatalf("Failed to create Docker runtime: %s", err)
	}

	// Create provisioner
	provisioner := NewTerraformDockerProvisioner(dockerRuntime)

	// Test provisioning (this will fail due to AWS credentials, but should get past initial setup)
	spec := &blueprint.Spec{
		Scaffold: blueprint.Scaffold{
			Destination: scaffoldDir,
		},
		Cloud: blueprint.CloudProvider{
			Region: "us-east-1",
		},
	}

	err = provisioner.Provision(spec)
	// We expect this to fail due to AWS credentials or other infrastructure issues
	// But it should not fail due to Docker connectivity issues
	if err != nil && strings.Contains(err.Error(), "failed to create Docker") {
		t.Errorf("Unexpected Docker connectivity error: %s", err)
	}

	t.Logf("Provision result (expected to fail due to AWS setup): %v", err)
}
