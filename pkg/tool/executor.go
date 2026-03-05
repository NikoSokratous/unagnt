package tool

import (
	"context"
	"encoding/json"

	"github.com/NikoSokratous/unagnt/pkg/runtime"
)

// Executor wraps a Registry to implement runtime.ToolExecutor.
type Executor struct {
	Registry *Registry
}

// NewExecutor creates an executor from a registry.
func NewExecutor(r *Registry) *Executor {
	return &Executor{Registry: r}
}

// Execute implements runtime.ToolExecutor.
func (e *Executor) Execute(ctx context.Context, tool, version string, input json.RawMessage) (*runtime.ToolResult, error) {
	return e.Registry.Execute(ctx, tool, version, input)
}
