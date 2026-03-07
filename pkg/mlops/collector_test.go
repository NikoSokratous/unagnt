package mlops

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupCollectorDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	_, err = db.Exec(`
		CREATE TABLE model_performance_snapshots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			model_id TEXT NOT NULL,
			version TEXT NOT NULL,
			provider TEXT NOT NULL,
			latency_p50_ms INTEGER NOT NULL,
			latency_p95_ms INTEGER NOT NULL,
			latency_p99_ms INTEGER NOT NULL,
			error_rate REAL NOT NULL,
			throughput REAL NOT NULL,
			sample_count INTEGER NOT NULL,
			timestamp TIMESTAMP NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	return db
}

func TestCollector_RecordAndFlush(t *testing.T) {
	db := setupCollectorDB(t)
	ctx := context.Background()
	c := NewCollector(db, 10)

	c.Record("openai", "gpt-4", 100, true)
	c.Record("openai", "gpt-4", 150, true)
	c.Record("openai", "gpt-4", 200, false)

	if err := c.Flush(ctx); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM model_performance_snapshots").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 snapshot, got %d", count)
	}

	var errRate float64
	var p99 int64
	if err := db.QueryRow("SELECT error_rate, latency_p99_ms FROM model_performance_snapshots").Scan(&errRate, &p99); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if errRate < 0.3 || errRate > 0.4 {
		t.Errorf("error_rate expected ~0.33, got %v", errRate)
	}
	if p99 != 200 {
		t.Errorf("p99 expected 200, got %d", p99)
	}
}

func TestPerformanceStore_DetectDrift(t *testing.T) {
	db := setupCollectorDB(t)
	ctx := context.Background()
	now := time.Now()

	// Insert baseline (7 days ago) and recent
	_, err := db.Exec(`
		INSERT INTO model_performance_snapshots (model_id, version, provider, latency_p50_ms, latency_p95_ms, latency_p99_ms, error_rate, throughput, sample_count, timestamp)
		VALUES ('gpt-4', 'latest', 'openai', 50, 100, 150, 0.01, 100, 100, ?),
		       ('gpt-4', 'latest', 'openai', 80, 180, 300, 0.08, 90, 100, ?)
	`, now.Add(-7*24*time.Hour).Format(time.RFC3339), now.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	ps := NewPerformanceStore(db)
	result, err := ps.DetectDrift(ctx, "openai", "gpt-4", 1.5, 0.05)
	if err != nil {
		t.Fatalf("DetectDrift: %v", err)
	}
	if !result.Drifted {
		t.Error("expected drift (latency 300 > 1.5*150, error 0.08-0.01 > 0.05)")
	}
}
