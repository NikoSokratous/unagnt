package testing

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/NikoSokratous/unagnt/pkg/tool"
)

// MockTool is a tool implementation with configurable behavior.
type MockTool struct {
	NameStr        string
	VersionStr     string
	DescStr        string
	Output         map[string]any
	Err            error
	InputValidator func(json.RawMessage) error
	mu             sync.Mutex
}

// NewMockTool creates a mock tool with default behavior.
func NewMockTool(name, version string) *MockTool {
	return &MockTool{
		NameStr:    name,
		VersionStr: version,
		DescStr:    "mock tool for testing",
		Output:     map[string]any{"ok": true},
	}
}

// SetOutput sets the output returned by Execute.
func (m *MockTool) SetOutput(out map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Output = out
}

// SetError sets the error returned by Execute.
func (m *MockTool) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Err = err
}

func (m *MockTool) Name() string      { return m.NameStr }
func (m *MockTool) Version() string   { return m.VersionStr }
func (m *MockTool) Description() string { return m.DescStr }

func (m *MockTool) InputSchema() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"input": map[string]string{"type": "string"},
		},
	})
}

func (m *MockTool) Permissions() []tool.Permission { return nil }

func (m *MockTool) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
	m.mu.Lock()
	out := m.Output
	err := m.Err
	validator := m.InputValidator
	m.mu.Unlock()

	if validator != nil {
		if vErr := validator(input); vErr != nil {
			return nil, vErr
		}
	}
	if err != nil {
		return nil, err
	}
	if out == nil {
		return map[string]any{}, nil
	}
	result := make(map[string]any, len(out))
	for k, v := range out {
		result[k] = v
	}
	return result, nil
}
