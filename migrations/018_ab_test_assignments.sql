-- Migration 018: A/B Test Assignments (v4)
-- Records which run used which model in an A/B test

CREATE TABLE IF NOT EXISTS ab_test_assignments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ab_test_id TEXT NOT NULL,
    run_id TEXT NOT NULL,
    model_chosen TEXT NOT NULL,  -- model_a or model_b
    timestamp TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_ab_assignments_test ON ab_test_assignments(ab_test_id);
CREATE INDEX IF NOT EXISTS idx_ab_assignments_run ON ab_test_assignments(run_id);
