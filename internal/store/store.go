package store

import (
	"context"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/observe"
	"github.com/NikoSokratous/unagnt/pkg/runtime"
)

// Store is the interface for persistent storage of runs and events.
type Store interface {
	// SaveRun stores run metadata.
	SaveRun(ctx context.Context, run *RunMeta) error
	// GetRun retrieves run metadata by ID.
	GetRun(ctx context.Context, runID string) (*RunMeta, error)
	// SaveEvent appends an event to the log.
	SaveEvent(ctx context.Context, runID string, evt *observe.Event) error
	// GetEvents returns all events for a run, in order.
	GetEvents(ctx context.Context, runID string) ([]*observe.Event, error)
	// GetHistory returns step history for a run (for replay).
	GetHistory(ctx context.Context, runID string) ([]runtime.StepRecord, error)
	// SaveHistory stores step history for a run.
	SaveHistory(ctx context.Context, runID string, history []runtime.StepRecord) error
	// ListRuns returns run IDs (optionally filtered).
	ListRuns(ctx context.Context, limit int) ([]string, error)
}

// RunMeta is run metadata.
type RunMeta struct {
	RunID     string    `json:"run_id"`
	AgentName string    `json:"agent_name"`
	Goal      string    `json:"goal"`
	State     string    `json:"state"`
	StepCount int       `json:"step_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DeadLetter captures terminal failures for later inspection/replay.
type DeadLetter struct {
	RunID      string    `json:"run_id"`
	AgentName  string    `json:"agent_name"`
	Goal       string    `json:"goal"`
	Source     string    `json:"source"`
	Error      string    `json:"error"`
	Payload    string    `json:"payload,omitempty"`
	Attempt    int       `json:"attempt"`
	MaxRetries int       `json:"max_retries"`
	FailedAt   time.Time `json:"failed_at"`
}
