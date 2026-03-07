package cost

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// CostTracker tracks costs per agent, tenant, and user
type CostTracker struct {
	db            *sql.DB
	providers     map[string]PricingProvider
	currentCosts  map[string]*CostAccumulator
	mu            sync.RWMutex
	flushInterval time.Duration
	stopChan      chan struct{}
}

// CostAccumulator accumulates costs in memory before flushing
type CostAccumulator struct {
	AgentID      string
	TenantID     string
	UserID       string
	Provider     string
	Model        string
	WorkflowID   string
	WorkflowName string
	InputTokens  int64
	OutputTokens int64
	TotalCost    float64
	CallCount    int64
	LastUpdated  time.Time
}

// CostEntry represents a cost record in the database
type CostEntry struct {
	ID           string
	AgentID      string
	TenantID     string
	UserID       string
	Provider     string
	Model        string
	WorkflowID   string
	WorkflowName string
	InputTokens  int64
	OutputTokens int64
	Cost         float64
	CallCount    int64
	Timestamp    time.Time
}

// PricingProvider defines pricing information for an LLM provider
type PricingProvider interface {
	GetModelPricing(model string) (*ModelPricing, error)
	CalculateCost(model string, inputTokens, outputTokens int64) (float64, error)
}

// ModelPricing contains pricing information for a specific model
type ModelPricing struct {
	Provider        string
	Model           string
	InputPricePerM  float64 // Price per million input tokens
	OutputPricePerM float64 // Price per million output tokens
	MinimumCharge   float64
	Currency        string
}

// NewCostTracker creates a new cost tracker
func NewCostTracker(db *sql.DB) *CostTracker {
	return &CostTracker{
		db:            db,
		providers:     make(map[string]PricingProvider),
		currentCosts:  make(map[string]*CostAccumulator),
		flushInterval: 30 * time.Second,
		stopChan:      make(chan struct{}),
	}
}

// RegisterProvider registers a pricing provider
func (ct *CostTracker) RegisterProvider(name string, provider PricingProvider) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.providers[name] = provider
}

// Start starts the cost tracker background tasks
func (ct *CostTracker) Start(ctx context.Context) {
	go ct.flushLoop(ctx)
}

// Stop stops the cost tracker
func (ct *CostTracker) Stop() {
	close(ct.stopChan)
}

// TrackLLMCall tracks the cost of an LLM API call.
// workflowID and workflowName are optional; pass "" when not in a workflow.
func (ct *CostTracker) TrackLLMCall(ctx context.Context, agentID, tenantID, userID, provider, model string, inputTokens, outputTokens int64, workflowID, workflowName string) error {
	// Calculate cost
	cost, err := ct.calculateCost(provider, model, inputTokens, outputTokens)
	if err != nil {
		return fmt.Errorf("calculate cost: %w", err)
	}

	// Create accumulator key (include workflow for granularity)
	key := fmt.Sprintf("%s:%s:%s:%s:%s:%s", agentID, tenantID, userID, provider, model, workflowID)

	ct.mu.Lock()
	defer ct.mu.Unlock()

	// Get or create accumulator
	acc, exists := ct.currentCosts[key]
	if !exists {
		acc = &CostAccumulator{
			AgentID:      agentID,
			TenantID:     tenantID,
			UserID:       userID,
			Provider:     provider,
			Model:        model,
			WorkflowID:   workflowID,
			WorkflowName: workflowName,
			LastUpdated:  time.Now(),
		}
		ct.currentCosts[key] = acc
	}

	// Accumulate
	acc.InputTokens += inputTokens
	acc.OutputTokens += outputTokens
	acc.TotalCost += cost
	acc.CallCount++
	acc.LastUpdated = time.Now()

	return nil
}

// calculateCost calculates the cost for a given provider and model
func (ct *CostTracker) calculateCost(provider, model string, inputTokens, outputTokens int64) (float64, error) {
	ct.mu.RLock()
	pricingProvider, exists := ct.providers[provider]
	ct.mu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("provider %s not registered", provider)
	}

	return pricingProvider.CalculateCost(model, inputTokens, outputTokens)
}

