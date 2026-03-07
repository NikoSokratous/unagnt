package cost

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Schema from migrations 012 + 015
	_, err = db.Exec(`
		CREATE TABLE cost_entries (
			id TEXT PRIMARY KEY,
			agent_id TEXT NOT NULL,
			tenant_id TEXT NOT NULL,
			user_id TEXT,
			provider TEXT NOT NULL,
			model TEXT NOT NULL,
			workflow_id TEXT,
			workflow_name TEXT,
			input_tokens INTEGER NOT NULL,
			output_tokens INTEGER NOT NULL,
			cost REAL NOT NULL,
			call_count INTEGER DEFAULT 1,
			timestamp TIMESTAMP NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	return db
}

func TestGetCostsByWorkflow(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	ct := NewCostTracker(db)

	// Insert test data directly (skip accumulator/flush)
	now := time.Now()
	for _, row := range []struct {
		id, agentID, tenantID, wfID, wfName string
		cost                                float64
	}{
		{"1", "a1", "t1", "wf1", "Workflow A", 1.5},
		{"2", "a2", "t1", "wf1", "Workflow A", 2.0},
		{"3", "a1", "t1", "wf2", "Workflow B", 0.5},
		{"4", "a1", "t1", "", "", 1.0},
	} {
		_, err := db.Exec(`
			INSERT INTO cost_entries (id, agent_id, tenant_id, user_id, provider, model, workflow_id, workflow_name, input_tokens, output_tokens, cost, call_count, timestamp)
			VALUES (?, ?, ?, '', 'openai', 'gpt-4', ?, ?, 100, 50, ?, 1, ?)
		`, row.id, row.agentID, row.tenantID, row.wfID, row.wfName, row.cost, now)
		if err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	costs, err := ct.GetCostsByWorkflow(ctx, "t1", now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil {
		t.Fatalf("GetCostsByWorkflow: %v", err)
	}
	if len(costs) != 3 {
		t.Errorf("expected 3 workflow buckets, got %d: %v", len(costs), costs)
	}
	if c := costs["wf1"]; c != 3.5 {
		t.Errorf("workflow wf1: expected 3.5, got %v", c)
	}
	if c := costs["wf2"]; c != 0.5 {
		t.Errorf("workflow wf2: expected 0.5, got %v", c)
	}
	if c := costs["(no workflow)"]; c != 1.0 {
		t.Errorf("no workflow: expected 1.0, got %v", c)
	}
}

func TestGetCostBreakdownWithFilter(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	ct := NewCostTracker(db)

	now := time.Now()
	_, err := db.Exec(`
		INSERT INTO cost_entries (id, agent_id, tenant_id, user_id, provider, model, workflow_id, workflow_name, input_tokens, output_tokens, cost, call_count, timestamp)
		VALUES ('1', 'a1', 't1', 'u1', 'openai', 'gpt-4', 'wf1', 'WF1', 100, 50, 1.0, 1, ?),
		       ('2', 'a2', 't1', 'u1', 'openai', 'gpt-4', 'wf1', 'WF1', 200, 100, 2.0, 1, ?),
		       ('3', 'a1', 't1', 'u1', 'openai', 'gpt-4', 'wf2', 'WF2', 100, 50, 0.5, 1, ?)
	`, now, now, now)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	// No filter
	entries, err := ct.GetCostBreakdown(ctx, "t1", now.Add(-time.Hour), now.Add(time.Hour), nil)
	if err != nil {
		t.Fatalf("GetCostBreakdown: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	// Filter by workflow
	entries, err = ct.GetCostBreakdown(ctx, "t1", now.Add(-time.Hour), now.Add(time.Hour), &CostBreakdownFilter{WorkflowID: "wf1"})
	if err != nil {
		t.Fatalf("GetCostBreakdown: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("filter workflow: expected 2 entries, got %d", len(entries))
	}
	for _, e := range entries {
		if e.WorkflowID != "wf1" {
			t.Errorf("expected workflow_id wf1, got %q", e.WorkflowID)
		}
	}

	// Filter by model
	entries, err = ct.GetCostBreakdown(ctx, "t1", now.Add(-time.Hour), now.Add(time.Hour), &CostBreakdownFilter{Model: "gpt-4"})
	if err != nil {
		t.Fatalf("GetCostBreakdown: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("filter model: expected 3 entries, got %d", len(entries))
	}
}
