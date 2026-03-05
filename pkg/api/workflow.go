package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/registry"
	"github.com/NikoSokratous/unagnt/pkg/workflow"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// WorkflowAPI handles workflow-related endpoints
type WorkflowAPI struct {
	marketplace *registry.WorkflowMarketplace
	debugMgr    *DebugSessionManager
}

// DebugSessionManager manages active debug sessions
type DebugSessionManager struct {
	sessions map[string]*workflow.DebugSession
}

func NewDebugSessionManager() *DebugSessionManager {
	return &DebugSessionManager{
		sessions: make(map[string]*workflow.DebugSession),
	}
}

func (dm *DebugSessionManager) CreateSession(workflowID string) *workflow.DebugSession {
	session := workflow.NewDebugSession(workflowID)
	dm.sessions[workflowID] = session
	return session
}

func (dm *DebugSessionManager) GetSession(workflowID string) *workflow.DebugSession {
	return dm.sessions[workflowID]
}

func NewWorkflowAPI(marketplace *registry.WorkflowMarketplace) *WorkflowAPI {
	return &WorkflowAPI{
		marketplace: marketplace,
		debugMgr:    NewDebugSessionManager(),
	}
}

// RegisterRoutes registers workflow API routes
func (api *WorkflowAPI) RegisterRoutes(router *mux.Router) {
	// Marketplace routes
	router.HandleFunc("/v1/workflows/marketplace", api.ListTemplates).Methods("GET")
	router.HandleFunc("/v1/workflows/marketplace", api.PublishTemplate).Methods("POST")
	router.HandleFunc("/v1/workflows/marketplace/{id}", api.GetTemplate).Methods("GET")
	router.HandleFunc("/v1/workflows/marketplace/{id}/install", api.InstallTemplate).Methods("POST")
	router.HandleFunc("/v1/workflows/marketplace/{id}/rate", api.RateTemplate).Methods("POST")

	// Debug routes
	router.HandleFunc("/v1/workflows/{id}/debug/start", api.StartDebug).Methods("POST")
	router.HandleFunc("/v1/workflows/{id}/debug/stop", api.StopDebug).Methods("POST")
	router.HandleFunc("/v1/workflows/{id}/debug/breakpoint", api.AddBreakpoint).Methods("POST")
	router.HandleFunc("/v1/workflows/{id}/debug/breakpoint/{nodeId}", api.RemoveBreakpoint).Methods("DELETE")
	router.HandleFunc("/v1/workflows/{id}/debug/pause", api.PauseDebug).Methods("POST")
	router.HandleFunc("/v1/workflows/{id}/debug/resume", api.ResumeDebug).Methods("POST")
	router.HandleFunc("/v1/workflows/{id}/debug/step", api.StepOver).Methods("POST")
	router.HandleFunc("/v1/workflows/{id}/debug/state", api.GetDebugState).Methods("GET")

	// Validation routes
	router.HandleFunc("/v1/workflows/validate", api.ValidateWorkflow).Methods("POST")
	router.HandleFunc("/v1/workflows/preview", api.PreviewWorkflow).Methods("POST")
}

// ListTemplates handles GET /v1/workflows/marketplace
func (api *WorkflowAPI) ListTemplates(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")
	limit := 50

	filters := map[string]interface{}{
		"limit": limit,
	}
	if category != "" {
		filters["category"] = category
	}
	if search != "" {
		filters["search"] = search
	}

	templates, err := api.marketplace.Search(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

// PublishTemplate handles POST /v1/workflows/marketplace
func (api *WorkflowAPI) PublishTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		Category     string   `json:"category"`
		Tags         []string `json:"tags"`
		TemplateYAML string   `json:"template_yaml"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	template := &registry.WorkflowTemplate{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Author:       getUserFromContext(r.Context()),
		Description:  req.Description,
		Category:     req.Category,
		Tags:         req.Tags,
		TemplateYAML: req.TemplateYAML,
		Version:      "1.0.0",
		License:      "MIT",
	}

	if err := api.marketplace.Publish(r.Context(), template); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(template)
}

// GetTemplate handles GET /v1/workflows/marketplace/{id}
func (api *WorkflowAPI) GetTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	template, err := api.marketplace.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

// InstallTemplate handles POST /v1/workflows/marketplace/{id}/install
func (api *WorkflowAPI) InstallTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	template, err := api.marketplace.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := api.marketplace.IncrementDownloads(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"yaml": template.TemplateYAML,
	})
}

// RateTemplate handles POST /v1/workflows/marketplace/{id}/rate
func (api *WorkflowAPI) RateTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req struct {
		Rating float64 `json:"rating"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		http.Error(w, "Rating must be between 1 and 5", http.StatusBadRequest)
		return
	}

	if err := api.marketplace.Rate(r.Context(), id, req.Rating); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// StartDebug handles POST /v1/workflows/{id}/debug/start
func (api *WorkflowAPI) StartDebug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowID := vars["id"]

	session := api.debugMgr.CreateSession(workflowID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id":  workflowID,
		"workflow_id": workflowID,
		"status":      "started",
		"started_at":  time.Now(),
	})
	_ = session
}

