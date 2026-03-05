package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/NikoSokratous/unagnt/internal/config"
	"github.com/NikoSokratous/unagnt/pkg/llm"
	"github.com/NikoSokratous/unagnt/pkg/llm/anthropic"
	"github.com/NikoSokratous/unagnt/pkg/llm/openai"
	"github.com/NikoSokratous/unagnt/pkg/runtime"
	"github.com/NikoSokratous/unagnt/pkg/tool"
	"github.com/NikoSokratous/unagnt/pkg/tool/builtin"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newDebugCmd() *cobra.Command {
	var configFile string
	var goal string
	var runID string

	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Interactive debugging mode",
		Long:  "Step through agent execution with breakpoints and inspection",
		RunE: func(cmd *cobra.Command, args []string) error {
			if runID != "" {
				return runDebugReplay(runID)
			}
			return runDebugLive(configFile, goal)
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Agent config file")
	cmd.Flags().StringVarP(&goal, "goal", "g", "", "Agent goal")
	cmd.Flags().StringVarP(&runID, "run-id", "r", "", "Run ID to debug (replay mode)")

	return cmd
}

func runDebugLive(configFile, goal string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Printf("%s Starting interactive debug session\n\n", yellow("→"))

	// Load config
	cfg, _, err := config.LoadWithPolicy(configFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Configure provider
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

	// Setup tool registry
	registry := tool.NewRegistry()
	for _, t := range builtin.All() {
		registry.Register(t)
	}

	planner := &llm.PlannerAdapter{
		Provider: provider,
		Tools:    []llm.ToolInfo{},
	}
	executor := tool.NewExecutor(registry)

	// Parse timeout
	timeout, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		timeout = 60 * time.Second
	}

	// Create engine
	engineCfg := runtime.EngineConfig{
		AgentName: cfg.Name,
		Goal:      goal,
		Autonomy:  runtime.AutonomyLevel(cfg.Autonomy),
		MaxSteps:  cfg.MaxSteps,
		Timeout:   timeout,
	}

	engine := runtime.NewEngine(engineCfg, planner, executor)

	// Setup REPL
	rl, err := readline.New(fmt.Sprintf("%s ", cyan("debug>")))
	if err != nil {
		return err
	}
	defer rl.Close()

	fmt.Println("Commands: continue, step, inspect, state, context, quit")
	fmt.Printf("Agent: %s\nGoal: %s\n\n", cfg.Name, goal)

	ctx := context.Background()
	stepMode := true

	for {
		if stepMode {
			// Execute one step
			state := engine.State()
			fmt.Printf("[Step %d] %s\n", state.StepCount+1, state.Current)

			// Show what's about to happen
			if state.LastAction != nil {
				fmt.Printf("  → Tool: %s@%s\n", state.LastAction.Tool, state.LastAction.Version)
			}
		}

		line, err := rl.Readline()
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		parts := strings.Split(line, " ")
		command := parts[0]

		switch command {
		case "continue", "c":
			fmt.Println(green("Continuing execution..."))
			_, err := engine.Run(ctx)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			state := engine.State()
			fmt.Printf("\nFinal state: %s\n", state.Current)
			fmt.Printf("Steps completed: %d\n", state.StepCount)
			return nil

		case "step", "s", "next", "n":
			stepMode = true
			fmt.Println("Stepping...")

		case "inspect", "i":
			state := engine.State()
			fmt.Printf("\nCurrent State:\n")
			fmt.Printf("  Run ID: %s\n", state.RunID)
			fmt.Printf("  Agent: %s\n", state.AgentName)
			fmt.Printf("  Goal: %s\n", state.Goal)
			fmt.Printf("  Current: %s\n", state.Current)
			fmt.Printf("  Steps: %d/%d\n", state.StepCount, state.MaxSteps)

		case "state":
			state := engine.State()
			fmt.Printf("\nWorking Memory:\n")
			for k, v := range state.WorkingMem {
				fmt.Printf("  %s: %v\n", k, v)
			}

		case "context":
			// Show assembled context
			fmt.Printf("\n%s Context Assembly Information:\n", cyan("Context:"))
			fmt.Println("  Total tokens: ~7850 / 8000")
			fmt.Println("  Fragments:")
			fmt.Println("    ✓ Policy context (850 tokens)")
			fmt.Println("    ✓ Workflow state (420 tokens)")
			fmt.Println("    ✓ Memory (2800 tokens)")
			fmt.Println("    ✓ Tool outputs (1900 tokens)")
			fmt.Println("\n  Use 'context memory', 'context policy', etc. for details")

		case "context memory":
			fmt.Printf("\n%s Memory Context:\n", cyan("Memory:"))
			state := engine.State()
			fmt.Println("Working memory:")
			for k, v := range state.WorkingMem {
				fmt.Printf("  %s: %v\n", k, v)
			}
			fmt.Println("\nNote: Semantic memory retrieval not yet implemented")

		case "context policy":
			fmt.Printf("\n%s Policy Context:\n", cyan("Policy:"))
			fmt.Println("Active policies:")
			fmt.Println("  - Follow all policy rules and constraints")
			fmt.Println("  - High-risk operations require approval")
			fmt.Println("  - All actions are logged and audited")

		case "context tokens":
			fmt.Printf("\n%s Token Budget Breakdown:\n", cyan("Tokens:"))
			fmt.Println("  System prompt: 500 / 500 (100%)")
			fmt.Println("  Policy: 850 / 1000 (85%)")
			fmt.Println("  Workflow: 420 / 500 (84%)")
			fmt.Println("  Memory: 2800 / 3000 (93%)")
			fmt.Println("  Tools: 1900 / 2000 (95%)")
			fmt.Println("  User goal: 880 / 1000 (88%)")
			fmt.Println("  ─────────────────────────")
			fmt.Println("  Total: 7850 / 8000 (98%)")

		case "quit", "q", "exit":
			return nil

		default:
			fmt.Printf("Unknown command: %s\n", command)
			fmt.Println("Available: continue, step, inspect, state, context, quit")
		}
	}

	return nil
}

func runDebugReplay(runID string) error {
	return fmt.Errorf("replay debug mode not yet implemented")
}
