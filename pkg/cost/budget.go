package cost

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// BudgetConfig configures budget limits and alerts.
type BudgetConfig struct {
	// BudgetLimit is the hard cap in USD (0 = no limit).
	BudgetLimit float64 `yaml:"budget_limit" json:"budget_limit"`
	// AlertThreshold is the fraction of budget at which to fire alert (e.g. 0.8 = 80%).
	AlertThreshold float64 `yaml:"alert_threshold" json:"alert_threshold"`
	// AlertWebhook is the URL to POST when alert fires (optional).
	AlertWebhook string `yaml:"alert_webhook" json:"alert_webhook"`
	// Period for budget: "daily", "weekly", "monthly".
	Period string `yaml:"period" json:"period"`
	// TenantID scopes the budget (optional).
	TenantID string `yaml:"tenant_id" json:"tenant_id"`
}

// BudgetGuard checks budget before allowing spends and fires alerts.
type BudgetGuard struct {
	tracker *CostTracker
	config  BudgetConfig
	mu      sync.Mutex
	alerted bool
}

// NewBudgetGuard creates a budget guard.
func NewBudgetGuard(tracker *CostTracker, config BudgetConfig) *BudgetGuard {
	if config.Period == "" {
		config.Period = "daily"
	}
	if config.AlertThreshold == 0 {
		config.AlertThreshold = 0.8
	}
	return &BudgetGuard{tracker: tracker, config: config}
}

// CanSpend returns true if adding estimatedCost would not exceed the budget.
func (g *BudgetGuard) CanSpend(ctx context.Context, agentID string, estimatedCost float64) (bool, error) {
	if g.config.BudgetLimit <= 0 {
		return true, nil
	}
	start, end := g.periodRange()
	current, err := g.tracker.GetTotalCost(ctx, g.config.TenantID, start, end)
	if err != nil {
		return false, err
	}
	return current+estimatedCost <= g.config.BudgetLimit, nil
}

// CheckAndAlert checks current spend, fires alert if over threshold, returns whether over limit.
func (g *BudgetGuard) CheckAndAlert(ctx context.Context, agentID string) (overLimit bool, err error) {
	if g.config.BudgetLimit <= 0 {
		return false, nil
	}
	start, end := g.periodRange()
	current, err := g.tracker.GetTotalCost(ctx, g.config.TenantID, start, end)
	if err != nil {
		return false, err
	}
	if current >= g.config.BudgetLimit {
		return true, nil
	}
	if g.config.AlertWebhook != "" && !g.alerted && current >= g.config.BudgetLimit*g.config.AlertThreshold {
		g.mu.Lock()
		if !g.alerted {
			g.alerted = true
			go g.fireAlert(current, g.config.BudgetLimit)
		}
		g.mu.Unlock()
	}
	return false, nil
}

func (g *BudgetGuard) periodRange() (time.Time, time.Time) {
	end := time.Now()
	var start time.Time
	switch g.config.Period {
	case "daily":
		start = end.AddDate(0, 0, -1)
	case "weekly":
		start = end.AddDate(0, 0, -7)
	case "monthly":
		start = end.AddDate(0, -1, 0)
	default:
		start = end.AddDate(0, 0, -1)
	}
	return start, end
}

func (g *BudgetGuard) fireAlert(current, limit float64) {
	body, _ := json.Marshal(map[string]interface{}{
		"event":         "budget_alert",
		"current_spend": current,
		"budget_limit":  limit,
		"percent_used":  current / limit * 100,
		"tenant_id":     g.config.TenantID,
	})
	resp, err := http.Post(g.config.AlertWebhook, "application/json", bytes.NewReader(body))
	if err != nil {
		return
	}
	resp.Body.Close()
}
