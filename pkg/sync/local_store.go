package sync

import (
	"context"
	"sync"
	"time"

	"github.com/NikoSokratous/unagnt/internal/store"
)

// LocalStore reads and writes runs for sync.
type LocalStore interface {
	ListRunIDs(ctx context.Context, limit int) ([]string, error)
	GetRun(ctx context.Context, runID string) (*store.RunMeta, error)
	SaveRun(ctx context.Context, run *store.RunMeta) error
}

// StoreAdapter adapts store.Store to LocalStore.
type StoreAdapter struct {
	Store store.Store
}

func (a *StoreAdapter) ListRunIDs(ctx context.Context, limit int) ([]string, error) {
	return a.Store.ListRuns(ctx, limit)
}

func (a *StoreAdapter) GetRun(ctx context.Context, runID string) (*store.RunMeta, error) {
	return a.Store.GetRun(ctx, runID)
}

func (a *StoreAdapter) SaveRun(ctx context.Context, run *store.RunMeta) error {
	return a.Store.SaveRun(ctx, run)
}

// LocalSyncStore tracks sync state and builds/applies bundles.
type LocalSyncStore struct {
	store    LocalStore
	lastPush time.Time
	lastPull time.Time
	mu       sync.RWMutex
}

// NewLocalSyncStore creates a sync store.
func NewLocalSyncStore(s LocalStore) *LocalSyncStore {
	return &LocalSyncStore{store: s}
}

// BuildBundle creates a delta bundle from local runs (since sinceTime if nonzero).
func (ls *LocalSyncStore) BuildBundle(ctx context.Context, sinceTime time.Time) (*DeltaBundle, error) {
	ids, err := ls.store.ListRunIDs(ctx, 500)
	if err != nil {
		return nil, err
	}
	records := make([]RunRecord, 0, len(ids))
	for _, id := range ids {
		r, err := ls.store.GetRun(ctx, id)
		if err != nil || r == nil {
			continue
		}
		if !sinceTime.IsZero() && r.UpdatedAt.Before(sinceTime) {
			continue
		}
		records = append(records, RunRecord{
			RunID:     r.RunID,
			AgentName: r.AgentName,
			Goal:      r.Goal,
			State:     r.State,
			StepCount: r.StepCount,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		})
	}
	return &DeltaBundle{Runs: records, Timestamp: time.Now()}, nil
}

// ApplyBundle applies a received bundle (last-write-wins per run_id).
func (ls *LocalSyncStore) ApplyBundle(ctx context.Context, bundle *DeltaBundle) error {
	for _, r := range bundle.Runs {
		run := &store.RunMeta{
			RunID:     r.RunID,
			AgentName: r.AgentName,
			Goal:      r.Goal,
			State:     r.State,
			StepCount: r.StepCount,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		}
		if err := ls.store.SaveRun(ctx, run); err != nil {
			return err
		}
	}
	ls.mu.Lock()
	ls.lastPull = bundle.Timestamp
	ls.mu.Unlock()
	return nil
}

// SetLastPush records the last push time.
func (ls *LocalSyncStore) SetLastPush(t time.Time) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.lastPush = t
}

// LastPush returns the last push time.
func (ls *LocalSyncStore) LastPush() time.Time {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.lastPush
}

// LastPull returns the last pull time.
func (ls *LocalSyncStore) LastPull() time.Time {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.lastPull
}

// Status returns sync status.
func (ls *LocalSyncStore) Status(ctx context.Context) (*SyncStatus, error) {
	ids, err := ls.store.ListRunIDs(ctx, 1000)
	if err != nil {
		return nil, err
	}
	st := &SyncStatus{LocalRuns: len(ids)}
	ls.mu.RLock()
	if !ls.lastPush.IsZero() {
		t := ls.lastPush
		st.LastPush = &t
	}
	if !ls.lastPull.IsZero() {
		t := ls.lastPull
		st.LastPull = &t
	}
	ls.mu.RUnlock()
	return st, nil
}
