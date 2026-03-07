package policy

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryApprovalQueue is an in-memory implementation of ApprovalQueue
type MemoryApprovalQueue struct {
	mu   sync.RWMutex
	reqs map[string]*ApprovalRequest
}

// NewMemoryApprovalQueue creates an in-memory approval queue
func NewMemoryApprovalQueue() *MemoryApprovalQueue {
	return &MemoryApprovalQueue{
		reqs: make(map[string]*ApprovalRequest),
	}
}

// Enqueue adds a new approval request
func (q *MemoryApprovalQueue) Enqueue(ctx context.Context, tool string, input map[string]any, approvers []string, runID, stepID string) (string, error) {
	id := uuid.New().String()
	q.mu.Lock()
	defer q.mu.Unlock()
	q.reqs[id] = &ApprovalRequest{
		ID:        id,
		Tool:      tool,
		Input:     input,
		Approvers: approvers,
		RunID:     runID,
		StepID:    stepID,
		CreatedAt: time.Now(),
		Status:    "pending",
	}
	return id, nil
}

// ListPending returns all pending requests
func (q *MemoryApprovalQueue) ListPending(ctx context.Context) ([]*ApprovalRequest, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	var out []*ApprovalRequest
	for _, r := range q.reqs {
		if r.Status == "pending" {
			out = append(out, r)
		}
	}
	return out, nil
}

// Get returns a request by ID
func (q *MemoryApprovalQueue) Get(ctx context.Context, id string) (*ApprovalRequest, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	r, ok := q.reqs[id]
	if !ok {
		return nil, nil
	}
	return r, nil
}

// Approve marks the request as approved
func (q *MemoryApprovalQueue) Approve(ctx context.Context, id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if r, ok := q.reqs[id]; ok && r.Status == "pending" {
		r.Status = "approved"
	}
	return nil
}

// Deny marks the request as denied
func (q *MemoryApprovalQueue) Deny(ctx context.Context, id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if r, ok := q.reqs[id]; ok && r.Status == "pending" {
		r.Status = "denied"
	}
	return nil
}
