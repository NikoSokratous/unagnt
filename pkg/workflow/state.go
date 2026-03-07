package workflow

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// WorkflowState represents the persistent state of a workflow.
type WorkflowState struct {
	WorkflowID   string                 `json:"workflow_id"`
	WorkflowName string                 `json:"workflow_name"`
	Status       string                 `json:"status"` // running, completed, failed, cancelled
	CurrentStep  string                 `json:"current_step"`
	Outputs      map[string]interface{} `json:"outputs"`
	StartedAt    time.Time              `json:"started_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// NodeState represents the state of a single node.
type NodeState struct {
	NodeID      string      `json:"node_id"`
	WorkflowID  string      `json:"workflow_id"`
	StepName    string      `json:"step_name"`
	Status      string      `json:"status"`
	Output      interface{} `json:"output,omitempty"`
	StartedAt   time.Time   `json:"started_at"`
	CompletedAt time.Time   `json:"completed_at"`
}

// StateStore manages workflow state persistence.
type StateStore struct {
	db *sql.DB
}

// NewStateStore creates a new state store.
func NewStateStore(db *sql.DB) *StateStore {
	return &StateStore{db: db}
}

// SaveCheckpoint saves a workflow checkpoint.
func (s *StateStore) SaveCheckpoint(ctx context.Context, state *WorkflowState) error {
	outputsJSON, err := json.Marshal(state.Outputs)
	if err != nil {
		return fmt.Errorf("marshal outputs: %w", err)
	}

	query := `
		INSERT INTO workflow_states (
			id, workflow_name, status, current_step, outputs, started_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			status = excluded.status,
			current_step = excluded.current_step,
			outputs = excluded.outputs,
			updated_at = excluded.updated_at
	`

	_, err = s.db.ExecContext(ctx, query,
		state.WorkflowID,
		state.WorkflowName,
		state.Status,
		state.CurrentStep,
		string(outputsJSON),
		state.StartedAt,
		state.UpdatedAt,
	)

	return err
}

// LoadCheckpoint loads a workflow checkpoint.
func (s *StateStore) LoadCheckpoint(ctx context.Context, workflowID string) (*WorkflowState, error) {
	query := `
		SELECT workflow_name, status, current_step, outputs, started_at, updated_at
		FROM workflow_states
		WHERE id = ?
	`

	var state WorkflowState
	var outputsJSON string

	err := s.db.QueryRowContext(ctx, query, workflowID).Scan(
		&state.WorkflowName,
		&state.Status,
		&state.CurrentStep,
		&outputsJSON,
		&state.StartedAt,
		&state.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	state.WorkflowID = workflowID

	// Unmarshal outputs
	if err := json.Unmarshal([]byte(outputsJSON), &state.Outputs); err != nil {
		return nil, fmt.Errorf("unmarshal outputs: %w", err)
	}

	return &state, nil
}

// SaveNodeState saves a node execution state.
func (s *StateStore) SaveNodeState(ctx context.Context, nodeState *NodeState) error {
	outputJSON, err := json.Marshal(nodeState.Output)
	if err != nil {
		return fmt.Errorf("marshal output: %w", err)
	}

	query := `
		INSERT INTO workflow_step_states (
			id, workflow_id, step_name, status, output, started_at, completed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			status = excluded.status,
			output = excluded.output,
			completed_at = excluded.completed_at
	`

	_, err = s.db.ExecContext(ctx, query,
		nodeStateRowID(nodeState.WorkflowID, nodeState.NodeID),
		nodeState.WorkflowID,
		nodeState.StepName,
		nodeState.Status,
		string(outputJSON),
		nodeState.StartedAt,
		nodeState.CompletedAt,
	)

	return err
}

// LoadNodeStates loads all node states for a workflow.
func (s *StateStore) LoadNodeStates(ctx context.Context, workflowID string) ([]NodeState, error) {
	query := `
		SELECT id, step_name, status, output, started_at, completed_at
		FROM workflow_step_states
		WHERE workflow_id = ?
		ORDER BY started_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states := make([]NodeState, 0)

	for rows.Next() {
		var state NodeState
		var outputJSON string

		var rowID string
		err := rows.Scan(
			&rowID,
			&state.StepName,
			&state.Status,
			&outputJSON,
			&state.StartedAt,
			&state.CompletedAt,
		)
		if err != nil {
			continue
		}

		state.WorkflowID = workflowID
		state.NodeID = nodeIDFromRowID(workflowID, rowID)

		// Unmarshal output
		if outputJSON != "" && outputJSON != "null" {
			var output interface{}
			if err := json.Unmarshal([]byte(outputJSON), &output); err == nil {
				state.Output = output
			}
		}

		states = append(states, state)
	}

	return states, nil
}

func nodeStateRowID(workflowID, nodeID string) string {
	if workflowID == "" {
		return nodeID
	}
	return workflowID + ":" + nodeID
}

func nodeIDFromRowID(workflowID, rowID string) string {
	prefix := workflowID + ":"
	if strings.HasPrefix(rowID, prefix) {
		return strings.TrimPrefix(rowID, prefix)
	}
	return rowID
}

// ListWorkflows lists all workflow states.
func (s *StateStore) ListWorkflows(ctx context.Context, status string, limit int) ([]WorkflowState, error) {
	query := `
		SELECT id, workflow_name, status, current_step, started_at, updated_at
		FROM workflow_states
	`

	args := make([]interface{}, 0)
	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}

	query += " ORDER BY updated_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	workflows := make([]WorkflowState, 0, limit)

	for rows.Next() {
		var wf WorkflowState

		err := rows.Scan(
			&wf.WorkflowID,
			&wf.WorkflowName,
			&wf.Status,
			&wf.CurrentStep,
			&wf.StartedAt,
			&wf.UpdatedAt,
		)
		if err != nil {
			continue
		}

		workflows = append(workflows, wf)
	}

	return workflows, nil
}

// DeleteWorkflowState removes a workflow state.
func (s *StateStore) DeleteWorkflowState(ctx context.Context, workflowID string) error {
	// Delete node states first (foreign key)
	_, err := s.db.ExecContext(ctx, "DELETE FROM workflow_step_states WHERE workflow_id = ?", workflowID)
	if err != nil {
		return err
	}

	// Delete workflow state
	_, err = s.db.ExecContext(ctx, "DELETE FROM workflow_states WHERE id = ?", workflowID)
	return err
}
