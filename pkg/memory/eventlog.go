package memory

import (
	"context"

	"github.com/NikoSokratous/unagnt/pkg/observe"
)

// EventLog is the interface for the immutable append-only execution log.
type EventLog interface {
	Append(ctx context.Context, runID string, evt *observe.Event) error
	GetRun(ctx context.Context, runID string) ([]*observe.Event, error)
}
