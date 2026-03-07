package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/NikoSokratous/unagnt/pkg/tool"
)

// ToolHarness wraps a Tool for testing with assertion helpers.
type ToolHarness struct {
	Tool   tool.Tool
	Schema *tool.InputSchema
}

// NewToolHarness creates a harness for a tool.
func NewToolHarness(t tool.Tool) *ToolHarness {
	raw, _ := t.InputSchema()
	sch, _ := tool.NewInputSchema(raw)
	return &ToolHarness{Tool: t, Schema: sch}
}

// ExecuteWithInput runs the tool with the given input and returns output and error.
func (h *ToolHarness) ExecuteWithInput(ctx context.Context, input json.RawMessage) (map[string]any, error) {
	return h.Tool.Execute(ctx, input)
}

// ExecuteWithMap marshals the input map to JSON and runs the tool.
func (h *ToolHarness) ExecuteWithMap(ctx context.Context, input map[string]any) (map[string]any, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	return h.Tool.Execute(ctx, raw)
}

// AssertNoError fails the test if err is non-nil.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AssertOutputContains fails if output does not contain the expected key, or if the
// string representation of its value does not contain expectedSubstr.
func AssertOutputContains(t *testing.T, output map[string]any, key, expectedSubstr string) {
	t.Helper()
	v, ok := output[key]
	if !ok {
		t.Fatalf("output missing key %q", key)
	}
	s := fmt.Sprint(v)
	if !strings.Contains(s, expectedSubstr) {
		t.Fatalf("output[%q] = %q does not contain %q", key, s, expectedSubstr)
	}
}

// AssertOutputEqual fails if output[key] != expected (using fmt.Sprintf for comparison).
func AssertOutputEqual(t *testing.T, output map[string]any, key string, expected any) {
	t.Helper()
	v, ok := output[key]
	if !ok {
		t.Fatalf("output missing key %q", key)
	}
	if fmt.Sprint(v) != fmt.Sprint(expected) {
		t.Fatalf("output[%q] = %v, want %v", key, v, expected)
	}
}

// AssertSchemaValid validates input against the tool's schema. Fails test if invalid.
func (h *ToolHarness) AssertSchemaValid(t *testing.T, input json.RawMessage) {
	t.Helper()
	if h.Schema == nil {
		return
	}
	if err := h.Schema.Validate(input); err != nil {
		t.Fatalf("schema validation failed: %v", err)
	}
}
