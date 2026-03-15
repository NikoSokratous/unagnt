package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/policy"
)

// AgentNodeExecutor executes an agent node (runs a real agent). When nil, executeNode uses a placeholder.
type AgentNodeExecutor interface {
	Execute(ctx context.Context, agentName, goal string, outputs map[string]interface{}) (interface{}, error)
}

// ExecutionResult represents the result of a workflow execution.
type ExecutionResult struct {
	WorkflowID  string                 `json:"workflow_id"`
	Status      string                 `json:"status"` // running, completed, failed, partial
	Outputs     map[string]interface{} `json:"outputs"`
	NodeResults map[string]*NodeResult `json:"node_results"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt time.Time              `json:"completed_at"`
	Duration    time.Duration          `json:"duration"`
	Error       string                 `json:"error,omitempty"`
}

// NodeResult represents the result of a single node execution.
type NodeResult struct {
	NodeID      string                 `json:"node_id"`
	Status      string                 `json:"status"` // pending, running, completed, failed, skipped
	Output      interface{}            `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt time.Time              `json:"completed_at"`
	Duration    time.Duration          `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Executor executes DAG-based workflows.
type Executor struct {
	stateStore     *StateStore
	condEvaluator  ConditionEvaluator
	approvalQueue  policy.ApprovalQueue
	agentNodeExec  AgentNodeExecutor
	approvalPoll   time.Duration
	approvalExpiry time.Duration
}

// ConditionEvaluator evaluates conditional expressions.
type ConditionEvaluator interface {
	Evaluate(condition string, context map[string]interface{}) (bool, error)
}

// NewExecutor creates a new DAG executor.
func NewExecutor(stateStore *StateStore, condEvaluator ConditionEvaluator) *Executor {
	return &Executor{
		stateStore:     stateStore,
		condEvaluator:  condEvaluator,
		approvalPoll:   5 * time.Second,
		approvalExpiry: 24 * time.Hour,
	}
}

// NewExecutorWithApproval creates a DAG executor with approval queue for human-in-the-loop steps.
func NewExecutorWithApproval(stateStore *StateStore, condEvaluator ConditionEvaluator, approvalQueue policy.ApprovalQueue) *Executor {
	e := NewExecutor(stateStore, condEvaluator)
	e.approvalQueue = approvalQueue
	return e
}

// WithAgentNodeExecutor sets the executor for agent nodes (replaces placeholder).
func (e *Executor) WithAgentNodeExecutor(exec AgentNodeExecutor) *Executor {
	e.agentNodeExec = exec
	return e
}

// Execute runs a DAG workflow.
func (e *Executor) Execute(ctx context.Context, dag *DAG, workflowID string) (*ExecutionResult, error) {
	// Validate DAG
	if err := dag.Validate(); err != nil {
		return nil, fmt.Errorf("invalid DAG: %w", err)
	}

	// Initialize result
	result := &ExecutionResult{
		WorkflowID:  workflowID,
		Status:      "running",
		Outputs:     make(map[string]interface{}),
		NodeResults: make(map[string]*NodeResult),
		StartedAt:   time.Now(),
	}

	// Initialize node results
	for nodeID := range dag.Nodes {
		result.NodeResults[nodeID] = &NodeResult{
			NodeID: nodeID,
			Status: "pending",
		}
	}

	// Get execution levels (nodes that can run in parallel)
	levels, err := dag.GetExecutionLevels()
	if err != nil {
		return nil, fmt.Errorf("get execution levels: %w", err)
	}

	// Execute level by level
	for levelIdx, level := range levels {
		// Check context cancellation
		select {
		case <-ctx.Done():
			result.Status = "cancelled"
			result.Error = "workflow cancelled"
			result.CompletedAt = time.Now()
			result.Duration = result.CompletedAt.Sub(result.StartedAt)
			return result, ctx.Err()
		default:
		}

		// Execute all nodes in this level in parallel
		if err := e.executeLevel(ctx, dag, level, result); err != nil {
			result.Status = "failed"
			result.Error = fmt.Sprintf("level %d failed: %v", levelIdx, err)
			result.CompletedAt = time.Now()
			result.Duration = result.CompletedAt.Sub(result.StartedAt)
			return result, err
		}

		// Save checkpoint after each level
		if e.stateStore != nil {
			state := &WorkflowState{
				WorkflowID:   workflowID,
				WorkflowName: workflowID,
				Status:       "running",
				CurrentStep:  fmt.Sprintf("level-%d", levelIdx),
				Outputs:      result.Outputs,
				StartedAt:    result.StartedAt,
				UpdatedAt:    time.Now(),
			}
			if err := e.stateStore.SaveCheckpoint(ctx, state); err != nil {
				return result, fmt.Errorf("save checkpoint: %w", err)
			}
		}
	}

	// All levels completed
	result.Status = "completed"
	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	return result, nil
}

// executeLevel executes all nodes in a level in parallel.
func (e *Executor) executeLevel(ctx context.Context, dag *DAG, nodeIDs []string, result *ExecutionResult) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)

	for _, nodeID := range nodeIDs {
		wg.Add(1)

		go func(id string) {
			defer wg.Done()

			node := dag.Nodes[id]
			nodeResult := result.NodeResults[id]

			// Evaluate condition if present
			if node.Condition != "" && e.condEvaluator != nil {
				condCtx := map[string]interface{}{
					"outputs": result.Outputs,
				}

				shouldExecute, err := e.condEvaluator.Evaluate(node.Condition, condCtx)
				if err != nil {
					mu.Lock()
					nodeResult.Status = "failed"
					nodeResult.Error = fmt.Sprintf("condition evaluation failed: %v", err)
					errors = append(errors, err)
					mu.Unlock()
					return
				}

				if !shouldExecute {
					mu.Lock()
					nodeResult.Status = "skipped"
					nodeResult.CompletedAt = time.Now()
					mu.Unlock()
					return
				}
			}

			// Execute node
			mu.Lock()
			nodeResult.Status = "running"
			nodeResult.StartedAt = time.Now()
			mu.Unlock()

			var output interface{}
			var execErr error
			if node.Type == NodeTypeApproval {
				output, execErr = e.executeApprovalNode(ctx, node, id, result.WorkflowID, result.Outputs)
			} else {
				output, execErr = e.executeNode(ctx, node, result.Outputs)
			}

			mu.Lock()
			if execErr != nil {
				nodeResult.Status = "failed"
				nodeResult.Error = execErr.Error()
				errors = append(errors, execErr)
			} else {
				nodeResult.Status = "completed"
				nodeResult.Output = output

				// Store output if output_key is specified
				if node.OutputKey != "" {
					result.Outputs[node.OutputKey] = output
				}
			}

			nodeResult.CompletedAt = time.Now()
			nodeResult.Duration = nodeResult.CompletedAt.Sub(nodeResult.StartedAt)
			mu.Unlock()

			if e.stateStore != nil {
				_ = e.stateStore.SaveNodeState(ctx, &NodeState{
					NodeID:      id,
					WorkflowID:  result.WorkflowID,
					StepName:    node.Name,
					Status:      nodeResult.Status,
					Output:      nodeResult.Output,
					StartedAt:   nodeResult.StartedAt,
					CompletedAt: nodeResult.CompletedAt,
				})
			}
		}(nodeID)
	}

	wg.Wait()

	// Return first error if any
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

// executeApprovalNode runs a human-in-the-loop approval step.
func (e *Executor) executeApprovalNode(ctx context.Context, node *Node, nodeID, workflowID string, outputs map[string]interface{}) (interface{}, error) {
	if e.approvalQueue == nil {
		return nil, fmt.Errorf("approval step requires ApprovalQueue; use NewExecutorWithApproval")
	}
	approvers := node.Approvers
	if len(approvers) == 0 {
		approvers = []string{"default"}
	}
	msg := node.ApprovalMessage
	if msg == "" {
		msg = fmt.Sprintf("Approve workflow step: %s", node.Name)
	}
	input := map[string]any{
		"workflow_id": workflowID,
		"step_id":     nodeID,
		"message":     msg,
		"outputs":     outputs,
	}
	id, err := e.approvalQueue.Enqueue(ctx, "workflow-approval", input, approvers, workflowID, nodeID)
	if err != nil {
		return nil, fmt.Errorf("enqueue approval: %w", err)
	}
	deadline := time.Now().Add(e.approvalExpiry)
	ticker := time.NewTicker(e.approvalPoll)
	defer ticker.Stop()
	for {
		req, err := e.approvalQueue.Get(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("get approval: %w", err)
		}
		if req == nil {
			return nil, fmt.Errorf("approval request %s not found", id)
		}
		if req.Status == "approved" {
			return map[string]any{"approved": true, "approval_id": id}, nil
		}
		if req.Status == "denied" {
			return nil, fmt.Errorf("approval denied for step %s", node.Name)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("approval expired for step %s", node.Name)
			}
		}
	}
}

// executeNode executes a single node (agent or simulated).
func (e *Executor) executeNode(ctx context.Context, node *Node, outputs map[string]interface{}) (interface{}, error) {
	if e.agentNodeExec != nil && node.Agent != "" {
		goal := node.Goal
		if goal == "" {
			goal = "Complete the assigned task"
		}
		return e.agentNodeExec.Execute(ctx, node.Agent, goal, outputs)
	}
	// Fallback: simulate execution
	time.Sleep(50 * time.Millisecond)
	return map[string]interface{}{
		"node":   node.ID,
		"agent":  node.Agent,
		"status": "completed",
	}, nil
}

// Resume resumes a workflow from a checkpoint.
func (e *Executor) Resume(ctx context.Context, workflowID string, dag *DAG) (*ExecutionResult, error) {
	if e.stateStore == nil {
		return nil, fmt.Errorf("state store not configured")
	}

	// Load checkpoint
	state, err := e.stateStore.LoadCheckpoint(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("load checkpoint: %w", err)
	}

	// Reconstruct result
	result := &ExecutionResult{
		WorkflowID:  workflowID,
		Status:      "running",
		Outputs:     state.Outputs,
		NodeResults: make(map[string]*NodeResult),
		StartedAt:   state.StartedAt,
	}

	// Load node states
	nodeStates, err := e.stateStore.LoadNodeStates(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("load node states: %w", err)
	}

	for _, ns := range nodeStates {
		result.NodeResults[ns.NodeID] = &NodeResult{
			NodeID:      ns.NodeID,
			Status:      ns.Status,
			Output:      ns.Output,
			StartedAt:   ns.StartedAt,
			CompletedAt: ns.CompletedAt,
		}
	}

	// Initialize missing nodes
	for nodeID := range dag.Nodes {
		if _, exists := result.NodeResults[nodeID]; !exists {
			result.NodeResults[nodeID] = &NodeResult{
				NodeID: nodeID,
				Status: "pending",
			}
		}
	}

	// Continue execution from where it stopped
	// For simplicity, re-execute from the beginning but skip completed nodes
	return e.executeWithSkip(ctx, dag, result)
}

// executeWithSkip executes DAG, skipping already completed nodes.
func (e *Executor) executeWithSkip(ctx context.Context, dag *DAG, result *ExecutionResult) (*ExecutionResult, error) {
	levels, err := dag.GetExecutionLevels()
	if err != nil {
		return nil, err
	}

	for _, level := range levels {
		// Filter out completed/failed nodes
		activeNodes := make([]string, 0)
		for _, nodeID := range level {
			status := result.NodeResults[nodeID].Status
			if status != "completed" && status != "failed" {
				activeNodes = append(activeNodes, nodeID)
			}
		}

		if len(activeNodes) > 0 {
			if err := e.executeLevel(ctx, dag, activeNodes, result); err != nil {
				result.Status = "failed"
				result.Error = err.Error()
				result.CompletedAt = time.Now()
				result.Duration = result.CompletedAt.Sub(result.StartedAt)
				return result, err
			}
		}
	}

	result.Status = "completed"
	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	return result, nil
}
