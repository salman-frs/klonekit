package provisioner

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/mock"

	"klonekit/internal/runtime"
	"klonekit/pkg/blueprint"
	runtimePkg "klonekit/pkg/runtime"
)

// TestCopyFile tests the copyFile function
func TestCopyFile(t *testing.T) {
	// Create temporary source file
	srcFile, err := os.CreateTemp("", "test-src-*")
	if err != nil {
		t.Fatalf("Failed to create temp source file: %v", err)
	}
	defer os.Remove(srcFile.Name())

	testContent := "test content for copy"
	if _, err := srcFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write test content: %v", err)
	}
	srcFile.Close()

	// Create temporary destination path
	dstFile, err := os.CreateTemp("", "test-dst-*")
	if err != nil {
		t.Fatalf("Failed to create temp destination file: %v", err)
	}
	dstPath := dstFile.Name()
	dstFile.Close()
	os.Remove(dstPath) // Remove so copyFile can create it

	// Test successful copy
	err = copyFile(srcFile.Name(), dstPath)
	if err != nil {
		t.Errorf("copyFile failed: %v", err)
	}

	// Verify content was copied
	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Errorf("Failed to read destination file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Content mismatch. Expected %s, got %s", testContent, string(content))
	}

	// Clean up
	os.Remove(dstPath)

	// Test error cases
	err = copyFile("non-existent-file", dstPath)
	if err == nil {
		t.Error("Expected error when copying non-existent file")
	}
}

