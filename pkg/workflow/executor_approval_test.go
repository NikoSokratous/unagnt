package workflow

import (
	"context"
	"database/sql"
	"testing"

	"github.com/NikoSokratous/unagnt/pkg/policy"
	_ "modernc.org/sqlite"
)

func TestExecutorApprovalStepWithAutoApprove(t *testing.T) {
	db, err := sql.Open("sqlite", t.TempDir()+"/wfstate.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(context.Background(), testWorkflowStateSchema); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	store := NewStateStore(db)
	exec := NewExecutorWithApproval(store, nil, policy.NewAutoApprovalQueue())
	dag := NewDAG()

	if err := dag.AddNode(&Node{ID: "a", Name: "a", Agent: "agent-a", Goal: "goal-a", OutputKey: "oa"}); err != nil {
		t.Fatalf("add node a: %v", err)
	}
	if err := dag.AddNode(&Node{
		ID:              "approval",
		Name:            "approval",
		Type:            NodeTypeApproval,
		Agent:           "approval",
		Goal:            "sign-off",
		Approvers:       []string{"admin"},
		ApprovalMessage: "Approve step",
		OutputKey:       "approval_result",
		Dependencies:    []string{"a"},
	}); err != nil {
		t.Fatalf("add approval node: %v", err)
	}
	if err := dag.AddNode(&Node{ID: "b", Name: "b", Agent: "agent-b", Goal: "goal-b", Dependencies: []string{"approval"}, OutputKey: "ob"}); err != nil {
		t.Fatalf("add node b: %v", err)
	}
	if err := dag.AddEdge("a", "approval"); err != nil {
		t.Fatalf("add edge: %v", err)
	}
	if err := dag.AddEdge("approval", "b"); err != nil {
		t.Fatalf("add edge: %v", err)
	}

	ctx := context.Background()
	result, err := exec.Execute(ctx, dag, "wf-approval-test")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if result.Status != "completed" {
		t.Fatalf("expected completed, got %s", result.Status)
	}
	ar := result.NodeResults["approval"]
	if ar == nil || ar.Status != "completed" {
		t.Fatalf("approval node expected completed, got %v", ar)
	}
}
