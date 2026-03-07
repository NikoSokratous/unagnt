package policy

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// AutoApprovalQueue approves all requests immediately. Used for testing or --auto-approve CLI.
type AutoApprovalQueue struct {
	mu   sync.RWMutex
	reqs map[string]*ApprovalRequest
}

// NewAutoApprovalQueue creates an approval queue that auto-approves
func NewAutoApprovalQueue() *AutoApprovalQueue {
	return &AutoApprovalQueue{
		reqs: make(map[string]*ApprovalRequest),
	}
}

// Enqueue adds a request and immediately marks it approved
func (q *AutoApprovalQueue) Enqueue(ctx context.Context, tool string, input map[string]any, approvers []string, runID, stepID string) (string, error) {
	id := uuid.New().String()
	q.mu.Lock()
	q.reqs[id] = &ApprovalRequest{
		ID:        id,
		Tool:      tool,
		Input:     input,
		Approvers: approvers,
		RunID:     runID,
		StepID:    stepID,
		CreatedAt: time.Now(),
		Status:    "approved",
	}
	q.mu.Unlock()
	return id, nil
}

// ListPending returns empty (nothing stays pending)
func (q *AutoApprovalQueue) ListPending(ctx context.Context) ([]*ApprovalRequest, error) {
	return nil, nil
}

// Get returns the request (or nil if not found)
func (q *AutoApprovalQueue) Get(ctx context.Context, id string) (*ApprovalRequest, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	r, _ := q.reqs[id]
	return r, nil
}

// Approve no-op (already approved)
func (q *AutoApprovalQueue) Approve(ctx context.Context, id string) error {
	return nil
}

// Deny no-op
func (q *AutoApprovalQueue) Deny(ctx context.Context, id string) error {
	return nil
}
