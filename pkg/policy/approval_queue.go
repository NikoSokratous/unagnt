package policy

import (
	"context"
	"time"
)

// ApprovalRequest represents a pending approval request
type ApprovalRequest struct {
	ID        string
	Tool      string
	Input     map[string]any
	Approvers []string
	RunID     string
	StepID    string
	CreatedAt time.Time
	Status    string // "pending", "approved", "denied"
}

// ApprovalQueue manages pending approval requests
type ApprovalQueue interface {
	Enqueue(ctx context.Context, tool string, input map[string]any, approvers []string, runID, stepID string) (string, error)
	ListPending(ctx context.Context) ([]*ApprovalRequest, error)
	Get(ctx context.Context, id string) (*ApprovalRequest, error)
	Approve(ctx context.Context, id string) error
	Deny(ctx context.Context, id string) error
}
