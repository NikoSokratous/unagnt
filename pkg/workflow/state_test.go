package workflow

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestStateStoreNodeStateIDsAreWorkflowScoped(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(context.Background(), testWorkflowStateSchema); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	store := NewStateStore(db)
	now := time.Now()

	// Same node ID in two different workflows should not conflict.
	if err := store.SaveNodeState(context.Background(), &NodeState{
		NodeID:      "analyze",
		WorkflowID:  "wf-a",
		StepName:    "Analyze",
		Status:      "completed",
		Output:      map[string]any{"wf": "a"},
		StartedAt:   now,
		CompletedAt: now,
	}); err != nil {
		t.Fatalf("save node state wf-a: %v", err)
	}

	if err := store.SaveNodeState(context.Background(), &NodeState{
		NodeID:      "analyze",
		WorkflowID:  "wf-b",
		StepName:    "Analyze",
		Status:      "completed",
		Output:      map[string]any{"wf": "b"},
		StartedAt:   now,
		CompletedAt: now,
	}); err != nil {
		t.Fatalf("save node state wf-b: %v", err)
	}

	aStates, err := store.LoadNodeStates(context.Background(), "wf-a")
	if err != nil {
		t.Fatalf("load wf-a node states: %v", err)
	}
	bStates, err := store.LoadNodeStates(context.Background(), "wf-b")
	if err != nil {
		t.Fatalf("load wf-b node states: %v", err)
	}

	if len(aStates) != 1 || len(bStates) != 1 {
		t.Fatalf("expected 1 state each, got wf-a=%d wf-b=%d", len(aStates), len(bStates))
	}
	if aStates[0].NodeID != "analyze" || bStates[0].NodeID != "analyze" {
		t.Fatalf("unexpected node IDs: wf-a=%q wf-b=%q", aStates[0].NodeID, bStates[0].NodeID)
	}
}
