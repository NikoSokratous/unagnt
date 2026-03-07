package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/observe"
	"github.com/NikoSokratous/unagnt/pkg/runtime"
	_ "modernc.org/sqlite"
)

// SQLite implements Store using SQLite.
type SQLite struct {
	db *sql.DB
}

// NewSQLite creates a SQLite store.
func NewSQLite(path string) (*SQLite, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	s := &SQLite{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *SQLite) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS runs (
			run_id TEXT PRIMARY KEY,
			agent_name TEXT,
			goal TEXT,
			state TEXT,
			step_count INTEGER,
			created_at TEXT,
			updated_at TEXT
		);
		CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			run_id TEXT NOT NULL,
			step_id TEXT,
			timestamp TEXT,
			type TEXT,
			agent TEXT,
			data TEXT,
			model_provider TEXT,
			model_name TEXT,
			FOREIGN KEY (run_id) REFERENCES runs(run_id)
		);
		CREATE TABLE IF NOT EXISTS history (
			run_id TEXT NOT NULL,
			step_id TEXT,
			timestamp TEXT,
			state TEXT,
			action_tool TEXT,
			action_version TEXT,
			action_input TEXT,
			result_tool_id TEXT,
			result_output TEXT,
			result_error TEXT,
			result_duration INTEGER,
			reasoning TEXT,
			metadata TEXT,
			FOREIGN KEY (run_id) REFERENCES runs(run_id)
		);
		CREATE TABLE IF NOT EXISTS dead_letters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			run_id TEXT NOT NULL,
			agent_name TEXT NOT NULL,
			goal TEXT NOT NULL,
			source TEXT NOT NULL,
			error TEXT NOT NULL,
			payload TEXT,
			attempt INTEGER NOT NULL,
			max_retries INTEGER NOT NULL,
			failed_at TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_events_run ON events(run_id);
		CREATE INDEX IF NOT EXISTS idx_history_run ON history(run_id);
		CREATE INDEX IF NOT EXISTS idx_dead_letters_failed_at ON dead_letters(failed_at DESC);
	`)
	return err
}

// SaveRun implements Store.
func (s *SQLite) SaveRun(ctx context.Context, run *RunMeta) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO runs (run_id, agent_name, goal, state, step_count, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		run.RunID, run.AgentName, run.Goal, run.State, run.StepCount,
		run.CreatedAt.Format(time.RFC3339), run.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetRun implements Store.
func (s *SQLite) GetRun(ctx context.Context, runID string) (*RunMeta, error) {
	var r RunMeta
	var ca, ua string
	err := s.db.QueryRowContext(ctx,
		`SELECT run_id, agent_name, goal, state, step_count, created_at, updated_at FROM runs WHERE run_id = ?`,
		runID,
	).Scan(&r.RunID, &r.AgentName, &r.Goal, &r.State, &r.StepCount, &ca, &ua)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.CreatedAt, _ = time.Parse(time.RFC3339, ca)
	r.UpdatedAt, _ = time.Parse(time.RFC3339, ua)
	return &r, nil
}

// SaveEvent implements Store.
func (s *SQLite) SaveEvent(ctx context.Context, runID string, evt *observe.Event) error {
	data, _ := json.Marshal(evt.Data)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO events (run_id, step_id, timestamp, type, agent, data, model_provider, model_name) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		runID, evt.StepID, evt.Timestamp.Format(time.RFC3339), evt.Type, evt.Agent, string(data), evt.Model.Provider, evt.Model.Name,
	)
	return err
}

// GetEvents implements Store.
func (s *SQLite) GetEvents(ctx context.Context, runID string) ([]*observe.Event, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT run_id, step_id, timestamp, type, agent, data, model_provider, model_name FROM events WHERE run_id = ? ORDER BY id`,
		runID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*observe.Event
	for rows.Next() {
		var e observe.Event
		var dataJSON, ts string
		if err := rows.Scan(&e.RunID, &e.StepID, &ts, &e.Type, &e.Agent, &dataJSON, &e.Model.Provider, &e.Model.Name); err != nil {
			return nil, err
		}
		e.Timestamp, _ = time.Parse(time.RFC3339, ts)
		_ = json.Unmarshal([]byte(dataJSON), &e.Data)
		out = append(out, &e)
	}
	return out, rows.Err()
}

