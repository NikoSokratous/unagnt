package orchestrate

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/NikoSokratous/unagnt/internal/config"
	"github.com/NikoSokratous/unagnt/pkg/llm"
	"github.com/NikoSokratous/unagnt/pkg/llm/anthropic"
	"github.com/NikoSokratous/unagnt/pkg/llm/ollama"
	"github.com/NikoSokratous/unagnt/pkg/llm/openai"
	"github.com/NikoSokratous/unagnt/pkg/mcp"
	"github.com/NikoSokratous/unagnt/pkg/policy"
	"github.com/NikoSokratous/unagnt/pkg/runtime"
	"github.com/NikoSokratous/unagnt/pkg/tool"
	"github.com/NikoSokratous/unagnt/pkg/tool/builtin"
	_ "modernc.org/sqlite"
)

// RuntimeStepExecutor executes workflow steps with the real runtime engine.
type RuntimeStepExecutor struct {
	AllowSimulatedFallback bool
	StorePath              string
	ApprovalWebhook        string
	ApprovalQueue          policy.ApprovalQueue
}

// ExecuteStep runs a real agent runtime from an agent config path/name.
func (e *RuntimeStepExecutor) ExecuteStep(ctx context.Context, agentName, goal string, outputs map[string]interface{}) (*StepResult, error) {
	startedAt := time.Now()

	agentPath, err := resolveAgentConfigPath(agentName)
	if err != nil {
		if e.AllowSimulatedFallback {
			return (SimulatedExecutor{}).ExecuteStep(ctx, agentName, goal, outputs)
		}
		return nil, err
	}

	cfg, policyPath, err := config.LoadWithPolicy(agentPath)
	if err != nil {
		if e.AllowSimulatedFallback {
			return (SimulatedExecutor{}).ExecuteStep(ctx, agentName, goal, outputs)
		}
		return nil, fmt.Errorf("load agent config %s: %w", agentPath, err)
	}

	selectedModel := pickModelConfig(cfg, goal)
	provider, err := newLLMProvider(selectedModel)
	if err != nil {
		return nil, err
	}

	reg := tool.NewRegistry()
	for _, t := range builtin.All() {
		reg.Register(t)
	}

	mcpClients := make([]*mcp.MCPClient, 0, len(cfg.MCPSources))
	for _, mcpCfg := range cfg.MCPSources {
		client, loadErr := mcp.LoadMCPSource(ctx, reg, mcp.MCPSourceConfig{
			Type:       mcpCfg.Type,
			Command:    mcpCfg.Command,
			Args:       mcpCfg.Args,
			URL:        mcpCfg.URL,
			ToolPrefix: mcpCfg.ToolPrefix,
		})
		if loadErr != nil {
			return nil, fmt.Errorf("load MCP source %s: %w", mcpCfg.Type, loadErr)
		}
		mcpClients = append(mcpClients, client)
	}
	defer closeMCPClients(mcpClients)

	toolInfos := listToolInfos(cfg, reg)
	planner := &llm.PlannerAdapter{Provider: provider, Tools: toolInfos}
	baseExec := tool.NewExecutor(reg)

	var exec runtime.ToolExecutor = baseExec
	if policyPath != "" {
		pol, loadErr := policy.LoadEngine(policyPath)
		if loadErr != nil {
			return nil, fmt.Errorf("load policy: %w", loadErr)
		}
		approvalCb := webhookApprovalCallback(e.ApprovalWebhook)
		if e.ApprovalQueue != nil {
			approvalCb = queueApprovalCallback(e.ApprovalQueue)
		}
		exec = &tool.PolicyExecutor{
			Inner:       baseExec,
			Policy:      pol,
			RiskScorer:  policy.NewDefaultRiskScorer(),
			Approval:    policy.NewApprovalGate(approvalCb),
			Environment: "production",
		}
	}

	engine := runtime.NewEngine(runtime.EngineConfig{
		AgentName: cfg.Name,
		Goal:      goal,
		MaxSteps:  cfg.MaxSteps,
		Timeout:   cfg.TimeoutDuration(),
		Autonomy:  cfg.AutonomyLevel(),
	}, planner, exec)

	state, runErr := engine.Run(ctx)
	if runErr != nil {
		return &StepResult{
			Name:        agentName,
			Agent:       cfg.Name,
			Status:      "failed",
			RunID:       state.RunID,
			Error:       runErr.Error(),
			StartedAt:   startedAt,
			CompletedAt: time.Now(),
			Duration:    time.Since(startedAt),
		}, runErr
	}

	if e.StorePath != "" {
		_ = persistRunState(ctx, e.StorePath, state)
	}

	return &StepResult{
		Name:   agentName,
		Agent:  cfg.Name,
		Status: "completed",
		RunID:  state.RunID,
		Output: map[string]interface{}{
			"state":      state.Current,
			"step_count": state.StepCount,
			"model": map[string]any{
				"provider": selectedModel.Provider,
				"name":     selectedModel.Name,
			},
			"last_result": func() map[string]any {
				if state.LastResult == nil {
					return map[string]any{}
				}
				return state.LastResult.Output
			}(),
		},
		StartedAt:   startedAt,
		CompletedAt: time.Now(),
		Duration:    time.Since(startedAt),
	}, nil
}

