package testing

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/runtime"
)

// MockExecutor implements runtime.ToolExecutor with configurable results per tool+input.
type MockExecutor struct {
	responses map[string]*runtime.ToolResult
	mu        sync.RWMutex
}

// MockResponseKey returns a key for tool+version+input for MockExecutor.
func MockResponseKey(tool, version string, input json.RawMessage) string {
	return tool + "@" + version + ":" + string(input)
}

// NewMockExecutor creates a mock executor.
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		responses: make(map[string]*runtime.ToolResult),
	}
}

// SetResponse sets the result for a given tool, version, and input.
func (m *MockExecutor) SetResponse(tool, version string, input json.RawMessage, result *runtime.ToolResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := MockResponseKey(tool, version, input)
	m.responses[key] = result
}

// SetResponseForAnyInput sets the result for any input to this tool+version.
func (m *MockExecutor) SetResponseForAnyInput(tool, version string, result *runtime.ToolResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := tool + "@" + version + ":*"
	m.responses[key] = result
}

// Execute implements runtime.ToolExecutor.
func (m *MockExecutor) Execute(ctx context.Context, toolName, version string, input json.RawMessage) (*runtime.ToolResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := MockResponseKey(toolName, version, input)
	if r, ok := m.responses[key]; ok {
		return r, nil
	}
	keyAny := toolName + "@" + version + ":*"
	if r, ok := m.responses[keyAny]; ok {
		return r, nil
	}
	return &runtime.ToolResult{
		ToolID:   toolName + "@" + version,
		Error:    "mock: no response configured",
		Duration: time.Millisecond,
	}, nil
}
