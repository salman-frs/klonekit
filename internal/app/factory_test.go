package app

import (
	"strings"
	"testing"

	"klonekit/internal/scm"
)

func TestProviderFactory_GetScmProvider(t *testing.T) {
	factory := NewProviderFactory()

	tests := []struct {
		name         string
		providerName string
		expectError  bool
		errorMsg     string
		expectType   string
	}{
		{
			name:         "Valid GitLab provider",
			providerName: "gitlab",
			expectError:  true, // Expected in test environment due to missing GITLAB_PRIVATE_TOKEN
			errorMsg:     "failed to create GitLab provider",
		},
		{
			name:         "Unsupported provider",
			providerName: "github",
			expectError:  true,
			errorMsg:     "unsupported SCM provider: github",
		},
		{
			name:         "Empty provider name",
			providerName: "",
			expectError:  true,
			errorMsg:     "unsupported SCM provider:",
		},
		{
			name:         "Invalid provider name",
			providerName: "invalid-provider",
			expectError:  true,
			errorMsg:     "unsupported SCM provider: invalid-provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := factory.GetScmProvider(tt.providerName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.errorMsg, err.Error())
				}
				if provider != nil {
					t.Errorf("Expected provider to be nil on error, got: %T", provider)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}

			if provider == nil {
				t.Error("Expected provider to be non-nil")
				return
			}

			// Provider is already of type scm.ScmProvider (returned from GetScmProvider)
			// Verify it's the expected concrete implementation type

			// For GitLab, we can check the specific type
			if tt.providerName == "gitlab" {
				if _, ok := provider.(*scm.GitLabProvider); !ok {
					t.Errorf("Expected *scm.GitLabProvider, got: %T", provider)
				}
			}
		})
	}
}

func TestProviderFactory_GetProvisioner(t *testing.T) {
	factory := NewProviderFactory()

	tests := []struct {
		name         string
		providerName string
		expectError  bool
		errorMsg     string
		skipReason   string
	}{
		{
			name:         "Valid AWS provider",
			providerName: "aws",
			expectError:  false,
		},
		{
			name:         "Unsupported provider",
			providerName: "azure",
			expectError:  true,
			errorMsg:     "unsupported provisioner: azure",
		},
		{
			name:         "Empty provider name",
			providerName: "",
			expectError:  true,
			errorMsg:     "unsupported provisioner:",
		},
		{
			name:         "Invalid provider name",
			providerName: "invalid-provider",
			expectError:  true,
			errorMsg:     "unsupported provisioner: invalid-provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provisioner, err := factory.GetProvisioner(tt.providerName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.errorMsg, err.Error())
				}
				if provisioner != nil {
					t.Errorf("Expected provisioner to be nil on error, got: %T", provisioner)
				}
				return
			}

			// For AWS provisioner, Docker runtime is required
			if tt.providerName == "aws" {
				// Docker may not be available in test environments
				if err != nil && strings.Contains(err.Error(), "failed to create Docker runtime") {
					t.Skipf("Skipping test: Docker not available in test environment: %v", err)
					return
				}
			}

			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}

			if provisioner == nil {
				t.Error("Expected provisioner to be non-nil")
				return
			}

			// We can't easily test the interface implementation without importing the provisioner package
			// But we can verify the provisioner is not nil and has expected behavior
			// The integration tests will verify the actual functionality
		})
	}
}

func TestNewProviderFactory(t *testing.T) {
	factory := NewProviderFactory()

	if factory == nil {
		t.Error("Expected factory to be non-nil")
		return
	}

	// Verify factory can create providers
	scmProvider, err := factory.GetScmProvider("gitlab")
	if err != nil && !strings.Contains(err.Error(), "GITLAB_PRIVATE_TOKEN") {
		t.Errorf("Unexpected error from factory: %s", err)
	}

	// We expect GitLab provider creation to fail with token error in test environment
	if err == nil && scmProvider == nil {
		t.Error("Expected either provider or error, got neither")
	}
}

// TestProviderFactory_Integration tests that the factory works with actual app orchestrator
func TestProviderFactory_Integration(t *testing.T) {
	// This test verifies the factory integrates correctly with the app package
	// It creates a factory and ensures the providers it returns are compatible

	factory := NewProviderFactory()

	// Test that all supported providers can be created (even if they fail due to missing credentials)
	supportedScmProviders := []string{"gitlab"}
	for _, provider := range supportedScmProviders {
		_, err := factory.GetScmProvider(provider)
		// We expect GitLab to fail with authentication error in test environment
		if err != nil && !strings.Contains(err.Error(), "GITLAB_PRIVATE_TOKEN") {
			t.Errorf("Unexpected error for SCM provider %s: %s", provider, err)
		}
	}

	supportedProvisioners := []string{"aws"}
	for _, provider := range supportedProvisioners {
		provisioner, err := factory.GetProvisioner(provider)
		// Docker may not be available in test environments
		if err != nil && strings.Contains(err.Error(), "failed to create Docker runtime") {
			t.Skipf("Skipping test: Docker not available for provider %s: %v", provider, err)
			continue
		}
		if err != nil {
			t.Errorf("Unexpected error for provisioner %s: %s", provider, err)
			continue
		}
		if provisioner == nil {
			t.Errorf("Expected provisioner for %s to be non-nil", provider)
		}
	}
}