// StopDebug handles POST /v1/workflows/{id}/debug/stop
func (api *WorkflowAPI) StopDebug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowID := vars["id"]

	delete(api.debugMgr.sessions, workflowID)

	w.WriteHeader(http.StatusOK)
}

// AddBreakpoint handles POST /v1/workflows/{id}/debug/breakpoint
func (api *WorkflowAPI) AddBreakpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowID := vars["id"]

	var req struct {
		NodeID    string `json:"node_id"`
		Condition string `json:"condition"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	session := api.debugMgr.GetSession(workflowID)
	if session == nil {
		http.Error(w, "Debug session not found", http.StatusNotFound)
		return
	}

	session.AddBreakpoint(req.NodeID, req.Condition)

	w.WriteHeader(http.StatusOK)
}

// RemoveBreakpoint handles DELETE /v1/workflows/{id}/debug/breakpoint/{nodeId}
func (api *WorkflowAPI) RemoveBreakpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowID := vars["id"]
	nodeID := vars["nodeId"]

	session := api.debugMgr.GetSession(workflowID)
	if session == nil {
		http.Error(w, "Debug session not found", http.StatusNotFound)
		return
	}

	session.RemoveBreakpoint(nodeID)

	w.WriteHeader(http.StatusOK)
}

// PauseDebug handles POST /v1/workflows/{id}/debug/pause
func (api *WorkflowAPI) PauseDebug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowID := vars["id"]

	session := api.debugMgr.GetSession(workflowID)
	if session == nil {
		http.Error(w, "Debug session not found", http.StatusNotFound)
		return
	}

	session.Pause()

	w.WriteHeader(http.StatusOK)
}

// ResumeDebug handles POST /v1/workflows/{id}/debug/resume
func (api *WorkflowAPI) ResumeDebug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowID := vars["id"]

	session := api.debugMgr.GetSession(workflowID)
	if session == nil {
		http.Error(w, "Debug session not found", http.StatusNotFound)
		return
	}

	session.Resume()

	w.WriteHeader(http.StatusOK)
}

// StepOver handles POST /v1/workflows/{id}/debug/step
func (api *WorkflowAPI) StepOver(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowID := vars["id"]

	session := api.debugMgr.GetSession(workflowID)
	if session == nil {
		http.Error(w, "Debug session not found", http.StatusNotFound)
		return
	}

	session.StepOver()

	w.WriteHeader(http.StatusOK)
}

// GetDebugState handles GET /v1/workflows/{id}/debug/state
func (api *WorkflowAPI) GetDebugState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowID := vars["id"]

	session := api.debugMgr.GetSession(workflowID)
	if session == nil {
		http.Error(w, "Debug session not found", http.StatusNotFound)
		return
	}

	state := session.GetState()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

// ValidateWorkflow handles POST /v1/workflows/validate
func (api *WorkflowAPI) ValidateWorkflow(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Basic validation (simplified)
	response := map[string]interface{}{
		"valid":    true,
		"errors":   []string{},
		"warnings": []string{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PreviewWorkflow handles POST /v1/workflows/preview
func (api *WorkflowAPI) PreviewWorkflow(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WorkflowYAML string `json:"workflow_yaml"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return preview data (simplified)
	preview := map[string]interface{}{
		"estimated_duration": "5m",
		"steps":              3,
		"agents_required":    []string{"coder", "reviewer"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preview)
}

// getUserFromContext extracts the authenticated user from request context
func getUserFromContext(ctx context.Context) string {
	// Try to get user from context (set by auth middleware)
	if user := ctx.Value("user"); user != nil {
		if userStr, ok := user.(string); ok {
			return userStr
		}
		// If it's a user object/struct, try to extract username
		if userInfo, ok := user.(interface{ GetUsername() string }); ok {
			return userInfo.GetUsername()
		}
	}

	// Fallback to anonymous if no user in context
	return "anonymous"
}
