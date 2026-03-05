package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/policy"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

const policyVersioningSchema = `
CREATE TABLE IF NOT EXISTS policy_versions (
    id TEXT PRIMARY KEY,
    policy_name TEXT NOT NULL,
    version TEXT NOT NULL,
    content BLOB NOT NULL,
    format TEXT DEFAULT 'yaml',
    author TEXT,
    changelog TEXT,
    effective_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    supersedes TEXT,
    active BOOLEAN DEFAULT false,
    metadata JSON,
    UNIQUE(policy_name, version)
);
CREATE INDEX IF NOT EXISTS idx_policy_versions_name ON policy_versions(policy_name);
CREATE INDEX IF NOT EXISTS idx_policy_versions_active ON policy_versions(policy_name, active);
`

func newPolicyCmd(db *sql.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage policies and policy versions",
		Long:  "Commands for managing policy versions, testing, and simulation",
	}

	cmd.PersistentFlags().String("db", "agent.db", "Database path for policy store")

	cmd.AddCommand(
		newPolicyApplyCmd(),
		newPolicyListCmd(db),
		newPolicyVersionCmd(db),
		newPolicyActivateCmd(db),
		newPolicyTestCmd(db),
		newPolicySimulateCmd(db),
		newPolicyAuditCmd(db),
	)

	return cmd
}

func getPolicyDB(cmd *cobra.Command, fallback *sql.DB) (*sql.DB, error) {
	if fallback != nil {
		return fallback, nil
	}
	dbPath := "agent.db"
	p := cmd
	for p != nil && p.Name() != "policy" {
		p = p.Parent()
	}
	if p != nil {
		if f := p.PersistentFlags().Lookup("db"); f != nil {
			dbPath = f.Value.String()
		}
	}
	return openPolicyDB(dbPath)
}

func openPolicyDB(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = "agent.db"
	}
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", absPath)
	if err != nil {
		return nil, err
	}
	if _, err := db.ExecContext(context.Background(), policyVersioningSchema); err != nil {
		db.Close()
		return nil, fmt.Errorf("ensure policy schema: %w", err)
	}
	return db, nil
}

func newPolicyApplyCmd() *cobra.Command {
	var dbPath, policyName, version, author, changelog string
	var activate bool

	cmd := &cobra.Command{
		Use:   "apply <file|url>",
		Short: "Apply a policy from file or URL (GitOps)",
		Long:  "Load policy from a local file or URL, validate, and persist to the version store. Use --activate to set as active.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := args[0]
			content, err := loadPolicyContent(source)
			if err != nil {
				return fmt.Errorf("load policy: %w", err)
			}

			if _, err := policy.LoadEngineFromBytes(content); err != nil {
				return fmt.Errorf("validate policy: %w", err)
			}

			db, err := openPolicyDB(dbPath)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer db.Close()

			store, err := policy.NewVersionStore(db, "policies")
			if err != nil {
				return err
			}

			name := policyName
			if name == "" {
				name = derivePolicyName(source)
			}
			ver := version
			if ver == "" {
				ver = "1.0.0"
			}

			pv := &policy.PolicyVersion{
				PolicyName: name,
				Version:    ver,
				Content:    content,
				Format:     "yaml",
				Author:     author,
				Changelog:  changelog,
				Active:     activate,
			}

			ctx := context.Background()
			if err := store.SaveVersion(ctx, pv); err != nil {
				return fmt.Errorf("save version: %w", err)
			}

			fmt.Printf("✓ Policy %s@%s applied", name, ver)
			if activate {
				fmt.Print(" (active)")
			}
			fmt.Println()
			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "agent.db", "Database path for policy version store")
	cmd.Flags().StringVar(&policyName, "policy-name", "", "Policy name (default: derived from source)")
	cmd.Flags().StringVar(&version, "version", "1.0.0", "Policy version")
	cmd.Flags().StringVar(&author, "author", "", "Author (optional)")
	cmd.Flags().StringVar(&changelog, "changelog", "", "Changelog (optional)")
	cmd.Flags().BoolVar(&activate, "activate", true, "Set as active version")
	return cmd
}

func loadPolicyContent(source string) ([]byte, error) {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		resp, err := http.Get(source)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	}
	return os.ReadFile(source)
}

