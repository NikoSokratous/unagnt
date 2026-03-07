package orchestrate

import "context"

// QueueBackend is the abstraction for run request queues (in-memory or durable).
type QueueBackend interface {
	Enqueue(ctx context.Context, req RunRequest) error
	Dequeue(ctx context.Context) (RunRequest, bool, error)
	Len() int
	Close() error
}
