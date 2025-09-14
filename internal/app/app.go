package app

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"klonekit/internal/parser"
	"klonekit/internal/provisioner"
	"klonekit/internal/runtime"
	"klonekit/internal/scaffolder"
	"klonekit/internal/scm"
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

// Apply orchestrates the complete KloneKit workflow using a stateful execution engine.
// This function implements the Facade pattern over all internal components with resume capability.
func Apply(blueprintPath string, isDryRun bool, retainState bool) error {
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

	// Stage 1: Scaffold Terraform files
	if !state.shouldSkipStage(StageScaffold) {
		fmt.Printf("%s🚧 Stage 1: Scaffolding Terraform files%s\n", ColorCyan, ColorReset)
		if err := executeScaffoldStage(blueprint, isDryRun); err != nil {
			return fmt.Errorf("scaffolding failed: %w", err)
		}

		// Update state after successful completion
		state.LastSuccessfulStage = StageScaffold
		if !isDryRun {
			if err := saveState(state); err != nil {
				return fmt.Errorf("failed to save state after scaffolding: %w", err)
			}
		}
		fmt.Println()
	} else {
		fmt.Printf("%s⏭️  Stage 1: Scaffolding (skipped - already completed)%s\n", ColorGreen, ColorReset)
		fmt.Println()
	}

	// Stage 2: Source Control Management
	if !state.shouldSkipStage(StageSCM) {
		fmt.Printf("%s📱 Stage 2: Creating GitLab repository%s\n", ColorPurple, ColorReset)
		if err := executeSCMStage(blueprint, isDryRun); err != nil {
			return fmt.Errorf("SCM stage failed: %w", err)
		}

		// Update state after successful completion
		state.LastSuccessfulStage = StageSCM
		if !isDryRun {
			if err := saveState(state); err != nil {
				return fmt.Errorf("failed to save state after SCM: %w", err)
			}
		}
		fmt.Println()
	} else {
		fmt.Printf("%s⏭️  Stage 2: SCM (skipped - already completed)%s\n", ColorGreen, ColorReset)
		fmt.Println()
	}

	// Stage 3: Infrastructure Provisioning
	if !state.shouldSkipStage(StageProvision) {
		fmt.Printf("%s🏗️  Stage 3: Provisioning infrastructure%s\n", ColorRed, ColorReset)
		if err := executeProvisionStage(blueprint, isDryRun); err != nil {
			return fmt.Errorf("provisioning stage failed: %w", err)
		}

		// Update state after successful completion
		state.LastSuccessfulStage = StageProvision
		if !isDryRun {
			if err := saveState(state); err != nil {
				return fmt.Errorf("failed to save state after provisioning: %w", err)
			}
		}
		fmt.Println()
	} else {
		fmt.Printf("%s⏭️  Stage 3: Provisioning (skipped - already completed)%s\n", ColorGreen, ColorReset)
		fmt.Println()
	}

	// Mark workflow as completed and clean up state file
	state.LastSuccessfulStage = StageCompleted
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

// executeScaffoldStage handles the scaffolding stage of the workflow
func executeScaffoldStage(blueprint *blueprint.Blueprint, isDryRun bool) error {
	if err := scaffolder.Scaffold(&blueprint.Spec, isDryRun); err != nil {
		return err
	}

	if isDryRun {
		fmt.Printf("%s✅ Scaffolding simulation completed successfully%s\n", ColorGreen, ColorReset)
	} else {
		fmt.Printf("%s✅ Terraform files scaffolded to: %s%s\n", ColorGreen, blueprint.Spec.Scaffold.Destination, ColorReset)
	}
	slog.Info("Scaffolding completed successfully", "destination", blueprint.Spec.Scaffold.Destination, "dryRun", isDryRun)
	return nil
}

// executeSCMStage handles the source control management stage of the workflow
func executeSCMStage(blueprint *blueprint.Blueprint, isDryRun bool) error {
	if isDryRun {
		fmt.Printf("%s🔍 DRY RUN: Would create GitLab repository '%s' in namespace '%s'%s\n",
			ColorYellow, blueprint.Spec.SCM.Project.Name, blueprint.Spec.SCM.Project.Namespace, ColorReset)
		fmt.Printf("%s🔍 DRY RUN: Would push scaffolded files to repository%s\n", ColorYellow, ColorReset)
	} else {
		provider, err := scm.NewGitLabProvider()
		if err != nil {
			return fmt.Errorf("SCM provider initialization failed: %w", err)
		}

		if err := provider.CreateRepo(&blueprint.Spec); err != nil {
			return fmt.Errorf("GitLab repository creation failed: %w", err)
		}
	}

	if isDryRun {
		fmt.Printf("%s✅ SCM simulation completed successfully%s\n", ColorGreen, ColorReset)
	} else {
		fmt.Printf("%s✅ GitLab repository created: %s%s\n", ColorGreen, blueprint.Spec.SCM.Project.Name, ColorReset)
	}
	slog.Info("SCM stage completed successfully", "repoName", blueprint.Spec.SCM.Project.Name, "dryRun", isDryRun)
	return nil
}

// executeProvisionStage handles the infrastructure provisioning stage of the workflow
func executeProvisionStage(blueprint *blueprint.Blueprint, isDryRun bool) error {
	if isDryRun {
		fmt.Printf("%s🔍 DRY RUN: Would pull Terraform Docker image%s\n", ColorYellow, ColorReset)
		fmt.Printf("%s🔍 DRY RUN: Would execute 'terraform init' in container%s\n", ColorYellow, ColorReset)
		fmt.Printf("%s🔍 DRY RUN: Would execute 'terraform apply -auto-approve' in container%s\n", ColorYellow, ColorReset)
		fmt.Printf("%s🔍 DRY RUN: Would provision infrastructure in %s region%s\n",
			ColorYellow, blueprint.Spec.Cloud.Region, ColorReset)
	} else {
		// Create Docker runtime instance
		dockerRuntime, err := runtime.NewDockerRuntime()
		if err != nil {
			return fmt.Errorf("failed to create Docker runtime: %w", err)
		}

		// Create provisioner with the runtime
		terraformProvisioner := provisioner.NewTerraformDockerProvisioner(dockerRuntime)

		if err := terraformProvisioner.Provision(&blueprint.Spec); err != nil {
			return fmt.Errorf("infrastructure provisioning failed: %w", err)
		}
	}

	if isDryRun {
		fmt.Printf("%s✅ Provisioning simulation completed successfully%s\n", ColorGreen, ColorReset)
	} else {
		fmt.Printf("%s✅ Infrastructure provisioned successfully in %s%s\n", ColorGreen, blueprint.Spec.Cloud.Region, ColorReset)
	}
	slog.Info("Provisioning stage completed successfully", "region", blueprint.Spec.Cloud.Region, "dryRun", isDryRun)
	return nil
}

// ValidatePrerequisites checks that all required external dependencies are available.
func ValidatePrerequisites() error {
	slog.Info("Validating KloneKit prerequisites")

	// Check if Docker is available (required for provisioning)
	if _, err := runtime.NewDockerRuntime(); err != nil {
		return fmt.Errorf("Docker prerequisite check failed: %w", err)
	}

	slog.Info("All prerequisites validated successfully")
	return nil
}
