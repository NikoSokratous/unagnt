package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/risk"
	"github.com/spf13/cobra"
)

func newRiskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "risk",
		Short: "Risk scoring and analytics",
		Long:  "Commands for risk assessment, analytics, and compliance reporting",
	}

	cmd.AddCommand(
		newRiskAssessCmd(),
		newRiskStatsCmd(),
		newRiskReportCmd(),
		newRiskTopCmd(),
		newRiskTrendCmd(),
	)

	return cmd
}

func newRiskAssessCmd() *cobra.Command {
	var toolName string
	var environment string
	var permissions []string

	cmd := &cobra.Command{
		Use:   "assess",
		Short: "Assess risk for an action",
		Long:  "Calculate risk score for a proposed action",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create risk engine with default config
			config := risk.DefaultRiskConfig()
			engine := risk.NewRiskEngine(config)

			// Create action context
			actionCtx := risk.ActionContext{
				ToolName:      toolName,
				Permissions:   permissions,
				Environment:   environment,
				RecentActions: 5, // Mock value
			}

			// Assess risk
			assessment, err := engine.Assess(context.Background(), actionCtx)
			if err != nil {
				return fmt.Errorf("assess risk: %w", err)
			}

			// Print results
			printRiskAssessment(assessment)

			return nil
		},
	}

	cmd.Flags().StringVar(&toolName, "tool", "", "Tool name (required)")
	cmd.Flags().StringVar(&environment, "env", "dev", "Environment (dev, staging, production)")
	cmd.Flags().StringSliceVar(&permissions, "permission", []string{}, "Required permissions")
	cmd.MarkFlagRequired("tool")

	return cmd
}

func newRiskStatsCmd() *cobra.Command {
	var days int
	var environment string

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "View risk statistics",
		Long:  "Display aggregated risk statistics for a time period",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Risk Statistics (Last %d days)\n\n", days)

			// Mock stats
			fmt.Println("=== Overview ===")
			fmt.Println("Total Assessments: 1,234")
			fmt.Println("Average Risk Score: 0.42 (medium)")
			fmt.Println("Compliance Rate: 92.5%")
			fmt.Println()

			fmt.Println("=== By Risk Level ===")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "LEVEL\tCOUNT\tPERCENTAGE")
			fmt.Fprintln(w, "Low\t850\t68.9%")
			fmt.Fprintln(w, "Medium\t290\t23.5%")
			fmt.Fprintln(w, "High\t82\t6.6%")
			fmt.Fprintln(w, "Critical\t12\t1.0%")
			w.Flush()
			fmt.Println()

			fmt.Println("=== By Category ===")
			w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "CATEGORY\tAVG SCORE")
			fmt.Fprintln(w, "Security\t0.45")
			fmt.Fprintln(w, "Privacy\t0.38")
			fmt.Fprintln(w, "Cost\t0.25")
			fmt.Fprintln(w, "Impact\t0.52")
			fmt.Fprintln(w, "Compliance\t0.35")
			w.Flush()
			fmt.Println()

			fmt.Println("=== By Decision ===")
			w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "DECISION\tCOUNT")
			fmt.Fprintln(w, "Allow\t950")
			fmt.Fprintln(w, "Allow with Log\t192")
			fmt.Fprintln(w, "Require Review\t74")
			fmt.Fprintln(w, "Deny\t18")
			w.Flush()

			return nil
		},
	}

	cmd.Flags().IntVar(&days, "days", 7, "Number of days to analyze")
	cmd.Flags().StringVar(&environment, "env", "", "Filter by environment")

	return cmd
}

func newRiskReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Compliance reporting",
	}

	cmd.AddCommand(
		newRiskReportGenerateCmd(),
		newRiskReportListCmd(),
		newRiskReportViewCmd(),
		newRiskReportExportCmd(),
	)

	return cmd
}

