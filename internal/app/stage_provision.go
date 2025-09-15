package app

import (
	"context"
	"fmt"
	"log/slog"

	"klonekit/pkg/blueprint"
)

// ProvisionStage implements the Stage interface for the infrastructure provisioning stage
type ProvisionStage struct {
	blueprint       *blueprint.Blueprint
	providerFactory *ProviderFactory
	isDryRun        bool
	autoApprove     bool
}

// NewProvisionStage creates a new provision stage instance
func NewProvisionStage(blueprint *blueprint.Blueprint, providerFactory *ProviderFactory, isDryRun bool, autoApprove bool) *ProvisionStage {
	return &ProvisionStage{
		blueprint:       blueprint,
		providerFactory: providerFactory,
		isDryRun:        isDryRun,
		autoApprove:     autoApprove,
	}
}

// Name returns the name of the stage
func (s *ProvisionStage) Name() string {
	return "provision"
}

// Execute performs the provisioning stage logic
func (s *ProvisionStage) Execute(ctx context.Context, state *ExecutionState) error {
	if s.isDryRun {
		fmt.Printf("%s🔍 DRY RUN: Would pull Terraform Docker image%s\n", ColorYellow, ColorReset)
		fmt.Printf("%s🔍 DRY RUN: Would execute 'terraform init' in container%s\n", ColorYellow, ColorReset)
		fmt.Printf("%s🔍 DRY RUN: Would execute 'terraform plan' in container%s\n", ColorYellow, ColorReset)
		if s.autoApprove {
			fmt.Printf("%s🔍 DRY RUN: Would execute 'terraform apply -auto-approve' in container%s\n", ColorYellow, ColorReset)
			fmt.Printf("%s🔍 DRY RUN: Would provision infrastructure using %s provider in %s region%s\n",
				ColorYellow, s.blueprint.Spec.Cloud.Provider, s.blueprint.Spec.Cloud.Region, ColorReset)
		} else {
			fmt.Printf("%s🔍 DRY RUN: Would validate infrastructure (no apply without --auto-approve)%s\n", ColorYellow, ColorReset)
		}
	} else {
		provisioner, err := s.providerFactory.GetProvisioner(s.blueprint.Spec.Cloud.Provider)
		if err != nil {
			return fmt.Errorf("provisioner initialization failed: %w", err)
		}

		if err := provisioner.Provision(&s.blueprint.Spec, s.autoApprove); err != nil {
			return fmt.Errorf("infrastructure provisioning failed: %w", err)
		}
	}

	if s.isDryRun {
		fmt.Printf("%s✅ Provisioning simulation completed successfully%s\n", ColorGreen, ColorReset)
	} else if s.autoApprove {
		fmt.Printf("%s✅ Infrastructure provisioned successfully using %s provider in %s%s\n", ColorGreen, s.blueprint.Spec.Cloud.Provider, s.blueprint.Spec.Cloud.Region, ColorReset)
	} else {
		fmt.Printf("%s✅ Infrastructure validated successfully (use --auto-approve to provision)%s\n", ColorGreen, ColorReset)
	}
	slog.Info("Provisioning stage completed successfully", "provider", s.blueprint.Spec.Cloud.Provider, "region", s.blueprint.Spec.Cloud.Region, "dryRun", s.isDryRun)
	return nil
}