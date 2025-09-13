package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"klonekit/internal/parser"
	"klonekit/internal/scaffolder"
)

var rootCmd = &cobra.Command{
	Use:   "klonekit",
	Short: "KloneKit - Infrastructure provisioning and GitLab project setup tool",
	Long: `KloneKit is a CLI tool that helps DevOps engineers provision infrastructure
and set up GitLab projects using blueprint configurations.`,
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a blueprint configuration",
	Long: `Apply processes a blueprint YAML file and executes the infrastructure
provisioning and GitLab setup according to the configuration.`,
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

		fmt.Printf("Successfully parsed blueprint: %s\n", blueprint.Metadata.Name)
		fmt.Printf("  Kind: %s\n", blueprint.Kind)
		fmt.Printf("  API Version: %s\n", blueprint.APIVersion)
		fmt.Printf("  Description: %s\n", blueprint.Metadata.Description)
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

func init() {
	applyCmd.Flags().StringP("file", "f", "", "Path to the blueprint YAML file (required)")
	applyCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(applyCmd)

	scaffoldCmd.Flags().StringP("file", "f", "", "Path to the blueprint YAML file (required)")
	scaffoldCmd.Flags().Bool("dry-run", false, "Print files that would be created without actually writing them")
	scaffoldCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(scaffoldCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
