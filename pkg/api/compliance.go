package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/risk"
	"github.com/gorilla/mux"
)

// ComplianceAPI handles compliance report endpoints
type ComplianceAPI struct {
	generator *risk.ReportGenerator
}

// NewComplianceAPI creates a compliance API
func NewComplianceAPI(generator *risk.ReportGenerator) *ComplianceAPI {
	return &ComplianceAPI{generator: generator}
}

// RegisterRoutes registers compliance routes
func (api *ComplianceAPI) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/v1/compliance/reports/generate", api.GenerateReport).Methods("POST")
	router.HandleFunc("/v1/compliance/reports", api.ListReports).Methods("GET")
	router.HandleFunc("/v1/compliance/reports/{id}", api.GetReport).Methods("GET")
	router.HandleFunc("/v1/compliance/reports/{id}/export", api.ExportReport).Methods("GET")
}

// GenerateReport handles POST /v1/compliance/reports/generate
func (api *ComplianceAPI) GenerateReport(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type  string `json:"type"` // daily, weekly, monthly, custom
		Date  string `json:"date"`
		Start string `json:"start"`
		End   string `json:"end"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Type = "daily"
	}

	ctx := r.Context()
	var report *risk.ComplianceReport
	var err error

	switch req.Type {
	case "daily":
		date := time.Now()
		if req.Date != "" {
			date, _ = time.Parse("2006-01-02", req.Date)
		}
		report, err = api.generator.GenerateDailyReport(ctx, date)
	case "weekly":
		weekStart := time.Now()
		if req.Date != "" {
			weekStart, _ = time.Parse("2006-01-02", req.Date)
		}
		report, err = api.generator.GenerateWeeklyReport(ctx, weekStart)
	case "monthly":
		now := time.Now()
		if req.Date != "" {
			t, _ := time.Parse("2006-01-02", req.Date)
			now = t
		}
		report, err = api.generator.GenerateMonthlyReport(ctx, now.Year(), now.Month())
	case "custom":
		start, e1 := time.Parse(time.RFC3339, req.Start)
		end, e2 := time.Parse(time.RFC3339, req.End)
		if e1 != nil || e2 != nil {
			http.Error(w, "custom requires start and end in RFC3339", http.StatusBadRequest)
			return
		}
		report, err = api.generator.GenerateCustomReport(ctx, start, end)
	default:
		report, err = api.generator.GenerateDailyReport(ctx, time.Now())
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// ListReports handles GET /v1/compliance/reports
func (api *ComplianceAPI) ListReports(w http.ResponseWriter, r *http.Request) {
	reportType := r.URL.Query().Get("type")
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	reports, err := api.generator.ListReports(r.Context(), reportType, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"reports": reports})
}

// GetReport handles GET /v1/compliance/reports/{id}
func (api *ComplianceAPI) GetReport(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	report, err := api.generator.GetReport(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if report == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// ExportReport handles GET /v1/compliance/reports/{id}/export
func (api *ComplianceAPI) ExportReport(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	ef := risk.ExportFormat(format)
	data, err := api.generator.ExportReport(r.Context(), id, ef)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch ef {
	case risk.ExportFormatCEF:
		w.Header().Set("Content-Type", "application/cef")
	case risk.ExportFormatCSV:
		w.Header().Set("Content-Type", "text/csv")
	default:
		w.Header().Set("Content-Type", "application/json")
	}
	w.Write(data)
}
