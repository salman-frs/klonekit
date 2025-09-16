package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"klonekit/internal/parser"
	"klonekit/pkg/blueprint"
)

const (
	// Color codes for console output
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

// Apply orchestrates the complete KloneKit workflow using a dynamic stage runner.
// This function implements the Facade pattern over all internal components with resume capability.
func Apply(blueprintPath string, isDryRun bool, retainState bool, autoApprove bool) error {
	slog.Info("Starting KloneKit apply workflow", "blueprintPath", blueprintPath, "dryRun", isDryRun)

	// Load existing state or create new state
	state, err := loadState()
	if err != nil {
		return fmt.Errorf("failed to load execution state: %w", err)
	}

	var isResume bool
	if state == nil {
		// Fresh start - create new state
		runID := uuid.New().String()
		state = newState(blueprintPath, runID)
		slog.Info("Starting new KloneKit workflow", "runId", runID, "blueprintPath", blueprintPath)
	} else {
		// Resume existing run
		isResume = true
		nextStage := state.getNextStage()
		fmt.Printf("%s📋 State file found. Resuming from stage: %s%s\n", ColorYellow, nextStage, ColorReset)
		slog.Info("Resuming KloneKit workflow", "runId", state.RunID, "nextStage", nextStage, "lastStage", state.LastSuccessfulStage)
		fmt.Println()
	}

	if isDryRun {
		fmt.Printf("%s🔍 DRY RUN MODE - No actual changes will be made%s\n", ColorYellow, ColorReset)
		if isResume {
			fmt.Printf("%s🔍 DRY RUN: Simulating resume from stage: %s%s\n", ColorYellow, state.getNextStage(), ColorReset)
		}
		fmt.Println()
	}

	// Parse blueprint (needed for all stages)
	blueprint, err := parser.Parse(blueprintPath)
	if err != nil {
		return fmt.Errorf("blueprint parsing failed: %w", err)
	}
	slog.Info("Blueprint parsed successfully", "name", blueprint.Metadata.Name, "kind", blueprint.Kind)

	// Build the stages slice
	providerFactory := NewProviderFactory()
	stages := buildStages(blueprint, providerFactory, isDryRun, autoApprove)

	// Execute stages using the dynamic stage runner
	ctx := context.Background()
	if err := runStages(ctx, stages, state, isDryRun); err != nil {
		return fmt.Errorf("stage execution failed: %w", err)
	}

	// Mark workflow as completed and clean up state file
	state.LastSuccessfulStage = StageCompleted
	state.LastCompletedStage = "completed"
	if !isDryRun {
		if retainState {
			// Save final state for auditing purposes
			if err := saveState(state); err != nil {
				slog.Warn("Failed to save final state", "error", err)
			} else {
				slog.Info("State file retained for auditing", "file", StateFileName)
			}
		} else {
			// Remove state file on successful completion
			if err := removeStateFile(); err != nil {
				slog.Warn("Failed to clean up state file", "error", err)
			}
		}
	}

	// Workflow completion
	if isDryRun {
		fmt.Printf("%s🎉 DRY RUN COMPLETED - All stages simulated successfully!%s\n", ColorGreen, ColorReset)
		fmt.Printf("%sNo actual resources were created or modified.%s\n", ColorYellow, ColorReset)
	} else {
		fmt.Printf("%s🎉 KLONEKIT APPLY COMPLETED SUCCESSFULLY!%s\n", ColorGreen, ColorReset)
		fmt.Printf("%s✨ Your infrastructure project '%s' is ready!%s\n", ColorWhite, blueprint.Metadata.Name, ColorReset)
	}

	slog.Info("KloneKit apply workflow completed successfully", "blueprintName", blueprint.Metadata.Name, "dryRun", isDryRun)
	return nil
}

// buildStages constructs the slice of stages to be executed based on the blueprint
func buildStages(blueprint *blueprint.Blueprint, providerFactory *ProviderFactory, isDryRun bool, autoApprove bool) []Stage {
	stages := []Stage{
		NewScaffoldStage(blueprint, isDryRun),
		NewScmStage(blueprint, providerFactory, isDryRun),
		NewProvisionStage(blueprint, providerFactory, isDryRun, autoApprove),
	}
	return stages
}

// runStages executes the stages in order, skipping those already completed
func runStages(ctx context.Context, stages []Stage, state *ExecutionState, isDryRun bool) error {
	for i, stage := range stages {
		stageName := stage.Name()

		// Check if this stage should be skipped
		if shouldSkipStage(state, stageName) {
			fmt.Printf("%s⏭️  Stage %d: %s (skipped - already completed)%s\n", ColorGreen, i+1, stageName, ColorReset)
			fmt.Println()
			continue
		}

		// Execute the stage
		fmt.Printf("%s🔄 Stage %d: %s%s\n", getStageColor(stageName), i+1, stageName, ColorReset)
		if err := stage.Execute(ctx, state); err != nil {
			return fmt.Errorf("stage '%s' failed: %w", stageName, err)
		}

		// Update state after successful completion
		state.LastCompletedStage = stageName
		// Update legacy field for backward compatibility
		switch stageName {
		case "scaffold":
			state.LastSuccessfulStage = StageScaffold
		case "scm":
			state.LastSuccessfulStage = StageSCM
		case "provision":
			state.LastSuccessfulStage = StageProvision
		}

		if !isDryRun {
			if err := saveState(state); err != nil {
				return fmt.Errorf("failed to save state after stage '%s': %w", stageName, err)
			}
		}
		fmt.Println()
	}
	return nil
}

// shouldSkipStage determines if a stage should be skipped based on the current state
func shouldSkipStage(state *ExecutionState, stageName string) bool {
	if state == nil || state.LastCompletedStage == "" {
		return false // Fresh start, don't skip any stage
	}

	// If the stage was already completed, skip it
	completedStages := getCompletedStages(state.LastCompletedStage)
	for _, completed := range completedStages {
		if completed == stageName {
			return true
		}
	}
	return false
}

// getCompletedStages returns the list of stages that are considered completed
// based on the last completed stage
func getCompletedStages(lastCompletedStage string) []string {
	switch lastCompletedStage {
	case "completed":
		return []string{"scaffold", "scm", "provision"}
	case "provision":
		return []string{"scaffold", "scm", "provision"}
	case "scm":
		return []string{"scaffold", "scm"}
	case "scaffold":
		return []string{"scaffold"}
	default:
		return []string{}
	}
}

// getStageColor returns the appropriate color for each stage
func getStageColor(stageName string) string {
	switch stageName {
	case "scaffold":
		return ColorCyan
	case "scm":
		return ColorPurple
	case "provision":
		return ColorRed
	default:
		return ColorWhite
	}
}


// ValidatePrerequisites checks that all required external dependencies are available.
func ValidatePrerequisites() error {
	slog.Info("Validating KloneKit prerequisites")

	// Check if Docker is available (required for provisioning) by attempting to create factory and provisioner
	factory := NewProviderFactory()
	_, err := factory.GetProvisioner("aws")
	if err != nil {
		return fmt.Errorf("Docker prerequisite check failed: %w", err)
	}

	slog.Info("All prerequisites validated successfully")
	return nil
}
