package risk

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"time"
)

// RiskScore represents a calculated risk score.
type RiskScore struct {
	Score      float64            `json:"score"`      // 0.0 to 1.0
	Level      RiskLevel          `json:"level"`      // low, medium, high, critical
	Factors    []RiskFactor       `json:"factors"`    // Contributing factors
	Breakdown  map[string]float64 `json:"breakdown"`  // Category scores
	Confidence float64            `json:"confidence"` // Confidence in assessment
	Timestamp  time.Time          `json:"timestamp"`
	Metadata   map[string]any     `json:"metadata,omitempty"`
}

// RiskLevel categorizes risk severity.
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"      // 0.0 - 0.3
	RiskLevelMedium   RiskLevel = "medium"   // 0.3 - 0.6
	RiskLevelHigh     RiskLevel = "high"     // 0.6 - 0.85
	RiskLevelCritical RiskLevel = "critical" // 0.85 - 1.0
)

// RiskFactor represents a component of risk.
type RiskFactor struct {
	Category     string  `json:"category"` // security, privacy, cost, etc.
	Description  string  `json:"description"`
	Score        float64 `json:"score"`        // 0.0 to 1.0
	Weight       float64 `json:"weight"`       // Importance multiplier
	Contribution float64 `json:"contribution"` // score * weight
}

// RiskCategory defines risk assessment categories.
type RiskCategory string

const (
	CategorySecurity    RiskCategory = "security"    // Data access, permissions
	CategoryPrivacy     RiskCategory = "privacy"     // PII, sensitive data
	CategoryCost        RiskCategory = "cost"        // API costs, resource usage
	CategoryReversible  RiskCategory = "reversible"  // Can action be undone?
	CategoryImpact      RiskCategory = "impact"      // Blast radius
	CategoryCompliance  RiskCategory = "compliance"  // Regulatory requirements
	CategoryReliability RiskCategory = "reliability" // Failure likelihood
)

