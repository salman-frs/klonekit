package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse_ValidBlueprint(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "klonekit-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid blueprint file
	validYaml := `apiVersion: v1
kind: Blueprint
metadata:
  name: test-project
  description: A test project
  labels:
    team: engineering
spec:
  scm:
    provider: gitlab
    url: https://gitlab.example.com
    token: glpat-token123
    project:
      name: my-project
      namespace: my-org
      description: Test project
      visibility: private
  cloud:
    provider: aws
    region: us-east-1
  scaffold:
    source: ./terraform
    destination: ./output
  variables:
    vpc_cidr: "10.0.0.0/16"
    instance_type: t3.micro
`

	filePath := filepath.Join(tmpDir, "valid-blueprint.yaml")
	if err := os.WriteFile(filePath, []byte(validYaml), 0644); err != nil {
		t.Fatal(err)
	}

	// Test parsing
	bp, err := Parse(filePath)
	if err != nil {
		t.Fatalf("Expected successful parsing, got error: %v", err)
	}

	// Verify the parsed content
	if bp.APIVersion != "v1" {
		t.Errorf("Expected APIVersion 'v1', got '%s'", bp.APIVersion)
	}
	if bp.Kind != "Blueprint" {
		t.Errorf("Expected Kind 'Blueprint', got '%s'", bp.Kind)
	}
	if bp.Metadata.Name != "test-project" {
		t.Errorf("Expected Name 'test-project', got '%s'", bp.Metadata.Name)
	}
	if bp.Spec.SCM.Provider != "gitlab" {
		t.Errorf("Expected SCM provider 'gitlab', got '%s'", bp.Spec.SCM.Provider)
	}
	if bp.Spec.Cloud.Provider != "aws" {
		t.Errorf("Expected Cloud provider 'aws', got '%s'", bp.Spec.Cloud.Provider)
	}
}

func TestParse_FileNotFound(t *testing.T) {
	_, err := Parse("nonexistent-file.yaml")
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
	}
	if !strings.Contains(err.Error(), "blueprint file not found") {
		t.Errorf("Expected 'file not found' error, got: %v", err)
	}
}

func TestParse_MalformedYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "klonekit-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a malformed YAML file
	malformedYaml := `apiVersion: v1
kind: Blueprint
metadata:
  name: test
  description: "unclosed quote
spec:
  invalid yaml structure
`

	filePath := filepath.Join(tmpDir, "malformed.yaml")
	if err := os.WriteFile(filePath, []byte(malformedYaml), 0644); err != nil {
		t.Fatal(err)
	}

	_, err = Parse(filePath)
	if err == nil {
		t.Fatal("Expected error for malformed YAML, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read blueprint file") {
		t.Errorf("Expected 'failed to read blueprint file' error, got: %v", err)
	}
}

func TestParse_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name          string
		yaml          string
		expectedError string
	}{
		{
			name: "missing apiVersion",
			yaml: `kind: Blueprint
metadata:
  name: test
spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: token
    project:
      name: test
      namespace: test
  cloud:
    provider: aws
    region: us-east-1
  scaffold:
    source: ./src
    destination: ./dst
`,
			expectedError: "field 'APIVersion' is required but missing",
		},
		{
			name: "wrong kind value",
			yaml: `apiVersion: v1
kind: WrongKind
metadata:
  name: test
spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: token
    project:
      name: test
      namespace: test
  cloud:
    provider: aws
    region: us-east-1
  scaffold:
    source: ./src
    destination: ./dst
`,
			expectedError: "field 'Kind' must be 'Blueprint'",
		},
		{
			name: "missing metadata name",
			yaml: `apiVersion: v1
kind: Blueprint
metadata:
  description: test
spec:
  scm:
    provider: gitlab
    url: https://gitlab.com
    token: token
    project:
      name: test
      namespace: test
  cloud:
    provider: aws
    region: us-east-1
  scaffold:
    source: ./src
    destination: ./dst
`,
			expectedError: "field 'Name' is required but missing",
		},
		{
			name: "missing scm provider",
			yaml: `apiVersion: v1
kind: Blueprint
metadata:
  name: test
spec:
  scm:
    url: https://gitlab.com
    token: token
    project:
      name: test
      namespace: test
  cloud:
    provider: aws
    region: us-east-1
  scaffold:
    source: ./src
    destination: ./dst
`,
			expectedError: "field 'Provider' is required but missing",
		},
		{
			name: "invalid scm provider",
			yaml: `apiVersion: v1
kind: Blueprint
metadata:
  name: test
spec:
  scm:
    provider: github
    url: https://gitlab.com
    token: token
    project:
      name: test
      namespace: test
  cloud:
    provider: aws
    region: us-east-1
  scaffold:
    source: ./src
    destination: ./dst
`,
			expectedError: "field 'Provider' must be one of: gitlab",
		},
		{
			name: "invalid URL",
			yaml: `apiVersion: v1
kind: Blueprint
metadata:
  name: test
spec:
  scm:
    provider: gitlab
    url: not-a-url
    token: token
    project:
      name: test
      namespace: test
  cloud:
    provider: aws
    region: us-east-1
  scaffold:
    source: ./src
    destination: ./dst
`,
			expectedError: "field 'URL' must be a valid URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "klonekit-test-")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			filePath := filepath.Join(tmpDir, "test.yaml")
			if err := os.WriteFile(filePath, []byte(tt.yaml), 0644); err != nil {
				t.Fatal(err)
			}

			_, err = Parse(filePath)
			if err == nil {
				t.Fatal("Expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
			}
		})
	}
}
