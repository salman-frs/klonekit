package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApply_DryRun(t *testing.T) {
	// Create a temporary blueprint file
	tempDir, err := os.MkdirTemp("", "klonekit-app-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %s", err)
	}
	defer os.RemoveAll(tempDir)

	blueprintContent := `
apiVersion: v1
kind: Blueprint
metadata:
  name: test-infrastructure
  description: Test blueprint for integration testing
spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: test-token
    project:
      name: test-repo
      namespace: test-user
      description: Test repository
      visibility: private
  cloud:
    provider: aws
    region: us-east-1
  scaffold:
    source: /test/source
    destination: /test/destination
  variables:
    instance_type: t3.micro
`

	blueprintFile := filepath.Join(tempDir, "test-blueprint.yaml")
	if err := os.WriteFile(blueprintFile, []byte(blueprintContent), 0644); err != nil {
		t.Fatalf("Failed to create test blueprint file: %s", err)
	}

	// Create test directories that would be referenced
	sourceDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "destination")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %s", err)
	}

	// Create a test Terraform file in source
	testTfFile := filepath.Join(sourceDir, "main.tf")
	if err := os.WriteFile(testTfFile, []byte("# Test terraform file"), 0644); err != nil {
		t.Fatalf("Failed to create test terraform file: %s", err)
	}

	// Update blueprint to use actual temp directories
	blueprintContent = strings.ReplaceAll(blueprintContent, "/test/source", sourceDir)
	blueprintContent = strings.ReplaceAll(blueprintContent, "/test/destination", destDir)
	if err := os.WriteFile(blueprintFile, []byte(blueprintContent), 0644); err != nil {
		t.Fatalf("Failed to update test blueprint file: %s", err)
	}

	tests := []struct {
		name        string
		isDryRun    bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Dry run mode - should simulate all stages",
			isDryRun:    true,
			expectError: false,
		},
		{
			name:        "Normal mode - will fail on GitLab auth (expected)",
			isDryRun:    false,
			expectError: true,
			errorMsg:    "SCM provider initialization failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Apply(blueprintFile, tt.isDryRun)

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

			// In dry run mode, destination directory should not be created
			if tt.isDryRun {
				if _, err := os.Stat(destDir); !os.IsNotExist(err) {
					t.Errorf("Expected destination directory not to be created in dry run mode, but it exists")
				}
			}
		})
	}
}

func TestApply_InvalidBlueprint(t *testing.T) {
	tests := []struct {
		name          string
		blueprintPath string
		expectError   bool
		errorMsg      string
	}{
		{
			name:          "Non-existent blueprint file",
			blueprintPath: "/nonexistent/blueprint.yaml",
			expectError:   true,
			errorMsg:      "blueprint parsing failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Apply(tt.blueprintPath, false)

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

func TestValidatePrerequisites(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Prerequisites validation",
			expectError: false, // Will depend on Docker availability
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePrerequisites()

			// Docker may not be available in test environments
			if err != nil && strings.Contains(err.Error(), "failed to connect to Docker daemon") {
				t.Skipf("Skipping test: Docker not available in test environment: %v", err)
				return
			}

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

func TestApply_StageFailureHandling(t *testing.T) {
	// Test that failure in scaffolding prevents subsequent stages
	tempDir, err := os.MkdirTemp("", "klonekit-app-failure-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %s", err)
	}
	defer os.RemoveAll(tempDir)

	// Create invalid blueprint with non-existent source directory
	blueprintContent := `
apiVersion: v1
kind: Blueprint
metadata:
  name: test-failure
  description: Test blueprint for failure handling
spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: test-token
    project:
      name: test-repo
      namespace: test-user
      description: Test repository
      visibility: private
  cloud:
    provider: aws
    region: us-east-1
  scaffold:
    source: /nonexistent/source/directory
    destination: /test/destination
  variables:
    instance_type: t3.micro
`

	blueprintFile := filepath.Join(tempDir, "invalid-blueprint.yaml")
	if err := os.WriteFile(blueprintFile, []byte(blueprintContent), 0644); err != nil {
		t.Fatalf("Failed to create invalid blueprint file: %s", err)
	}

	// This should fail at scaffolding stage
	err = Apply(blueprintFile, false)
	if err == nil {
		t.Error("Expected error due to invalid source directory, but got none")
		return
	}

	if !strings.Contains(err.Error(), "scaffolding failed") {
		t.Errorf("Expected scaffolding failure, got: %s", err.Error())
	}
}

// Helper function to create a complete valid blueprint for testing
func createValidTestBlueprint(tempDir string) (string, error) {
	sourceDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "destination")

	// Create source directory with test terraform file
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create source directory: %w", err)
	}

	testTfFile := filepath.Join(sourceDir, "main.tf")
	tfContent := `
resource "local_file" "test" {
  content  = "Hello from KloneKit integration test"
  filename = "test_output.txt"
}
`
	if err := os.WriteFile(testTfFile, []byte(tfContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create test terraform file: %w", err)
	}

	blueprintContent := fmt.Sprintf(`
apiVersion: v1
kind: Blueprint
metadata:
  name: integration-test
  description: Complete integration test blueprint
spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: test-token
    project:
      name: integration-test-repo
      namespace: test-user
      description: Integration test repository
      visibility: private
  cloud:
    provider: aws
    region: us-east-1
  scaffold:
    source: %s
    destination: %s
  variables:
    test_variable: integration_test
`, sourceDir, destDir)

	blueprintFile := filepath.Join(tempDir, "integration-blueprint.yaml")
	if err := os.WriteFile(blueprintFile, []byte(blueprintContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create integration blueprint file: %w", err)
	}

	return blueprintFile, nil
}

func TestApply_FullWorkflowDryRun(t *testing.T) {
	// Test the complete workflow in dry-run mode
	tempDir, err := os.MkdirTemp("", "klonekit-app-full-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %s", err)
	}
	defer os.RemoveAll(tempDir)

	blueprintFile, err := createValidTestBlueprint(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test blueprint: %s", err)
	}

	// Execute full workflow in dry-run mode
	err = Apply(blueprintFile, true)
	if err != nil {
		t.Errorf("Unexpected error in dry-run mode: %s", err)
	}

	// Verify that no actual files were created (dry-run should not modify filesystem)
	destDir := filepath.Join(tempDir, "destination")
	if _, err := os.Stat(destDir); !os.IsNotExist(err) {
		t.Error("Destination directory should not exist after dry-run")
	}
}
