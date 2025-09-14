package app

import (
	"fmt"
	"log/slog"

	"klonekit/internal/parser"
	"klonekit/internal/provisioner"
	"klonekit/internal/runtime"
	"klonekit/internal/scaffolder"
	"klonekit/internal/scm"
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

// Apply orchestrates the complete KloneKit workflow: parse, scaffold, scm, provision.
// This function implements the Facade pattern over all internal components.
func Apply(blueprintPath string, isDryRun bool) error {
	slog.Info("Starting KloneKit apply workflow", "blueprintPath", blueprintPath, "dryRun", isDryRun)

	if isDryRun {
		fmt.Printf("%s🔍 DRY RUN MODE - No actual changes will be made%s\n", ColorYellow, ColorReset)
		fmt.Println()
	}

	// Stage 1: Parse and validate blueprint
	fmt.Printf("%s📋 Stage 1: Parsing blueprint configuration%s\n", ColorBlue, ColorReset)
	blueprint, err := parser.Parse(blueprintPath)
	if err != nil {
		return fmt.Errorf("blueprint parsing failed: %w", err)
	}
	fmt.Printf("%s✅ Blueprint parsed successfully: %s%s\n", ColorGreen, blueprint.Metadata.Name, ColorReset)
	slog.Info("Blueprint parsed successfully", "name", blueprint.Metadata.Name, "kind", blueprint.Kind)
	fmt.Println()

	// Stage 2: Scaffold Terraform files
	fmt.Printf("%s🚧 Stage 2: Scaffolding Terraform files%s\n", ColorCyan, ColorReset)
	if err := scaffolder.Scaffold(&blueprint.Spec, isDryRun); err != nil {
		return fmt.Errorf("scaffolding failed: %w", err)
	}

	if isDryRun {
		fmt.Printf("%s✅ Scaffolding simulation completed successfully%s\n", ColorGreen, ColorReset)
	} else {
		fmt.Printf("%s✅ Terraform files scaffolded to: %s%s\n", ColorGreen, blueprint.Spec.Scaffold.Destination, ColorReset)
	}
	slog.Info("Scaffolding completed successfully", "destination", blueprint.Spec.Scaffold.Destination, "dryRun", isDryRun)
	fmt.Println()

	// Stage 3: Source Control Management
	fmt.Printf("%s📱 Stage 3: Creating GitLab repository%s\n", ColorPurple, ColorReset)
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
	fmt.Println()

	// Stage 4: Infrastructure Provisioning
	fmt.Printf("%s🏗️  Stage 4: Provisioning infrastructure%s\n", ColorRed, ColorReset)
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
	fmt.Println()

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
