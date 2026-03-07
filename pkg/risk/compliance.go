package risk

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// ComplianceReport represents a compliance audit report.
type ComplianceReport struct {
	ID              string                 `json:"id"`
	ReportType      string                 `json:"report_type"` // daily, weekly, monthly, audit
	PeriodStart     time.Time              `json:"period_start"`
	PeriodEnd       time.Time              `json:"period_end"`
	GeneratedAt     time.Time              `json:"generated_at"`
	TotalActions    int                    `json:"total_actions"`
	HighRiskCount   int                    `json:"high_risk_count"`
	DeniedCount     int                    `json:"denied_count"`
	ApprovalCount   int                    `json:"approval_count"`
	Summary         ComplianceSummary      `json:"summary"`
	Findings        []ComplianceFinding    `json:"findings"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceSummary provides high-level compliance metrics.
type ComplianceSummary struct {
	ComplianceRate   float64            `json:"compliance_rate"` // % of actions that passed
	AverageRiskScore float64            `json:"average_risk_score"`
	PolicyViolations int                `json:"policy_violations"`
	CriticalIssues   int                `json:"critical_issues"`
	ByCategory       map[string]float64 `json:"by_category"`
	ByEnvironment    map[string]int     `json:"by_environment"`
}

// ComplianceFinding represents a compliance issue.
type ComplianceFinding struct {
	Severity    string   `json:"severity"` // low, medium, high, critical
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Count       int      `json:"count"`
	Examples    []string `json:"examples"` // Sample assessment IDs
	Impact      string   `json:"impact"`
	Remediation string   `json:"remediation"`
}

// ReportGenerator generates compliance reports.
type ReportGenerator struct {
	db       *sql.DB
	analyzer *RiskAnalyzer
}

// NewReportGenerator creates a new report generator.
func NewReportGenerator(db *sql.DB) *ReportGenerator {
	return &ReportGenerator{
		db:       db,
		analyzer: NewRiskAnalyzer(db),
	}
}

// GenerateDailyReport creates a daily compliance report.
func (g *ReportGenerator) GenerateDailyReport(ctx context.Context, date time.Time) (*ComplianceReport, error) {
	start := date.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)

	return g.generateReport(ctx, "daily", start, end)
}

// GenerateWeeklyReport creates a weekly compliance report.
func (g *ReportGenerator) GenerateWeeklyReport(ctx context.Context, weekStart time.Time) (*ComplianceReport, error) {
	end := weekStart.Add(7 * 24 * time.Hour)
	return g.generateReport(ctx, "weekly", weekStart, end)
}

// GenerateMonthlyReport creates a monthly compliance report.
func (g *ReportGenerator) GenerateMonthlyReport(ctx context.Context, year int, month time.Month) (*ComplianceReport, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	return g.generateReport(ctx, "monthly", start, end)
}

// GenerateCustomReport creates a custom period report.
func (g *ReportGenerator) GenerateCustomReport(ctx context.Context, start, end time.Time) (*ComplianceReport, error) {
	return g.generateReport(ctx, "custom", start, end)
}

// generateReport creates a compliance report for the given period.
func (g *ReportGenerator) generateReport(ctx context.Context, reportType string, start, end time.Time) (*ComplianceReport, error) {
	report := &ComplianceReport{
		ID:          generateReportID(),
		ReportType:  reportType,
		PeriodStart: start,
		PeriodEnd:   end,
		GeneratedAt: time.Now(),
	}

	// Get aggregated stats
	stats, err := g.analyzer.GetStats(ctx, start, end)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}

	report.TotalActions = stats.TotalAssessments
	report.HighRiskCount = stats.ByLevel[RiskLevelHigh] + stats.ByLevel[RiskLevelCritical]
	report.DeniedCount = stats.ByDecision[DecisionDeny]
	report.ApprovalCount = stats.ByDecision[DecisionRequireReview]

	// Calculate compliance rate
	compliantActions := stats.ByDecision[DecisionAllow] + stats.ByDecision[DecisionAllowWithLog]
	complianceRate := 0.0
	if stats.TotalAssessments > 0 {
		complianceRate = float64(compliantActions) / float64(stats.TotalAssessments)
	}

	// Build summary
	report.Summary = ComplianceSummary{
		ComplianceRate:   complianceRate,
		AverageRiskScore: stats.AverageScore,
		PolicyViolations: stats.ByDecision[DecisionDeny],
		CriticalIssues:   stats.ByLevel[RiskLevelCritical],
		ByCategory:       stats.ByCategory,
		ByEnvironment:    make(map[string]int),
	}

	// Get top risky actions
	riskyActions, err := g.analyzer.GetTopRiskyActions(ctx, 10, 0.7)
	if err == nil && len(riskyActions) > 0 {
		// Analyze findings
		report.Findings = g.analyzeFindings(riskyActions)
		report.Recommendations = g.generateRecommendations(report)
	}

	// Save report to database
	if err := g.saveReport(ctx, report); err != nil {
		return nil, fmt.Errorf("save report: %w", err)
	}

	return report, nil
}

// analyzeFindings analyzes risk assessments for patterns.
func (g *ReportGenerator) analyzeFindings(assessments []RiskAssessment) []ComplianceFinding {
	findings := make([]ComplianceFinding, 0)

	// Group by risk level
	criticalCount := 0
	highCount := 0
	criticalExamples := make([]string, 0)

	for _, assessment := range assessments {
		if assessment.RiskScore.Level == RiskLevelCritical {
			criticalCount++
			if len(criticalExamples) < 3 {
				criticalExamples = append(criticalExamples, fmt.Sprintf("%s: %.2f",
					assessment.ActionContext.ToolName,
					assessment.RiskScore.Score))
			}
		} else if assessment.RiskScore.Level == RiskLevelHigh {
			highCount++
		}
	}

	// Critical risk finding
	if criticalCount > 0 {
		findings = append(findings, ComplianceFinding{
			Severity:    "critical",
			Category:    "high_risk_actions",
			Description: fmt.Sprintf("%d critical risk actions detected", criticalCount),
			Count:       criticalCount,
			Examples:    criticalExamples,
			Impact:      "Potential security or compliance violations",
			Remediation: "Review and update policies to prevent high-risk actions",
		})
	}

	// High risk finding
	if highCount > 0 {
		findings = append(findings, ComplianceFinding{
			Severity:    "high",
			Category:    "elevated_risk",
			Description: fmt.Sprintf("%d high risk actions detected", highCount),
			Count:       highCount,
			Impact:      "Increased exposure to security incidents",
			Remediation: "Implement additional approval gates",
		})
	}

	return findings
}

// generateRecommendations creates actionable recommendations.
func (g *ReportGenerator) generateRecommendations(report *ComplianceReport) []string {
	recommendations := make([]string, 0)

	// Low compliance rate
	if report.Summary.ComplianceRate < 0.8 {
		recommendations = append(recommendations,
			"Compliance rate below 80% - review and update policies")
	}

	// High average risk
	if report.Summary.AverageRiskScore > 0.6 {
		recommendations = append(recommendations,
			"Average risk score elevated - implement stricter controls")
	}

	// Many denials
	if report.DeniedCount > report.TotalActions/10 {
		recommendations = append(recommendations,
			"High denial rate - policies may be too restrictive or agents need guidance")
	}

	// Critical issues
	if report.Summary.CriticalIssues > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("%d critical issues detected - immediate review required", report.Summary.CriticalIssues))
	}

	// Default recommendation
	if len(recommendations) == 0 {
		recommendations = append(recommendations,
			"Compliance within acceptable ranges - continue monitoring")
	}

	return recommendations
}

// saveReport persists a compliance report.
func (g *ReportGenerator) saveReport(ctx context.Context, report *ComplianceReport) error {
	summaryJSON, _ := json.Marshal(report.Summary)
	findingsJSON, _ := json.Marshal(report.Findings)
	recommendationsJSON, _ := json.Marshal(report.Recommendations)
	metadataJSON, _ := json.Marshal(report.Metadata)

	query := `
		INSERT INTO compliance_reports (
			id, report_type, period_start, period_end, generated_at,
			total_actions, high_risk_count, denied_count, approval_count,
			summary, findings, recommendations, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := g.db.ExecContext(ctx, query,
		report.ID,
		report.ReportType,
		report.PeriodStart,
		report.PeriodEnd,
		report.GeneratedAt,
		report.TotalActions,
		report.HighRiskCount,
		report.DeniedCount,
		report.ApprovalCount,
		string(summaryJSON),
		string(findingsJSON),
		string(recommendationsJSON),
		string(metadataJSON),
	)

	return err
}

// GetReport retrieves a compliance report by ID.
func (g *ReportGenerator) GetReport(ctx context.Context, id string) (*ComplianceReport, error) {
	query := `
		SELECT 
			report_type, period_start, period_end, generated_at,
			total_actions, high_risk_count, denied_count, approval_count,
			summary, findings, recommendations, metadata
		FROM compliance_reports
		WHERE id = ?
	`

	var reportType string
	var periodStart, periodEnd, generatedAt time.Time
	var totalActions, highRiskCount, deniedCount, approvalCount int
	var summaryJSON, findingsJSON, recommendationsJSON, metadataJSON string

	err := g.db.QueryRowContext(ctx, query, id).Scan(
		&reportType, &periodStart, &periodEnd, &generatedAt,
		&totalActions, &highRiskCount, &deniedCount, &approvalCount,
		&summaryJSON, &findingsJSON, &recommendationsJSON, &metadataJSON,
	)
	if err != nil {
		return nil, err
	}

	report := &ComplianceReport{
		ID:            id,
		ReportType:    reportType,
		PeriodStart:   periodStart,
		PeriodEnd:     periodEnd,
		GeneratedAt:   generatedAt,
		TotalActions:  totalActions,
		HighRiskCount: highRiskCount,
		DeniedCount:   deniedCount,
		ApprovalCount: approvalCount,
	}

	json.Unmarshal([]byte(summaryJSON), &report.Summary)
	json.Unmarshal([]byte(findingsJSON), &report.Findings)
	json.Unmarshal([]byte(recommendationsJSON), &report.Recommendations)
	json.Unmarshal([]byte(metadataJSON), &report.Metadata)

	return report, nil
}

// ListReports lists compliance reports.
func (g *ReportGenerator) ListReports(ctx context.Context, reportType string, limit int) ([]ComplianceReport, error) {
	query := `
		SELECT 
			id, report_type, period_start, period_end, generated_at,
			total_actions, high_risk_count, denied_count, approval_count
		FROM compliance_reports
	`

	args := make([]interface{}, 0)
	if reportType != "" {
		query += " WHERE report_type = ?"
		args = append(args, reportType)
	}

	query += " ORDER BY generated_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := g.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reports := make([]ComplianceReport, 0, limit)

	for rows.Next() {
		var report ComplianceReport
		err := rows.Scan(
			&report.ID,
			&report.ReportType,
			&report.PeriodStart,
			&report.PeriodEnd,
			&report.GeneratedAt,
			&report.TotalActions,
			&report.HighRiskCount,
			&report.DeniedCount,
			&report.ApprovalCount,
		)
		if err != nil {
			continue
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// generateReportID generates a unique report ID.
func generateReportID() string {
	return fmt.Sprintf("report-%d", time.Now().UnixNano())
}

// ExportFormat defines export formats for reports.
type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatCEF  ExportFormat = "cef"
	ExportFormatPDF  ExportFormat = "pdf"
)

// ExportReport exports a compliance report in the specified format.
func (g *ReportGenerator) ExportReport(ctx context.Context, reportID string, format ExportFormat) ([]byte, error) {
	report, err := g.GetReport(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("get report: %w", err)
	}

	switch format {
	case ExportFormatJSON:
		return json.MarshalIndent(report, "", "  ")

	case ExportFormatCSV:
		return g.exportCSV(report)

	case ExportFormatCEF:
		return g.exportCEF(report)

	case ExportFormatPDF:
		return nil, fmt.Errorf("PDF export not yet implemented")

	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// exportCSV exports report as CSV.
func (g *ReportGenerator) exportCSV(report *ComplianceReport) ([]byte, error) {
	csv := "Report ID,Type,Period Start,Period End,Total Actions,High Risk,Denied,Compliance Rate\n"
	csv += fmt.Sprintf("%s,%s,%s,%s,%d,%d,%d,%.2f%%\n",
		report.ID,
		report.ReportType,
		report.PeriodStart.Format("2006-01-02"),
		report.PeriodEnd.Format("2006-01-02"),
		report.TotalActions,
		report.HighRiskCount,
		report.DeniedCount,
		report.Summary.ComplianceRate*100,
	)

	return []byte(csv), nil
}

// exportCEF exports report in CEF format for SIEM
func (g *ReportGenerator) exportCEF(report *ComplianceReport) ([]byte, error) {
	// CEF: CEF:Version|Device Vendor|Device Product|Device Version|Signature ID|Name|Severity|Extension
	cef := "CEF:0|Unagnt|AgentRuntime|1.0|compliance_report|" + report.ReportType + "|" + fmt.Sprintf("%.0f", report.Summary.ComplianceRate*10) + "|"
	cef += "rt=" + fmt.Sprintf("%d", report.GeneratedAt.UnixMilli()) + " "
	cef += "msg=" + fmt.Sprintf("Compliance report %s: %d actions, %.1f%% compliant", report.ID, report.TotalActions, report.Summary.ComplianceRate*100)
	return []byte(cef), nil
}
