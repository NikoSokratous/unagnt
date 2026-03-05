package integration

import (
	"context"
	"testing"
	"time"

	"github.com/NikoSokratous/unagnt/internal/store"
	"github.com/NikoSokratous/unagnt/pkg/observe"
	"github.com/NikoSokratous/unagnt/pkg/runtime"
	"github.com/NikoSokratous/unagnt/pkg/tool"
	"github.com/NikoSokratous/unagnt/pkg/tool/builtin"
)

// TestEndToEndAgentExecution tests a full agent run
func TestEndToEndAgentExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create temporary store
	tmpDB := t.TempDir() + "/test.db"
	st, err := store.NewSQLite(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer st.Close()

	// Register tools
	registry := tool.NewRegistry()
	for _, t := range builtin.All() {
		registry.Register(t)
	}

	// Create mock planner that returns echo action
	mockPlanner := &mockPlanner{
		actions: []runtime.PlannedAction{
			{
				Tool:      "echo",
				Version:   "1",
				Input:     map[string]any{"message": "test"},
				Reasoning: "Testing echo tool",
			},
		},
	}

	// Create engine with correct signature
	cfg := runtime.EngineConfig{
		AgentName: "test-agent",
		Goal:      "test goal",
		Autonomy:  runtime.AutonomyStandard,
		MaxSteps:  5,
	}

	engine := runtime.NewEngine(cfg, mockPlanner, tool.NewExecutor(registry))

	// Run engine
	ctx := context.Background()
	_, err = engine.Run(ctx)
	if err != nil {
		t.Fatalf("Engine run failed: %v", err)
	}

	// Verify state
	state := engine.State()
	t.Logf("Final state: %+v", state)
	t.Logf("Step count: %d", state.StepCount)
	t.Logf("Final status: %v", state.Current)

	if state.Current != runtime.StateCompleted && state.Current != runtime.StateFailed {
		t.Errorf("Expected state Completed or Failed, got %v", state.Current)
	}
	if state.StepCount == 0 {
		t.Error("Expected at least one step")
	}
}

// mockPlanner provides predetermined actions
type mockPlanner struct {
	actions []runtime.PlannedAction
	index   int
}

func (m *mockPlanner) Plan(ctx context.Context, input runtime.StepInput) (*runtime.PlannedAction, error) {
	if m.index >= len(m.actions) {
		// Return empty action to signal completion
		return &runtime.PlannedAction{
			Tool:      "",
			Reasoning: "Task completed",
		}, nil
	}
	action := m.actions[m.index]
	m.index++
	return &action, nil
}

// TestStoreOperations tests store CRUD operations
func TestStoreOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	tmpDB := t.TempDir() + "/test.db"
	st, err := store.NewSQLite(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	now := time.Now()

	// Save run
	run := &store.RunMeta{
		RunID:     "test-run-1",
		AgentName: "test-agent",
		Goal:      "test goal",
		State:     "running",
		StepCount: 1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = st.SaveRun(ctx, run)
	if err != nil {
		t.Fatalf("SaveRun failed: %v", err)
	}

	// Get run
	retrieved, err := st.GetRun(ctx, "test-run-1")
	if err != nil {
		t.Fatalf("GetRun failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Run not found")
	}
	if retrieved.RunID != "test-run-1" {
		t.Errorf("RunID = %v, want test-run-1", retrieved.RunID)
	}

	// Save event
	evt := &observe.Event{
		RunID:     "test-run-1",
		StepID:    "step-1",
		Timestamp: now,
		Type:      observe.EventInit,
		Agent:     "test-agent",
		Model: observe.ModelMeta{
			Provider: "mock",
			Name:     "test-model",
		},
	}

	err = st.SaveEvent(ctx, "test-run-1", evt)
	if err != nil {
		t.Fatalf("SaveEvent failed: %v", err)
	}

	// Get events
	events, err := st.GetEvents(ctx, "test-run-1")
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	// List runs
	runs, err := st.ListRuns(ctx, 10)
	if err != nil {
		t.Fatalf("ListRuns failed: %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("Expected 1 run, got %d", len(runs))
	}
}
