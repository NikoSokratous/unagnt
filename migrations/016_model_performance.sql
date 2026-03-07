-- Migration 016: Model Performance Snapshots (v4)
-- Stores model performance metrics for drift detection

CREATE TABLE IF NOT EXISTS model_performance_snapshots (
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

CREATE INDEX IF NOT EXISTS idx_model_perf_model ON model_performance_snapshots(model_id, provider);
CREATE INDEX IF NOT EXISTS idx_model_perf_timestamp ON model_performance_snapshots(timestamp DESC);
