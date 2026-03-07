package orchestrate

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestWorkflowEngineSequential(t *testing.T) {
	// Create a mock workflow
	workflow := &WorkflowConfig{
		Name:        "test-sequential",
		Description: "Test sequential workflow",
		Steps: []WorkflowStep{
			{
				Name:      "step1",
				Agent:     "agent1",
				Goal:      "Goal 1",
				OutputKey: "result1",
			},
			{
				Name:      "step2",
				Agent:     "agent2",
				Goal:      "Goal 2 with {{.result1}}",
				OutputKey: "result2",
			},
		},
		Timeout: "5m",
		OnError: "stop",
	}

	// Create engine
	engine := NewWorkflowEngineWithExecutor(nil, SimulatedExecutor{})

	// Execute
	ctx := context.Background()
	result, err := engine.Execute(ctx, workflow)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Status != "completed" && result.Status != "partial" {
		t.Errorf("Expected completed or partial, got %s", result.Status)
	}

	if len(result.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(result.Steps))
	}
}

func TestWorkflowEngineParallel(t *testing.T) {
	workflow := &WorkflowConfig{
		Name:        "test-parallel",
		Description: "Test parallel workflow",
		Parallel: []WorkflowStep{
			{
				Name:      "parallel1",
				Agent:     "agent1",
				Goal:      "Goal 1",
				OutputKey: "result1",
			},
			{
				Name:      "parallel2",
				Agent:     "agent2",
				Goal:      "Goal 2",
				OutputKey: "result2",
			},
			{
				Name:      "parallel3",
				Agent:     "agent3",
				Goal:      "Goal 3",
				OutputKey: "result3",
			},
		},
		Timeout: "5m",
		OnError: "continue",
	}

	engine := NewWorkflowEngineWithExecutor(nil, SimulatedExecutor{})

	ctx := context.Background()
	result, err := engine.Execute(ctx, workflow)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(result.Steps) != 3 {
		t.Errorf("Expected 3 steps, got %d", len(result.Steps))
	}

	// Check that parallel execution was faster than sequential would be
	// Each step simulates 100ms, so parallel should be ~100ms, not 300ms
	if result.Duration > 500*time.Millisecond {
		t.Errorf("Parallel execution took too long: %v", result.Duration)
	}
}

func TestWorkflowValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  WorkflowConfig
		wantErr bool
	}{
		{
			name: "valid sequential",
			config: WorkflowConfig{
				Name: "test",
				Steps: []WorkflowStep{
					{Agent: "agent1", Goal: "goal1"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid parallel",
			config: WorkflowConfig{
				Name: "test",
				Parallel: []WorkflowStep{
					{Agent: "agent1", Goal: "goal1"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			config: WorkflowConfig{
				Steps: []WorkflowStep{
					{Agent: "agent1", Goal: "goal1"},
				},
			},
			wantErr: true,
		},
		{
			name: "no steps",
			config: WorkflowConfig{
				Name: "test",
			},
			wantErr: true,
		},
		{
			name: "both steps and parallel",
			config: WorkflowConfig{
				Name: "test",
				Steps: []WorkflowStep{
					{Agent: "agent1", Goal: "goal1"},
				},
				Parallel: []WorkflowStep{
					{Agent: "agent2", Goal: "goal2"},
				},
			},
			wantErr: true,
		},
		{
			name: "step missing agent",
			config: WorkflowConfig{
				Name: "test",
				Steps: []WorkflowStep{
					{Goal: "goal1"},
				},
			},
			wantErr: true,
		},
		{
			name: "step missing goal",
			config: WorkflowConfig{
				Name: "test",
				Steps: []WorkflowStep{
					{Agent: "agent1"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadWorkflow(t *testing.T) {
	// Create a temporary workflow file
	tmpfile, err := os.CreateTemp("", "workflow-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	workflowYAML := `
name: test-workflow
description: Test workflow
steps:
  - name: step1
    agent: agent1
    goal: "Test goal"
    output_key: result
timeout: 5m
on_error: stop
`

	if _, err := tmpfile.Write([]byte(workflowYAML)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	workflow, err := LoadWorkflow(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadWorkflow failed: %v", err)
	}

	if workflow.Name != "test-workflow" {
		t.Errorf("Expected name 'test-workflow', got '%s'", workflow.Name)
	}

	if len(workflow.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(workflow.Steps))
	}
}
