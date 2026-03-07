package sync

import "time"

// DeltaBundle is a snapshot of data for push/pull.
type DeltaBundle struct {
	Runs      []RunRecord `json:"runs"`
	Timestamp time.Time   `json:"timestamp"`
}

// RunRecord is run metadata for sync.
type RunRecord struct {
	RunID     string    `json:"run_id"`
	AgentName string    `json:"agent_name"`
	Goal      string    `json:"goal"`
	State     string    `json:"state"`
	StepCount int       `json:"step_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SyncStatus is the result of a sync status check.
type SyncStatus struct {
	LastPush   *time.Time `json:"last_push,omitempty"`
	LastPull   *time.Time `json:"last_pull,omitempty"`
	LocalRuns  int       `json:"local_runs"`
	RemoteRuns int       `json:"remote_runs,omitempty"`
	PendingPush int      `json:"pending_push,omitempty"`
}
