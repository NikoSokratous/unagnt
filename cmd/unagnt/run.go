package main

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
	"github.com/NikoSokratous/unagnt/internal/store"
	"github.com/NikoSokratous/unagnt/pkg/cost"
	"github.com/NikoSokratous/unagnt/pkg/llm"
	"github.com/NikoSokratous/unagnt/pkg/llm/anthropic"
	"github.com/NikoSokratous/unagnt/pkg/llm/openai"
	"github.com/NikoSokratous/unagnt/pkg/mcp"
	"github.com/NikoSokratous/unagnt/pkg/observe"
	"github.com/NikoSokratous/unagnt/pkg/policy"
	"github.com/NikoSokratous/unagnt/pkg/runtime"
	"github.com/NikoSokratous/unagnt/pkg/tool"
	"github.com/NikoSokratous/unagnt/pkg/tool/builtin"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

func newRunCmd() *cobra.Command {
	var configPath, goal, logFile, storePath, approvalWebhook, costDB string
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run an agent with a goal",
		RunE: func(cmd *cobra.Command, args []string) error {
			if goal == "" && len(args) > 0 {
				goal = args[0]
			}
			return runRun(configPath, goal, logFile, storePath, approvalWebhook, costDB)
		},
	}
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to agent config YAML")
	cmd.Flags().StringVarP(&goal, "goal", "g", "", "Goal for the agent")
	cmd.Flags().StringVar(&logFile, "log-file", "", "Write events to file (in addition to stdout)")
	cmd.Flags().StringVar(&storePath, "store", "agent.db", "SQLite store path (empty to disable persistence)")
	cmd.Flags().StringVar(&approvalWebhook, "approval-webhook", "", "URL for HITL approval (POST request, blocks until approved/denied)")
	cmd.Flags().StringVar(&costDB, "cost-db", "", "Path to cost DB for budget checks (defaults to --store when set)")
	_ = cmd.MarkFlagRequired("config")
	_ = cmd.MarkFlagRequired("goal")
	return cmd
}

