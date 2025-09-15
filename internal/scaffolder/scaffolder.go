package scaffolder

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"klonekit/pkg/blueprint"
)

// Scaffold processes a blueprint spec and generates Terraform files.
// It copies the source module directory to the destination and creates terraform.tfvars.json.
func Scaffold(spec *blueprint.Spec, isDryRun bool) error {
	if spec == nil {
		return fmt.Errorf("spec cannot be nil")
	}

	sourcePath := spec.Scaffold.Source
	destPath := spec.Scaffold.Destination

	// Validate source path exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source module directory not found: %s", sourcePath)
	}

	if isDryRun {
		return performDryRun(spec)
	}

	// Create destination directory
	if err := os.MkdirAll(destPath, 0750); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Copy source directory to destination
	if err := copyDirectory(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to copy source directory: %w", err)
	}

	// Generate terraform.tfvars.json file
	if err := generateTerraformVars(spec, destPath); err != nil {
		return fmt.Errorf("failed to generate terraform.tfvars.json: %w", err)
	}

	return nil
}

// performDryRun logs what would be done without actually performing the operations.
func performDryRun(spec *blueprint.Spec) error {
	sourcePath := spec.Scaffold.Source
	destPath := spec.Scaffold.Destination

	fmt.Printf("DRY RUN: Would copy directory from %s to %s\n", sourcePath, destPath)

	// Walk through source directory to show what would be copied
	err := filepath.WalkDir(sourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}

		destFile := filepath.Join(destPath, relPath)
		if d.IsDir() {
			fmt.Printf("DRY RUN: Would create directory: %s\n", destFile)
		} else {
			fmt.Printf("DRY RUN: Would copy file: %s\n", destFile)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk source directory: %w", err)
	}

	// Show terraform.tfvars.json that would be generated
	tfvarsPath := filepath.Join(destPath, "terraform.tfvars.json")
	fmt.Printf("DRY RUN: Would create file: %s\n", tfvarsPath)

	// Use only user-defined variables
	allVars := spec.Variables
	if len(allVars) > 0 {
		fmt.Println("DRY RUN: terraform.tfvars.json content would be:")
		if jsonBytes, err := json.MarshalIndent(allVars, "", "  "); err == nil {
			fmt.Println(string(jsonBytes))
		}
	}

	return nil
}

// copyDirectory recursively copies a directory from src to dst.
func copyDirectory(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0750)
		}

		return copyFile(path, destPath)
	})
}

// validatePath ensures the path is safe and doesn't contain directory traversal sequences
func validatePath(path string) error {
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path contains directory traversal: %s", path)
	}
	return nil
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	// Validate paths to prevent directory traversal
	if err := validatePath(src); err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}
	if err := validatePath(dst); err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Copy file permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// generateTerraformVars creates a terraform.tfvars.json file with the variables from the blueprint.
func generateTerraformVars(spec *blueprint.Spec, destPath string) error {
	// Use only user-defined variables
	allVars := spec.Variables

	if len(allVars) == 0 {
		return nil
	}

	tfvarsPath := filepath.Join(destPath, "terraform.tfvars.json")

	jsonBytes, err := json.MarshalIndent(allVars, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal variables to JSON: %w", err)
	}

	if err := os.WriteFile(tfvarsPath, jsonBytes, 0600); err != nil {
		return fmt.Errorf("failed to write terraform.tfvars.json: %w", err)
	}

	return nil
}

