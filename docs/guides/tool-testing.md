# Tool Testing Guide

How to test tools in isolation using the tool testing harness and mocks.

## Overview

The `pkg/tool/testing` package provides:

- **ToolHarness** – Wraps a tool with assertion helpers for unit tests
- **MockTool** – Configurable tool implementation for testing
- **MockExecutor** – Configurable `runtime.ToolExecutor` for integration tests

## ToolHarness

Use `ToolHarness` to test a tool directly without a full runtime:

```go
import (
    "context"
    "testing"
    "github.com/NikoSokratous/unagnt/pkg/tool/builtin"
    "github.com/NikoSokratous/unagnt/pkg/tool/testing"
)

func TestEchoTool(t *testing.T) {
    ctx := context.Background()
    harness := testing.NewToolHarness(&builtin.Echo{})

    out, err := harness.ExecuteWithMap(ctx, map[string]any{"message": "hello"})
    testing.AssertNoError(t, err)
    testing.AssertOutputContains(t, out, "echoed", "hello")
}
```

### Helpers

| Helper | Description |
|--------|-------------|
| `AssertNoError(t, err)` | Fails if err is non-nil |
| `AssertOutputContains(t, output, key, substr)` | Fails if output[key] (string) doesn't contain substr |
| `AssertOutputEqual(t, output, key, expected)` | Fails if output[key] != expected |
| `harness.AssertSchemaValid(t, input)` | Validates input against the tool's JSON schema |

## MockTool

`MockTool` implements `tool.Tool` with configurable output and error:

```go
mock := testing.NewMockTool("my_tool", "1")
mock.SetOutput(map[string]any{"result": "ok"})
mock.SetError(nil)  // or errors.New("fail")

out, err := mock.Execute(ctx, json.RawMessage(`{}`))
```

Use `SetOutput` and `SetError` to simulate different behaviors in tests.

## MockExecutor

`MockExecutor` implements `runtime.ToolExecutor` for integration tests that need to mock tool execution:

```go
exec := testing.NewMockExecutor()
exec.SetResponse("tool", "1", json.RawMessage(`{"x":1}`), &runtime.ToolResult{
    ToolID:   "tool@1",
    Output:   map[string]any{"out": "ok"},
    Duration: time.Millisecond,
})

// Or for any input to a tool:
exec.SetResponseForAnyInput("echo", "1", &runtime.ToolResult{
    ToolID:   "echo@1",
    Output:   map[string]any{"echoed": "anything"},
    Duration: time.Millisecond,
})

// Use with PolicyExecutor, Engine, etc.
policyExec := &tool.PolicyExecutor{Inner: exec}
res, _ := policyExec.Execute(ctx, "tool", "1", input)
```

## Full Example

```go
func TestMyToolWithHarness(t *testing.T) {
    ctx := context.Background()
    reg := tool.NewRegistry()
    reg.Register(&builtin.Echo{})
    reg.Register(&builtin.Calc{})

    calc, _ := reg.Get("calc", "1")
    harness := testing.NewToolHarness(calc)

    out, err := harness.ExecuteWithMap(ctx, map[string]any{
        "op": "add", "a": 5.0, "b": 3.0,
    })
    testing.AssertNoError(t, err)
    if out["result"].(float64) != 8 {
        t.Errorf("expected 8, got %v", out["result"])
    }
}
```

## See Also

- [Tool Development Guide](tool-development.md) – Creating custom tools
- [Policy Testing](policy-testing.md) – Testing policies with tools
