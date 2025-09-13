package scaffolder

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"klonekit/pkg/blueprint"
)

func TestScaffold_ValidSpec(t *testing.T) {
	// Create temporary directories for testing
	tmpDir, err := os.MkdirTemp("", "klonekit-scaffold-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "source")
	dstDir := filepath.Join(tmpDir, "destination")

	// Create source directory with test files
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test terraform files in source
	testFiles := map[string]string{
		"main.tf":      "resource \"aws_instance\" \"test\" {}",
		"variables.tf": "variable \"instance_type\" {}",
		"outputs.tf":   "output \"instance_id\" {}",
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(srcDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create test spec
	spec := &blueprint.Spec{
		Scaffold: blueprint.Scaffold{
			Source:      srcDir,
			Destination: dstDir,
		},
		Variables: map[string]interface{}{
			"instance_type": "t3.micro",
			"vpc_cidr":      "10.0.0.0/16",
		},
	}

	// Execute scaffold
	err = Scaffold(spec, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify destination directory exists
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		t.Fatal("Destination directory was not created")
	}

	// Verify all source files were copied
	for filename, expectedContent := range testFiles {
		filePath := filepath.Join(dstDir, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Failed to read copied file %s: %v", filename, err)
			continue
		}

		if string(content) != expectedContent {
			t.Errorf("File %s content mismatch. Expected: %s, Got: %s", filename, expectedContent, string(content))
		}
	}

	// Verify terraform.tfvars.json was created
	tfvarsPath := filepath.Join(dstDir, "terraform.tfvars.json")
	tfvarsContent, err := os.ReadFile(tfvarsPath)
	if err != nil {
		t.Fatalf("terraform.tfvars.json not created: %v", err)
	}

	// Parse and verify JSON content
	var variables map[string]interface{}
	if err := json.Unmarshal(tfvarsContent, &variables); err != nil {
		t.Fatalf("Invalid JSON in terraform.tfvars.json: %v", err)
	}

	if variables["instance_type"] != "t3.micro" {
		t.Errorf("Expected instance_type 't3.micro', got: %v", variables["instance_type"])
	}
	if variables["vpc_cidr"] != "10.0.0.0/16" {
		t.Errorf("Expected vpc_cidr '10.0.0.0/16', got: %v", variables["vpc_cidr"])
	}
}

func TestScaffold_DryRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "klonekit-scaffold-dryrun-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "source")
	dstDir := filepath.Join(tmpDir, "destination")

	// Create source directory with test files
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(srcDir, "main.tf")
	if err := os.WriteFile(testFile, []byte("resource \"aws_instance\" \"test\" {}"), 0644); err != nil {
		t.Fatal(err)
	}

	spec := &blueprint.Spec{
		Scaffold: blueprint.Scaffold{
			Source:      srcDir,
			Destination: dstDir,
		},
		Variables: map[string]interface{}{
			"instance_type": "t3.micro",
		},
	}

	// Execute dry run
	err = Scaffold(spec, true)
	if err != nil {
		t.Fatalf("Expected no error from dry run, got: %v", err)
	}

	// Verify destination directory was NOT created
	if _, err := os.Stat(dstDir); !os.IsNotExist(err) {
		t.Error("Destination directory should not be created during dry run")
	}

	// Verify no files were actually written
	tfvarsPath := filepath.Join(dstDir, "terraform.tfvars.json")
	if _, err := os.Stat(tfvarsPath); !os.IsNotExist(err) {
		t.Error("terraform.tfvars.json should not be created during dry run")
	}
}

func TestScaffold_SourceNotFound(t *testing.T) {
	spec := &blueprint.Spec{
		Scaffold: blueprint.Scaffold{
			Source:      "/nonexistent/path",
			Destination: "/tmp/test",
		},
	}

	err := Scaffold(spec, false)
	if err == nil {
		t.Fatal("Expected error for non-existent source directory, got nil")
	}

	if !strings.Contains(err.Error(), "source module directory not found") {
		t.Errorf("Expected 'source module directory not found' error, got: %v", err)
	}
}

func TestScaffold_NilSpec(t *testing.T) {
	err := Scaffold(nil, false)
	if err == nil {
		t.Fatal("Expected error for nil spec, got nil")
	}

	if !strings.Contains(err.Error(), "spec cannot be nil") {
		t.Errorf("Expected 'spec cannot be nil' error, got: %v", err)
	}
}

func TestScaffold_NoVariables(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "klonekit-scaffold-novars-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "source")
	dstDir := filepath.Join(tmpDir, "destination")

	// Create source directory with test file
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(srcDir, "main.tf")
	if err := os.WriteFile(testFile, []byte("resource \"aws_instance\" \"test\" {}"), 0644); err != nil {
		t.Fatal(err)
	}

	spec := &blueprint.Spec{
		Scaffold: blueprint.Scaffold{
			Source:      srcDir,
			Destination: dstDir,
		},
		Variables: nil, // No variables
	}

	// Execute scaffold
	err = Scaffold(spec, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify terraform.tfvars.json was NOT created (no variables)
	tfvarsPath := filepath.Join(dstDir, "terraform.tfvars.json")
	if _, err := os.Stat(tfvarsPath); !os.IsNotExist(err) {
		t.Error("terraform.tfvars.json should not be created when no variables are provided")
	}

	// Verify source file was still copied
	copiedFile := filepath.Join(dstDir, "main.tf")
	if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
		t.Error("Source file should still be copied even without variables")
	}
}

func TestScaffold_NestedDirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "klonekit-scaffold-nested-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "source")
	dstDir := filepath.Join(tmpDir, "destination")

	// Create nested directory structure in source
	nestedDir := filepath.Join(srcDir, "modules", "vpc")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create files in nested directory
	nestedFile := filepath.Join(nestedDir, "vpc.tf")
	if err := os.WriteFile(nestedFile, []byte("resource \"aws_vpc\" \"main\" {}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create root level file
	rootFile := filepath.Join(srcDir, "main.tf")
	if err := os.WriteFile(rootFile, []byte("module \"vpc\" { source = \"./modules/vpc\" }"), 0644); err != nil {
		t.Fatal(err)
	}

	spec := &blueprint.Spec{
		Scaffold: blueprint.Scaffold{
			Source:      srcDir,
			Destination: dstDir,
		},
		Variables: map[string]interface{}{
			"vpc_cidr": "10.0.0.0/16",
		},
	}

	// Execute scaffold
	err = Scaffold(spec, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify nested directory structure was preserved
	copiedNestedFile := filepath.Join(dstDir, "modules", "vpc", "vpc.tf")
	if _, err := os.Stat(copiedNestedFile); os.IsNotExist(err) {
		t.Error("Nested file was not copied")
	}

	// Verify root file was copied
	copiedRootFile := filepath.Join(dstDir, "main.tf")
	if _, err := os.Stat(copiedRootFile); os.IsNotExist(err) {
		t.Error("Root file was not copied")
	}
}
