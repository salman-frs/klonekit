package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"klonekit/internal/app"
	"klonekit/internal/errors"
	"klonekit/internal/parser"
	"klonekit/internal/provisioner"
	"klonekit/internal/runtime"
	"klonekit/internal/scaffolder"
	"klonekit/internal/scm"
)

// findBlueprintFile searches for klonekit.yml or klonekit.yaml in the current directory
func findBlueprintFile() string {
	files := []string{"klonekit.yml", "klonekit.yaml"}
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			return file
		}
	}
	return ""
}

// getFileFlag gets the file flag value, falling back to auto-detection if not provided
func getFileFlag(cmd *cobra.Command) (string, error) {
	file, _ := cmd.Flags().GetString("file")
	if file != "" {
		return file, nil
	}

	// Try to auto-detect blueprint file
	autoDetected := findBlueprintFile()
	if autoDetected == "" {
		return "", errors.NewBlueprintError(
			"Failed to locate blueprint file",
			"No klonekit.yml or klonekit.yaml file found in current directory",
			"Create a blueprint file (klonekit.yml or klonekit.yaml) or specify one with -f flag",
			fmt.Errorf("no blueprint file found in current directory"),
		)
	}

	return autoDetected, nil
}

// version is set at build time via ldflags
var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "klonekit",
	Short:   "KloneKit - Infrastructure provisioning and GitLab project setup tool",
	Version: version,
	Long: `KloneKit is a CLI tool that helps DevOps engineers provision infrastructure
and set up GitLab projects using blueprint configurations.`,
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a complete blueprint workflow",
	Long: `Apply executes the complete KloneKit workflow: scaffolding Terraform files,
creating GitLab repositories, and provisioning infrastructure - all from a single command.

This orchestrates all individual commands (scaffold, scm, provision) in the correct sequence.`,
	Run: func(cmd *cobra.Command, args []string) {
		file, err := getFileFlag(cmd)
		if err != nil {
			errors.HandleError(err)
			os.Exit(1)
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		retainState, _ := cmd.Flags().GetBool("retain-state")
		autoApprove, _ := cmd.Flags().GetBool("auto-approve")

		// Execute the complete workflow via app orchestrator
		if err := app.Apply(file, dryRun, retainState, autoApprove); err != nil {
			errors.HandleError(err)
			os.Exit(1)
		}
	},
}

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Generate Terraform files from a blueprint",
	Long: `Scaffold processes a blueprint YAML file and generates the complete set
of Terraform files locally for verification before infrastructure creation.`,
	Run: func(cmd *cobra.Command, args []string) {
		file, err := getFileFlag(cmd)
		if err != nil {
			errors.HandleError(err)
			os.Exit(1)
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Parse and validate the blueprint file
		blueprint, err := parser.Parse(file)
		if err != nil {
			errors.HandleError(err)
			os.Exit(1)
		}

		// Process the blueprint with the scaffolder
		fmt.Printf("Scaffolding blueprint: %s\n", blueprint.Metadata.Name)

		if err := scaffolder.Scaffold(&blueprint.Spec, dryRun); err != nil {
			errors.HandleError(err)
			os.Exit(1)
		}

		if dryRun {
			fmt.Println("Dry run completed successfully.")
		} else {
			fmt.Printf("Scaffolding completed successfully. Files written to: %s\n", blueprint.Spec.Scaffold.Destination)
		}
	},
}

var scmCmd = &cobra.Command{
	Use:   "scm",
	Short: "Create GitLab repository from scaffolded project",
	Long: `SCM processes a scaffolded project directory and publishes it to a new
GitLab repository using the GitLab API and git operations.`,
	Run: func(cmd *cobra.Command, args []string) {
		file, err := getFileFlag(cmd)
		if err != nil {
			errors.HandleError(err)
			os.Exit(1)
		}

		// Parse and validate the blueprint file
		blueprint, err := parser.Parse(file)
		if err != nil {
			errors.HandleError(err)
			os.Exit(1)
		}

		// Create GitLab repository and push scaffolded files
		fmt.Printf("Creating GitLab repository for: %s\n", blueprint.Metadata.Name)

		provider, err := scm.NewGitLabProvider()
		if err != nil {
			errors.HandleError(err)
			os.Exit(1)
		}

		if err := provider.CreateRepo(&blueprint.Spec); err != nil {
			errors.HandleError(err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created GitLab repository: %s\n", blueprint.Spec.SCM.Project.Name)
	},
}

var provisionCmd = &cobra.Command{
	Use:   "provision",
	Short: "Provision infrastructure using containerized Terraform",
	Long: `Provision executes Terraform commands within a Docker container to provision
infrastructure defined in the scaffolded Terraform files. This ensures a consistent
and isolated environment for infrastructure provisioning.`,
	Run: func(cmd *cobra.Command, args []string) {
		file, err := getFileFlag(cmd)
		if err != nil {
			errors.HandleError(err)
			os.Exit(1)
		}

		autoApprove, _ := cmd.Flags().GetBool("auto-approve")

		// Parse and validate the blueprint file
		blueprint, err := parser.Parse(file)
		if err != nil {
			errors.HandleError(err)
			os.Exit(1)
		}

		// Provision infrastructure using Docker
		fmt.Printf("Provisioning infrastructure for: %s\n", blueprint.Metadata.Name)

		// Create Docker runtime instance
		dockerRuntime, err := runtime.NewDockerRuntime()
		if err != nil {
			errors.HandleError(err)
			os.Exit(1)
		}

		// Create provisioner with the runtime
		terraformProvisioner := provisioner.NewTerraformDockerProvisioner(dockerRuntime)

		if err := terraformProvisioner.Provision(&blueprint.Spec, autoApprove); err != nil {
			errors.HandleError(err)
			os.Exit(1)
		}

		if autoApprove {
			fmt.Printf("Successfully provisioned infrastructure for: %s\n", blueprint.Metadata.Name)
		} else {
			fmt.Printf("Successfully validated infrastructure for: %s (use --auto-approve to provision)\n", blueprint.Metadata.Name)
		}
	},
}

func init() {
	applyCmd.Flags().StringP("file", "f", "", "Path to the blueprint YAML file (auto-detects klonekit.yml/klonekit.yaml if not specified)")
	applyCmd.Flags().Bool("dry-run", false, "Simulate the workflow without making any changes")
	applyCmd.Flags().Bool("retain-state", false, "Keep the state file after successful completion for auditing purposes")
	applyCmd.Flags().Bool("auto-approve", false, "Automatically approve terraform apply without prompting")
	rootCmd.AddCommand(applyCmd)

	scaffoldCmd.Flags().StringP("file", "f", "", "Path to the blueprint YAML file (auto-detects klonekit.yml/klonekit.yaml if not specified)")
	scaffoldCmd.Flags().Bool("dry-run", false, "Print files that would be created without actually writing them")
	rootCmd.AddCommand(scaffoldCmd)

	scmCmd.Flags().StringP("file", "f", "", "Path to the blueprint YAML file (auto-detects klonekit.yml/klonekit.yaml if not specified)")
	rootCmd.AddCommand(scmCmd)

	provisionCmd.Flags().StringP("file", "f", "", "Path to the blueprint YAML file (auto-detects klonekit.yml/klonekit.yaml if not specified)")
	provisionCmd.Flags().Bool("auto-approve", false, "Automatically approve terraform apply without prompting")
	rootCmd.AddCommand(provisionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		errors.HandleError(err)
		os.Exit(1)
	}
}
