package testing_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/runtime"
	"github.com/NikoSokratous/unagnt/pkg/tool"
	"github.com/NikoSokratous/unagnt/pkg/tool/builtin"
	testingpkg "github.com/NikoSokratous/unagnt/pkg/tool/testing"
)

func TestExampleBuiltinWithHarness(t *testing.T) {
	ctx := context.Background()
	reg := tool.NewRegistry()
	reg.Register(&builtin.Echo{})
	reg.Register(&builtin.Calc{})

	// Test echo via harness
	echo, _ := reg.Get("echo", "1")
	harness := testingpkg.NewToolHarness(echo)
	out, err := harness.ExecuteWithMap(ctx, map[string]any{"message": "hello"})
	testingpkg.AssertNoError(t, err)
	testingpkg.AssertOutputContains(t, out, "echoed", "hello")

	// Test calc via harness
	calc, _ := reg.Get("calc", "1")
	harness2 := testingpkg.NewToolHarness(calc)
	out2, err := harness2.ExecuteWithMap(ctx, map[string]any{"op": "add", "a": 5.0, "b": 3.0})
	testingpkg.AssertNoError(t, err)
	if out2["result"].(float64) != 8 {
		t.Errorf("expected 8, got %v", out2["result"])
	}
}

func TestExampleMockExecutorWithPolicyWrapper(t *testing.T) {
	// MockExecutor wrapped by PolicyExecutor (nil policy = pass-through)
	ctx := context.Background()
	mock := testingpkg.NewMockExecutor()
	input := json.RawMessage(`{"x":1}`)
	mock.SetResponse("my_tool", "1", input, &runtime.ToolResult{
		ToolID:   "my_tool@1",
		Output:   map[string]any{"done": true},
		Duration: 5 * time.Millisecond,
	})

	// Wrap with PolicyExecutor - nil policy means no checks, just delegate
	policyExec := &tool.PolicyExecutor{
		Inner: mock,
		// Policy, RiskScorer, Approval nil = pass through
	}

	res, err := policyExec.Execute(ctx, "my_tool", "1", input)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !res.Output["done"].(bool) {
		t.Error("expected done=true")
	}
}