// ActionContext provides context for risk assessment.
type ActionContext struct {
	ToolName        string                 `json:"tool_name"`
	Input           map[string]interface{} `json:"input"`
	Permissions     []string               `json:"permissions"`
	AgentID         string                 `json:"agent_id"`
	Environment     string                 `json:"environment"`    // dev, staging, prod
	RecentActions   int                    `json:"recent_actions"` // Actions in last 5 min
	PreviousDenials int                    `json:"previous_denials"`
	UserApproved    bool                   `json:"user_approved"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// RiskConfig configures the risk engine.
type RiskConfig struct {
	Enabled         bool                        `json:"enabled"`
	DefaultWeights  map[RiskCategory]float64    `json:"default_weights"`
	CategoryRules   map[RiskCategory][]RiskRule `json:"category_rules"`
	ThresholdLow    float64                     `json:"threshold_low"`    // 0.3
	ThresholdMedium float64                     `json:"threshold_medium"` // 0.6
	ThresholdHigh   float64                     `json:"threshold_high"`   // 0.85
	RequireApproval float64                     `json:"require_approval"` // 0.8
	AutoDeny        float64                     `json:"auto_deny"`        // 0.95
}

// RiskRule defines a risk assessment rule.
type RiskRule struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Condition   string                 `json:"condition"` // CEL expression
	Score       float64                `json:"score"`     // 0.0 to 1.0
	Weight      float64                `json:"weight"`    // Importance
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RiskAssessment contains detailed risk analysis.
type RiskAssessment struct {
	ActionContext ActionContext `json:"action_context"`
	RiskScore     RiskScore     `json:"risk_score"`
	Decision      RiskDecision  `json:"decision"`
	Timestamp     time.Time     `json:"timestamp"`
	AssessorID    string        `json:"assessor_id"`
	Version       string        `json:"version"` // Risk engine version
}

// RiskDecision represents the outcome of risk assessment.
type RiskDecision string

const (
	DecisionAllow         RiskDecision = "allow"          // Proceed
	DecisionAllowWithLog  RiskDecision = "allow_with_log" // Proceed but log
	DecisionRequireReview RiskDecision = "require_review" // Human approval needed
	DecisionDeny          RiskDecision = "deny"           // Block action
)

// RiskEngine calculates risk scores.
type RiskEngine struct {
	config  RiskConfig
	version string
}

// NewRiskEngine creates a new risk engine.
func NewRiskEngine(config RiskConfig) *RiskEngine {
	// Set default thresholds if not provided
	if config.ThresholdLow == 0 {
		config.ThresholdLow = 0.3
	}
	if config.ThresholdMedium == 0 {
		config.ThresholdMedium = 0.6
	}
	if config.ThresholdHigh == 0 {
		config.ThresholdHigh = 0.85
	}
	if config.RequireApproval == 0 {
		config.RequireApproval = 0.8
	}
	if config.AutoDeny == 0 {
		config.AutoDeny = 0.95
	}

	// Set default weights
	if config.DefaultWeights == nil {
		config.DefaultWeights = map[RiskCategory]float64{
			CategorySecurity:    1.5, // Higher weight for security
			CategoryPrivacy:     1.2,
			CategoryCost:        0.3,
			CategoryReversible:  1.0,
			CategoryImpact:      1.3, // Higher weight for impact
			CategoryCompliance:  1.1,
			CategoryReliability: 0.4,
		}
	}

	return &RiskEngine{
		config:  config,
		version: "1.0.0",
	}
}

// Assess calculates risk score for an action.
func (e *RiskEngine) Assess(ctx context.Context, actionCtx ActionContext) (*RiskAssessment, error) {
	if !e.config.Enabled {
		return &RiskAssessment{
			ActionContext: actionCtx,
			RiskScore: RiskScore{
				Score: 0.0,
				Level: RiskLevelLow,
			},
			Decision:  DecisionAllow,
			Timestamp: time.Now(),
			Version:   e.version,
		}, nil
	}

	// Calculate risk factors for each category
	factors := make([]RiskFactor, 0)

	// Security risk
	securityScore := e.assessSecurity(actionCtx)
	factors = append(factors, RiskFactor{
		Category:     string(CategorySecurity),
		Description:  "Permissions and data access risk",
		Score:        securityScore,
		Weight:       e.config.DefaultWeights[CategorySecurity],
		Contribution: securityScore * e.config.DefaultWeights[CategorySecurity],
	})

	// Privacy risk
	privacyScore := e.assessPrivacy(actionCtx)
	factors = append(factors, RiskFactor{
		Category:     string(CategoryPrivacy),
		Description:  "PII and sensitive data risk",
		Score:        privacyScore,
		Weight:       e.config.DefaultWeights[CategoryPrivacy],
		Contribution: privacyScore * e.config.DefaultWeights[CategoryPrivacy],
	})

	// Cost risk
	costScore := e.assessCost(actionCtx)
	factors = append(factors, RiskFactor{
		Category:     string(CategoryCost),
		Description:  "Resource and API cost risk",
		Score:        costScore,
		Weight:       e.config.DefaultWeights[CategoryCost],
		Contribution: costScore * e.config.DefaultWeights[CategoryCost],
	})

	// Reversibility risk
	reversibleScore := e.assessReversibility(actionCtx)
	factors = append(factors, RiskFactor{
		Category:     string(CategoryReversible),
		Description:  "Ability to undo action",
		Score:        reversibleScore,
		Weight:       e.config.DefaultWeights[CategoryReversible],
		Contribution: reversibleScore * e.config.DefaultWeights[CategoryReversible],
	})

	// Impact risk
	impactScore := e.assessImpact(actionCtx)
	factors = append(factors, RiskFactor{
		Category:     string(CategoryImpact),
		Description:  "Blast radius and downstream effects",
		Score:        impactScore,
		Weight:       e.config.DefaultWeights[CategoryImpact],
		Contribution: impactScore * e.config.DefaultWeights[CategoryImpact],
	})

	// Compliance risk
	complianceScore := e.assessCompliance(actionCtx)
	factors = append(factors, RiskFactor{
		Category:     string(CategoryCompliance),
		Description:  "Regulatory and policy compliance",
		Score:        complianceScore,
		Weight:       e.config.DefaultWeights[CategoryCompliance],
		Contribution: complianceScore * e.config.DefaultWeights[CategoryCompliance],
	})

	// Calculate weighted score
	totalWeight := 0.0
	totalScore := 0.0
	breakdown := make(map[string]float64)

	for _, factor := range factors {
		totalScore += factor.Contribution
		totalWeight += factor.Weight
		breakdown[factor.Category] = factor.Score
	}

	finalScore := totalScore / totalWeight
	if math.IsNaN(finalScore) {
		finalScore = 0.5 // Default to medium risk if calculation fails
	}

	// Determine risk level
	level := e.scoreToLevel(finalScore)

	// Calculate confidence
	confidence := e.calculateConfidence(actionCtx, factors)

	riskScore := RiskScore{
		Score:      finalScore,
		Level:      level,
		Factors:    factors,
		Breakdown:  breakdown,
		Confidence: confidence,
		Timestamp:  time.Now(),
	}

	// Make decision
	decision := e.makeDecision(riskScore, actionCtx)

	assessment := &RiskAssessment{
		ActionContext: actionCtx,
		RiskScore:     riskScore,
		Decision:      decision,
		Timestamp:     time.Now(),
		AssessorID:    "risk-engine-v1",
		Version:       e.version,
	}

	return assessment, nil
}

// assessSecurity evaluates security risk.
func (e *RiskEngine) assessSecurity(ctx ActionContext) float64 {
	score := 0.0

	// High-risk tools
	highRiskTools := map[string]float64{
		"delete_file":     0.8,
		"execute_command": 0.95,
		"http_request":    0.4,
		"write_file":      0.5,
		"read_file":       0.2,
	}

	if baseScore, exists := highRiskTools[ctx.ToolName]; exists {
		score = baseScore
	} else {
		score = 0.3 // Default for unknown tools
	}

	// Permissions boost risk significantly
	dangerousPerms := map[string]float64{
		"exec":         0.3,
		"fs:delete":    0.25,
		"fs:write":     0.15,
		"net:external": 0.1,
	}

	for _, perm := range ctx.Permissions {
		if boost, exists := dangerousPerms[perm]; exists {
			score += boost
		}
	}

	// Production environment significantly increases risk
	if ctx.Environment == "production" {
		score += 0.3
	} else if ctx.Environment == "staging" {
		score += 0.1
	}

	return math.Min(score, 1.0)
}

// assessPrivacy evaluates privacy risk.
func (e *RiskEngine) assessPrivacy(ctx ActionContext) float64 {
	score := 0.0

	// Check for PII-related tools
	piiTools := map[string]float64{
		"read_file":    0.4,
		"http_request": 0.5,
		"database":     0.6,
	}

	if baseScore, exists := piiTools[ctx.ToolName]; exists {
		score = baseScore
	}

	// Check input for PII patterns (simplified)
	if ctx.Input != nil {
		inputStr := fmt.Sprintf("%v", ctx.Input)
		if containsPII(inputStr) {
			score += 0.3
		}
	}

	return math.Min(score, 1.0)
}

// assessCost evaluates cost risk.
func (e *RiskEngine) assessCost(ctx ActionContext) float64 {
	score := 0.0

	// High-cost tools
	costlyTools := map[string]float64{
		"http_request": 0.3,
		"model_call":   0.4,
		"database":     0.2,
	}

	if baseScore, exists := costlyTools[ctx.ToolName]; exists {
		score = baseScore
	}

	// Rapid actions increase cost risk
	if ctx.RecentActions > 10 {
		score += 0.2
	}
	if ctx.RecentActions > 50 {
		score += 0.3
	}

	return math.Min(score, 1.0)
}

// assessReversibility evaluates if action can be undone.
func (e *RiskEngine) assessReversibility(ctx ActionContext) float64 {
	// Irreversible actions have high risk
	irreversibleTools := map[string]float64{
		"delete_file":     1.0,
		"execute_command": 0.9,
		"http_request":    0.7, // External calls may not be reversible
		"send_email":      1.0,
	}

	if score, exists := irreversibleTools[ctx.ToolName]; exists {
		return score
	}

	// Reversible by default
	return 0.2
}

// assessImpact evaluates blast radius.
func (e *RiskEngine) assessImpact(ctx ActionContext) float64 {
	score := 0.0

	// Production has much higher impact
	envImpact := map[string]float64{
		"production": 0.9,
		"staging":    0.5,
		"dev":        0.1,
	}

	if impact, exists := envImpact[ctx.Environment]; exists {
		score = impact
	}

	// Tools with wide impact
	highImpactTools := map[string]float64{
		"delete_file":     0.4,
		"execute_command": 0.5,
		"database":        0.4,
		"write_file":      0.3,
	}

	if toolImpact, exists := highImpactTools[ctx.ToolName]; exists {
		score += toolImpact
	}

	return math.Min(score, 1.0)
}

// assessCompliance evaluates regulatory risk.
func (e *RiskEngine) assessCompliance(ctx ActionContext) float64 {
	score := 0.0

	// Check for compliance-sensitive operations
	complianceTools := map[string]float64{
		"read_file":    0.3, // May access regulated data
		"write_file":   0.4,
		"database":     0.5,
		"http_request": 0.4, // External data transfer
	}

	if baseScore, exists := complianceTools[ctx.ToolName]; exists {
		score = baseScore
	}

	// Production operations need compliance checks
	if ctx.Environment == "production" {
		score += 0.2
	}

	return math.Min(score, 1.0)
}

// scoreToLevel converts numeric score to risk level.
func (e *RiskEngine) scoreToLevel(score float64) RiskLevel {
	if score < e.config.ThresholdLow {
		return RiskLevelLow
	}
	if score < e.config.ThresholdMedium {
		return RiskLevelMedium
	}
	if score < e.config.ThresholdHigh {
		return RiskLevelHigh
	}
	return RiskLevelCritical
}

// calculateConfidence estimates assessment confidence.
func (e *RiskEngine) calculateConfidence(ctx ActionContext, factors []RiskFactor) float64 {
	confidence := 0.7 // Base confidence

	// More factors = higher confidence
	if len(factors) > 5 {
		confidence += 0.1
	}

	// User approval increases confidence
	if ctx.UserApproved {
		confidence += 0.1
	}

	// Production environment = more data = higher confidence
	if ctx.Environment == "production" {
		confidence += 0.1
	}

	return math.Min(confidence, 1.0)
}

// makeDecision determines action based on risk score.
func (e *RiskEngine) makeDecision(score RiskScore, ctx ActionContext) RiskDecision {
	// Auto-deny critical risk
	if score.Score >= e.config.AutoDeny {
		return DecisionDeny
	}

	// Require approval for high risk
	if score.Score >= e.config.RequireApproval {
		return DecisionRequireReview
	}

	// Log medium risk
	if score.Level == RiskLevelMedium || score.Level == RiskLevelHigh {
		return DecisionAllowWithLog
	}

	// Allow low risk
	return DecisionAllow
}

// containsPII checks for common PII patterns using regex.
// Uses heuristics for SSN, email, credit card, phone; production may integrate dedicated PII detection.
var (
	piiSSN        = regexp.MustCompile(`\b\d{3}[-\s]?\d{2}[-\s]?\d{4}\b`)
	piiEmail      = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	piiCreditCard = regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`)
	piiPhone      = regexp.MustCompile(`(?:\+1[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}\b`)
	piiKeywords   = regexp.MustCompile(`(?i)\b(?:ssn|social\s*security|credit\s*card|password|secret|api[_-]?key|token)\b`)
)

func containsPII(input string) bool {
	if piiSSN.MatchString(input) || piiEmail.MatchString(input) ||
		piiCreditCard.MatchString(input) || piiPhone.MatchString(input) ||
		piiKeywords.MatchString(input) {
		return true
	}
	return false
}

// DefaultRiskConfig returns a sensible default configuration.
func DefaultRiskConfig() RiskConfig {
	return RiskConfig{
		Enabled: true,
		DefaultWeights: map[RiskCategory]float64{
			CategorySecurity:    1.5,
			CategoryPrivacy:     1.2,
			CategoryCost:        0.3,
			CategoryReversible:  1.0,
			CategoryImpact:      1.3,
			CategoryCompliance:  1.1,
			CategoryReliability: 0.4,
		},
		ThresholdLow:    0.3,
		ThresholdMedium: 0.6,
		ThresholdHigh:   0.85,
		RequireApproval: 0.8,
		AutoDeny:        0.95,
	}
}