func derivePolicyName(source string) string {
	base := filepath.Base(source)
	if i := strings.Index(base, "."); i > 0 {
		return base[:i]
	}
	return base
}

func newPolicyListCmd(db *sql.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all policies",
		RunE: func(cmd *cobra.Command, args []string) error {
			useDb, err := getPolicyDB(cmd, db)
			if err != nil {
				return err
			}
			if db == nil {
				defer useDb.Close()
			}
			ctx := context.Background()
			store, err := policy.NewVersionStore(useDb, "policies")
			if err != nil {
				return err
			}

			policies, err := store.ListPolicies(ctx)
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "POLICY NAME\tACTIVE VERSION\tLAST UPDATED")

			for _, policyName := range policies {
				active, err := store.GetActiveVersion(ctx, policyName)
				if err != nil {
					fmt.Fprintf(w, "%s\t%s\t%s\n", policyName, "none", "")
					continue
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n",
					policyName,
					active.Version,
					active.CreatedAt.Format("2006-01-02 15:04"))
			}

			w.Flush()
			return nil
		},
	}
}

func newPolicyVersionCmd(db *sql.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "versions <policy-name>",
		Short: "List versions of a policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			useDb, err := getPolicyDB(cmd, db)
			if err != nil {
				return err
			}
			if db == nil {
				defer useDb.Close()
			}
			ctx := context.Background()
			store, err := policy.NewVersionStore(useDb, "policies")
			if err != nil {
				return err
			}

			policyName := args[0]
			versions, err := store.ListVersions(ctx, policyName)
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "VERSION\tAUTHOR\tCREATED\tACTIVE\tCHANGELOG")

			for _, v := range versions {
				active := ""
				if v.Active {
					active = "✓"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					v.Version,
					v.Author,
					v.CreatedAt.Format("2006-01-02 15:04"),
					active,
					truncate(v.Changelog, 50))
			}

			w.Flush()
			return nil
		},
	}

	return cmd
}

func newPolicyActivateCmd(db *sql.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "activate <policy-name> <version>",
		Short: "Activate a specific policy version",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			useDb, err := getPolicyDB(cmd, db)
			if err != nil {
				return err
			}
			if db == nil {
				defer useDb.Close()
			}
			ctx := context.Background()
			store, err := policy.NewVersionStore(useDb, "policies")
			if err != nil {
				return err
			}

			policyName := args[0]
			version := args[1]

			if err := store.SetActiveVersion(ctx, policyName, version); err != nil {
				return err
			}

			fmt.Printf("✓ Activated %s@%s\n", policyName, version)
			return nil
		},
	}
}

