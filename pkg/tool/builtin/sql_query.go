package builtin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NikoSokratous/unagnt/pkg/tool"
	_ "modernc.org/sqlite"
)

// SQLQuery is a built-in tool for read-only and write SQL queries against SQLite.
// Used for DB agent demos; policy should restrict dangerous operations.
type SQLQuery struct{}

func (SQLQuery) Name() string    { return "sql_query" }
func (SQLQuery) Version() string { return "1" }
func (SQLQuery) Description() string {
	return "Execute a SQL query against a SQLite database. Use for SELECT, INSERT, UPDATE, DELETE. Database path is optional (default: demo.db in current directory)."
}
func (SQLQuery) Permissions() []tool.Permission {
	return []tool.Permission{{Scope: "db", Required: true}}
}

func (SQLQuery) InputSchema() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":     "object",
		"required": []string{"query"},
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "SQL query to execute (SELECT, INSERT, UPDATE, DELETE)",
			},
			"database": map[string]any{
				"type":        "string",
				"description": "Database file path (default: demo.db)",
			},
		},
	})
}

func (h SQLQuery) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
	var req struct {
		Query    string `json:"query"`
		Database string `json:"database"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if req.Query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if req.Database == "" {
		req.Database = "demo.db"
	}
	// Resolve path relative to cwd
	path := req.Database
	if !filepath.IsAbs(path) {
		cwd, _ := os.Getwd()
		path = filepath.Join(cwd, path)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	query := strings.TrimSpace(req.Query)
	upper := strings.ToUpper(query)

	if strings.HasPrefix(upper, "SELECT") {
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("query: %w", err)
		}
		defer rows.Close()

		cols, _ := rows.Columns()
		var results []map[string]any
		for rows.Next() {
			vals := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				return nil, fmt.Errorf("scan: %w", err)
			}
			row := make(map[string]any)
			for i, c := range cols {
				v := vals[i]
				if b, ok := v.([]byte); ok {
					v = string(b)
				}
				row[c] = v
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return map[string]any{
			"rows":      results,
			"row_count": len(results),
		}, nil
	}

	// INSERT, UPDATE, DELETE
	res, err := db.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("exec: %w", err)
	}
	affected, _ := res.RowsAffected()
	return map[string]any{
		"rows_affected": affected,
		"ok":            true,
	}, nil
}
