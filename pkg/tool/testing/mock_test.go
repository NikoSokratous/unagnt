package testing

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestMockTool_Execute(t *testing.T) {
	ctx := context.Background()
	m := NewMockTool("mock", "1")
	m.SetOutput(map[string]any{"result": "ok"})

	out, err := m.Execute(ctx, json.RawMessage(`{}`))
	AssertNoError(t, err)
	AssertOutputEqual(t, out, "result", "ok")
}

func TestMockTool_ExecuteWithError(t *testing.T) {
	ctx := context.Background()
	m := NewMockTool("mock", "1")
	m.SetError(errors.New("mock error"))

	_, err := m.Execute(ctx, json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "mock error" {
		t.Errorf("got %v", err)
	}
}
