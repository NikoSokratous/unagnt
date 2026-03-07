package mlops

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// PerformanceSnapshot represents a row from model_performance_snapshots
type PerformanceSnapshot struct {
	ID           int64
	ModelID      string
	Version      string
	Provider     string
	LatencyP50Ms int64
	LatencyP95Ms int64
	LatencyP99Ms int64
	ErrorRate    float64
	Throughput   float64
	SampleCount  int
	Timestamp    time.Time
}

// DriftResult indicates model drift status
type DriftResult struct {
	ModelID      string
	Provider     string
	Drifted      bool
	LatencyDelta float64  // recent p99 vs baseline p99 (ratio)
	ErrorDelta   float64  // recent error rate - baseline
	Message      string
}

// PerformanceStore reads model performance data
type PerformanceStore struct {
	db *sql.DB
}

// NewPerformanceStore creates a performance store
func NewPerformanceStore(db *sql.DB) *PerformanceStore {
	return &PerformanceStore{db: db}
}

// GetRecentSnapshots returns recent snapshots for a model
func (ps *PerformanceStore) GetRecentSnapshots(ctx context.Context, provider, modelID string, limit int) ([]PerformanceSnapshot, error) {
	if limit <= 0 {
		limit = 10
	}
	query := `
		SELECT id, model_id, version, provider, latency_p50_ms, latency_p95_ms, latency_p99_ms, error_rate, throughput, sample_count, timestamp
		FROM model_performance_snapshots
		WHERE provider = ? AND model_id = ?
		ORDER BY timestamp DESC LIMIT ?
	`
	rows, err := ps.db.QueryContext(ctx, query, provider, modelID, limit)
	if err != nil {
		return nil, fmt.Errorf("query snapshots: %w", err)
	}
	defer rows.Close()

	var out []PerformanceSnapshot
	for rows.Next() {
		var s PerformanceSnapshot
		var ts string
		if err := rows.Scan(&s.ID, &s.ModelID, &s.Version, &s.Provider, &s.LatencyP50Ms, &s.LatencyP95Ms, &s.LatencyP99Ms, &s.ErrorRate, &s.Throughput, &s.SampleCount, &ts); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		s.Timestamp, _ = time.Parse(time.RFC3339, ts)
		out = append(out, s)
	}
	return out, rows.Err()
}

// GetBaselineSnapshot returns the oldest snapshot in the window (baseline)
func (ps *PerformanceStore) GetBaselineSnapshot(ctx context.Context, provider, modelID string, since time.Time) (*PerformanceSnapshot, error) {
	query := `
		SELECT id, model_id, version, provider, latency_p50_ms, latency_p95_ms, latency_p99_ms, error_rate, throughput, sample_count, timestamp
		FROM model_performance_snapshots
		WHERE provider = ? AND model_id = ? AND timestamp >= ?
		ORDER BY timestamp ASC LIMIT 1
	`
	var s PerformanceSnapshot
	var ts string
	err := ps.db.QueryRowContext(ctx, query, provider, modelID, since).Scan(
		&s.ID, &s.ModelID, &s.Version, &s.Provider, &s.LatencyP50Ms, &s.LatencyP95Ms, &s.LatencyP99Ms, &s.ErrorRate, &s.Throughput, &s.SampleCount, &ts,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query baseline: %w", err)
	}
	s.Timestamp, _ = time.Parse(time.RFC3339, ts)
	return &s, nil
}

// DetectDrift compares recent performance vs baseline and flags drift
func (ps *PerformanceStore) DetectDrift(ctx context.Context, provider, modelID string, latencyThresholdRatio, errorThreshold float64) (*DriftResult, error) {
	recent, err := ps.GetRecentSnapshots(ctx, provider, modelID, 5)
	if err != nil {
		return nil, err
	}
	if len(recent) == 0 {
		return &DriftResult{ModelID: modelID, Provider: provider, Drifted: false, Message: "no data"}, nil
	}

	// Baseline = 7 days ago
	baselineSince := time.Now().Add(-7 * 24 * time.Hour)
	baseline, err := ps.GetBaselineSnapshot(ctx, provider, modelID, baselineSince)
	if err != nil || baseline == nil {
		return &DriftResult{ModelID: modelID, Provider: provider, Drifted: false, Message: "no baseline"}, nil
	}

	// Use most recent snapshot
	cur := recent[0]
	latencyRatio := 1.0
	if baseline.LatencyP99Ms > 0 {
		latencyRatio = float64(cur.LatencyP99Ms) / float64(baseline.LatencyP99Ms)
	}
	errorDelta := cur.ErrorRate - baseline.ErrorRate

	if latencyThresholdRatio <= 0 {
		latencyThresholdRatio = 1.5
	}
	if errorThreshold <= 0 {
		errorThreshold = 0.05
	}

	drifted := latencyRatio > latencyThresholdRatio || errorDelta > errorThreshold
	msg := "ok"
	if drifted {
		msg = fmt.Sprintf("latency_ratio=%.2f, error_delta=%.4f", latencyRatio, errorDelta)
	}

	return &DriftResult{
		ModelID:      modelID,
		Provider:     provider,
		Drifted:      drifted,
		LatencyDelta: latencyRatio - 1.0,
		ErrorDelta:   errorDelta,
		Message:      msg,
	}, nil
}
