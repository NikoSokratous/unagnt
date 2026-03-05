package tool

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/runtime"
)

// Tool is the interface for executable tools.
type Tool interface {
	Name() string
	Version() string
	Description() string
	InputSchema() ([]byte, error)
	Permissions() []Permission
	Execute(ctx context.Context, input json.RawMessage) (output map[string]any, err error)
}

// Registry holds registered tools and executes them.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool // key: "name@version"
}

// NewRegistry creates a new tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry.
func (r *Registry) Register(t Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := t.Name() + "@" + t.Version()
	r.tools[key] = t
}

// Get returns a tool by name and version.
func (r *Registry) Get(name, version string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := name + "@" + version
	t, ok := r.tools[key]
	return t, ok
}

// List returns all registered tool names and versions.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []string
	for k := range r.tools {
		out = append(out, k)
	}
	return out
}

// Execute implements runtime.ToolExecutor.
func (r *Registry) Execute(ctx context.Context, tool, version string, input json.RawMessage) (*runtime.ToolResult, error) {
	t, ok := r.Get(tool, version)
	if !ok {
		return &runtime.ToolResult{
			Error: "tool not found: " + tool + "@" + version,
		}, nil
	}

	start := time.Now()
	output, err := t.Execute(ctx, input)
	duration := time.Since(start)

	if err != nil {
		return &runtime.ToolResult{
			ToolID:   tool + "@" + version,
			Error:    err.Error(),
			Duration: duration,
		}, err
	}

	if output == nil {
		output = make(map[string]any)
	}

	return &runtime.ToolResult{
		ToolID:   tool + "@" + version,
		Output:   output,
		Duration: duration,
	}, nil
}