// flushLoop periodically flushes accumulated costs to the database
func (ct *CostTracker) flushLoop(ctx context.Context) {
	ticker := time.NewTicker(ct.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ct.flush()
			return
		case <-ct.stopChan:
			ct.flush()
			return
		case <-ticker.C:
			ct.flush()
		}
	}
}

// flush writes accumulated costs to the database
func (ct *CostTracker) flush() {
	ct.mu.Lock()
	costs := make([]*CostAccumulator, 0, len(ct.currentCosts))
	for _, acc := range ct.currentCosts {
		costs = append(costs, acc)
	}
	ct.currentCosts = make(map[string]*CostAccumulator)
	ct.mu.Unlock()

	if len(costs) == 0 {
		return
	}

	// Batch insert
	tx, err := ct.db.Begin()
	if err != nil {
		fmt.Printf("Failed to begin transaction: %v\n", err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO cost_entries (
			id, agent_id, tenant_id, user_id, provider, model,
			workflow_id, workflow_name,
			input_tokens, output_tokens, cost, call_count, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		fmt.Printf("Failed to prepare statement: %v\n", err)
		return
	}
	defer stmt.Close()

	for _, acc := range costs {
		id := fmt.Sprintf("%d", time.Now().UnixNano())
		_, err := stmt.Exec(
			id, acc.AgentID, acc.TenantID, acc.UserID,
			acc.Provider, acc.Model,
			acc.WorkflowID, acc.WorkflowName,
			acc.InputTokens, acc.OutputTokens,
			acc.TotalCost, acc.CallCount,
			acc.LastUpdated,
		)
		if err != nil {
			fmt.Printf("Failed to insert cost entry: %v\n", err)
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		fmt.Printf("Failed to commit transaction: %v\n", err)
	}
}

// GetCostsByAgent returns costs grouped by agent
func (ct *CostTracker) GetCostsByAgent(ctx context.Context, tenantID string, start, end time.Time) (map[string]float64, error) {
	query := `
		SELECT agent_id, SUM(cost) as total_cost
		FROM cost_entries
		WHERE tenant_id = ? AND timestamp BETWEEN ? AND ?
		GROUP BY agent_id
	`

	rows, err := ct.db.QueryContext(ctx, query, tenantID, start, end)
	if err != nil {
		return nil, fmt.Errorf("query costs: %w", err)
	}
	defer rows.Close()

	costs := make(map[string]float64)
	for rows.Next() {
		var agentID string
		var totalCost float64
		if err := rows.Scan(&agentID, &totalCost); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		costs[agentID] = totalCost
	}

	return costs, rows.Err()
}

// GetCostsByTenant returns costs grouped by tenant
func (ct *CostTracker) GetCostsByTenant(ctx context.Context, start, end time.Time) (map[string]float64, error) {
	query := `
		SELECT tenant_id, SUM(cost) as total_cost
		FROM cost_entries
		WHERE timestamp BETWEEN ? AND ?
		GROUP BY tenant_id
	`

	rows, err := ct.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("query costs: %w", err)
	}
	defer rows.Close()

	costs := make(map[string]float64)
	for rows.Next() {
		var tenantID string
		var totalCost float64
		if err := rows.Scan(&tenantID, &totalCost); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		costs[tenantID] = totalCost
	}

	return costs, rows.Err()
}

// GetCostsByUser returns costs grouped by user
func (ct *CostTracker) GetCostsByUser(ctx context.Context, tenantID string, start, end time.Time) (map[string]float64, error) {
	query := `
		SELECT user_id, SUM(cost) as total_cost
		FROM cost_entries
		WHERE tenant_id = ? AND timestamp BETWEEN ? AND ?
		GROUP BY user_id
	`

	rows, err := ct.db.QueryContext(ctx, query, tenantID, start, end)
	if err != nil {
		return nil, fmt.Errorf("query costs: %w", err)
	}
	defer rows.Close()

	costs := make(map[string]float64)
	for rows.Next() {
		var userID string
		var totalCost float64
		if err := rows.Scan(&userID, &totalCost); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		costs[userID] = totalCost
	}

	return costs, rows.Err()
}

// GetTotalCost returns the total cost for a given time range.
// If tenantID is empty, returns cost across all tenants.
func (ct *CostTracker) GetTotalCost(ctx context.Context, tenantID string, start, end time.Time) (float64, error) {
	var query string
	var args []interface{}
	if tenantID == "" {
		query = `SELECT COALESCE(SUM(cost), 0) as total_cost FROM cost_entries WHERE timestamp BETWEEN ? AND ?`
		args = []interface{}{start, end}
	} else {
		query = `SELECT COALESCE(SUM(cost), 0) as total_cost FROM cost_entries WHERE tenant_id = ? AND timestamp BETWEEN ? AND ?`
		args = []interface{}{tenantID, start, end}
	}
	var totalCost float64
	err := ct.db.QueryRowContext(ctx, query, args...).Scan(&totalCost)
	if err != nil {
		return 0, fmt.Errorf("query total cost: %w", err)
	}
	return totalCost, nil
}

// GetCostsByWorkflow returns costs grouped by workflow
func (ct *CostTracker) GetCostsByWorkflow(ctx context.Context, tenantID string, start, end time.Time) (map[string]float64, error) {
	query := `
		SELECT COALESCE(workflow_id, '') as wf_id, SUM(cost) as total_cost
		FROM cost_entries
		WHERE tenant_id = ? AND timestamp BETWEEN ? AND ?
		GROUP BY COALESCE(workflow_id, '')
	`

	rows, err := ct.db.QueryContext(ctx, query, tenantID, start, end)
	if err != nil {
		return nil, fmt.Errorf("query costs by workflow: %w", err)
	}
	defer rows.Close()

	costs := make(map[string]float64)
	for rows.Next() {
		var wfID string
		var totalCost float64
		if err := rows.Scan(&wfID, &totalCost); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		key := wfID
		if key == "" {
			key = "(no workflow)"
		}
		costs[key] = totalCost
	}

	return costs, rows.Err()
}

// CostBreakdownFilter optionally filters GetCostBreakdown by workflow or model
type CostBreakdownFilter struct {
	WorkflowID string
	Model      string // provider:model or exact model name
}

// GetCostBreakdown returns detailed cost breakdown.
// Pass nil filter for no filtering.
func (ct *CostTracker) GetCostBreakdown(ctx context.Context, tenantID string, start, end time.Time, filter *CostBreakdownFilter) ([]*CostEntry, error) {
	query := `
		SELECT id, agent_id, tenant_id, user_id, provider, model,
		       COALESCE(workflow_id, ''), COALESCE(workflow_name, ''),
		       input_tokens, output_tokens, cost, call_count, timestamp
		FROM cost_entries
		WHERE tenant_id = ? AND timestamp BETWEEN ? AND ?
	`
	args := []interface{}{tenantID, start, end}

	if filter != nil {
		if filter.WorkflowID != "" {
			query += ` AND (workflow_id = ? OR (workflow_id IS NULL AND ? = ''))`
			args = append(args, filter.WorkflowID, filter.WorkflowID)
		}
		if filter.Model != "" {
			query += ` AND (model = ? OR (provider || ':' || model) = ?)`
			args = append(args, filter.Model, filter.Model)
		}
	}

	query += ` ORDER BY timestamp DESC`

	rows, err := ct.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query cost breakdown: %w", err)
	}
	defer rows.Close()

	var entries []*CostEntry
	for rows.Next() {
		entry := &CostEntry{}
		var wfID, wfName string
		if err := rows.Scan(
			&entry.ID, &entry.AgentID, &entry.TenantID, &entry.UserID,
			&entry.Provider, &entry.Model,
			&wfID, &wfName,
			&entry.InputTokens, &entry.OutputTokens,
			&entry.Cost, &entry.CallCount, &entry.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		entry.WorkflowID = wfID
		entry.WorkflowName = wfName
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}
