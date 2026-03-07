package testing

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/runtime"
)

func TestMockExecutor_Execute(t *testing.T) {
	ctx := context.Background()
	exec := NewMockExecutor()
	input := json.RawMessage(`{"x":1}`)
	exec.SetResponse("tool", "1", input, &runtime.ToolResult{
		ToolID:   "tool@1",
		Output:   map[string]any{"out": "ok"},
		Duration: 10 * time.Millisecond,
	})

	res, err := exec.Execute(ctx, "tool", "1", input)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.Output["out"] != "ok" {
		t.Errorf("got output %v", res.Output)
	}
}

func TestMockExecutor_ExecuteForAnyInput(t *testing.T) {
	ctx := context.Background()
	exec := NewMockExecutor()
	exec.SetResponseForAnyInput("echo", "1", &runtime.ToolResult{
		ToolID:   "echo@1",
		Output:   map[string]any{"echoed": "any"},
		Duration: time.Millisecond,
	})

	res, _ := exec.Execute(ctx, "echo", "1", json.RawMessage(`{}`))
	if res.Output["echoed"] != "any" {
		t.Errorf("got %v", res.Output)
	}
}
