package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"klonekit/internal/app"
	"klonekit/internal/parser"
	"klonekit/internal/provisioner"
	"klonekit/internal/runtime"
	"klonekit/internal/scaffolder"
	"klonekit/internal/scm"
)

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
		file, _ := cmd.Flags().GetString("file")
		if file == "" {
			fmt.Fprintln(os.Stderr, "Error: --file flag is required")
			os.Exit(1)
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		retainState, _ := cmd.Flags().GetBool("retain-state")

		// Execute the complete workflow via app orchestrator
		if err := app.Apply(file, dryRun, retainState); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
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
		file, _ := cmd.Flags().GetString("file")
		if file == "" {
			fmt.Fprintln(os.Stderr, "Error: --file flag is required")
			os.Exit(1)
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Parse and validate the blueprint file
		blueprint, err := parser.Parse(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		// Process the blueprint with the scaffolder
		fmt.Printf("Scaffolding blueprint: %s\n", blueprint.Metadata.Name)

		if err := scaffolder.Scaffold(&blueprint.Spec, dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
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
		file, _ := cmd.Flags().GetString("file")
		if file == "" {
			fmt.Fprintln(os.Stderr, "Error: --file flag is required")
			os.Exit(1)
		}

		// Parse and validate the blueprint file
		blueprint, err := parser.Parse(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		// Create GitLab repository and push scaffolded files
		fmt.Printf("Creating GitLab repository for: %s\n", blueprint.Metadata.Name)

		provider, err := scm.NewGitLabProvider()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		if err := provider.CreateRepo(&blueprint.Spec); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
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
		file, _ := cmd.Flags().GetString("file")
		if file == "" {
			fmt.Fprintln(os.Stderr, "Error: --file flag is required")
			os.Exit(1)
		}

		// Parse and validate the blueprint file
		blueprint, err := parser.Parse(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		// Provision infrastructure using Docker
		fmt.Printf("Provisioning infrastructure for: %s\n", blueprint.Metadata.Name)

		// Create Docker runtime instance
		dockerRuntime, err := runtime.NewDockerRuntime()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating Docker runtime: %s\n", err)
			os.Exit(1)
		}

		// Create provisioner with the runtime
		terraformProvisioner := provisioner.NewTerraformDockerProvisioner(dockerRuntime)

		if err := terraformProvisioner.Provision(&blueprint.Spec); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully provisioned infrastructure for: %s\n", blueprint.Metadata.Name)
	},
}

func init() {
	applyCmd.Flags().StringP("file", "f", "", "Path to the blueprint YAML file (required)")
	applyCmd.Flags().Bool("dry-run", false, "Simulate the workflow without making any changes")
	applyCmd.Flags().Bool("retain-state", false, "Keep the state file after successful completion for auditing purposes")
	if err := applyCmd.MarkFlagRequired("file"); err != nil {
		slog.Error("Failed to mark file flag as required for apply command", "error", err)
	}
	rootCmd.AddCommand(applyCmd)

	scaffoldCmd.Flags().StringP("file", "f", "", "Path to the blueprint YAML file (required)")
	scaffoldCmd.Flags().Bool("dry-run", false, "Print files that would be created without actually writing them")
	if err := scaffoldCmd.MarkFlagRequired("file"); err != nil {
		slog.Error("Failed to mark file flag as required for scaffold command", "error", err)
	}
	rootCmd.AddCommand(scaffoldCmd)

	scmCmd.Flags().StringP("file", "f", "", "Path to the blueprint YAML file (required)")
	if err := scmCmd.MarkFlagRequired("file"); err != nil {
		slog.Error("Failed to mark file flag as required for scm command", "error", err)
	}
	rootCmd.AddCommand(scmCmd)

	provisionCmd.Flags().StringP("file", "f", "", "Path to the blueprint YAML file (required)")
	if err := provisionCmd.MarkFlagRequired("file"); err != nil {
		slog.Error("Failed to mark file flag as required for provision command", "error", err)
	}
	rootCmd.AddCommand(provisionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
