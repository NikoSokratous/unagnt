package api

import (
	"context"
	"errors"
	"sync"

	"github.com/NikoSokratous/unagnt/pkg/replay"
)

var errNotFound = errors.New("snapshot not found")

// MemoryReplayStore is an in-memory ReplayStore for testing.
type MemoryReplayStore struct {
	mu        sync.RWMutex
	snapshots map[string]*replay.RunSnapshot
}

// NewMemoryReplayStore creates an in-memory replay store.
func NewMemoryReplayStore() *MemoryReplayStore {
	return &MemoryReplayStore{
		snapshots: make(map[string]*replay.RunSnapshot),
	}
}

// Save adds a snapshot (for testing).
func (s *MemoryReplayStore) Save(snap *replay.RunSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshots[snap.ID] = snap
}

// LoadSnapshot implements ReplayStore.
func (s *MemoryReplayStore) LoadSnapshot(ctx context.Context, id string) (*replay.RunSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap, ok := s.snapshots[id]
	if !ok {
		return nil, errNotFound
	}
	return snap, nil
}

// ListSnapshots implements ReplayStore.
func (s *MemoryReplayStore) ListSnapshots(ctx context.Context, runID string, limit int) ([]replay.SnapshotMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []replay.SnapshotMetadata
	for _, snap := range s.snapshots {
		if runID != "" && snap.RunID != runID {
			continue
		}
		list = append(list, replay.SnapshotMetadata{
			ID:         snap.ID,
			RunID:      snap.RunID,
			AgentName:  snap.AgentName,
			Goal:       snap.Goal,
			CreatedAt:  snap.CreatedAt,
			FinalState: snap.FinalState,
		})
		if len(list) >= limit {
			break
		}
	}
	return list, nil
}
