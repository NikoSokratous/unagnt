package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/cost"
	"github.com/NikoSokratous/unagnt/pkg/mlops"
	"github.com/NikoSokratous/unagnt/pkg/monitoring"
	"github.com/NikoSokratous/unagnt/pkg/policy"
	"github.com/gorilla/mux"
)

// AnalyticsAPI handles analytics endpoints
type AnalyticsAPI struct {
	costTracker *cost.CostTracker
	slaMonitor  *monitoring.SLAMonitor
	auditLogger *policy.AuditLogger     // optional
	perfStore   *mlops.PerformanceStore // optional; enables model-performance and model-drift
}

// NewAnalyticsAPI creates a new analytics API
func NewAnalyticsAPI(costTracker *cost.CostTracker, slaMonitor *monitoring.SLAMonitor) *AnalyticsAPI {
	return &AnalyticsAPI{
		costTracker: costTracker,
		slaMonitor:  slaMonitor,
	}
}

// NewAnalyticsAPIWithAudit creates analytics API with policy audit support
func NewAnalyticsAPIWithAudit(costTracker *cost.CostTracker, slaMonitor *monitoring.SLAMonitor, audit *policy.AuditLogger) *AnalyticsAPI {
	return &AnalyticsAPI{
		costTracker: costTracker,
		slaMonitor:  slaMonitor,
		auditLogger: audit,
	}
}

// SetModelPerformanceStore enables model performance and drift endpoints
func (api *AnalyticsAPI) SetModelPerformanceStore(store *mlops.PerformanceStore) {
	api.perfStore = store
}

// RegisterRoutes registers analytics API routes
func (api *AnalyticsAPI) RegisterRoutes(router *mux.Router) {
	// Cost endpoints
	router.HandleFunc("/v1/analytics/costs", api.GetCosts).Methods("GET")
	router.HandleFunc("/v1/analytics/costs/agents", api.GetCostsByAgent).Methods("GET")
	router.HandleFunc("/v1/analytics/costs/tenants", api.GetCostsByTenant).Methods("GET")
	router.HandleFunc("/v1/analytics/costs/breakdown", api.GetCostBreakdown).Methods("GET")
	router.HandleFunc("/v1/analytics/costs/workflows", api.GetCostsByWorkflow).Methods("GET")

	// SLA endpoints
	router.HandleFunc("/v1/analytics/sla", api.GetSLA).Methods("GET")
	router.HandleFunc("/v1/analytics/sla/report", api.GenerateSLAReport).Methods("GET")
	router.HandleFunc("/v1/analytics/sla/targets", api.SetSLATarget).Methods("POST")

	// Model performance and drift (v4)
	router.HandleFunc("/v1/analytics/model-performance", api.GetModelPerformance).Methods("GET")
	router.HandleFunc("/v1/analytics/model-drift", api.GetModelDrift).Methods("GET")

	// Performance endpoints
	router.HandleFunc("/v1/analytics/performance", api.GetPerformance).Methods("GET")
	router.HandleFunc("/v1/analytics/metrics", api.GetMetrics).Methods("GET")
	router.HandleFunc("/v1/analytics/denials", api.GetPolicyDenials).Methods("GET")
	router.HandleFunc("/v1/analytics/denials/stats", api.GetPolicyDenialsStats).Methods("GET")

	// Compliance / SIEM audit export (v2.0)
	router.HandleFunc("/v1/compliance/audit/export", api.ExportAuditLogs).Methods("GET")
}

