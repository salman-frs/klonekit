package provisioner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker/client"

	"klonekit/pkg/blueprint"
)

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
					Source:      "/source/path",
					Destination: "/nonexistent/path",
				},
			},
			expectError: true,
			errorMsg:    "scaffold directory does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a provisioner with minimal dependencies for basic validation
			provisioner := &TerraformDockerProvisioner{}

			err := provisioner.Provision(tt.spec)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %s", err)
			}
		})
	}
}

func TestTerraformDockerProvisioner_createContainerConfig(t *testing.T) {
	provisioner := &TerraformDockerProvisioner{}

	scaffoldDir := "/test/scaffold"
	awsCredsDir := "/test/aws"

	containerConfig, hostConfig, err := provisioner.createContainerConfig(scaffoldDir, awsCredsDir)

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	// Verify container configuration
	if containerConfig.Image != TerraformDockerImage {
		t.Errorf("Expected image '%s', got '%s'", TerraformDockerImage, containerConfig.Image)
	}

	if containerConfig.WorkingDir != WorkingDirectory {
		t.Errorf("Expected working directory '%s', got '%s'", WorkingDirectory, containerConfig.WorkingDir)
	}

	if !containerConfig.Tty {
		t.Error("Expected TTY to be enabled")
	}

	// Verify host configuration mounts
	if len(hostConfig.Mounts) != 2 {
		t.Errorf("Expected 2 mounts, got %d", len(hostConfig.Mounts))
		return
	}

	// Check scaffold directory mount
	scaffoldMount := hostConfig.Mounts[0]
	if scaffoldMount.Source != scaffoldDir {
		t.Errorf("Expected scaffold mount source '%s', got '%s'", scaffoldDir, scaffoldMount.Source)
	}
	if scaffoldMount.Target != WorkingDirectory {
		t.Errorf("Expected scaffold mount target '%s', got '%s'", WorkingDirectory, scaffoldMount.Target)
	}

	// Check AWS credentials mount
	awsMount := hostConfig.Mounts[1]
	if awsMount.Source != awsCredsDir {
		t.Errorf("Expected AWS mount source '%s', got '%s'", awsCredsDir, awsMount.Source)
	}
	if awsMount.Target != "/root/.aws" {
		t.Errorf("Expected AWS mount target '/root/.aws', got '%s'", awsMount.Target)
	}
	if !awsMount.ReadOnly {
		t.Error("Expected AWS mount to be read-only")
	}

	// Verify environment variables
	expectedEnvVars := []string{
		"AWS_SHARED_CREDENTIALS_FILE=/root/.aws/credentials",
		"AWS_CONFIG_FILE=/root/.aws/config",
	}

	for _, expectedVar := range expectedEnvVars {
		found := false
		for _, actualVar := range containerConfig.Env {
			if actualVar == expectedVar {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected environment variable '%s' not found", expectedVar)
		}
	}
}

func TestTerraformDockerProvisioner_getAWSCredentialsDir(t *testing.T) {
	provisioner := &TerraformDockerProvisioner{}

	awsDir, err := provisioner.getAWSCredentialsDir()

	// This test is environment-dependent
	// In test environments without AWS credentials, it's expected to fail
	if err != nil && strings.Contains(err.Error(), "AWS credentials directory not found") {
		t.Skipf("Skipping test: AWS credentials not configured in test environment: %v", err)
		return
	}

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	if awsDir == "" {
		t.Error("Expected AWS directory path to be non-empty")
	}

	// Verify the directory exists
	if _, err := os.Stat(awsDir); os.IsNotExist(err) {
		t.Errorf("AWS directory does not exist: %s", awsDir)
	}
}

// E2E test that requires Docker daemon to be running
func TestTerraformDockerProvisioner_E2E_Local(t *testing.T) {
	// Skip this test if SHORT flag is set or in CI
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Check if Docker is available
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		t.Skipf("Skipping E2E test: Docker not available: %v", err)
	}

	ctx := context.Background()
	if _, err := dockerClient.Ping(ctx); err != nil {
		t.Skipf("Skipping E2E test: Docker daemon not accessible: %v", err)
	}

	// Create temporary scaffold directory
	tempDir, err := os.MkdirTemp("", "klonekit-e2e-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %s", err)
	}
	defer os.RemoveAll(tempDir)

	// Create simple Terraform configuration using local provider (no AWS needed)
	terraformConfig := `
terraform {
  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "~> 2.0"
    }
  }
}

resource "local_file" "test" {
  content  = "Hello from KloneKit E2E test!"
  filename = "klonekit_e2e_output.txt"
}
`

	tfFile := filepath.Join(tempDir, "main.tf")
	if err := os.WriteFile(tfFile, []byte(terraformConfig), 0644); err != nil {
		t.Fatalf("Failed to create Terraform file: %s", err)
	}

	// Create temporary AWS credentials directory for the test (even though we're using local provider)
	awsDir, err := os.MkdirTemp("", "test-aws-e2e-*")
	if err != nil {
		t.Fatalf("Failed to create temp AWS directory: %s", err)
	}
	defer os.RemoveAll(awsDir)

	credentialsFile := filepath.Join(awsDir, "credentials")
	configFile := filepath.Join(awsDir, "config")

	if err := os.WriteFile(credentialsFile, []byte("[default]\naws_access_key_id = test\naws_secret_access_key = test"), 0644); err != nil {
		t.Fatalf("Failed to create test credentials file: %s", err)
	}

	if err := os.WriteFile(configFile, []byte("[default]\nregion = us-east-1"), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %s", err)
	}

	// Create provisioner
	provisioner, err := NewTerraformDockerProvisioner()
	if err != nil {
		t.Fatalf("Failed to create provisioner: %s", err)
	}

	// Create a modified provisioner that uses our test AWS directory
	// We can't modify the struct directly, so we'll create a wrapper
	testProvisioner := &testTerraformDockerProvisioner{
		TerraformDockerProvisioner: provisioner,
		testAWSDir:                 awsDir,
	}

	spec := &blueprint.Spec{
		Scaffold: blueprint.Scaffold{
			Source:      "/source/path",
			Destination: tempDir,
		},
	}

	// Run provisioning
	err = testProvisioner.Provision(spec)
	if err != nil {
		t.Errorf("Provisioning failed: %s", err)
		return
	}

	// Verify that the local file was created by Terraform
	expectedFile := filepath.Join(tempDir, "klonekit_e2e_output.txt")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected output file was not created: %s", expectedFile)
	} else {
		// Read the file content to verify
		content, err := os.ReadFile(expectedFile)
		if err != nil {
			t.Errorf("Failed to read output file: %s", err)
		} else if string(content) != "Hello from KloneKit E2E test!" {
			t.Errorf("Unexpected file content: %s", string(content))
		}
	}
}

