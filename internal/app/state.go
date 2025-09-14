package app

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ExecutionStage represents the stages of the apply workflow
type ExecutionStage string

const (
	StageScaffold  ExecutionStage = "scaffold"
	StageSCM       ExecutionStage = "scm"
	StageProvision ExecutionStage = "provision"
	StageCompleted ExecutionStage = "completed"
)

// ExecutionState represents the state of a KloneKit apply run
type ExecutionState struct {
	SchemaVersion       string         `json:"schema_version"`
	RunID               string         `json:"run_id"`
	LastSuccessfulStage ExecutionStage `json:"last_successful_stage"`
	BlueprintPath       string         `json:"blueprint_path"`
	CreatedAt           time.Time      `json:"created_at"`
	LastUpdatedAt       time.Time      `json:"last_updated_at"`
}

const (
	StateFileName      = ".klonekit.state.json"
	StateSchemaVersion = "1.0"
)

// loadState attempts to load the execution state from the state file.
// Returns nil if the file doesn't exist (fresh start).
func loadState() (*ExecutionState, error) {
	if _, err := os.Stat(StateFileName); os.IsNotExist(err) {
		return nil, nil // Fresh start - no state file exists
	}

	data, err := os.ReadFile(StateFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state ExecutionState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &state, nil
}

// saveState persists the execution state to the state file.
func saveState(state *ExecutionState) error {
	state.LastUpdatedAt = time.Now()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize state: %w", err)
	}

	if err := os.WriteFile(StateFileName, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// newState creates a new execution state for a fresh run
func newState(blueprintPath, runID string) *ExecutionState {
	now := time.Now()
	return &ExecutionState{
		SchemaVersion:       StateSchemaVersion,
		RunID:               runID,
		LastSuccessfulStage: "", // No stage completed yet
		BlueprintPath:       blueprintPath,
		CreatedAt:           now,
		LastUpdatedAt:       now,
	}
}

// shouldSkipStage determines if a stage should be skipped based on the current state
func (s *ExecutionState) shouldSkipStage(stage ExecutionStage) bool {
	if s == nil || s.LastSuccessfulStage == "" {
		return false // Fresh start, don't skip any stage
	}

	switch stage {
	case StageScaffold:
		return s.LastSuccessfulStage == StageScaffold ||
			s.LastSuccessfulStage == StageSCM ||
			s.LastSuccessfulStage == StageProvision ||
			s.LastSuccessfulStage == StageCompleted
	case StageSCM:
		return s.LastSuccessfulStage == StageSCM ||
			s.LastSuccessfulStage == StageProvision ||
			s.LastSuccessfulStage == StageCompleted
	case StageProvision:
		return s.LastSuccessfulStage == StageProvision ||
			s.LastSuccessfulStage == StageCompleted
	default:
		return false
	}
}

// getNextStage returns the next stage to execute based on the current state
func (s *ExecutionState) getNextStage() ExecutionStage {
	if s == nil || s.LastSuccessfulStage == "" {
		return StageScaffold
	}

	switch s.LastSuccessfulStage {
	case StageScaffold:
		return StageSCM
	case StageSCM:
		return StageProvision
	case StageProvision:
		return StageCompleted
	default:
		return StageScaffold
	}
}

// removeStateFile removes the state file after successful completion
func removeStateFile() error {
	if _, err := os.Stat(StateFileName); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to remove
	}

	if err := os.Remove(StateFileName); err != nil {
		return fmt.Errorf("failed to remove state file: %w", err)
	}

	return nil
}
