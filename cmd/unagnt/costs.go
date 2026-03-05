package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

const costEntriesSchema = `
CREATE TABLE IF NOT EXISTS cost_entries (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    user_id TEXT,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    input_tokens INTEGER NOT NULL,
    output_tokens INTEGER NOT NULL,
    cost REAL NOT NULL,
    call_count INTEGER DEFAULT 1,
    timestamp TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_cost_entries_agent ON cost_entries(agent_id);
CREATE INDEX IF NOT EXISTS idx_cost_entries_tenant ON cost_entries(tenant_id);
CREATE INDEX IF NOT EXISTS idx_cost_entries_timestamp ON cost_entries(timestamp DESC);
`

func newCostsCmd() *cobra.Command {
	var dbPath string
	var by string
	var dateRange string
	var tenant string

	cmd := &cobra.Command{
		Use:   "costs",
		Short: "View cost data",
		Long:  "Display cost data from the database. Use --by to group by agent, tenant, or user.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				dbPath = "agent.db"
			}
			absPath, err := filepath.Abs(dbPath)
			if err != nil {
				return err
			}
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				return fmt.Errorf("database not found at %s (run unagntd or unagnt run first to create cost data)", absPath)
			}

			db, err := sql.Open("sqlite", absPath)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer db.Close()

			if _, err := db.ExecContext(context.Background(), costEntriesSchema); err != nil {
				return fmt.Errorf("ensure schema: %w", err)
			}

			start, end := parseDateRange(dateRange)

			switch by {
			case "agent":
				return printCostsByAgent(db, tenant, start, end)
			case "tenant":
				return printCostsByTenant(db, start, end)
			case "user":
				return printCostsByUser(db, tenant, start, end)
			case "total", "":
				return printTotalCost(db, tenant, start, end)
			default:
				return fmt.Errorf("invalid --by: use agent, tenant, user, or total")
			}
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: agent.db)")
	cmd.Flags().StringVar(&by, "by", "total", "Group by: agent, tenant, user, or total")
	cmd.Flags().StringVar(&dateRange, "range", "24h", "Time range: 1h, 24h, 7d, 30d")
	cmd.Flags().StringVar(&tenant, "tenant", "", "Filter by tenant (optional)")
	return cmd
}

func parseDateRange(r string) (time.Time, time.Time) {
	end := time.Now()
	var start time.Time
	switch r {
	case "1h":
		start = end.Add(-1 * time.Hour)
	case "24h":
		start = end.Add(-24 * time.Hour)
	case "7d":
		start = end.Add(-7 * 24 * time.Hour)
	case "30d":
		start = end.Add(-30 * 24 * time.Hour)
	default:
		start = end.Add(-24 * time.Hour)
	}
	return start, end
}

func printTotalCost(db *sql.DB, tenant string, start, end time.Time) error {
	ctx := context.Background()
	var total float64
	var err error

	if tenant != "" {
		err = db.QueryRowContext(ctx,
			`SELECT COALESCE(SUM(cost), 0) FROM cost_entries WHERE tenant_id = ? AND timestamp BETWEEN ? AND ?`,
			tenant, start, end).Scan(&total)
	} else {
		err = db.QueryRowContext(ctx,
			`SELECT COALESCE(SUM(cost), 0) FROM cost_entries WHERE timestamp BETWEEN ? AND ?`,
			start, end).Scan(&total)
	}
	if err != nil {
		return err
	}

	fmt.Printf("Total cost: $%.4f (range: %s to %s)\n", total, start.Format("2006-01-02 15:04"), end.Format("2006-01-02 15:04"))
	return nil
}

func printCostsByAgent(db *sql.DB, tenant string, start, end time.Time) error {
	ctx := context.Background()
	var rows *sql.Rows
	var err error

	if tenant != "" {
		rows, err = db.QueryContext(ctx,
			`SELECT agent_id, SUM(cost) as total FROM cost_entries WHERE tenant_id = ? AND timestamp BETWEEN ? AND ? GROUP BY agent_id ORDER BY total DESC`,
			tenant, start, end)
	} else {
		rows, err = db.QueryContext(ctx,
			`SELECT agent_id, SUM(cost) as total FROM cost_entries WHERE timestamp BETWEEN ? AND ? GROUP BY agent_id ORDER BY total DESC`,
			start, end)
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "AGENT_ID\tCOST")
	for rows.Next() {
		var agentID string
		var total float64
		if err := rows.Scan(&agentID, &total); err != nil {
			return err
		}
		fmt.Fprintf(w, "%s\t$%.4f\n", agentID, total)
	}
	w.Flush()
	return rows.Err()
}

func printCostsByTenant(db *sql.DB, start, end time.Time) error {
	ctx := context.Background()
	rows, err := db.QueryContext(ctx,
		`SELECT tenant_id, SUM(cost) as total FROM cost_entries WHERE timestamp BETWEEN ? AND ? GROUP BY tenant_id ORDER BY total DESC`,
		start, end)
	if err != nil {
		return err
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TENANT_ID\tCOST")
	for rows.Next() {
		var tenantID string
		var total float64
		if err := rows.Scan(&tenantID, &total); err != nil {
			return err
		}
		fmt.Fprintf(w, "%s\t$%.4f\n", tenantID, total)
	}
	w.Flush()
	return rows.Err()
}

func printCostsByUser(db *sql.DB, tenant string, start, end time.Time) error {
	ctx := context.Background()
	var rows *sql.Rows
	var err error

	if tenant != "" {
		rows, err = db.QueryContext(ctx,
			`SELECT user_id, SUM(cost) as total FROM cost_entries WHERE tenant_id = ? AND timestamp BETWEEN ? AND ? GROUP BY user_id ORDER BY total DESC`,
			tenant, start, end)
	} else {
		rows, err = db.QueryContext(ctx,
			`SELECT user_id, SUM(cost) as total FROM cost_entries WHERE timestamp BETWEEN ? AND ? GROUP BY user_id ORDER BY total DESC`,
			start, end)
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "USER_ID\tCOST")
	for rows.Next() {
		var userID sql.NullString
		var total float64
		if err := rows.Scan(&userID, &total); err != nil {
			return err
		}
		id := ""
		if userID.Valid {
			id = userID.String
		}
		fmt.Fprintf(w, "%s\t$%.4f\n", id, total)
	}
	w.Flush()
	return rows.Err()
}
