package app

import (
	"context"
)

// Stage represents a single stage in the KloneKit apply workflow.
// Each stage implements this interface to provide a name and execution logic.
type Stage interface {
	Name() string
	Execute(ctx context.Context, state *ExecutionState) error
}