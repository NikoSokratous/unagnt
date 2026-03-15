package replay

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// SQLiteSnapshotStore persists snapshots to SQLite using run_snapshots (migration 006).
type SQLiteSnapshotStore struct {
	db *sql.DB
}

// NewSQLiteSnapshotStore creates a SQLite-backed snapshot store.
func NewSQLiteSnapshotStore(db *sql.DB) *SQLiteSnapshotStore {
	return &SQLiteSnapshotStore{db: db}
}

// sqliteTimeFormat is a format SQLite's julianday() and datetime() understand.
const sqliteTimeFormat = "2006-01-02 15:04:05.000"

// SaveSnapshot persists a snapshot to run_snapshots.
func (s *SQLiteSnapshotStore) SaveSnapshot(ctx context.Context, snapshot *RunSnapshot) error {
	modelCalls, _ := json.Marshal(snapshot.ModelCalls)
	toolCalls, _ := json.Marshal(snapshot.ToolCalls)
	agentConfig, _ := json.Marshal(snapshot.AgentConfig)
	env, _ := json.Marshal(snapshot.Environment)
	checksums, _ := json.Marshal(snapshot.Checksums)

	createdAt := snapshot.CreatedAt.Format(sqliteTimeFormat)
	startTime := snapshot.StartTime.Format(sqliteTimeFormat)
	endTime := snapshot.EndTime.Format(sqliteTimeFormat)

	_, err := s.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO run_snapshots (
			id, run_id, version, created_at, agent_name, goal, agent_config,
			model_calls, tool_calls, environment, start_time, end_time,
			final_state, checksums, compressed, encrypted, size_bytes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		snapshot.ID, snapshot.RunID, snapshot.Version, createdAt,
		snapshot.AgentName, snapshot.Goal, agentConfig,
		modelCalls, toolCalls, env,
		startTime, endTime,
		snapshot.FinalState, checksums,
		snapshot.Compressed, snapshot.Encrypted, snapshot.SizeBytes,
	)

	return err
}

// parseSQLiteTime parses timestamps stored by SaveSnapshot.
func parseSQLiteTime(s string) time.Time {
	t, _ := time.ParseInLocation(sqliteTimeFormat, s, time.UTC)
	if t.IsZero() {
		t, _ = time.Parse(time.RFC3339, s)
	}
	return t
}

// LoadSnapshot retrieves a snapshot by ID.
func (s *SQLiteSnapshotStore) LoadSnapshot(ctx context.Context, snapshotID string) (*RunSnapshot, error) {
	var modelCalls, toolCalls, agentConfig, env, checksums []byte
	var createdStr, startStr, endStr string
	snap := &RunSnapshot{}

	err := s.db.QueryRowContext(ctx, `
		SELECT id, run_id, version, created_at, agent_name, goal, agent_config,
		       model_calls, tool_calls, environment, start_time, end_time,
		       final_state, checksums, compressed, encrypted, size_bytes
		FROM run_snapshots WHERE id = ?`, snapshotID,
	).Scan(
		&snap.ID, &snap.RunID, &snap.Version, &createdStr,
		&snap.AgentName, &snap.Goal, &agentConfig,
		&modelCalls, &toolCalls, &env,
		&startStr, &endStr,
		&snap.FinalState, &checksums,
		&snap.Compressed, &snap.Encrypted, &snap.SizeBytes,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("snapshot not found: %s", snapshotID)
	}
	if err != nil {
		return nil, err
	}

	snap.CreatedAt = parseSQLiteTime(createdStr)
	snap.StartTime = parseSQLiteTime(startStr)
	snap.EndTime = parseSQLiteTime(endStr)

	_ = json.Unmarshal(modelCalls, &snap.ModelCalls)
	_ = json.Unmarshal(toolCalls, &snap.ToolCalls)
	_ = json.Unmarshal(agentConfig, &snap.AgentConfig)
	_ = json.Unmarshal(env, &snap.Environment)
	_ = json.Unmarshal(checksums, &snap.Checksums)

	return snap, nil
}

// ListSnapshots lists snapshots. If runID is empty, returns all (up to limit).
func (s *SQLiteSnapshotStore) ListSnapshots(ctx context.Context, runID string, limit int) ([]SnapshotMetadata, error) {
	query := `SELECT id, run_id, agent_name, goal, created_at,
	         COALESCE((julianday(end_time) - julianday(start_time)) * 86400, 0),
	         final_state, COALESCE(size_bytes,0), compressed
	         FROM run_snapshots`
	args := []interface{}{}
	if runID != "" {
		query += ` WHERE run_id = ?`
		args = append(args, runID)
	}
	query += ` ORDER BY created_at DESC LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []SnapshotMetadata
	for rows.Next() {
		var m SnapshotMetadata
		var durSec float64
		var createdStr string
		if err := rows.Scan(&m.ID, &m.RunID, &m.AgentName, &m.Goal, &createdStr,
			&durSec, &m.FinalState, &m.SizeBytes, &m.Compressed); err != nil {
			return nil, err
		}
		m.CreatedAt = parseSQLiteTime(createdStr)
		m.Duration = time.Duration(durSec * 1e9)
		list = append(list, m)
	}

	return list, rows.Err()
}

// DeleteSnapshot removes a snapshot.
func (s *SQLiteSnapshotStore) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM run_snapshots WHERE id = ?`, snapshotID)
	return err
}
