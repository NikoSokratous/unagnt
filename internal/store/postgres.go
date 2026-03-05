package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/observe"
	"github.com/NikoSokratous/unagnt/pkg/runtime"
	_ "github.com/lib/pq"
)

// Postgres implements Store using PostgreSQL.
type Postgres struct {
	db *sql.DB
}

// NewPostgres creates a PostgreSQL store.
// connStr example: "postgres://user:password@localhost/dbname?sslmode=disable"
func NewPostgres(connStr string) (*Postgres, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("postgres ping failed: %w", err)
	}

	p := &Postgres{db: db}
	if err := p.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return p, nil
}

func (p *Postgres) migrate() error {
	_, err := p.db.Exec(`
		CREATE TABLE IF NOT EXISTS runs (
			run_id TEXT PRIMARY KEY,
			agent_name TEXT,
			goal TEXT,
			state TEXT,
			step_count INTEGER,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		);
		
		CREATE TABLE IF NOT EXISTS events (
			id SERIAL PRIMARY KEY,
			run_id TEXT NOT NULL,
			step_id TEXT,
			timestamp TIMESTAMP,
			type TEXT,
			agent TEXT,
			data JSONB,
			model_provider TEXT,
			model_name TEXT,
			FOREIGN KEY (run_id) REFERENCES runs(run_id)
		);
		
		CREATE TABLE IF NOT EXISTS history (
			run_id TEXT NOT NULL,
			step_id TEXT,
			timestamp TIMESTAMP,
			state TEXT,
			action_tool TEXT,
			action_version TEXT,
			action_input JSONB,
			result_tool_id TEXT,
			result_output JSONB,
			result_error TEXT,
			result_duration BIGINT,
			reasoning TEXT,
			metadata JSONB,
			FOREIGN KEY (run_id) REFERENCES runs(run_id)
		);
		
		CREATE INDEX IF NOT EXISTS idx_events_run ON events(run_id);
		CREATE INDEX IF NOT EXISTS idx_history_run ON history(run_id);
		CREATE INDEX IF NOT EXISTS idx_runs_updated ON runs(updated_at DESC);
	`)
	return err
}

// SaveRun implements Store.
func (p *Postgres) SaveRun(ctx context.Context, run *RunMeta) error {
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO runs (run_id, agent_name, goal, state, step_count, created_at, updated_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (run_id) DO UPDATE SET
		 	agent_name = EXCLUDED.agent_name,
		 	goal = EXCLUDED.goal,
		 	state = EXCLUDED.state,
		 	step_count = EXCLUDED.step_count,
		 	updated_at = EXCLUDED.updated_at`,
		run.RunID, run.AgentName, run.Goal, run.State, run.StepCount,
		run.CreatedAt, run.UpdatedAt,
	)
	return err
}

// GetRun implements Store.
func (p *Postgres) GetRun(ctx context.Context, runID string) (*RunMeta, error) {
	var r RunMeta
	err := p.db.QueryRowContext(ctx,
		`SELECT run_id, agent_name, goal, state, step_count, created_at, updated_at 
		 FROM runs WHERE run_id = $1`,
		runID,
	).Scan(&r.RunID, &r.AgentName, &r.Goal, &r.State, &r.StepCount, &r.CreatedAt, &r.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// SaveEvent implements Store.
func (p *Postgres) SaveEvent(ctx context.Context, runID string, evt *observe.Event) error {
	data, _ := json.Marshal(evt.Data)
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO events (run_id, step_id, timestamp, type, agent, data, model_provider, model_name) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		runID, evt.StepID, evt.Timestamp, evt.Type, evt.Agent, data, evt.Model.Provider, evt.Model.Name,
	)
	return err
}

// GetEvents implements Store.
func (p *Postgres) GetEvents(ctx context.Context, runID string) ([]*observe.Event, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT run_id, step_id, timestamp, type, agent, data, model_provider, model_name 
		 FROM events WHERE run_id = $1 ORDER BY id`,
		runID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*observe.Event
	for rows.Next() {
		var e observe.Event
		var dataJSON []byte
		if err := rows.Scan(&e.RunID, &e.StepID, &e.Timestamp, &e.Type, &e.Agent, &dataJSON, &e.Model.Provider, &e.Model.Name); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(dataJSON, &e.Data)
		out = append(out, &e)
	}
	return out, rows.Err()
}

// GetHistory implements Store.
func (p *Postgres) GetHistory(ctx context.Context, runID string) ([]runtime.StepRecord, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT step_id, timestamp, state, action_tool, action_version, action_input, 
		        result_tool_id, result_output, result_error, result_duration, reasoning, metadata 
		 FROM history WHERE run_id = $1 ORDER BY ctid`,
		runID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []runtime.StepRecord
	for rows.Next() {
		var r runtime.StepRecord
		var actionInput, resultOutput, metadata []byte
		var actionTool, actionVersion, resultToolID *string
		var resultError *string
		var resultDuration *int64

		if err := rows.Scan(&r.StepID, &r.Timestamp, &r.State, &actionTool, &actionVersion, &actionInput,
			&resultToolID, &resultOutput, &resultError, &resultDuration, &r.Reasoning, &metadata); err != nil {
			return nil, err
		}

		if actionTool != nil {
			var input map[string]any
			_ = json.Unmarshal(actionInput, &input)
			r.Action = &runtime.ToolCall{
				Tool:    *actionTool,
				Version: *actionVersion,
				Input:   input,
			}
		}

		if resultToolID != nil {
			var output map[string]any
			_ = json.Unmarshal(resultOutput, &output)
			r.Result = &runtime.ToolResult{
				ToolID: *resultToolID,
				Output: output,
			}
			if resultError != nil {
				r.Result.Error = *resultError
			}
			if resultDuration != nil {
				r.Result.Duration = time.Duration(*resultDuration)
			}
		}

		_ = json.Unmarshal(metadata, &r.Metadata)
		out = append(out, r)
	}
	return out, rows.Err()
}

// SaveHistory implements Store.
func (p *Postgres) SaveHistory(ctx context.Context, runID string, history []runtime.StepRecord) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, h := range history {
		var actionTool, actionVersion *string
		var actionInput []byte
		var resultToolID *string
		var resultOutput []byte
		var resultError *string
		var resultDuration *int64

		if h.Action != nil {
			actionTool = &h.Action.Tool
			actionVersion = &h.Action.Version
			actionInput, _ = json.Marshal(h.Action.Input)
		}

		if h.Result != nil {
			resultToolID = &h.Result.ToolID
			resultOutput, _ = json.Marshal(h.Result.Output)
			if h.Result.Error != "" {
				resultError = &h.Result.Error
			}
			d := int64(h.Result.Duration)
			resultDuration = &d
		}

		metadata, _ := json.Marshal(h.Metadata)
		_, err := tx.ExecContext(ctx,
			`INSERT INTO history (run_id, step_id, timestamp, state, action_tool, action_version, action_input, 
			                     result_tool_id, result_output, result_error, result_duration, reasoning, metadata) 
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
			runID, h.StepID, h.Timestamp, h.State, actionTool, actionVersion, actionInput,
			resultToolID, resultOutput, resultError, resultDuration, h.Reasoning, metadata,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ListRuns implements Store.
func (p *Postgres) ListRuns(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := p.db.QueryContext(ctx,
		`SELECT run_id FROM runs ORDER BY updated_at DESC LIMIT $1`,
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

// Close closes the database connection.
func (p *Postgres) Close() error {
	return p.db.Close()
}
