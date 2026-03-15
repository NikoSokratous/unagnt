package workflow

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

const testWorkflowStateSchema = `
CREATE TABLE IF NOT EXISTS workflow_states (
    id TEXT PRIMARY KEY,
    workflow_name TEXT NOT NULL,
    status TEXT NOT NULL,
    current_step TEXT,
    outputs TEXT,
    started_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL
);
CREATE TABLE IF NOT EXISTS workflow_step_states (
    id TEXT PRIMARY KEY,
    workflow_id TEXT NOT NULL,
    step_name TEXT NOT NULL,
    status TEXT NOT NULL,
    output TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);
`

func TestExecutorResumeFromCheckpoint(t *testing.T) {
	db, err := sql.Open("sqlite", t.TempDir()+"/wfstate.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(context.Background(), testWorkflowStateSchema); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	store := NewStateStore(db)
	exec := NewExecutor(store, nil)
	dag := NewDAG()
	if err := dag.AddNode(&Node{ID: "a", Name: "a", Agent: "agent-a", Goal: "goal-a", OutputKey: "oa"}); err != nil {
		t.Fatalf("add node a: %v", err)
	}
	if err := dag.AddNode(&Node{ID: "b", Name: "b", Agent: "agent-b", Goal: "goal-b", Dependencies: []string{"a"}, OutputKey: "ob"}); err != nil {
		t.Fatalf("add node b: %v", err)
	}
	if err := dag.AddEdge("a", "b"); err != nil {
		t.Fatalf("add edge: %v", err)
	}

	workflowID := "wf-resume-test"
	if _, err := exec.Execute(context.Background(), dag, workflowID); err != nil {
		t.Fatalf("execute: %v", err)
	}

	resumed, err := exec.Resume(context.Background(), workflowID, dag)
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if resumed.Status != "completed" {
		t.Fatalf("expected completed, got %s", resumed.Status)
	}
}
