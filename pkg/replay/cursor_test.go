package replay

import (
	"encoding/json"
	"testing"
	"time"
)

func TestReplayCursor_StepForwardBack(t *testing.T) {
	snap := &RunSnapshot{
		ID:      "s1",
		ToolCalls: []ToolExecution{
			{Sequence: 1, ToolName: "echo", Input: json.RawMessage(`{}`)},
			{Sequence: 2, ToolName: "calc", Input: json.RawMessage(`{"op":"add"}`)},
		},
	}

	c := NewReplayCursor(snap)
	if c.Position() != 0 {
		t.Errorf("initial position want 0, got %d", c.Position())
	}
	if !c.CanStepForward() {
		t.Error("expected CanStepForward true at start")
	}
	if c.CanStepBack() {
		t.Error("expected CanStepBack false at start")
	}

	if !c.StepForward() {
		t.Fatal("StepForward should succeed")
	}
	if c.Position() != 1 {
		t.Errorf("position after step want 1, got %d", c.Position())
	}

	c.StepForward()
	if c.Position() != 2 {
		t.Errorf("position want 2, got %d", c.Position())
	}
	if c.CanStepForward() {
		t.Error("at end, CanStepForward should be false")
	}

	c.StepBack()
	if c.Position() != 1 {
		t.Errorf("after StepBack want 1, got %d", c.Position())
	}
}

func TestReplayCursor_GetStateAt(t *testing.T) {
	snap := &RunSnapshot{
		ID:        "s1",
		StartTime: time.Now(),
		EndTime:   time.Now(),
		ToolCalls: []ToolExecution{
			{Sequence: 1, ToolName: "echo", Input: json.RawMessage(`{"x":1}`)},
		},
	}

	c := NewReplayCursor(snap)
	st := c.GetStateAt(1)
	if st.Position != 1 {
		t.Errorf("Position want 1, got %d", st.Position)
	}
	if st.CurrentAction == nil {
		t.Fatal("CurrentAction should be set at seq 1")
	}
	if st.CurrentAction.ToolName != "echo" {
		t.Errorf("ToolName want echo, got %s", st.CurrentAction.ToolName)
	}
	if !st.CanStepBack {
		t.Error("at seq 1, CanStepBack should be true")
	}
}
