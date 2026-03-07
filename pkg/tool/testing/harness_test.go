package testing

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/NikoSokratous/unagnt/pkg/tool/builtin"
)

func TestToolHarness_ExecuteWithMap(t *testing.T) {
	ctx := context.Background()
	echo := &builtin.Echo{}
	harness := NewToolHarness(echo)

	out, err := harness.ExecuteWithMap(ctx, map[string]any{"message": "hello"})
	AssertNoError(t, err)
	AssertOutputContains(t, out, "echoed", "hello")
}

func TestToolHarness_ExecuteWithInput(t *testing.T) {
	ctx := context.Background()
	echo := &builtin.Echo{}
	harness := NewToolHarness(echo)

	raw := json.RawMessage(`{"message":"world"}`)
	out, err := harness.ExecuteWithInput(ctx, raw)
	AssertNoError(t, err)
	if _, ok := out["echoed"]; !ok {
		t.Error("expected echoed key in output")
	}
}
