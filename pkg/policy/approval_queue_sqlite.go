package policy

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SQLiteApprovalQueue is a SQLite-backed approval queue
type SQLiteApprovalQueue struct {
	db *sql.DB
}

// NewSQLiteApprovalQueue creates a SQLite-backed approval queue
func NewSQLiteApprovalQueue(db *sql.DB) *SQLiteApprovalQueue {
	return &SQLiteApprovalQueue{db: db}
}

func (q *SQLiteApprovalQueue) Enqueue(ctx context.Context, tool string, input map[string]any, approvers []string, runID, stepID string) (string, error) {
	id := uuid.New().String()
	inputJSON, _ := json.Marshal(input)
	approversJSON, _ := json.Marshal(approvers)

	_, err := q.db.ExecContext(ctx,
		`INSERT INTO approval_requests (id, tool, input, approvers, run_id, step_id, status, created_at) VALUES (?, ?, ?, ?, ?, ?, 'pending', ?)`,
		id, tool, string(inputJSON), string(approversJSON), runID, stepID, time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return "", fmt.Errorf("enqueue: %w", err)
	}
	return id, nil
}

func (q *SQLiteApprovalQueue) ListPending(ctx context.Context) ([]*ApprovalRequest, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT id, tool, input, approvers, run_id, step_id, created_at FROM approval_requests WHERE status = 'pending' ORDER BY created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}
	defer rows.Close()

	var out []*ApprovalRequest
	for rows.Next() {
		var inputJSON, approversJSON, createdAt string
		r := &ApprovalRequest{Status: "pending"}
		if err := rows.Scan(&r.ID, &r.Tool, &inputJSON, &approversJSON, &r.RunID, &r.StepID, &createdAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(inputJSON), &r.Input)
		_ = json.Unmarshal([]byte(approversJSON), &r.Approvers)
		r.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		out = append(out, r)
	}
	return out, rows.Err()
}

func (q *SQLiteApprovalQueue) Get(ctx context.Context, id string) (*ApprovalRequest, error) {
	var inputJSON, approversJSON, createdAt string
	r := &ApprovalRequest{}
	err := q.db.QueryRowContext(ctx, `SELECT id, tool, input, approvers, run_id, step_id, status, created_at FROM approval_requests WHERE id = ?`, id).Scan(
		&r.ID, &r.Tool, &inputJSON, &approversJSON, &r.RunID, &r.StepID, &r.Status, &createdAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}
	_ = json.Unmarshal([]byte(inputJSON), &r.Input)
	_ = json.Unmarshal([]byte(approversJSON), &r.Approvers)
	r.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return r, nil
}

func (q *SQLiteApprovalQueue) Approve(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, `UPDATE approval_requests SET status = 'approved' WHERE id = ? AND status = 'pending'`, id)
	if err != nil {
		return fmt.Errorf("approve: %w", err)
	}
	return nil
}

func (q *SQLiteApprovalQueue) Deny(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, `UPDATE approval_requests SET status = 'denied' WHERE id = ? AND status = 'pending'`, id)
	if err != nil {
		return fmt.Errorf("deny: %w", err)
	}
	return nil
}
