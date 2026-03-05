package api

import (
	"encoding/json"
	"net/http"

	"gopkg.in/yaml.v3"

	"github.com/NikoSokratous/unagnt/pkg/policy"
)

// PolicyPlaygroundAPI serves the policy playground endpoint.
type PolicyPlaygroundAPI struct{}

// CheckRequest is the request body for policy check.
type CheckRequest struct {
	PolicyYAML  string                 `json:"policy_yaml"`
	Tool        string                 `json:"tool"`
	Input       map[string]interface{} `json:"input"`
	Environment string                 `json:"environment"`
	RiskScore   float64                `json:"risk_score"`
}

// CheckResponse is the response for policy check.
type CheckResponse struct {
	Allow           bool     `json:"allow"`
	Deny            bool     `json:"deny"`
	RequireApproval bool     `json:"require_approval"`
	Message         string   `json:"message,omitempty"`
	Approvers       []string `json:"approvers,omitempty"`
}

// HandlePolicyCheck handles POST /v1/policy/check
func HandlePolicyCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.PolicyYAML == "" {
		http.Error(w, "policy_yaml is required", http.StatusBadRequest)
		return
	}
	if req.Tool == "" {
		http.Error(w, "tool is required", http.StatusBadRequest)
		return
	}

	if req.Input == nil {
		req.Input = make(map[string]interface{})
	}
	if req.RiskScore == 0 {
		req.RiskScore = 0.3
	}

	var cfg policy.PolicyConfig
	if err := yaml.Unmarshal([]byte(req.PolicyYAML), &cfg); err != nil {
		http.Error(w, "invalid policy YAML: "+err.Error(), http.StatusBadRequest)
		return
	}

	engine := policy.NewEngine(&cfg)
	inputMap := make(map[string]any)
	for k, v := range req.Input {
		inputMap[k] = v
	}

	ctx := policy.EvalContext{
		Tool:        req.Tool,
		Input:       inputMap,
		Environment: req.Environment,
		RiskScore:   req.RiskScore,
	}

	result := engine.Check(ctx)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CheckResponse{
		Allow:           result.Allow,
		Deny:            result.Deny,
		RequireApproval: result.RequireApproval,
		Message:         result.Message,
		Approvers:       result.Approvers,
	})
}