func resolveAgentConfigPath(agentName string) (string, error) {
	candidates := []string{
		agentName,
		agentName + ".yaml",
		filepath.Join("examples", agentName, "agent.yaml"),
		filepath.Join("examples", agentName+".yaml"),
	}

	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c, nil
		}
	}

	return "", fmt.Errorf("agent config not found for %q", agentName)
}

func newLLMProvider(model config.ModelConfig) (llm.Provider, error) {
	switch model.Provider {
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is required")
		}
		return openai.NewClient(apiKey, model.Name), nil
	case "anthropic":
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY is required")
		}
		return anthropic.NewClient(apiKey, model.Name), nil
	case "ollama":
		return ollama.NewClient(model.Name), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", model.Provider)
	}
}

func listToolInfos(cfg *config.AgentConfig, reg *tool.Registry) []llm.ToolInfo {
	if len(cfg.Tools) == 0 {
		keys := reg.List()
		out := make([]llm.ToolInfo, 0, len(keys))
		for _, key := range keys {
			name := key
			version := "1"
			for i := len(key) - 1; i >= 0; i-- {
				if key[i] == '@' {
					name = key[:i]
					version = key[i+1:]
					break
				}
			}
			if t, ok := reg.Get(name, version); ok {
				schema, _ := t.InputSchema()
				out = append(out, llm.ToolInfo{Name: t.Name(), Version: t.Version(), Description: t.Description(), InputSchema: schema})
			}
		}
		return out
	}

	out := make([]llm.ToolInfo, 0, len(cfg.Tools))
	for _, tr := range cfg.Tools {
		desc := ""
		var schema []byte
		if t, ok := reg.Get(tr.Name, tr.Version); ok {
			desc = t.Description()
			schema, _ = t.InputSchema()
		}
		out = append(out, llm.ToolInfo{Name: tr.Name, Version: tr.Version, Description: desc, InputSchema: schema})
	}
	return out
}

func closeMCPClients(clients []*mcp.MCPClient) {
	for _, c := range clients {
		_ = c.Close()
	}
}

func persistRunState(ctx context.Context, storePath string, state *runtime.AgentState) error {
	db, err := sql.Open("sqlite", storePath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, _ = db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS runs (
    run_id TEXT PRIMARY KEY,
    agent_name TEXT NOT NULL,
    goal TEXT NOT NULL,
    state TEXT NOT NULL,
    step_count INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
`)

	_, err = db.ExecContext(ctx,
		`INSERT OR REPLACE INTO runs (run_id, agent_name, goal, state, step_count, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		state.RunID, state.AgentName, state.Goal, string(state.Current), state.StepCount, state.CreatedAt, state.UpdatedAt,
	)
	return err
}

func queueApprovalCallback(q policy.ApprovalQueue) policy.ApprovalCallback {
	return func(ctx context.Context, tool string, input map[string]any, approvers []string) (bool, error) {
		id, err := q.Enqueue(ctx, tool, input, approvers, "", "")
		if err != nil {
			return false, err
		}
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return false, ctx.Err()
			case <-ticker.C:
				req, err := q.Get(ctx, id)
				if err != nil {
					return false, err
				}
				if req == nil {
					return false, nil
				}
				switch req.Status {
				case "approved":
					return true, nil
				case "denied":
					return false, nil
				}
			}
		}
	}
}

func webhookApprovalCallback(url string) policy.ApprovalCallback {
	if url == "" {
		return func(ctx context.Context, tool string, input map[string]any, approvers []string) (bool, error) {
			return true, nil
		}
	}

	return func(ctx context.Context, tool string, input map[string]any, approvers []string) (bool, error) {
		body, _ := json.Marshal(map[string]any{"tool": tool, "input": input, "approvers": approvers})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return false, err
		}
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: 5 * time.Minute}
		resp, err := client.Do(req)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()
		var result struct {
			Approved bool `json:"approved"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return false, err
		}
		return result.Approved, nil
	}
}
