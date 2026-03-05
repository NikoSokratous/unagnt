# Risk Scoring Guide

Real-time risk assessment for agent actions with automated decision-making.

## Table of Contents

1. [Overview](#overview)
2. [Risk Categories](#risk-categories)
3. [Scoring Algorithm](#scoring-algorithm)
4. [Decision Making](#decision-making)
5. [CLI Commands](#cli-commands)
6. [Configuration](#configuration)
7. [Integration](#integration)
8. [Analytics](#analytics)
9. [Compliance Reporting](#compliance-reporting)
10. [Best Practices](#best-practices)

## Overview

The risk scoring engine provides:

- **Real-time Assessment**: Evaluate risk before action execution
- **Multi-Category Analysis**: Security, privacy, cost, impact, compliance, reversibility
- **Automated Decisions**: Allow, log, require approval, or deny based on risk
- **Analytics**: Aggregate statistics, trends, and reports
- **Compliance**: Generate audit-ready compliance reports

## Risk Categories

### 1. Security Risk

Evaluates permission requirements and tool danger level.

**Factors:**
- Tool type (execute_command: 0.95, delete_file: 0.8)
- Permissions (exec: +0.3, fs:delete: +0.25)
- Environment (production: +0.3)

**Weight:** 1.5 (highest)

### 2. Privacy Risk

Assesses PII and sensitive data exposure.

**Factors:**
- PII-related tools (database: 0.6, read_file: 0.4)
- Input contains PII patterns (+0.3)
- Data access scope

**Weight:** 1.2

### 3. Cost Risk

Estimates resource and API costs.

**Factors:**
- High-cost tools (model_call: 0.4, http_request: 0.3)
- Action frequency (>10 recent: +0.2, >50: +0.3)
- Resource consumption

**Weight:** 0.3

### 4. Reversibility Risk

Measures if action can be undone.

**Factors:**
- Irreversible tools (delete_file: 1.0, send_email: 1.0)
- External API calls (0.7)
- State mutations

**Weight:** 1.0

### 5. Impact Risk

Assesses blast radius and downstream effects.

**Factors:**
- Environment (production: 0.9, staging: 0.5, dev: 0.1)
- Tool scope (execute_command: +0.5, database: +0.4)
- Affected systems

**Weight:** 1.3

### 6. Compliance Risk

Evaluates regulatory and policy requirements.

**Factors:**
- Regulated operations (database: 0.5, http_request: 0.4)
- Production compliance checks (+0.2)
- Data transfer operations

**Weight:** 1.1

### 7. Reliability Risk

Measures failure likelihood and stability.

**Factors:**
- External dependencies
- Historical failure rates
- Retry patterns

**Weight:** 0.4 (lowest)

## Scoring Algorithm

### Calculation

```go
For each category:
  categoryScore = assessCategory(context)
  contribution = categoryScore * categoryWeight

finalScore = sum(contributions) / sum(weights)
```

### Example

Action: `delete_file` in production with `fs:delete` permission

```
Security:      0.80 * 1.5 = 1.20
Privacy:       0.40 * 1.2 = 0.48
Cost:          0.20 * 0.3 = 0.06
Reversibility: 1.00 * 1.0 = 1.00
Impact:        0.90 * 1.3 = 1.17
Compliance:    0.40 * 1.1 = 0.44
Reliability:   0.30 * 0.4 = 0.12

Total: 4.47 / 6.8 = 0.66 (HIGH)
```

### Risk Levels

| Score Range | Level | Description |
|-------------|-------|-------------|
| 0.0 - 0.3 | Low | Safe operations |
| 0.3 - 0.6 | Medium | Moderate risk |
| 0.6 - 0.85 | High | Risky operations |
| 0.85 - 1.0 | Critical | Dangerous operations |

## Decision Making

### Decision Types

| Decision | Score Range | Action |
|----------|-------------|--------|
| `allow` | < 0.3 | Proceed without logging |
| `allow_with_log` | 0.3 - 0.8 | Proceed but audit |
| `require_review` | 0.8 - 0.95 | Human approval needed |
| `deny` | ≥ 0.95 | Block action |

### Thresholds

Configure in `RiskConfig`:

```yaml
risk:
  threshold_low: 0.3        # Low → Medium
  threshold_medium: 0.6     # Medium → High
  threshold_high: 0.85      # High → Critical
  require_approval: 0.8     # Require human review
  auto_deny: 0.95           # Automatic denial
```

## CLI Commands

### Assess Action Risk

```bash
unagnt risk assess \
  --tool delete_file \
  --env production \
  --permission fs:delete
```

Output:
```
=== Risk Assessment ===
Tool: delete_file
Environment: production
Permissions: [fs:delete]

Risk Score: 0.66 (high)
Confidence: 90.0%
Decision: allow_with_log

=== Risk Factors ===
CATEGORY      SCORE  WEIGHT  CONTRIBUTION
security      0.80   1.50    1.20
privacy       0.40   1.20    0.48
cost          0.20   0.30    0.06
reversible    1.00   1.00    1.00
impact        0.90   1.30    1.17
compliance    0.40   1.10    0.44
reliability   0.30   0.40    0.12

=== Score Breakdown ===
security        [████████████████░░░░] 0.80
privacy         [████████░░░░░░░░░░░░] 0.40
cost            [████░░░░░░░░░░░░░░░░] 0.20
reversible      [████████████████████] 1.00
impact          [██████████████████░░] 0.90
compliance      [████████░░░░░░░░░░░░] 0.40
reliability     [██████░░░░░░░░░░░░░░] 0.30

⚠ Action allowed but logged - medium risk
```

### View Statistics

```bash
unagnt risk stats --days 7
```

Output:
```
Risk Statistics (Last 7 days)

=== Overview ===
Total Assessments: 1,234
Average Risk Score: 0.42 (medium)
Compliance Rate: 92.5%

=== By Risk Level ===
LEVEL     COUNT  PERCENTAGE
Low       850    68.9%
Medium    290    23.5%
High      82     6.6%
Critical  12     1.0%

=== By Category ===
CATEGORY      AVG SCORE
Security      0.45
Privacy       0.38
Cost          0.25
Impact        0.52
Compliance    0.35

=== By Decision ===
DECISION          COUNT
Allow             950
Allow with Log    192
Require Review    74
Deny              18
```

### Top Risky Actions

```bash
unagnt risk top --limit 5 --min-score 0.7
```

Output:
```
Top 5 Risky Actions (score >= 0.70)

RANK  TOOL              SCORE  LEVEL     DECISION         TIMESTAMP
1     execute_command   0.92   critical  deny             2026-02-26 14:23
2     delete_file       0.89   critical  require_review   2026-02-26 13:15
3     http_request      0.78   high      require_review   2026-02-26 12:45
4     write_file        0.72   high      allow_with_log   2026-02-26 11:30
```

### Risk Trends

```bash
unagnt risk trend --days 7 --interval day
```

Output:
```
Risk Score Trend (Last 7 days, by day)

PERIOD       AVG SCORE  COUNT  TREND
2026-02-26   0.38       450    ↓
2026-02-25   0.42       523    ↑
2026-02-24   0.35       412    ↓
2026-02-23   0.39       389    ↑
2026-02-22   0.36       445    →
```

### Generate Compliance Report

```bash
unagnt risk report generate --type daily --date 2026-02-26
```

Output:
```
Generating daily compliance report...

✓ Collecting risk assessments
✓ Analyzing patterns
✓ Calculating metrics
✓ Generating findings

Report generated: report-1709049600
```

### List Reports

```bash
unagnt risk report list --type daily --limit 10
```

Output:
```
REPORT ID    TYPE    PERIOD      ACTIONS  HIGH RISK  DENIED  COMPLIANCE
report-001   daily   2026-02-26  450      23         8       94.2%
report-002   daily   2026-02-25  523      31         12      92.8%
report-003   weekly  2026-W08    3241     187        56      93.5%
```

### View Report

```bash
unagnt risk report view report-001
```

Output:
```
=== Compliance Report: report-001 ===

Type: Daily
Period: 2026-02-26
Generated: 2026-02-27 00:00:00

=== Summary ===
Total Actions: 450
High Risk: 23 (5.1%)
Denied: 8 (1.8%)
Required Approval: 15 (3.3%)
Compliance Rate: 94.2%
Average Risk Score: 0.38 (low-medium)

=== Findings ===
1. [HIGH] 12 critical risk actions detected
   Examples: delete_file (0.89), execute_command (0.92)
   Remediation: Review and update policies

2. [MEDIUM] Elevated risk in production environment
   Impact: 45% of high-risk actions in production
   Remediation: Implement additional approval gates

=== Recommendations ===
• Compliance rate within acceptable range
• Monitor critical risk actions closely
• Consider stricter controls for production
```

### Export Report

```bash
unagnt risk report export report-001 --format json --output report.json
unagnt risk report export report-001 --format csv --output report.csv
```

## Configuration

### Agent Configuration

```yaml
# agent.yaml
risk_scoring:
  enabled: true
  require_approval: 0.8
  auto_deny: 0.95
  
  # Category weights (higher = more important)
  weights:
    security: 1.5
    privacy: 1.2
    impact: 1.3
    compliance: 1.1
    reversible: 1.0
    reliability: 0.4
    cost: 0.3
  
  # Thresholds
  thresholds:
    low: 0.3
    medium: 0.6
    high: 0.85
```

### Programmatic Usage

```go
import "github.com/Unagnt/Unagnt/pkg/risk"

// Create engine
config := risk.DefaultRiskConfig()
engine := risk.NewRiskEngine(config)

// Assess action
actionCtx := risk.ActionContext{
    ToolName:    "delete_file",
    Environment: "production",
    Permissions: []string{"fs:delete"},
}

assessment, err := engine.Assess(ctx, actionCtx)

fmt.Printf("Risk: %.2f (%s)\n", 
    assessment.RiskScore.Score,
    assessment.RiskScore.Level)

switch assessment.Decision {
case risk.DecisionAllow:
    // Proceed
case risk.DecisionRequireReview:
    // Get approval
case risk.DecisionDeny:
    // Block
}
```

## Integration

### With Policy Engine

Risk scores can trigger policy rules:

```yaml
# policy.yaml
rules:
  - name: high-risk-approval
    match:
      risk_score: ">= 0.8"
    effect: require_approval
    
  - name: critical-risk-deny
    match:
      risk_score: ">= 0.95"
    effect: deny
```

### With Agent Runtime

```go
// In executor
assessment, _ := riskEngine.Assess(ctx, actionContext)

if assessment.Decision == risk.DecisionDeny {
    return fmt.Errorf("action denied: risk too high (%.2f)", 
        assessment.RiskScore.Score)
}

if assessment.Decision == risk.DecisionRequireReview {
    // Pause for approval
    approval := waitForApproval(assessment)
    if !approval.Approved {
        return fmt.Errorf("action not approved")
    }
}

// Proceed with action
```

### With Audit Logger

```go
// Log risk assessment
auditLog := AuditLog{
    ActionName: actionCtx.ToolName,
    Decision:   string(assessment.Decision),
    RiskScore:  assessment.RiskScore.Score,
    RiskLevel:  string(assessment.RiskScore.Level),
}
auditLogger.Log(ctx, auditLog)
```

## Analytics

### Aggregated Statistics

```go
analyzer := risk.NewRiskAnalyzer(db)

stats, err := analyzer.GetStats(ctx, startTime, endTime)

fmt.Printf("Total: %d\n", stats.TotalAssessments)
fmt.Printf("Average: %.2f\n", stats.AverageScore)
fmt.Printf("High Risk: %d\n", stats.ByLevel[risk.RiskLevelHigh])
```

### Top Risky Actions

```go
actions, err := analyzer.GetTopRiskyActions(ctx, 10, 0.7)

for _, action := range actions {
    fmt.Printf("%s: %.2f (%s)\n",
        action.ActionContext.ToolName,
        action.RiskScore.Score,
        action.RiskScore.Level)
}
```

### Trend Analysis

```go
trends, err := analyzer.GetTrend(ctx, start, end, "day")

for _, point := range trends {
    fmt.Printf("%s: %.2f (%d actions)\n",
        point.Timestamp.Format("2006-01-02"),
        point.AverageScore,
        point.Count)
}
```

## Compliance Reporting

### Generate Reports

```go
generator := risk.NewReportGenerator(db)

// Daily report
report, err := generator.GenerateDailyReport(ctx, time.Now())

// Weekly report
report, err := generator.GenerateWeeklyReport(ctx, weekStart)

// Monthly report
report, err := generator.GenerateMonthlyReport(ctx, 2026, time.February)

// Custom period
report, err := generator.GenerateCustomReport(ctx, start, end)
```

### Report Structure

```go
type ComplianceReport struct {
    ID            string
    ReportType    string  // daily, weekly, monthly
    PeriodStart   time.Time
    PeriodEnd     time.Time
    TotalActions  int
    HighRiskCount int
    DeniedCount   int
    Summary       ComplianceSummary
    Findings      []ComplianceFinding
    Recommendations []string
}
```

### Export Reports

```go
// Export as JSON
data, err := generator.ExportReport(ctx, reportID, risk.ExportFormatJSON)

// Export as CSV
data, err := generator.ExportReport(ctx, reportID, risk.ExportFormatCSV)
```

## Best Practices

### 1. Configure Weights for Your Environment

```yaml
# High-security environment
weights:
  security: 2.0      # Emphasize security
  compliance: 1.8    # Emphasize compliance
  cost: 0.2          # De-emphasize cost

# Cost-conscious environment
weights:
  cost: 1.5          # Emphasize cost
  security: 1.0      # Standard security
```

### 2. Set Environment-Specific Thresholds

```yaml
# Production - strict
production:
  require_approval: 0.6
  auto_deny: 0.85

# Dev - permissive
dev:
  require_approval: 0.9
  auto_deny: 0.98
```

### 3. Monitor Trends

```bash
# Weekly check
unagnt risk trend --days 7 --interval day

# Watch for:
# - Increasing average scores
# - Spike in high-risk actions
# - Unusual patterns
```

### 4. Regular Compliance Reports

```bash
# Automate daily reports
0 0 * * * unagnt risk report generate --type daily

# Weekly review
0 0 * * 0 unagnt risk report generate --type weekly
```

### 5. Tune Based on Analytics

```bash
# Get stats
unagnt risk stats --days 30

# If denial rate > 10%:
# - Policies too strict
# - Adjust thresholds

# If average score > 0.7:
# - Actions too risky
# - Review agent behavior
```

## Use Cases

### 1. Pre-Execution Gating

```go
// Before executing tool
assessment, _ := engine.Assess(ctx, actionContext)

if assessment.Decision == risk.DecisionDeny {
    return errors.New("action denied by risk engine")
}

if assessment.Decision == risk.DecisionRequireReview {
    // Pause and request approval
    if !getApproval(assessment) {
        return errors.New("approval not granted")
    }
}

// Proceed with action
```

### 2. Audit Trail Enrichment

```go
// Add risk scores to audit logs
auditLogger.Log(AuditLog{
    Action:     "delete_file",
    RiskScore:  0.66,
    RiskLevel:  "high",
    Decision:   "allow_with_log",
})
```

### 3. Compliance Monitoring

```bash
# Generate monthly report
unagnt risk report generate --type monthly

# Check compliance rate
unagnt risk stats --days 30

# Export for auditors
unagnt risk report export report-feb --format pdf
```

### 4. Incident Investigation

```bash
# Find high-risk actions
unagnt risk top --min-score 0.8

# Review decisions
unagnt policy audit query --decision deny --days 7

# Check trends around incident
unagnt risk trend --days 7 --interval hour
```

### 5. Policy Tuning

```bash
# Assess current policy effectiveness
unagnt risk stats

# If too many denials:
# - Lower auto_deny threshold
# - Adjust category weights

# If too few denials:
# - Increase security weight
# - Lower approval threshold
```

## Performance

### Assessment Speed

- **Average**: 0.5-2ms per assessment
- **Benchmark**: 500,000+ assessments/second
- **Overhead**: <1% additional latency

### Storage

- **Assessment Size**: ~500 bytes (compressed)
- **Report Size**: ~50KB per daily report
- **Retention**: 90 days recommended

## Security Considerations

### PII Protection

Risk assessments may contain sensitive data:

```yaml
risk:
  encrypt_pii: true        # Encrypt prompts/responses
  redact_inputs: true      # Redact sensitive inputs
  storage_encryption: true # Encrypt at rest
```

### Access Control

Restrict risk data access:

```bash
# Only admins can view compliance reports
rbac:
  - role: admin
    permissions: [risk:read, risk:export]
  - role: user
    permissions: [risk:read_own]
```

## Troubleshooting

### High Denial Rate

```bash
# Check denial reasons
unagnt policy audit query --decision deny --days 7

# Review thresholds
# If too many denials, increase auto_deny threshold
```

### Low Risk Scores

```bash
# Check category weights
unagnt risk assess --tool <tool> --env production

# Adjust weights in config if needed
```

### Missing Assessments

```bash
# Verify risk engine is enabled
# Check agent config:
risk_scoring:
  enabled: true
```

## Advanced Features

### Custom Risk Rules

```go
config := risk.RiskConfig{
    CategoryRules: map[risk.RiskCategory][]risk.RiskRule{
        risk.CategorySecurity: {
            {
                Name: "production-delete-critical",
                Condition: "tool == 'delete_file' && env == 'production'",
                Score: 0.9,
                Weight: 2.0,
            },
        },
    },
}
```

### Dynamic Weight Adjustment

```go
// Adjust weights based on context
if isBusinessHours() {
    config.DefaultWeights[risk.CategoryImpact] = 1.5 // Higher during work
} else {
    config.DefaultWeights[risk.CategoryImpact] = 0.8 // Lower overnight
}
```

### Anomaly Detection

Future: Detect unusual risk patterns

```go
// Detect if current risk significantly higher than baseline
anomaly := detectAnomaly(currentScore, historicalAverage)
if anomaly.Severity == "high" {
    alertSecurityTeam(anomaly)
}
```

## Database Schema

```sql
-- Risk assessments
CREATE TABLE risk_assessments (
    id TEXT PRIMARY KEY,
    action_context JSON NOT NULL,
    risk_score JSON NOT NULL,
    decision TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL
);

-- Compliance reports
CREATE TABLE compliance_reports (
    id TEXT PRIMARY KEY,
    report_type TEXT NOT NULL,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    total_actions INTEGER NOT NULL,
    high_risk_count INTEGER NOT NULL,
    summary JSON NOT NULL
);

-- Risk anomalies
CREATE TABLE risk_anomalies (
    id TEXT PRIMARY KEY,
    assessment_id TEXT NOT NULL,
    anomaly_type TEXT NOT NULL,
    severity TEXT NOT NULL,
    detected_at TIMESTAMP NOT NULL
);
```

See `migrations/007_risk_scoring.sql` for full schema.

## API Reference

### RiskScore

```go
type RiskScore struct {
    Score       float64            // 0.0 to 1.0
    Level       RiskLevel          // low, medium, high, critical
    Factors     []RiskFactor       // Contributing factors
    Breakdown   map[string]float64 // Per-category scores
    Confidence  float64            // Assessment confidence
    Timestamp   time.Time
}
```

### ActionContext

```go
type ActionContext struct {
    ToolName       string
    Input          map[string]interface{}
    Permissions    []string
    AgentID        string
    Environment    string  // dev, staging, production
    RecentActions  int     // Actions in last 5 min
}
```

### RiskAssessment

```go
type RiskAssessment struct {
    ActionContext ActionContext
    RiskScore     RiskScore
    Decision      RiskDecision
    Timestamp     time.Time
    Version       string
}
```

## Examples

### Example 1: Risk-Based Approval

```go
engine := risk.NewRiskEngine(config)

// Assess before action
assessment, _ := engine.Assess(ctx, risk.ActionContext{
    ToolName:    "delete_file",
    Environment: "production",
    Permissions: []string{"fs:delete"},
})

if assessment.RiskScore.Score >= 0.8 {
    fmt.Println("High risk! Requesting approval...")
    approval := requestApproval(assessment)
    
    if !approval.Granted {
        return errors.New("approval denied")
    }
}

// Proceed with action
```

### Example 2: Risk Monitoring

```bash
# Morning routine
unagnt risk stats --days 1
unagnt risk top --limit 10

# Weekly review
unagnt risk report generate --type weekly
unagnt risk trend --days 30
```

### Example 3: Compliance Audit

```bash
# Generate quarterly report
unagnt risk report generate --type custom \
  --start 2026-01-01 --end 2026-03-31

# Export for auditors
unagnt risk report export report-q1 --format pdf

# Review findings
unagnt risk report view report-q1
```

## Next Steps

- [Policy Engine Guide](policy-versioning.md)
- [Deterministic Replay Guide](deterministic-replay.md)
- [API Documentation](api-reference.md)
