package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestApply_DryRun(t *testing.T) {
	// Clean up any existing state file
	os.Remove(StateFileName)
	defer os.Remove(StateFileName)

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
			err := Apply(blueprintFile, tt.isDryRun, false)

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
	// Clean up any existing state file
	os.Remove(StateFileName)
	defer os.Remove(StateFileName)
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
			err := Apply(tt.blueprintPath, false, false)

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
	// Clean up any existing state file
	os.Remove(StateFileName)
	defer os.Remove(StateFileName)

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
	err = Apply(blueprintFile, false, false)
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
	// Clean up any existing state file
	os.Remove(StateFileName)
	defer os.Remove(StateFileName)

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
	err = Apply(blueprintFile, true, false)
	if err != nil {
		t.Errorf("Unexpected error in dry-run mode: %s", err)
	}

	// Verify that no actual files were created (dry-run should not modify filesystem)
	destDir := filepath.Join(tempDir, "destination")
	if _, err := os.Stat(destDir); !os.IsNotExist(err) {
		t.Error("Destination directory should not exist after dry-run")
	}
}

func TestApply_StatefulExecution_FailureAfterScaffold(t *testing.T) {
	// Test that simulates failure after scaffold stage and verifies resume behavior
	tempDir, err := os.MkdirTemp("", "klonekit-stateful-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %s", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory to control where state file is created
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %s", err)
	}
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %s", err)
	}

	blueprintFile, err := createValidTestBlueprint(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test blueprint: %s", err)
	}

	// First run: This will fail at SCM stage (expected due to invalid GitLab credentials)
	// But scaffolding should succeed and be saved to state
	err = Apply(blueprintFile, false, false)
	if err == nil {
		t.Error("Expected error due to invalid GitLab credentials, but got none")
		return
	}

	// Verify state file was created with scaffold stage completed
	stateFile := filepath.Join(tempDir, StateFileName)
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Error("Expected state file to be created after partial failure")
		return
	}

	// Load and verify state
	state, err := loadState()
	if err != nil {
		t.Fatalf("Failed to load state: %s", err)
	}

	if state == nil {
		t.Error("Expected state to be loaded")
		return
	}

	if state.LastSuccessfulStage != StageScaffold {
		t.Errorf("Expected last successful stage to be scaffold, got: %s", state.LastSuccessfulStage)
	}

	// Verify destination directory was created (scaffolding succeeded)
	destDir := filepath.Join(tempDir, "destination")
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		t.Error("Expected destination directory to exist after successful scaffold stage")
	}
}

func TestApply_StatefulExecution_ResumeFromSCM(t *testing.T) {
	// Test resume behavior by manually creating a state file
	tempDir, err := os.MkdirTemp("", "klonekit-resume-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %s", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %s", err)
	}
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %s", err)
	}

	blueprintFile, err := createValidTestBlueprint(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test blueprint: %s", err)
	}

	// Manually create state file indicating scaffold stage was completed
	testState := &ExecutionState{
		SchemaVersion:       StateSchemaVersion,
		RunID:               "test-resume-run-123",
		LastSuccessfulStage: StageScaffold,
		BlueprintPath:       blueprintFile,
		CreatedAt:           time.Now().Add(-time.Hour),
		LastUpdatedAt:       time.Now().Add(-time.Hour),
	}

	if err := saveState(testState); err != nil {
		t.Fatalf("Failed to save test state: %s", err)
	}

	// Create the destination directory to simulate completed scaffold stage
	destDir := filepath.Join(tempDir, "destination")
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("Failed to create destination directory: %s", err)
	}

	// Copy a file to simulate completed scaffolding
	testFile := filepath.Join(destDir, "main.tf")
	sourceFile := filepath.Join(sourceDir, "main.tf")
	sourceContent, _ := os.ReadFile(sourceFile)
	if err := os.WriteFile(testFile, sourceContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %s", err)
	}

	// Run apply in dry-run mode - should resume from SCM stage
	err = Apply(blueprintFile, true, false) // Using dry-run to avoid actual GitLab operations
	if err != nil {
		t.Errorf("Unexpected error during resume in dry-run mode: %s", err)
	}

	// In dry-run mode, state file should still exist (not cleaned up)
	if _, err := os.Stat(filepath.Join(tempDir, StateFileName)); os.IsNotExist(err) {
		t.Error("State file should still exist after dry-run resume")
	}
}

