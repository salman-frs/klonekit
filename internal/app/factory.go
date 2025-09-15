package app

import (
	"fmt"

	"klonekit/internal/provisioner"
	"klonekit/internal/runtime"
	"klonekit/internal/scm"
)

// ProviderFactory provides methods to create SCM and provisioning providers
// based on string identifiers. This implements the Factory pattern to decouple
// the application orchestrator from concrete provider implementations.
type ProviderFactory struct{}

// NewProviderFactory creates a new instance of ProviderFactory.
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{}
}

// GetScmProvider returns the appropriate SCM provider implementation
// based on the provider name from the blueprint configuration.
func (f *ProviderFactory) GetScmProvider(providerName string) (scm.ScmProvider, error) {
	switch providerName {
	case "gitlab":
		provider, err := scm.NewGitLabProvider()
		if err != nil {
			return nil, fmt.Errorf("failed to create GitLab provider: %w", err)
		}
		return provider, nil
	default:
		return nil, fmt.Errorf("unsupported SCM provider: %s", providerName)
	}
}

// GetProvisioner returns the appropriate provisioner implementation
// based on the provider name from the blueprint configuration.
func (f *ProviderFactory) GetProvisioner(providerName string) (provisioner.Provisioner, error) {
	switch providerName {
	case "aws":
		// Create Docker runtime instance for Terraform
		dockerRuntime, err := runtime.NewDockerRuntime()
		if err != nil {
			return nil, fmt.Errorf("failed to create Docker runtime: %w", err)
		}
		return provisioner.NewTerraformDockerProvisioner(dockerRuntime), nil
	default:
		return nil, fmt.Errorf("unsupported provisioner: %s", providerName)
	}
}