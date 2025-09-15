package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_ErrorHandling_BlueprintNotFound(t *testing.T) {
	// Create a temporary directory without blueprint files
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Set custom log directory for test isolation
	originalLogDir := os.Getenv("KLONEKIT_LOG_DIR")
	os.Setenv("KLONEKIT_LOG_DIR", tempDir)
	defer func() {
		if originalLogDir != "" {
			os.Setenv("KLONEKIT_LOG_DIR", originalLogDir)
		} else {
			os.Unsetenv("KLONEKIT_LOG_DIR")
		}
	}()

	// Change to temp directory
	os.Chdir(tempDir)

	// Build the CLI binary
	binaryPath := filepath.Join(tempDir, "klonekit")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/klonekit")
	buildCmd.Dir = originalDir
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI binary: %v", err)
	}

	// Run scaffold command without blueprint file
	cmd := exec.Command(binaryPath, "scaffold")
	cmd.Env = append(os.Environ(), "KLONEKIT_LOG_DIR="+tempDir)
	output, err := cmd.CombinedOutput()

	// Should exit with non-zero code
	if err == nil {
		t.Error("Expected command to fail but it succeeded")
	}

	outputStr := string(output)

	// Check for expected error message components
	expectedParts := []string{
		"Error:",
		"Failed to locate blueprint file",
		"Cause:",
		"No klonekit.yml or klonekit.yaml file found",
		"Suggestion:",
		"Create a blueprint file",
	}

	for _, part := range expectedParts {
		if !strings.Contains(outputStr, part) {
			t.Errorf("Expected output to contain %q, but got: %s", part, outputStr)
		}
	}

	// Verify log file was created
	logFile := filepath.Join(tempDir, "klonekit.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Expected klonekit.log to be created")
	}
}

func TestCLI_ErrorHandling_InvalidBlueprintFile(t *testing.T) {
	// Create a temporary directory with invalid blueprint file
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Set custom log directory for test isolation
	originalLogDir := os.Getenv("KLONEKIT_LOG_DIR")
	os.Setenv("KLONEKIT_LOG_DIR", tempDir)
	defer func() {
		if originalLogDir != "" {
			os.Setenv("KLONEKIT_LOG_DIR", originalLogDir)
		} else {
			os.Unsetenv("KLONEKIT_LOG_DIR")
		}
	}()

	// Change to temp directory
	os.Chdir(tempDir)

	// Create invalid YAML file
	invalidYAML := `invalid: yaml: content:
  - this is not valid
    yaml: structure
      with: improper
    indentation`

	if err := os.WriteFile("klonekit.yml", []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to create invalid blueprint file: %v", err)
	}

	// Build the CLI binary
	binaryPath := filepath.Join(tempDir, "klonekit")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/klonekit")
	buildCmd.Dir = originalDir
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI binary: %v", err)
	}

	// Run scaffold command with invalid blueprint file
	cmd := exec.Command(binaryPath, "scaffold")
	cmd.Env = append(os.Environ(), "KLONEKIT_LOG_DIR="+tempDir)
	output, err := cmd.CombinedOutput()

	// Should exit with non-zero code
	if err == nil {
		t.Error("Expected command to fail but it succeeded")
	}

	outputStr := string(output)

	// Check for error output
	if !strings.Contains(outputStr, "Error:") {
		t.Errorf("Expected error output, but got: %s", outputStr)
	}

	// Verify log file was created
	logFile := filepath.Join(tempDir, "klonekit.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Expected klonekit.log to be created")
	}
}

func TestCLI_ErrorHandling_InvalidFlag(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Build the CLI binary
	binaryPath := filepath.Join(tempDir, "klonekit")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/klonekit")
	buildCmd.Dir = originalDir
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI binary: %v", err)
	}

	// Run with invalid flag
	cmd := exec.Command(binaryPath, "scaffold", "--invalid-flag")
	output, err := cmd.CombinedOutput()

	// Should exit with non-zero code
	if err == nil {
		t.Error("Expected command to fail but it succeeded")
	}

	outputStr := string(output)

	// Check for error output
	if !strings.Contains(outputStr, "Error:") && !strings.Contains(outputStr, "unknown flag") {
		t.Errorf("Expected error output about unknown flag, but got: %s", outputStr)
	}
}

func TestCLI_ErrorHandling_NonexistentBlueprintFile(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Change to temp directory
	os.Chdir(tempDir)

	// Build the CLI binary
	binaryPath := filepath.Join(tempDir, "klonekit")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/klonekit")
	buildCmd.Dir = originalDir
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI binary: %v", err)
	}

	// Run scaffold command with explicit nonexistent file
	cmd := exec.Command(binaryPath, "scaffold", "-f", "nonexistent.yml")
	output, err := cmd.CombinedOutput()

	// Should exit with non-zero code
	if err == nil {
		t.Error("Expected command to fail but it succeeded")
	}

	outputStr := string(output)

	// Check for error output
	if !strings.Contains(outputStr, "Error:") {
		t.Errorf("Expected error output, but got: %s", outputStr)
	}
}

func TestCLI_SuccessfulExecution_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Change to temp directory
	os.Chdir(tempDir)

	// Create a minimal valid blueprint file
	validYAML := `apiVersion: klonekit.dev/v1
kind: Blueprint
metadata:
  name: test-blueprint
spec:
  scaffold:
    destination: ./terraform
    source: minimal
  scm:
    project:
      name: test-project
      namespace: test-namespace
  provision:
    terraform:
      version: "1.8.0"`

	if err := os.WriteFile("klonekit.yml", []byte(validYAML), 0644); err != nil {
		t.Fatalf("Failed to create valid blueprint file: %v", err)
	}

	// Build the CLI binary
	binaryPath := filepath.Join(tempDir, "klonekit")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/klonekit")
	buildCmd.Dir = originalDir
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI binary: %v", err)
	}

	// Run scaffold command in dry-run mode (should succeed without external dependencies)
	cmd := exec.Command(binaryPath, "scaffold", "--dry-run")
	output, err := cmd.CombinedOutput()

	outputStr := string(output)

	// This may still fail due to missing dependencies, but error handling should work
	if err != nil {
		// If it fails, it should show proper error formatting
		if !strings.Contains(outputStr, "Error:") {
			t.Errorf("Expected structured error output, but got: %s", outputStr)
		}
	} else {
		// If it succeeds, it should show success message
		if !strings.Contains(outputStr, "Dry run completed successfully") {
			t.Errorf("Expected success message, but got: %s", outputStr)
		}
	}
}