func newPolicyTestCmd(db *sql.DB) *cobra.Command {
	var testFile string

	cmd := &cobra.Command{
		Use:   "test <policy-name>",
		Short: "Run policy tests",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			useDb, err := getPolicyDB(cmd, db)
			if err != nil {
				return err
			}
			if db == nil {
				defer useDb.Close()
			}
			ctx := context.Background()
			store, err := policy.NewVersionStore(useDb, "policies")
			if err != nil {
				return err
			}

			simulator := policy.NewSimulator(store, policy.NewExecutor())
			runner := policy.NewTestRunner(simulator, store)

			result, err := runner.RunTestFile(ctx, testFile)
			if err != nil {
				return err
			}

			// Print results
			fmt.Printf("\n%s\n", result.TestFile)
			fmt.Printf("Tests: %d total, %d passed, %d failed, %d skipped\n",
				result.TotalTests, result.Passed, result.Failed, result.Skipped)
			fmt.Printf("Duration: %s\n\n", result.Duration)

			for _, tc := range result.TestCases {
				status := "✓"
				if tc.Status == "failed" {
					status = "✗"
				} else if tc.Status == "skipped" {
					status = "○"
				}

				fmt.Printf("%s %s", status, tc.Name)
				if tc.Message != "" {
					fmt.Printf(" - %s", tc.Message)
				}
				fmt.Println()

				if tc.Status == "failed" {
					if tc.Expected != "" {
						fmt.Printf("    Expected: %s\n", tc.Expected)
					}
					if tc.Actual != "" {
						fmt.Printf("    Actual:   %s\n", tc.Actual)
					}
				}
			}

			if result.Failed > 0 {
				return fmt.Errorf("tests failed")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&testFile, "file", "f", "policy_test.yaml", "Test file path")
	return cmd
}

func newPolicySimulateCmd(db *sql.DB) *cobra.Command {
	var runID string
	var mode string

	cmd := &cobra.Command{
		Use:   "simulate <policy-name> <version>",
		Short: "Simulate policy against historical run",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			useDb, err := getPolicyDB(cmd, db)
			if err != nil {
				return err
			}
			if db == nil {
				defer useDb.Close()
			}
			ctx := context.Background()
			store, err := policy.NewVersionStore(useDb, "policies")
			if err != nil {
				return err
			}

			simulator := policy.NewSimulator(store, policy.NewExecutor())

			// Load actions from run history
			actions, err := loadActionsFromRun(useDb, runID)
			if err != nil {
				return fmt.Errorf("load actions from run: %w", err)
			}

			req := policy.SimulationRequest{
				PolicyName:    args[0],
				PolicyVersion: args[1],
				RunID:         runID,
				Mode:          policy.SimulationMode(mode),
				Actions:       actions,
			}

			result, err := simulator.Simulate(ctx, req)
			if err != nil {
				return err
			}

			// Print results
			fmt.Printf("\nSimulation Results\n")
			fmt.Printf("==================\n")
			fmt.Printf("Policy: %s@%s\n", result.PolicyName, result.PolicyVersion)
			fmt.Printf("Mode: %s\n", result.Mode)
			fmt.Printf("Duration: %s\n\n", result.CompletedAt.Sub(result.StartedAt))

			fmt.Printf("Total Actions: %d\n", result.TotalActions)
			fmt.Printf("Allowed: %d (%.1f%%)\n", result.Allowed, result.Summary.AllowRate*100)
			fmt.Printf("Denied: %d (%.1f%%)\n", result.Denied, result.Summary.DenyRate*100)
			fmt.Printf("Alerts: %d (%.1f%%)\n", result.Alerts, result.Summary.AlertRate*100)
			fmt.Printf("Avg Risk Score: %.2f\n\n", result.Summary.AvgRiskScore)

			if len(result.Summary.TopDenyReasons) > 0 {
				fmt.Println("Top Deny Reasons:")
				for _, reason := range result.Summary.TopDenyReasons {
					fmt.Printf("  - %s: %d\n", reason.Reason, reason.Count)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&runID, "run", "", "Run ID to simulate")
	cmd.Flags().StringVar(&mode, "mode", "simulation", "Simulation mode (audit, simulation, shadow)")
	cmd.MarkFlagRequired("run")

	return cmd
}

func newPolicyAuditCmd(db *sql.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Query policy audit logs",
	}

	cmd.AddCommand(newPolicyAuditQueryCmd(db))
	cmd.AddCommand(newPolicyAuditStatsCmd(db))
	cmd.AddCommand(newPolicyAuditExportCmd(db))

	return cmd
}

func newPolicyAuditQueryCmd(db *sql.DB) *cobra.Command {
	var agent, policyName, decision string
	var minRisk float64
	var limit int

	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query audit logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			useDb, err := getPolicyDB(cmd, db)
			if err != nil {
				return err
			}
			if db == nil {
				defer useDb.Close()
			}
			ctx := context.Background()
			logger := policy.NewAuditLogger(useDb)

			filter := policy.AuditFilter{
				AgentName:    agent,
				PolicyName:   policyName,
				Decision:     decision,
				MinRiskScore: minRisk,
				Limit:        limit,
			}

			logs, err := logger.Query(ctx, filter)
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "TIMESTAMP\tAGENT\tTOOL\tDECISION\tRISK\tREASON")

			for _, log := range logs {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%.2f\t%s\n",
					log.Timestamp.Format("2006-01-02 15:04"),
					log.AgentName,
					log.Tool,
					log.Decision,
					log.RiskScore,
					truncate(log.DenyReason, 40))
			}

			w.Flush()
			fmt.Printf("\nTotal: %d logs\n", len(logs))

			return nil
		},
	}

	cmd.Flags().StringVar(&agent, "agent", "", "Filter by agent name")
	cmd.Flags().StringVar(&policyName, "policy", "", "Filter by policy name")
	cmd.Flags().StringVar(&decision, "decision", "", "Filter by decision (allow, deny, alert)")
	cmd.Flags().Float64Var(&minRisk, "min-risk", 0, "Minimum risk score")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of results")

	return cmd
}