func newRiskReportGenerateCmd() *cobra.Command {
	var reportType string
	var date string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate compliance report",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Generating %s compliance report...\n\n", reportType)

			// Mock report generation
			fmt.Println("✓ Collecting risk assessments")
			fmt.Println("✓ Analyzing patterns")
			fmt.Println("✓ Calculating metrics")
			fmt.Println("✓ Generating findings")
			fmt.Println()

			reportID := fmt.Sprintf("report-%d", time.Now().Unix())
			fmt.Printf("Report generated: %s\n", reportID)

			return nil
		},
	}

	cmd.Flags().StringVar(&reportType, "type", "daily", "Report type (daily, weekly, monthly)")
	cmd.Flags().StringVar(&date, "date", "", "Date for report (YYYY-MM-DD)")

	return cmd
}

func newRiskReportListCmd() *cobra.Command {
	var reportType string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List compliance reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "REPORT ID\tTYPE\tPERIOD\tACTIONS\tHIGH RISK\tDENIED\tCOMPLIANCE")

			// Mock reports
			reports := []struct {
				id         string
				typ        string
				period     string
				actions    int
				highRisk   int
				denied     int
				compliance float64
			}{
				{"report-001", "daily", "2026-02-26", 450, 23, 8, 94.2},
				{"report-002", "daily", "2026-02-25", 523, 31, 12, 92.8},
				{"report-003", "weekly", "2026-W08", 3241, 187, 56, 93.5},
			}

			for _, r := range reports {
				if reportType != "" && r.typ != reportType {
					continue
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%d\t%.1f%%\n",
					r.id, r.typ, r.period, r.actions, r.highRisk, r.denied, r.compliance)
			}

			w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVar(&reportType, "type", "", "Filter by report type")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum reports to list")

	return cmd
}

func newRiskReportViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view <report-id>",
		Short: "View compliance report details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reportID := args[0]

			fmt.Printf("=== Compliance Report: %s ===\n\n", reportID)

			// Mock report
			fmt.Println("Type: Daily")
			fmt.Println("Period: 2026-02-26")
			fmt.Println("Generated: 2026-02-27 00:00:00")
			fmt.Println()

			fmt.Println("=== Summary ===")
			fmt.Println("Total Actions: 450")
			fmt.Println("High Risk: 23 (5.1%)")
			fmt.Println("Denied: 8 (1.8%)")
			fmt.Println("Required Approval: 15 (3.3%)")
			fmt.Println("Compliance Rate: 94.2%")
			fmt.Println("Average Risk Score: 0.38 (low-medium)")
			fmt.Println()

			fmt.Println("=== Findings ===")
			fmt.Println("1. [HIGH] 12 critical risk actions detected")
			fmt.Println("   Examples: delete_file (0.89), execute_command (0.92)")
			fmt.Println("   Remediation: Review and update policies")
			fmt.Println()
			fmt.Println("2. [MEDIUM] Elevated risk in production environment")
			fmt.Println("   Impact: 45% of high-risk actions in production")
			fmt.Println("   Remediation: Implement additional approval gates")
			fmt.Println()

			fmt.Println("=== Recommendations ===")
			fmt.Println("• Compliance rate within acceptable range")
			fmt.Println("• Monitor critical risk actions closely")
			fmt.Println("• Consider stricter controls for production")

			return nil
		},
	}
}

func newRiskReportExportCmd() *cobra.Command {
	var format string
	var output string

	cmd := &cobra.Command{
		Use:   "export <report-id>",
		Short: "Export compliance report",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reportID := args[0]

			fmt.Printf("Exporting report %s as %s...\n", reportID, format)

			if output == "" {
				output = fmt.Sprintf("%s.%s", reportID, format)
			}

			// Mock export
			fmt.Printf("Exported to: %s\n", output)

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "Export format (json, csv, pdf)")
	cmd.Flags().StringVar(&output, "output", "", "Output file path")

	return cmd
}

