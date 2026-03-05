package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newInitCmd() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new agent configuration",
		Long:  "Interactive wizard to create agent.yaml and policy.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(outputDir)
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory")
	return cmd
}

func runInit(outputDir string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Agent Runtime Configuration Wizard")
	fmt.Println("===================================")
	fmt.Println()

	// Agent name
	fmt.Print("Agent name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		name = "my-agent"
	}

	// LLM Provider
	fmt.Print("LLM provider (openai/anthropic/ollama) [openai]: ")
	provider, _ := reader.ReadString('\n')
	provider = strings.TrimSpace(provider)
	if provider == "" {
		provider = "openai"
	}

	// Model name
	defaultModel := "gpt-4o-mini"
	if provider == "anthropic" {
		defaultModel = "claude-3-5-sonnet-20241022"
	} else if provider == "ollama" {
		defaultModel = "llama2"
	}

	fmt.Printf("Model name [%s]: ", defaultModel)
	model, _ := reader.ReadString('\n')
	model = strings.TrimSpace(model)
	if model == "" {
		model = defaultModel
	}

	// Autonomy level
	fmt.Print("Autonomy level (0=manual, 1=cautious, 2=standard, 3=autonomous, 4=unrestricted) [2]: ")
	autonomyStr, _ := reader.ReadString('\n')
	autonomyStr = strings.TrimSpace(autonomyStr)
	autonomy := 2
	if autonomyStr != "" {
		fmt.Sscanf(autonomyStr, "%d", &autonomy)
	}

	// Tools
	fmt.Print("Include built-in tools? (y/n) [y]: ")
	includeTools, _ := reader.ReadString('\n')
	includeTools = strings.TrimSpace(strings.ToLower(includeTools))
	if includeTools == "" {
		includeTools = "y"
	}

	// Create agent config
	agentConfig := map[string]any{
		"name":        name,
		"version":     "1.0",
		"description": fmt.Sprintf("Agent: %s", name),
		"model": map[string]any{
			"provider":    provider,
			"name":        model,
			"temperature": 0.7,
		},
		"autonomy_level": autonomy,
		"max_steps":      10,
		"timeout":        "60s",
		"policy":         "./policy.yaml",
	}

	if includeTools == "y" {
		agentConfig["tools"] = []map[string]any{
			{"name": "echo", "version": "1"},
			{"name": "calc", "version": "1"},
			{"name": "http_request", "version": "1"},
		}
	}

	// Write agent config
	agentPath := fmt.Sprintf("%s/%s.yaml", outputDir, name)
	agentData, _ := yaml.Marshal(agentConfig)
	if err := os.WriteFile(agentPath, agentData, 0644); err != nil {
		return err
	}

	// Create basic policy
	policyConfig := map[string]any{
		"version": "1",
		"rules": []map[string]any{
			{
				"name": "high-risk-approval",
				"match": map[string]any{
					"risk_score": ">= 0.8",
				},
				"action":  "require_approval",
				"message": "High-risk action requires approval",
			},
		},
	}

	policyPath := fmt.Sprintf("%s/policy.yaml", outputDir)
	policyData, _ := yaml.Marshal(policyConfig)
	if err := os.WriteFile(policyPath, policyData, 0644); err != nil {
		return err
	}

	fmt.Printf("\n%s Created configuration files:\n", green("✓"))
	fmt.Printf("  - %s\n", agentPath)
	fmt.Printf("  - %s\n", policyPath)

	fmt.Printf("\n%s Next steps:\n", yellow("→"))
	fmt.Printf("  1. Set your API key: export %s_API_KEY=...\n", strings.ToUpper(provider))
	fmt.Printf("  2. Run your agent: unagnt run --config %s --goal \"<your goal>\"\n", agentPath)

	return nil
}
