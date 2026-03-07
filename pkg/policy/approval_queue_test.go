package policy

import (
	"context"
	"testing"
)

func TestMemoryApprovalQueue(t *testing.T) {
	ctx := context.Background()
	q := NewMemoryApprovalQueue()

	id, err := q.Enqueue(ctx, "deploy", map[string]any{"env": "prod"}, []string{"alice"}, "run1", "step1")
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty id")
	}

	pending, err := q.ListPending(ctx)
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending, got %d", len(pending))
	}
	if pending[0].Tool != "deploy" {
		t.Errorf("tool: expected deploy, got %s", pending[0].Tool)
	}

	if err := q.Approve(ctx, id); err != nil {
		t.Fatalf("Approve: %v", err)
	}

	pending, _ = q.ListPending(ctx)
	if len(pending) != 0 {
		t.Errorf("expected 0 pending after approve, got %d", len(pending))
	}

	req, _ := q.Get(ctx, id)
	if req == nil || req.Status != "approved" {
		t.Errorf("Get: expected approved, got %v", req)
	}
}