func newPolicyAuditStatsCmd(db *sql.DB) *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show audit statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			useDb, err := getPolicyDB(cmd, db)
			if err != nil {
				return err
			}
			if db == nil {
				defer useDb.Close()
			}
			ctx := context.Background()
			logger := policy.NewAuditLogger(useDb)

			endTime := time.Now()
			startTime := endTime.AddDate(0, 0, -days)

			stats, err := logger.GetStats(ctx, startTime, endTime)
			if err != nil {
				return err
			}

			fmt.Printf("\nAudit Statistics (Last %d days)\n", days)
			fmt.Println("=================================")
			fmt.Printf("Total Decisions: %d\n", stats.Total)
			fmt.Printf("Allowed: %d (%.1f%%)\n", stats.Allowed, float64(stats.Allowed)/float64(stats.Total)*100)
			fmt.Printf("Denied: %d (%.1f%%)\n", stats.Denied, float64(stats.Denied)/float64(stats.Total)*100)
			fmt.Printf("Alerts: %d (%.1f%%)\n", stats.Alerts, float64(stats.Alerts)/float64(stats.Total)*100)
			fmt.Printf("Avg Risk Score: %.2f\n", stats.AvgRiskScore)
			fmt.Printf("Max Risk Score: %.2f\n\n", stats.MaxRiskScore)

			if len(stats.TopDenyReasons) > 0 {
				fmt.Println("Top Deny Reasons:")
				for i, reason := range stats.TopDenyReasons {
					fmt.Printf("%d. %s: %d times\n", i+1, reason.Reason, reason.Count)
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&days, "days", 7, "Number of days to analyze")
	return cmd
}

func newPolicyAuditExportCmd(db *sql.DB) *cobra.Command {
	var format, output string
	var days int

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export audit logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			useDb, err := getPolicyDB(cmd, db)
			if err != nil {
				return err
			}
			if db == nil {
				defer useDb.Close()
			}
			ctx := context.Background()
			logger := policy.NewAuditLogger(useDb)

			endTime := time.Now()
			startTime := endTime.AddDate(0, 0, -days)

			filter := policy.AuditFilter{
				StartTime: startTime,
				EndTime:   endTime,
			}

			data, err := logger.Export(ctx, filter, format)
			if err != nil {
				return err
			}

			if output == "" {
				fmt.Println(string(data))
			} else {
				if err := os.WriteFile(output, data, 0644); err != nil {
					return err
				}
				fmt.Printf("✓ Exported to %s\n", output)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "Export format (json, csv)")
	cmd.Flags().StringVar(&output, "output", "", "Output file (default: stdout)")
	cmd.Flags().IntVar(&days, "days", 30, "Number of days to export")

	return cmd
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// loadActionsFromRun loads all actions from a historical run for policy simulation.
func loadActionsFromRun(db *sql.DB, runID string) ([]policy.ActionToSimulate, error) {
	// Query actions/events from the run history
	query := `
		SELECT tool_name, permissions, timestamp, context
		FROM run_events
		WHERE run_id = ?
		AND event_type = 'tool_execution'
		ORDER BY timestamp
	`

	rows, err := db.Query(query, runID)
	if err != nil {
		return nil, fmt.Errorf("query run events: %w", err)
	}
	defer rows.Close()

	actions := make([]policy.ActionToSimulate, 0)
	for rows.Next() {
		var toolName, permissions, contextJSON string
		var timestamp time.Time

		if err := rows.Scan(&toolName, &permissions, &timestamp, &contextJSON); err != nil {
			continue // Skip invalid rows
		}

		// Parse permissions
		var perms []string
		if permissions != "" {
			json.Unmarshal([]byte(permissions), &perms)
		}

		// Parse context
		var context map[string]interface{}
		if contextJSON != "" {
			json.Unmarshal([]byte(contextJSON), &context)
		}

		actions = append(actions, policy.ActionToSimulate{
			Sequence: len(actions) + 1,
			Tool:     toolName,
			Input:    context,
			Context:  context,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return actions, nil
}