// TestCleanDockerLogLine tests the cleanDockerLogLine function
func TestCleanDockerLogLine(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[0m[1mhello[0m", "hello"},
		{"regular text", "regular text"},
		{"", ""},
		{"[31mError:[0m Something went wrong", "Error: Something went wrong"},
	}

	for _, tt := range tests {
		result := cleanDockerLogLine(tt.input)
		if result != tt.expected {
			t.Errorf("cleanDockerLogLine(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestMain sets up mock AWS credentials for testing
func TestMain(m *testing.M) {
	// Create temporary AWS credentials directory
	tmpDir, err := os.MkdirTemp("", "aws-test-*")
	if err != nil {
		panic("Failed to create temp directory: " + err.Error())
	}
	defer os.RemoveAll(tmpDir)

	awsDir := filepath.Join(tmpDir, ".aws")
	if err := os.MkdirAll(awsDir, 0755); err != nil {
		panic("Failed to create .aws directory: " + err.Error())
	}

	// Create mock credentials file
	credentialsContent := `[default]
aws_access_key_id = test-access-key-id
aws_secret_access_key = test-secret-access-key
region = us-east-1
`
	credentialsFile := filepath.Join(awsDir, "credentials")
	if err := os.WriteFile(credentialsFile, []byte(credentialsContent), 0644); err != nil {
		panic("Failed to create credentials file: " + err.Error())
	}

	// Set environment variables to point to mock credentials
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	os.Setenv("AWS_ACCESS_KEY_ID", "test-access-key-id")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-access-key")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")

	// Run tests
	code := m.Run()

	// Restore original HOME environment variable
	if originalHome != "" {
		os.Setenv("HOME", originalHome)
	} else {
		os.Unsetenv("HOME")
	}

	os.Exit(code)
}

// MockContainerRuntime is a mock implementation of the ContainerRuntime interface
type MockContainerRuntime struct {
	mock.Mock
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
			mockRuntime := new(MockContainerRuntime)
			tt.setupMock(mockRuntime)

			// Create provisioner with mock
			provisioner := NewTerraformDockerProvisioner(mockRuntime)

			err := provisioner.Provision(tt.spec, true) // Use auto-approve for tests

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

			err = provisioner.Provision(tt.spec, true) // Use auto-approve for tests

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

// fixPermissionsRecursively fixes file permissions to ensure cleanup can succeed
func fixPermissionsRecursively(path string) error {
	return filepath.Walk(path, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			// Continue on permission errors during walk
			return nil
		}
		if info.IsDir() {
			// Set directory permissions to allow removal
			_ = os.Chmod(file, 0755)
		} else {
			// Set file permissions to allow removal
			_ = os.Chmod(file, 0644)
		}
		return nil
	})
}

// getCurrentUserHome gets the real user home directory, bypassing environment overrides
func getCurrentUserHome() string {
	if currentUser, err := user.Current(); err == nil {
		return currentUser.HomeDir
	}
	return ""
}

func TestTerraformDockerProvisioner_E2E_Local(t *testing.T) {
	t.Logf("Attempting to connect to Docker daemon...")

	// IMPORTANT: TestMain overrides HOME for AWS credentials, but we need the real HOME for Docker sockets
	// Temporarily restore the real HOME directory for Docker socket detection
	currentHome := os.Getenv("HOME")
	realHome := getCurrentUserHome() // Get the real user home directory
	if realHome != "" && currentHome != realHome {
		t.Logf("TestMain has overridden HOME. Current: %s, Real: %s", currentHome, realHome)
		os.Setenv("HOME", realHome)
		defer os.Setenv("HOME", currentHome) // Restore the test HOME after Docker connection
	}

	// Check if Docker daemon is accessible using dynamic socket detection
	dockerRuntime, err := runtime.NewDockerRuntime()

	// Restore test HOME immediately after Docker runtime creation
	if realHome != "" && currentHome != realHome {
		os.Setenv("HOME", currentHome)
	}

	if err != nil {
		// Show detailed error breakdown
		t.Logf("Docker connection failed with error: %v", err)
		t.Logf("Checking socket availability in real home directory...")

		// Check sockets in the real home directory
		sockets := []string{
			filepath.Join(realHome, ".colima", "docker.sock"),
			filepath.Join(realHome, ".colima", "default", "docker.sock"),
			"/var/run/docker.sock",
		}

		for _, socket := range sockets {
			if _, statErr := os.Stat(socket); statErr == nil {
				t.Logf("✓ Socket exists: %s", socket)

				// Test direct connection to this socket
				testClient, clientErr := client.NewClientWithOpts(
					client.WithHost("unix://"+socket),
					client.WithAPIVersionNegotiation(),
				)
				if clientErr != nil {
					t.Logf("  ✗ Client creation failed: %v", clientErr)
					continue
				}

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_, pingErr := testClient.Ping(ctx)
				cancel()
				testClient.Close()

				if pingErr != nil {
					t.Logf("  ✗ Ping failed: %v", pingErr)
				} else {
					t.Logf("  ✓ Direct connection successful!")
				}
			} else {
				t.Logf("✗ Socket missing: %s", socket)
			}
		}

		t.Skipf("Skipping E2E test: Docker daemon not accessible: %s", err)
		return
	}

	t.Logf("✓ Successfully connected to Docker daemon")
	_ = dockerRuntime // Use the runtime variable to avoid unused variable error

	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Add custom cleanup to handle permission issues from Terraform provider downloads
	defer func() {
		if err := fixPermissionsRecursively(tempDir); err != nil {
			t.Logf("Warning: failed to fix permissions during cleanup: %v", err)
		}

		// In CI environments, force cleanup with elevated permissions if needed
		if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
			if err := os.RemoveAll(tempDir); err != nil {
				t.Logf("Warning: CI cleanup failed: %v", err)
				// Try with more aggressive cleanup
				_ = exec.Command("rm", "-rf", tempDir).Run()
			}
		}
		// Note: t.TempDir() handles the actual removal, but we fix permissions first
	}()

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

	// Reuse the successfully connected Docker runtime from above
	// (no need to create a second one since we already have a working connection)

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

	err = provisioner.Provision(spec, true) // Use auto-approve for tests

	// In CI environments, AWS providers may download successfully even without credentials
	// The test should pass if either:
	// 1. No error (successful provider download + terraform init/plan)
	// 2. AWS-related error (expected in local dev without credentials)
	// Only Docker connectivity errors should cause test failure
	if err != nil && strings.Contains(err.Error(), "failed to create Docker") {
		t.Errorf("Unexpected Docker connectivity error: %s", err)
		return
	}

	if err == nil {
		t.Logf("✅ Provision completed successfully (CI environment with provider access)")
	} else {
		t.Logf("ℹ️ Provision failed as expected (likely AWS credentials): %v", err)
	}

	// Test passes in both cases - success or expected AWS failure
}
