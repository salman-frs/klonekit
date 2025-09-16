package app

import (
	"context"
	"fmt"
	"log/slog"

	"klonekit/pkg/blueprint"
)

// ScmStage implements the Stage interface for the source control management stage
type ScmStage struct {
	blueprint       *blueprint.Blueprint
	providerFactory *ProviderFactory
	isDryRun        bool
}

// NewScmStage creates a new SCM stage instance
func NewScmStage(blueprint *blueprint.Blueprint, providerFactory *ProviderFactory, isDryRun bool) *ScmStage {
	return &ScmStage{
		blueprint:       blueprint,
		providerFactory: providerFactory,
		isDryRun:        isDryRun,
	}
}

// Name returns the name of the stage
func (s *ScmStage) Name() string {
	return "scm"
}

// Execute performs the SCM stage logic
func (s *ScmStage) Execute(ctx context.Context, state *ExecutionState) error {
	if s.isDryRun {
		fmt.Printf("%s🔍 DRY RUN: Would create %s repository '%s' in namespace '%s'%s\n",
			ColorYellow, s.blueprint.Spec.SCM.Provider, s.blueprint.Spec.SCM.Project.Name, s.blueprint.Spec.SCM.Project.Namespace, ColorReset)
		fmt.Printf("%s🔍 DRY RUN: Would push scaffolded files to repository%s\n", ColorYellow, ColorReset)
	} else {
		provider, err := s.providerFactory.GetScmProvider(s.blueprint.Spec.SCM.Provider)
		if err != nil {
			return fmt.Errorf("SCM provider initialization failed: %w", err)
		}

		if err := provider.CreateRepo(&s.blueprint.Spec); err != nil {
			return fmt.Errorf("%s repository creation failed: %w", s.blueprint.Spec.SCM.Provider, err)
		}
	}

	if s.isDryRun {
		fmt.Printf("%s✅ SCM simulation completed successfully%s\n", ColorGreen, ColorReset)
	} else {
		fmt.Printf("%s✅ %s repository created: %s%s\n", ColorGreen, s.blueprint.Spec.SCM.Provider, s.blueprint.Spec.SCM.Project.Name, ColorReset)
	}
	slog.Info("SCM stage completed successfully", "provider", s.blueprint.Spec.SCM.Provider, "repoName", s.blueprint.Spec.SCM.Project.Name, "dryRun", s.isDryRun)
	return nil
}