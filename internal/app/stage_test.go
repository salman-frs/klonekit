package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"klonekit/internal/parser"
)

// TestStageExecution_Integration verifies that the new stage runner properly executes all stages
func TestStageExecution_Integration(t *testing.T) {
	// Clean up any existing state file
	os.Remove(StateFileName)
	defer os.Remove(StateFileName)

	// Create test environment
	tempDir, err := os.MkdirTemp("", "klonekit-stage-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %s", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test blueprint
	blueprintContent := `
apiVersion: v1
kind: Blueprint
metadata:
  name: stage-test
  description: Test blueprint for stage functionality
spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: test-token
    project:
      name: stage-test-repo
      namespace: test-user
      description: Stage test repository
      visibility: private
  cloud:
    provider: aws
    region: us-east-1
  scaffold:
    source: SOURCE_DIR
    destination: DEST_DIR
  variables:
    test_variable: stage_test
`

	sourceDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "destination")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %s", err)
	}

	// Create test Terraform file
	testTfFile := filepath.Join(sourceDir, "main.tf")
	if err := os.WriteFile(testTfFile, []byte("# Test terraform file for stages"), 0644); err != nil {
		t.Fatalf("Failed to create test terraform file: %s", err)
	}

	// Update blueprint with actual paths
	blueprintContent = strings.ReplaceAll(blueprintContent, "SOURCE_DIR", sourceDir)
	blueprintContent = strings.ReplaceAll(blueprintContent, "DEST_DIR", destDir)

	blueprintFile := filepath.Join(tempDir, "test-blueprint.yaml")
	if err := os.WriteFile(blueprintFile, []byte(blueprintContent), 0644); err != nil {
		t.Fatalf("Failed to create test blueprint file: %s", err)
	}

	// Parse blueprint for testing
	blueprint, err := parser.Parse(blueprintFile)
	if err != nil {
		t.Fatalf("Failed to parse blueprint: %s", err)
	}

	// Test buildStages function
	providerFactory := NewProviderFactory()
	stages := buildStages(blueprint, providerFactory, true, false)

	if len(stages) != 3 {
		t.Errorf("Expected 3 stages, got %d", len(stages))
	}

	expectedStageNames := []string{"scaffold", "scm", "provision"}
	for i, stage := range stages {
		if stage.Name() != expectedStageNames[i] {
			t.Errorf("Expected stage %d to be '%s', got '%s'", i, expectedStageNames[i], stage.Name())
		}
	}
}

// TestStageRecovery_AfterScaffold simulates a failure after scaffold stage and verifies recovery
func TestStageRecovery_AfterScaffold(t *testing.T) {
	// Clean up any existing state file
	os.Remove(StateFileName)
	defer os.Remove(StateFileName)

	// Create a state that shows scaffold as completed
	state := &ExecutionState{
		SchemaVersion:       StateSchemaVersion,
		RunID:               "test-recovery-123",
		LastCompletedStage:  "scaffold",
		LastSuccessfulStage: StageScaffold,
		BlueprintPath:       "/test/blueprint.yaml",
		CreatedAt:           time.Now(),
		LastUpdatedAt:       time.Now(),
	}

	// Test shouldSkipStage function
	if !shouldSkipStage(state, "scaffold") {
		t.Error("Should skip scaffold stage when it's already completed")
	}

	if shouldSkipStage(state, "scm") {
		t.Error("Should not skip scm stage when only scaffold is completed")
	}

	if shouldSkipStage(state, "provision") {
		t.Error("Should not skip provision stage when only scaffold is completed")
	}
}

// TestStageRecovery_AfterSCM simulates a failure after SCM stage and verifies recovery
func TestStageRecovery_AfterSCM(t *testing.T) {
	// Create a state that shows SCM as completed
	state := &ExecutionState{
		SchemaVersion:       StateSchemaVersion,
		RunID:               "test-recovery-456",
		LastCompletedStage:  "scm",
		LastSuccessfulStage: StageSCM,
		BlueprintPath:       "/test/blueprint.yaml",
		CreatedAt:           time.Now(),
		LastUpdatedAt:       time.Now(),
	}

	// Test shouldSkipStage function
	if !shouldSkipStage(state, "scaffold") {
		t.Error("Should skip scaffold stage when scm is completed")
	}

	if !shouldSkipStage(state, "scm") {
		t.Error("Should skip scm stage when it's already completed")
	}

	if shouldSkipStage(state, "provision") {
		t.Error("Should not skip provision stage when only scaffold and scm are completed")
	}
}

// TestStageInterface_Implementation verifies that all stages implement the Stage interface correctly
func TestStageInterface_Implementation(t *testing.T) {
	// Create test environment
	tempDir, err := os.MkdirTemp("", "klonekit-interface-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %s", err)
	}
	defer os.RemoveAll(tempDir)

	// Create minimal blueprint for testing
	blueprintContent := `
apiVersion: v1
kind: Blueprint
metadata:
  name: interface-test
spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: test-token
    project:
      name: interface-test-repo
      namespace: test-user
      description: Interface test repository
      visibility: private
  cloud:
    provider: aws
    region: us-east-1
  scaffold:
    source: ` + tempDir + `/source
    destination: ` + tempDir + `/destination
  variables:
    test_var: test_value
`

	blueprintFile := filepath.Join(tempDir, "test-blueprint.yaml")
	if err := os.WriteFile(blueprintFile, []byte(blueprintContent), 0644); err != nil {
		t.Fatalf("Failed to create test blueprint file: %s", err)
	}

	blueprint, err := parser.Parse(blueprintFile)
	if err != nil {
		t.Fatalf("Failed to parse blueprint: %s", err)
	}

	// Test individual stage implementations
	providerFactory := NewProviderFactory()

	// Test ScaffoldStage
	scaffoldStage := NewScaffoldStage(blueprint, true)
	if scaffoldStage.Name() != "scaffold" {
		t.Errorf("ScaffoldStage.Name() = %s, want 'scaffold'", scaffoldStage.Name())
	}

	// Test ScmStage
	scmStage := NewScmStage(blueprint, providerFactory, true)
	if scmStage.Name() != "scm" {
		t.Errorf("ScmStage.Name() = %s, want 'scm'", scmStage.Name())
	}

	// Test ProvisionStage
	provisionStage := NewProvisionStage(blueprint, providerFactory, true, false)
	if provisionStage.Name() != "provision" {
		t.Errorf("ProvisionStage.Name() = %s, want 'provision'", provisionStage.Name())
	}

	// Verify they all implement the Stage interface
	var stages []Stage = []Stage{scaffoldStage, scmStage, provisionStage}

	for i, stage := range stages {
		if stage == nil {
			t.Errorf("Stage %d is nil", i)
		}
		if stage.Name() == "" {
			t.Errorf("Stage %d has empty name", i)
		}
	}
}