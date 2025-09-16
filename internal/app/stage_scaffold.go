package app

import (
	"context"
	"fmt"
	"log/slog"

	"klonekit/internal/scaffolder"
	"klonekit/pkg/blueprint"
)

// ScaffoldStage implements the Stage interface for the scaffolding stage
type ScaffoldStage struct {
	blueprint *blueprint.Blueprint
	isDryRun  bool
}

// NewScaffoldStage creates a new scaffold stage instance
func NewScaffoldStage(blueprint *blueprint.Blueprint, isDryRun bool) *ScaffoldStage {
	return &ScaffoldStage{
		blueprint: blueprint,
		isDryRun:  isDryRun,
	}
}

// Name returns the name of the stage
func (s *ScaffoldStage) Name() string {
	return "scaffold"
}

// Execute performs the scaffolding stage logic
func (s *ScaffoldStage) Execute(ctx context.Context, state *ExecutionState) error {
	if err := scaffolder.Scaffold(&s.blueprint.Spec, s.isDryRun); err != nil {
		return fmt.Errorf("scaffolding failed: %w", err)
	}

	if s.isDryRun {
		fmt.Printf("%s✅ Scaffolding simulation completed successfully%s\n", ColorGreen, ColorReset)
	} else {
		fmt.Printf("%s✅ Terraform files scaffolded to: %s%s\n", ColorGreen, s.blueprint.Spec.Scaffold.Destination, ColorReset)
	}
	slog.Info("Scaffolding completed successfully", "destination", s.blueprint.Spec.Scaffold.Destination, "dryRun", s.isDryRun)
	return nil
}