func newRiskTopCmd() *cobra.Command {
	var limit int
	var minScore float64

	cmd := &cobra.Command{
		Use:   "top",
		Short: "Show top risky actions",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Top %d Risky Actions (score >= %.2f)\n\n", limit, minScore)

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "RANK\tTOOL\tSCORE\tLEVEL\tDECISION\tTIMESTAMP")

			// Mock top risky actions
			actions := []struct {
				tool     string
				score    float64
				level    string
				decision string
				time     string
			}{
				{"execute_command", 0.92, "critical", "deny", "2026-02-26 14:23"},
				{"delete_file", 0.89, "critical", "require_review", "2026-02-26 13:15"},
				{"http_request", 0.78, "high", "require_review", "2026-02-26 12:45"},
				{"write_file", 0.72, "high", "allow_with_log", "2026-02-26 11:30"},
				{"read_file", 0.68, "high", "allow_with_log", "2026-02-26 10:20"},
			}

			for i, action := range actions {
				if i >= limit {
					break
				}
				if action.score < minScore {
					continue
				}

				fmt.Fprintf(w, "%d\t%s\t%.2f\t%s\t%s\t%s\n",
					i+1, action.tool, action.score, action.level, action.decision, action.time)
			}

			w.Flush()
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "Number of actions to show")
	cmd.Flags().Float64Var(&minScore, "min-score", 0.6, "Minimum risk score")

	return cmd
}

func newRiskTrendCmd() *cobra.Command {
	var days int
	var interval string

	cmd := &cobra.Command{
		Use:   "trend",
		Short: "Show risk score trends",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Risk Score Trend (Last %d days, by %s)\n\n", days, interval)

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "PERIOD\tAVG SCORE\tCOUNT\tTREND")

			// Mock trend data
			trends := []struct {
				period string
				score  float64
				count  int
				trend  string
			}{
				{"2026-02-26", 0.38, 450, "↓"},
				{"2026-02-25", 0.42, 523, "↑"},
				{"2026-02-24", 0.35, 412, "↓"},
				{"2026-02-23", 0.39, 389, "↑"},
				{"2026-02-22", 0.36, 445, "→"},
			}

			for _, t := range trends {
				fmt.Fprintf(w, "%s\t%.2f\t%d\t%s\n",
					t.period, t.score, t.count, t.trend)
			}

			w.Flush()
			return nil
		},
	}

	cmd.Flags().IntVar(&days, "days", 7, "Number of days")
	cmd.Flags().StringVar(&interval, "interval", "day", "Interval (hour, day, week)")

	return cmd
}

func printRiskAssessment(assessment *risk.RiskAssessment) {
	fmt.Printf("\n=== Risk Assessment ===\n")
	fmt.Printf("Tool: %s\n", assessment.ActionContext.ToolName)
	fmt.Printf("Environment: %s\n", assessment.ActionContext.Environment)
	fmt.Printf("Permissions: %v\n\n", assessment.ActionContext.Permissions)

	score := assessment.RiskScore
	fmt.Printf("Risk Score: %.2f (%s)\n", score.Score, score.Level)
	fmt.Printf("Confidence: %.1f%%\n", score.Confidence*100)
	fmt.Printf("Decision: %s\n\n", assessment.Decision)

	// Print factors
	fmt.Println("=== Risk Factors ===")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CATEGORY\tSCORE\tWEIGHT\tCONTRIBUTION")

	for _, factor := range score.Factors {
		fmt.Fprintf(w, "%s\t%.2f\t%.2f\t%.2f\n",
			factor.Category,
			factor.Score,
			factor.Weight,
			factor.Contribution)
	}

	w.Flush()
	fmt.Println()

	// Print breakdown
	fmt.Println("=== Score Breakdown ===")
	for category, catScore := range score.Breakdown {
		bar := createBar(catScore, 20)
		fmt.Printf("%-15s [%s] %.2f\n", category, bar, catScore)
	}

	// Decision guidance
	fmt.Println()
	switch assessment.Decision {
	case risk.DecisionAllow:
		fmt.Println("✓ Action allowed - low risk")
	case risk.DecisionAllowWithLog:
		fmt.Println("⚠ Action allowed but logged - medium risk")
	case risk.DecisionRequireReview:
		fmt.Println("⏸ Action requires approval - high risk")
	case risk.DecisionDeny:
		fmt.Println("✗ Action denied - critical risk")
	}
}

func createBar(value float64, width int) string {
	filled := int(value * float64(width))
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}