func runRun(configPath, goal, logFile, storePath, approvalWebhook, costDBPath string) error {
	cfg, policyPath, err := config.LoadWithPolicy(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	var provider llm.Provider
	switch cfg.Model.Provider {
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("OPENAI_API_KEY is required")
		}
		provider = openai.NewClient(apiKey, cfg.Model.Name)
	case "anthropic":
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("ANTHROPIC_API_KEY is required")
		}
		provider = anthropic.NewClient(apiKey, cfg.Model.Name)
	default:
		return fmt.Errorf("unsupported provider: %s", cfg.Model.Provider)
	}

	reg := tool.NewRegistry()
	for _, t := range builtin.All() {
		reg.Register(t)
	}

	var mcpClients []*mcp.MCPClient
	if len(cfg.MCPSources) > 0 {
		ctx := context.Background()
		for _, mcpCfg := range cfg.MCPSources {
			mcpCfg2 := mcp.MCPSourceConfig{
				Type:       mcpCfg.Type,
				Command:    mcpCfg.Command,
				Args:       mcpCfg.Args,
				URL:        mcpCfg.URL,
				ToolPrefix: mcpCfg.ToolPrefix,
			}
			client, err := mcp.LoadMCPSource(ctx, reg, mcpCfg2)
			if err != nil {
				return fmt.Errorf("load MCP source %s: %w", mcpCfg.Type, err)
			}
			mcpClients = append(mcpClients, client)
		}
	}
	defer func() {
		for _, c := range mcpClients {
			_ = c.Close()
		}
	}()

	toolInfos := make([]llm.ToolInfo, 0, len(cfg.Tools)+20)
	if len(cfg.Tools) > 0 {
		for _, tr := range cfg.Tools {
			t, ok := reg.Get(tr.Name, tr.Version)
			desc := ""
			if ok {
				desc = t.Description()
			}
			toolInfos = append(toolInfos, llm.ToolInfo{Name: tr.Name, Version: tr.Version, Description: desc})
		}
	} else {
		for _, key := range reg.List() {
			var name, ver string
			for i := len(key) - 1; i >= 0; i-- {
				if key[i] == '@' {
					name = key[:i]
					ver = key[i+1:]
					break
				}
			}
			if name == "" {
				name = key
				ver = "1"
			}
			t, ok := reg.Get(name, ver)
			if ok {
				toolInfos = append(toolInfos, llm.ToolInfo{Name: t.Name(), Version: t.Version(), Description: t.Description()})
			}
		}
	}

	planner := &llm.PlannerAdapter{Provider: provider, Tools: toolInfos}
	baseExec := tool.NewExecutor(reg)
	var exec runtime.ToolExecutor = baseExec
	if policyPath != "" {
		pol, err := policy.LoadEngine(policyPath)
		if err != nil {
			return fmt.Errorf("load policy: %w", err)
		}
		var approval *policy.ApprovalGate
		if approvalWebhook != "" {
			approval = policy.NewApprovalGate(webhookApprovalCallback(approvalWebhook))
		} else {
			approval = policy.NewApprovalGate(func(ctx context.Context, tool string, input map[string]any, approvers []string) (bool, error) {
				fmt.Fprintf(os.Stderr, "Approval required for tool %q. Approve? [y/N]: ", tool)
				var answer string
				if _, err := fmt.Scanln(&answer); err != nil {
					return false, nil
				}
				return answer == "y" || answer == "Y", nil
			})
		}
		exec = &tool.PolicyExecutor{
			Inner:       baseExec,
			Policy:      pol,
			RiskScorer:  policy.NewDefaultRiskScorer(),
			Approval:    approval,
			Environment: "development",
		}
	}

	logger := observe.NewLogger(cfg.Name, os.Stdout)
	if logFile != "" {
		if _, err := logger.WithFile(logFile); err != nil {
			return fmt.Errorf("open log file: %w", err)
		}
	}

	if cfg.Budget != nil && cfg.Budget.Limit > 0 {
		dbPath := costDBPath
		if dbPath == "" {
			dbPath = storePath
		}
		if dbPath != "" {
			absPath, err := filepath.Abs(dbPath)
			if err == nil {
				db, err := sql.Open("sqlite", absPath)
				if err == nil {
					if _, err := db.ExecContext(context.Background(), costEntriesSchema); err == nil {
						ct := cost.NewCostTracker(db)
						ct.Start(context.Background())
						bg := cost.NewBudgetGuard(ct, cost.BudgetConfig{
							BudgetLimit:    cfg.Budget.Limit,
							AlertThreshold: cfg.Budget.AlertThreshold,
							AlertWebhook:   cfg.Budget.AlertWebhook,
							Period:         cfg.Budget.Period,
							TenantID:       cfg.Budget.TenantID,
						})
						ctx := context.Background()
						over, err := bg.CheckAndAlert(ctx, cfg.Name)
						if err == nil && over {
							return fmt.Errorf("budget limit exceeded (%.2f USD); run blocked", cfg.Budget.Limit)
						}
					}
					_ = db.Close()
				}
			}
		}
	}

	engineConfig := runtime.EngineConfig{
		AgentName: cfg.Name,
		Goal:      goal,
		MaxSteps:  cfg.MaxSteps,
		Timeout:   cfg.TimeoutDuration(),
		Autonomy:  cfg.AutonomyLevel(),
	}

	eng := runtime.NewEngine(engineConfig, planner, exec)
	logger.LogInit(eng.State().RunID, goal)

	ctx := context.Background()
	state, err := eng.Run(ctx)
	if err != nil {
		logger.LogError(state.RunID, "", err)
		return err
	}

	for _, step := range state.History {
		model := observe.ModelMeta{Provider: cfg.Model.Provider, Name: cfg.Model.Name}
		logger.LogPlan(state.RunID, step.StepID, step.Reasoning, model)
		if step.Action != nil {
			logger.LogToolCall(state.RunID, step.StepID, step.Action.Tool, step.Action.Version, step.Action.Input)
		}
		if step.Result != nil {
			duration := ""
			if step.Result.Duration > 0 {
				duration = step.Result.Duration.String()
			}
			logger.LogToolResult(state.RunID, step.StepID, step.Result.ToolID, step.Result.Output, step.Result.Error, duration)
		}
	}

	logger.LogCompleted(state.RunID)

	if storePath != "" {
		s, err := store.NewSQLite(storePath)
		if err == nil {
			defer s.Close()
			ctx := context.Background()
			_ = s.SaveRun(ctx, &store.RunMeta{
				RunID:     state.RunID,
				AgentName: state.AgentName,
				Goal:      state.Goal,
				State:     string(state.Current),
				StepCount: state.StepCount,
				CreatedAt: state.CreatedAt,
				UpdatedAt: state.UpdatedAt,
			})
			_ = s.SaveHistory(ctx, state.RunID, state.History)
		}
	}

	_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
		"run_id":     state.RunID,
		"state":      state.Current,
		"step_count": state.StepCount,
	})
	return nil
}

func webhookApprovalCallback(url string) policy.ApprovalCallback {
	return func(ctx context.Context, tool string, input map[string]any, approvers []string) (bool, error) {
		body, _ := json.Marshal(map[string]any{"tool": tool, "input": input, "approvers": approvers})
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
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