// GetCosts handles GET /v1/analytics/costs
func (api *AnalyticsAPI) GetCosts(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	timeRange := r.URL.Query().Get("range")

	start, end := parseTimeRange(timeRange)

	total, err := api.costTracker.GetTotalCost(r.Context(), tenantID, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	byAgent, err := api.costTracker.GetCostsByAgent(r.Context(), tenantID, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Format for frontend
	agentCosts := make([]map[string]interface{}, 0)
	for agentID, cost := range byAgent {
		agentCosts = append(agentCosts, map[string]interface{}{
			"agent_id": agentID,
			"cost":     cost,
			"requests": 0, // Would be queried separately
		})
	}

	response := map[string]interface{}{
		"total":    total,
		"by_agent": agentCosts,
		"period": map[string]string{
			"start": start.Format(time.RFC3339),
			"end":   end.Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCostsByAgent handles GET /v1/analytics/costs/agents
func (api *AnalyticsAPI) GetCostsByAgent(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	timeRange := r.URL.Query().Get("range")

	start, end := parseTimeRange(timeRange)

	costs, err := api.costTracker.GetCostsByAgent(r.Context(), tenantID, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(costs)
}

// GetCostsByTenant handles GET /v1/analytics/costs/tenants
func (api *AnalyticsAPI) GetCostsByTenant(w http.ResponseWriter, r *http.Request) {
	timeRange := r.URL.Query().Get("range")
	start, end := parseTimeRange(timeRange)

	costs, err := api.costTracker.GetCostsByTenant(r.Context(), start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(costs)
}

// GetCostBreakdown handles GET /v1/analytics/costs/breakdown
// Query params: tenant_id, range, workflow_id, model (optional filters)
func (api *AnalyticsAPI) GetCostBreakdown(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	timeRange := r.URL.Query().Get("range")
	workflowID := r.URL.Query().Get("workflow_id")
	model := r.URL.Query().Get("model")

	start, end := parseTimeRange(timeRange)

	var filter *cost.CostBreakdownFilter
	if workflowID != "" || model != "" {
		filter = &cost.CostBreakdownFilter{WorkflowID: workflowID, Model: model}
	}

	breakdown, err := api.costTracker.GetCostBreakdown(r.Context(), tenantID, start, end, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(breakdown)
}

// GetCostsByWorkflow handles GET /v1/analytics/costs/workflows
func (api *AnalyticsAPI) GetCostsByWorkflow(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	timeRange := r.URL.Query().Get("range")

	start, end := parseTimeRange(timeRange)

	costs, err := api.costTracker.GetCostsByWorkflow(r.Context(), tenantID, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(costs)
}

// GetModelPerformance handles GET /v1/analytics/model-performance
func (api *AnalyticsAPI) GetModelPerformance(w http.ResponseWriter, r *http.Request) {
	if api.perfStore == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"snapshots": []interface{}{}, "message": "model performance store not configured"})
		return
	}
	provider := r.URL.Query().Get("provider")
	modelID := r.URL.Query().Get("model_id")
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if provider == "" {
		provider = "openai"
	}
	if modelID == "" {
		modelID = "gpt-4"
	}

	snapshots, err := api.perfStore.GetRecentSnapshots(r.Context(), provider, modelID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out := make([]map[string]interface{}, 0, len(snapshots))
	for _, s := range snapshots {
		out = append(out, map[string]interface{}{
			"model_id":       s.ModelID,
			"provider":       s.Provider,
			"latency_p50_ms": s.LatencyP50Ms,
			"latency_p95_ms": s.LatencyP95Ms,
			"latency_p99_ms": s.LatencyP99Ms,
			"error_rate":     s.ErrorRate,
			"throughput":     s.Throughput,
			"sample_count":   s.SampleCount,
			"timestamp":      s.Timestamp.Format(time.RFC3339),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"snapshots": out})
}

// GetModelDrift handles GET /v1/analytics/model-drift
func (api *AnalyticsAPI) GetModelDrift(w http.ResponseWriter, r *http.Request) {
	if api.perfStore == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"drifted": false, "message": "model performance store not configured"})
		return
	}
	provider := r.URL.Query().Get("provider")
	modelID := r.URL.Query().Get("model_id")
	if provider == "" {
		provider = "openai"
	}
	if modelID == "" {
		modelID = "gpt-4"
	}

	result, err := api.perfStore.DetectDrift(r.Context(), provider, modelID, 1.5, 0.05)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"model_id":      result.ModelID,
		"provider":      result.Provider,
		"drifted":       result.Drifted,
		"latency_delta": result.LatencyDelta,
		"error_delta":   result.ErrorDelta,
		"message":       result.Message,
	})
}

// GetSLA handles GET /v1/analytics/sla
func (api *AnalyticsAPI) GetSLA(w http.ResponseWriter, r *http.Request) {
	// Mock SLA data for multiple services
	services := []map[string]interface{}{
		{
			"service_id":  "agent-runtime",
			"uptime":      99.95,
			"avg_latency": 145.0,
			"error_rate":  0.002,
		},
		{
			"service_id":  "workflow-engine",
			"uptime":      99.98,
			"avg_latency": 230.0,
			"error_rate":  0.001,
		},
		{
			"service_id":  "policy-engine",
			"uptime":      99.99,
			"avg_latency": 85.0,
			"error_rate":  0.0005,
		},
	}

	response := map[string]interface{}{
		"services": services,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GenerateSLAReport handles GET /v1/analytics/sla/report
func (api *AnalyticsAPI) GenerateSLAReport(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("service_id")
	tenantID := r.URL.Query().Get("tenant_id")
	timeRange := r.URL.Query().Get("range")

	start, end := parseTimeRange(timeRange)

	report, err := api.slaMonitor.GenerateReport(r.Context(), serviceID, tenantID, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// SetSLATarget handles POST /v1/analytics/sla/targets
func (api *AnalyticsAPI) SetSLATarget(w http.ResponseWriter, r *http.Request) {
	var target monitoring.SLATarget

	if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	api.slaMonitor.SetTarget(&target)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "created",
	})
}

// GetPerformance handles GET /v1/analytics/performance
func (api *AnalyticsAPI) GetPerformance(w http.ResponseWriter, r *http.Request) {
	// Mock performance data timeline
	timeline := []map[string]interface{}{
		{"timestamp": "2026-02-26T14:00:00Z", "cpu": 45.0, "memory": 62.0, "throughput": 150},
		{"timestamp": "2026-02-26T15:00:00Z", "cpu": 52.0, "memory": 68.0, "throughput": 180},
		{"timestamp": "2026-02-26T16:00:00Z", "cpu": 38.0, "memory": 55.0, "throughput": 120},
		{"timestamp": "2026-02-26T17:00:00Z", "cpu": 65.0, "memory": 75.0, "throughput": 220},
		{"timestamp": "2026-02-26T18:00:00Z", "cpu": 48.0, "memory": 64.0, "throughput": 165},
	}

	response := map[string]interface{}{
		"timeline": timeline,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMetrics handles GET /v1/analytics/metrics
func (api *AnalyticsAPI) GetMetrics(w http.ResponseWriter, r *http.Request) {
	// Real-time metrics
	metrics := map[string]interface{}{
		"active_agents":    12,
		"active_workflows": 5,
		"requests_per_min": 450,
		"avg_latency_ms":   145,
		"error_rate":       0.002,
		"cost_per_hour":    2.45,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetPolicyDenials handles GET /v1/analytics/denials
func (api *AnalyticsAPI) GetPolicyDenials(w http.ResponseWriter, r *http.Request) {
	if api.auditLogger == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"logs": []interface{}{},
		})
		return
	}
	timeRange := r.URL.Query().Get("range")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	start, end := parseTimeRange(timeRange)
	filter := policy.AuditFilter{
		Decision:  "deny",
		StartTime: start,
		EndTime:   end,
		Limit:     limit,
	}
	logs, err := api.auditLogger.Query(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"logs": logs})
}

// GetPolicyDenialsStats handles GET /v1/analytics/denials/stats
func (api *AnalyticsAPI) GetPolicyDenialsStats(w http.ResponseWriter, r *http.Request) {
	if api.auditLogger == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"denied":           0,
			"allowed":          0,
			"top_deny_reasons": []interface{}{},
		})
		return
	}
	timeRange := r.URL.Query().Get("range")
	start, end := parseTimeRange(timeRange)
	stats, err := api.auditLogger.GetStats(r.Context(), start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// ExportAuditLogs handles GET /v1/compliance/audit/export for SIEM and compliance reporting.
// Query params: format=json|csv|cef, range=1h|24h|7d|30d, limit=, agent_name=, decision=.
func (api *AnalyticsAPI) ExportAuditLogs(w http.ResponseWriter, r *http.Request) {
	if api.auditLogger == nil {
		http.Error(w, "audit logging not enabled", http.StatusNotImplemented)
		return
	}
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}
	timeRange := r.URL.Query().Get("range")
	start, end := parseTimeRange(timeRange)
	filter := policy.AuditFilter{
		AgentName:  r.URL.Query().Get("agent_name"),
		PolicyName: r.URL.Query().Get("policy_name"),
		Decision:   r.URL.Query().Get("decision"),
		StartTime:  start,
		EndTime:    end,
		Limit:      10000,
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			filter.Limit = l
		}
	}
	data, err := api.auditLogger.Export(r.Context(), filter, format)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch format {
	case "cef":
		w.Header().Set("Content-Type", "application/cef")
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
	default:
		w.Header().Set("Content-Type", "application/json")
	}
	w.Write(data)
}

// parseTimeRange parses a time range string and returns start/end times
func parseTimeRange(timeRange string) (time.Time, time.Time) {
	end := time.Now()
	var start time.Time

	switch timeRange {
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