// GetHistory implements Store.
func (s *SQLite) GetHistory(ctx context.Context, runID string) ([]runtime.StepRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT step_id, timestamp, state, action_tool, action_version, action_input, result_tool_id, result_output, result_error, result_duration, reasoning, metadata FROM history WHERE run_id = ? ORDER BY rowid`,
		runID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []runtime.StepRecord
	for rows.Next() {
		var r runtime.StepRecord
		var ts, actionInput, resultOutput, resultError, metadata string
		var actionTool, actionVersion, resultToolID *string
		var resultDuration *int64
		if err := rows.Scan(&r.StepID, &ts, &r.State, &actionTool, &actionVersion, &actionInput, &resultToolID, &resultOutput, &resultError, &resultDuration, &r.Reasoning, &metadata); err != nil {
			return nil, err
		}
		r.Timestamp, _ = time.Parse(time.RFC3339, ts)
		if actionTool != nil {
			r.Action = &runtime.ToolCall{Tool: *actionTool, Version: "", Input: nil}
			if actionVersion != nil {
				r.Action.Version = *actionVersion
			}
			_ = json.Unmarshal([]byte(actionInput), &r.Action.Input)
		}
		if resultToolID != nil {
			r.Result = &runtime.ToolResult{ToolID: *resultToolID}
			_ = json.Unmarshal([]byte(resultOutput), &r.Result.Output)
			if resultError != "" {
				r.Result.Error = resultError
			}
			if resultDuration != nil {
				r.Result.Duration = time.Duration(*resultDuration)
			}
		}
		_ = json.Unmarshal([]byte(metadata), &r.Metadata)
		out = append(out, r)
	}
	return out, rows.Err()
}

// SaveHistory implements Store.
func (s *SQLite) SaveHistory(ctx context.Context, runID string, history []runtime.StepRecord) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, h := range history {
		var actionTool, actionVersion, actionInput, resultToolID, resultOutput, resultError string
		var resultDuration int64
		if h.Action != nil {
			actionTool = h.Action.Tool
			actionVersion = h.Action.Version
			b, _ := json.Marshal(h.Action.Input)
			actionInput = string(b)
		}
		if h.Result != nil {
			resultToolID = h.Result.ToolID
			b, _ := json.Marshal(h.Result.Output)
			resultOutput = string(b)
			resultError = h.Result.Error
			resultDuration = int64(h.Result.Duration)
		}
		metadata, _ := json.Marshal(h.Metadata)
		_, err := tx.ExecContext(ctx,
			`INSERT INTO history (run_id, step_id, timestamp, state, action_tool, action_version, action_input, result_tool_id, result_output, result_error, result_duration, reasoning, metadata) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			runID, h.StepID, h.Timestamp.Format(time.RFC3339), h.State, actionTool, actionVersion, actionInput, resultToolID, resultOutput, resultError, resultDuration, h.Reasoning, string(metadata),
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ListRuns implements Store.
func (s *SQLite) ListRuns(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT run_id FROM runs ORDER BY updated_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// Close closes the database.
func (s *SQLite) Close() error {
	return s.db.Close()
}

// SaveDeadLetter persists a terminal run failure for investigation.
func (s *SQLite) SaveDeadLetter(ctx context.Context, dl *DeadLetter) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO dead_letters (run_id, agent_name, goal, source, error, payload, attempt, max_retries, failed_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		dl.RunID, dl.AgentName, dl.Goal, dl.Source, dl.Error, dl.Payload, dl.Attempt, dl.MaxRetries, dl.FailedAt.Format(time.RFC3339),
	)
	return err
}

// ListDeadLetters returns recent terminal failures.
func (s *SQLite) ListDeadLetters(ctx context.Context, limit int) ([]DeadLetter, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT run_id, agent_name, goal, source, error, payload, attempt, max_retries, failed_at FROM dead_letters ORDER BY failed_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]DeadLetter, 0, limit)
	for rows.Next() {
		var dl DeadLetter
		var failedAt string
		if err := rows.Scan(&dl.RunID, &dl.AgentName, &dl.Goal, &dl.Source, &dl.Error, &dl.Payload, &dl.Attempt, &dl.MaxRetries, &failedAt); err != nil {
			return nil, err
		}
		dl.FailedAt, _ = time.Parse(time.RFC3339, failedAt)
		out = append(out, dl)
	}
	return out, rows.Err()
}

// GetLatestDeadLetterByRunID returns the latest dead-letter entry for a run.
func (s *SQLite) GetLatestDeadLetterByRunID(ctx context.Context, runID string) (*DeadLetter, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT run_id, agent_name, goal, source, error, payload, attempt, max_retries, failed_at
		 FROM dead_letters
		 WHERE run_id = ?
		 ORDER BY failed_at DESC
		 LIMIT 1`,
		runID,
	)

	var dl DeadLetter
	var failedAt string
	if err := row.Scan(&dl.RunID, &dl.AgentName, &dl.Goal, &dl.Source, &dl.Error, &dl.Payload, &dl.Attempt, &dl.MaxRetries, &failedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	dl.FailedAt, _ = time.Parse(time.RFC3339, failedAt)
	return &dl, nil
}
