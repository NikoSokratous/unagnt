package policy

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AuditLogger logs all policy decisions for compliance.
type AuditLogger struct {
	db *sql.DB
}

// NewAuditLogger creates a new policy audit logger.
func NewAuditLogger(db *sql.DB) *AuditLogger {
	return &AuditLogger{db: db}
}

// AuditLog represents a policy audit log entry.
type AuditLog struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	RunID         string                 `json:"run_id,omitempty"`
	AgentName     string                 `json:"agent_name"`
	PolicyName    string                 `json:"policy_name"`
	PolicyVersion string                 `json:"policy_version"`
	Action        string                 `json:"action"`
	Tool          string                 `json:"tool"`
	Decision      string                 `json:"decision"` // "allow", "deny", "alert"
	RiskScore     float64                `json:"risk_score"`
	DenyReason    string                 `json:"deny_reason,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	ReviewedBy    string                 `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time             `json:"reviewed_at,omitempty"`
}

// Log logs a policy decision.
func (a *AuditLogger) Log(ctx context.Context, log *AuditLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	contextJSON, err := json.Marshal(log.Context)
	if err != nil {
		return fmt.Errorf("marshal context: %w", err)
	}

	query := `
		INSERT INTO policy_audit (
			id, timestamp, run_id, agent_name, policy_name, policy_version,
			action, tool, decision, risk_score, deny_reason, context,
			reviewed_by, reviewed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = a.db.ExecContext(ctx, query,
		log.ID,
		log.Timestamp,
		log.RunID,
		log.AgentName,
		log.PolicyName,
		log.PolicyVersion,
		log.Action,
		log.Tool,
		log.Decision,
		log.RiskScore,
		log.DenyReason,
		contextJSON,
		log.ReviewedBy,
		log.ReviewedAt,
	)

	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	return nil
}

// Query retrieves audit logs matching filters.
func (a *AuditLogger) Query(ctx context.Context, filter AuditFilter) ([]AuditLog, error) {
	query := `
		SELECT id, timestamp, run_id, agent_name, policy_name, policy_version,
		       action, tool, decision, risk_score, deny_reason, context,
		       reviewed_by, reviewed_at
		FROM policy_audit
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.AgentName != "" {
		query += " AND agent_name = ?"
		args = append(args, filter.AgentName)
	}
	if filter.PolicyName != "" {
		query += " AND policy_name = ?"
		args = append(args, filter.PolicyName)
	}
	if filter.Decision != "" {
		query += " AND decision = ?"
		args = append(args, filter.Decision)
	}
	if filter.MinRiskScore > 0 {
		query += " AND risk_score >= ?"
		args = append(args, filter.MinRiskScore)
	}
	if !filter.StartTime.IsZero() {
		query += " AND timestamp >= ?"
		args = append(args, filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		query += " AND timestamp <= ?"
		args = append(args, filter.EndTime)
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		var contextJSON []byte
		var runID, denyReason, reviewedBy sql.NullString
		var reviewedAt sql.NullTime

		err := rows.Scan(
			&log.ID,
			&log.Timestamp,
			&runID,
			&log.AgentName,
			&log.PolicyName,
			&log.PolicyVersion,
			&log.Action,
			&log.Tool,
			&log.Decision,
			&log.RiskScore,
			&denyReason,
			&contextJSON,
			&reviewedBy,
			&reviewedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}

		if runID.Valid {
			log.RunID = runID.String
		}
		if denyReason.Valid {
			log.DenyReason = denyReason.String
		}
		if reviewedBy.Valid {
			log.ReviewedBy = reviewedBy.String
		}
		if reviewedAt.Valid {
			log.ReviewedAt = &reviewedAt.Time
		}

		if len(contextJSON) > 0 {
			if err := json.Unmarshal(contextJSON, &log.Context); err != nil {
				return nil, fmt.Errorf("unmarshal context: %w", err)
			}
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// AuditFilter filters audit log queries.
type AuditFilter struct {
	AgentName    string
	PolicyName   string
	Decision     string
	MinRiskScore float64
	StartTime    time.Time
	EndTime      time.Time
	Limit        int
}

// GetStats returns audit statistics for a time range.
func (a *AuditLogger) GetStats(ctx context.Context, startTime, endTime time.Time) (*AuditStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN decision = 'allow' THEN 1 ELSE 0 END) as allowed,
			SUM(CASE WHEN decision = 'deny' THEN 1 ELSE 0 END) as denied,
			SUM(CASE WHEN decision = 'alert' THEN 1 ELSE 0 END) as alerts,
			AVG(risk_score) as avg_risk_score,
			MAX(risk_score) as max_risk_score
		FROM policy_audit
		WHERE timestamp >= ? AND timestamp <= ?
	`

	var stats AuditStats
	err := a.db.QueryRowContext(ctx, query, startTime, endTime).Scan(
		&stats.Total,
		&stats.Allowed,
		&stats.Denied,
		&stats.Alerts,
		&stats.AvgRiskScore,
		&stats.MaxRiskScore,
	)
	if err != nil {
		return nil, fmt.Errorf("query audit stats: %w", err)
	}

	// Get top deny reasons
	reasonQuery := `
		SELECT deny_reason, COUNT(*) as count
		FROM policy_audit
		WHERE decision = 'deny' 
		  AND timestamp >= ? 
		  AND timestamp <= ?
		  AND deny_reason IS NOT NULL
		GROUP BY deny_reason
		ORDER BY count DESC
		LIMIT 10
	`

	rows, err := a.db.QueryContext(ctx, reasonQuery, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("query deny reasons: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var reason string
		var count int
		if err := rows.Scan(&reason, &count); err != nil {
			return nil, fmt.Errorf("scan deny reason: %w", err)
		}
		stats.TopDenyReasons = append(stats.TopDenyReasons, DenyReasonCount{
			Reason: reason,
			Count:  count,
		})
	}

	return &stats, rows.Err()
}

// AuditStats contains audit statistics.
type AuditStats struct {
	Total          int               `json:"total"`
	Allowed        int               `json:"allowed"`
	Denied         int               `json:"denied"`
	Alerts         int               `json:"alerts"`
	AvgRiskScore   float64           `json:"avg_risk_score"`
	MaxRiskScore   float64           `json:"max_risk_score"`
	TopDenyReasons []DenyReasonCount `json:"top_deny_reasons"`
}

// MarkReviewed marks an audit log as reviewed by a human.
func (a *AuditLogger) MarkReviewed(ctx context.Context, logID, reviewedBy string) error {
	now := time.Now()
	_, err := a.db.ExecContext(ctx,
		"UPDATE policy_audit SET reviewed_by = ?, reviewed_at = ? WHERE id = ?",
		reviewedBy, now, logID)
	if err != nil {
		return fmt.Errorf("mark reviewed: %w", err)
	}
	return nil
}

// Export exports audit logs to a format for compliance reporting and SIEM.
// Supported formats: json, csv, cef (Common Event Format for SIEM).
func (a *AuditLogger) Export(ctx context.Context, filter AuditFilter, format string) ([]byte, error) {
	logs, err := a.Query(ctx, filter)
	if err != nil {
		return nil, err
	}

	switch format {
	case "json":
		return json.MarshalIndent(logs, "", "  ")
	case "csv":
		return a.exportCSV(logs)
	case "cef":
		return a.exportCEF(logs)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// exportCEF exports logs in Common Event Format for SIEM (e.g. ArcSight, Splunk).
// CEF:Version|Device Vendor|Device Product|Device Version|Signature ID|Name|Severity|Extension
func (a *AuditLogger) exportCEF(logs []AuditLog) ([]byte, error) {
	const (
		cefVersion    = "0"
		deviceVendor  = "AgentRuntime"
		deviceProduct = "PolicyEngine"
		deviceVersion = "1.0"
	)
	var out []byte
	for _, log := range logs {
		signatureID := "policy-" + log.Decision
		name := "Policy " + log.Decision
		severity := "3" // 1-10, 3 = low, 7 = medium, 10 = critical
		if log.Decision == "deny" {
			severity = "7"
		} else if log.Decision == "alert" {
			severity = "6"
		}
		ext := fmt.Sprintf("rt=%d msg=%s agent=%s policy=%s tool=%s action=%s riskScore=%.2f",
			log.Timestamp.UnixMilli(),
			escapeCEF(log.DenyReason),
			escapeCEF(log.AgentName),
			escapeCEF(log.PolicyName),
			escapeCEF(log.Tool),
			escapeCEF(log.Action),
			log.RiskScore,
		)
		if log.RunID != "" {
			ext += " runId=" + escapeCEF(log.RunID)
		}
		line := fmt.Sprintf("CEF:%s|%s|%s|%s|%s|%s|%s|%s\n",
			cefVersion, deviceVendor, deviceProduct, deviceVersion,
			signatureID, name, severity, ext)
		out = append(out, []byte(line)...)
	}
	return out, nil
}

func escapeCEF(s string) string {
	if s == "" {
		return ""
	}
	// CEF extension values: backslash and equals must be escaped
	r := strings.NewReplacer(`\`, `\\`, `=`, `\=`, "\n", `\n`, "\r", `\r`)
	return r.Replace(s)
}

// exportCSV exports logs as CSV.
func (a *AuditLogger) exportCSV(logs []AuditLog) ([]byte, error) {
	csv := "ID,Timestamp,Agent,Policy,Version,Action,Tool,Decision,RiskScore,DenyReason\n"
	for _, log := range logs {
		csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%.2f,%s\n",
			log.ID,
			log.Timestamp.Format(time.RFC3339),
			log.AgentName,
			log.PolicyName,
			log.PolicyVersion,
			log.Action,
			log.Tool,
			log.Decision,
			log.RiskScore,
			log.DenyReason,
		)
	}
	return []byte(csv), nil
}