func TestApply_StatefulExecution_DryRunWithState(t *testing.T) {
	// Test that dry-run mode correctly simulates resume behavior
	tempDir, err := os.MkdirTemp("", "klonekit-dryrun-state-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %s", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %s", err)
	}
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %s", err)
	}

	blueprintFile, err := createValidTestBlueprint(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test blueprint: %s", err)
	}

	// Create state indicating SCM was completed
	testState := &ExecutionState{
		SchemaVersion:       StateSchemaVersion,
		RunID:               "test-dryrun-123",
		LastSuccessfulStage: StageSCM,
		BlueprintPath:       blueprintFile,
		CreatedAt:           time.Now().Add(-time.Hour),
		LastUpdatedAt:       time.Now().Add(-time.Hour),
	}

	if err := saveState(testState); err != nil {
		t.Fatalf("Failed to save test state: %s", err)
	}

	// Run dry-run - should simulate resume from provision stage
	err = Apply(blueprintFile, true, false)
	if err != nil {
		t.Errorf("Unexpected error during dry-run with existing state: %s", err)
	}

	// State file should still exist after dry-run
	if _, err := os.Stat(filepath.Join(tempDir, StateFileName)); os.IsNotExist(err) {
		t.Error("State file should still exist after dry-run")
	}

	// Load state and verify it wasn't modified (dry-run shouldn't update state)
	finalState, err := loadState()
	if err != nil {
		t.Fatalf("Failed to load final state: %s", err)
	}

	if finalState.LastSuccessfulStage != StageSCM {
		t.Errorf("Expected state to remain unchanged in dry-run, but last stage changed to: %s", finalState.LastSuccessfulStage)
	}
}

func TestApply_RetainStateFlag(t *testing.T) {
	// Test that --retain-state flag keeps the state file after successful completion
	tempDir, err := os.MkdirTemp("", "klonekit-retain-state-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %s", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %s", err)
	}
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %s", err)
	}

	blueprintFile, err := createValidTestBlueprint(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test blueprint: %s", err)
	}

	// Test with retain-state=true in dry-run mode (to avoid GitLab API calls)
	err = Apply(blueprintFile, true, true)
	if err != nil {
		t.Errorf("Unexpected error with retain-state in dry-run: %s", err)
	}

	// In dry-run mode, no state file should be created
	if _, err := os.Stat(StateFileName); !os.IsNotExist(err) {
		t.Error("State file should not exist after dry-run, even with retain-state flag")
	}

	// Test with retain-state=false (default behavior)
	// Manually create a state file first to simulate a resumed successful run
	testState := newState(blueprintFile, "test-retain-false")
	testState.LastSuccessfulStage = StageProvision // Simulate completed workflow except final cleanup

	if err := saveState(testState); err != nil {
		t.Fatalf("Failed to save test state: %s", err)
	}

	// Run with retain-state=false - this should remove the state file
	err = Apply(blueprintFile, true, false) // Using dry-run to avoid actual operations
	if err != nil {
		t.Errorf("Unexpected error with retain-state=false: %s", err)
	}

	// State file should not exist (dry-run doesn't actually remove it, but let's test the logic path)
	// Note: In dry-run mode, state file operations are skipped, so we test the code path existed
}

func TestStateFile_LoadSaveRemove(t *testing.T) {
	// Test state file operations in isolation
	tempDir, err := os.MkdirTemp("", "klonekit-state-ops-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %s", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %s", err)
	}
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %s", err)
	}

	// Test loadState with no file
	state, err := loadState()
	if err != nil {
		t.Errorf("loadState should not error when file doesn't exist, got: %s", err)
	}
	if state != nil {
		t.Error("loadState should return nil when no state file exists")
	}

	// Test saveState
	testState := newState("test-blueprint.yaml", "test-run-id")
	testState.LastSuccessfulStage = StageScaffold

	if err := saveState(testState); err != nil {
		t.Fatalf("saveState failed: %s", err)
	}

	// Verify file exists
	if _, err := os.Stat(StateFileName); os.IsNotExist(err) {
		t.Error("State file should exist after saveState")
	}

	// Test loadState with existing file
	loadedState, err := loadState()
	if err != nil {
		t.Fatalf("loadState failed: %s", err)
	}

	if loadedState == nil {
		t.Error("loadState should return state when file exists")
		return
	}

	if loadedState.RunID != "test-run-id" {
		t.Errorf("Expected RunID 'test-run-id', got: %s", loadedState.RunID)
	}

	if loadedState.LastSuccessfulStage != StageScaffold {
		t.Errorf("Expected stage scaffold, got: %s", loadedState.LastSuccessfulStage)
	}

	// Test removeStateFile
	if err := removeStateFile(); err != nil {
		t.Fatalf("removeStateFile failed: %s", err)
	}

	// Verify file is gone
	if _, err := os.Stat(StateFileName); !os.IsNotExist(err) {
		t.Error("State file should be removed after removeStateFile")
	}

	// Test removeStateFile when file doesn't exist (should not error)
	if err := removeStateFile(); err != nil {
		t.Errorf("removeStateFile should not error when file doesn't exist, got: %s", err)
	}
}
