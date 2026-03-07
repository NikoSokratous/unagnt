package orchestrate

import (
	"context"
	"fmt"
	"sync"
)

// NewMemoryQueue creates an in-memory queue backend (non-durable).
func NewMemoryQueue(size int) QueueBackend {
	if size <= 0 {
		size = 256
	}
	return &memoryQueue{
		ch: make(chan RunRequest, size),
	}
}

type memoryQueue struct {
	ch   chan RunRequest
	once sync.Once
}

func (m *memoryQueue) Enqueue(_ context.Context, req RunRequest) error {
	select {
	case m.ch <- req:
		RunQueueDepth.Set(float64(len(m.ch)))
		return nil
	default:
		RunQueueRejected.Inc()
		return fmt.Errorf("runner queue is full")
	}
}

func (m *memoryQueue) Dequeue(ctx context.Context) (RunRequest, bool, error) {
	select {
	case <-ctx.Done():
		return RunRequest{}, false, ctx.Err()
	case req, ok := <-m.ch:
		if !ok {
			return RunRequest{}, false, nil
		}
		RunQueueDepth.Set(float64(len(m.ch)))
		return req, true, nil
	}
}

func (m *memoryQueue) Len() int {
	return len(m.ch)
}

func (m *memoryQueue) Close() error {
	m.once.Do(func() { close(m.ch) })
	return nil
}