// testTerraformDockerProvisioner wraps the regular provisioner for testing
type testTerraformDockerProvisioner struct {
	*TerraformDockerProvisioner
	testAWSDir string
}

// Override getAWSCredentialsDir for testing
func (t *testTerraformDockerProvisioner) getAWSCredentialsDir() (string, error) {
	return t.testAWSDir, nil
}

// Provision method that uses the overridden getAWSCredentialsDir
func (t *testTerraformDockerProvisioner) Provision(spec *blueprint.Spec) error {
	ctx := context.Background()

	// Validate that scaffold directory exists
	scaffoldDir := spec.Scaffold.Destination
	if _, err := os.Stat(scaffoldDir); os.IsNotExist(err) {
		return fmt.Errorf("scaffold directory does not exist: %s", scaffoldDir)
	}

	// Pull Terraform Docker image
	if err := t.pullTerraformImage(ctx); err != nil {
		return fmt.Errorf("failed to pull Terraform image: %w", err)
	}

	// Get absolute path of scaffold directory
	absScaffoldDir, err := filepath.Abs(scaffoldDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for scaffold directory: %w", err)
	}

	// Use test AWS credentials directory
	awsCredsDir := t.testAWSDir

	// Create container configuration
	containerConfig, hostConfig, err := t.createContainerConfig(absScaffoldDir, awsCredsDir)
	if err != nil {
		return fmt.Errorf("failed to create container configuration: %w", err)
	}

	// Execute Terraform init
	if err := t.runTerraformCommand(ctx, containerConfig, hostConfig, "init"); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	// Execute Terraform apply with auto-approve
	if err := t.runTerraformCommand(ctx, containerConfig, hostConfig, "apply", "-auto-approve"); err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	return nil
}
