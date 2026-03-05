package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/NikoSokratous/unagnt/internal/config"
	"github.com/NikoSokratous/unagnt/pkg/policy"
	"github.com/NikoSokratous/unagnt/pkg/tool"
	"github.com/NikoSokratous/unagnt/pkg/tool/builtin"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Agent management commands",
	}
	cmd.AddCommand(newAgentTestCmd())
	return cmd
}

func newAgentTestCmd() *cobra.Command {
	var configFile string
	var verbose bool

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test agent configuration",
		Long:  "Validates agent config and simulates execution without calling LLMs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAgentTest(configFile, verbose)
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Agent config file (required)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	cmd.MarkFlagRequired("config")

	return cmd
}

func runAgentTest(configFile string, verbose bool) error {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Printf("Testing agent config: %s\n\n", configFile)

	// Load config
	cfg, policyPath, err := config.LoadWithPolicy(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to load config: %v\n", red("✗"), err)
		return err
	}

	fmt.Printf("%s Config loaded successfully\n", green("✓"))

	// Validate agent config
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "%s Config validation failed: %v\n", red("✗"), err)
		return err
	}
	fmt.Printf("%s Agent config is valid\n", green("✓"))

	// Check model provider
	supportedProviders := []string{"openai", "anthropic", "ollama"}
	validProvider := false
	for _, p := range supportedProviders {
		if cfg.Model.Provider == p {
			validProvider = true
			break
		}
	}
	if !validProvider {
		fmt.Fprintf(os.Stderr, "%s Unsupported provider: %s\n", red("✗"), cfg.Model.Provider)
		return fmt.Errorf("unsupported provider")
	}
	fmt.Printf("%s Provider '%s' is supported\n", green("✓"), cfg.Model.Provider)

	// Check API key environment variable
	var apiKeyVar string
	switch cfg.Model.Provider {
	case "openai":
		apiKeyVar = "OPENAI_API_KEY"
	case "anthropic":
		apiKeyVar = "ANTHROPIC_API_KEY"
	}

	if apiKeyVar != "" {
		if os.Getenv(apiKeyVar) == "" {
			fmt.Fprintf(os.Stderr, "%s Environment variable %s not set\n", yellow("⚠"), apiKeyVar)
		} else {
			fmt.Printf("%s API key found (%s)\n", green("✓"), apiKeyVar)
		}
	}

	// Check tools
	registry := tool.NewRegistry()
	for _, t := range builtin.All() {
		registry.Register(t)
	}

	missingTools := []string{}
	for _, toolRef := range cfg.Tools {
		if _, ok := registry.Get(toolRef.Name, toolRef.Version); !ok {
			missingTools = append(missingTools, fmt.Sprintf("%s@%s", toolRef.Name, toolRef.Version))
		}
	}

	if len(missingTools) > 0 {
		fmt.Fprintf(os.Stderr, "%s Missing tools: %v\n", red("✗"), missingTools)
		return fmt.Errorf("tools not found")
	}
	fmt.Printf("%s All %d tools are available\n", green("✓"), len(cfg.Tools))

	// Check policy if specified
	if policyPath != "" {
		_, err := os.Stat(policyPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Policy file not found: %s\n", red("✗"), policyPath)
			return err
		}

		// Try to load policy
		_, err = policy.LoadEngine(policyPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Policy validation failed: %v\n", red("✗"), err)
			return err
		}
		fmt.Printf("%s Policy is valid\n", green("✓"))
	}

	// Simulate execution
	if verbose {
		fmt.Println("\nSimulated Execution Plan:")
		fmt.Printf("  Agent: %s\n", cfg.Name)
		fmt.Printf("  Model: %s/%s\n", cfg.Model.Provider, cfg.Model.Name)
		fmt.Printf("  Max Steps: %d\n", cfg.MaxSteps)
		fmt.Printf("  Autonomy: Level %d\n", cfg.Autonomy)
		fmt.Printf("  Timeout: %s\n", cfg.Timeout)
		fmt.Println("\n  Tools available:")
		for _, t := range cfg.Tools {
			fmt.Printf("    - %s@%s\n", t.Name, t.Version)
		}

		// Estimate costs (rough)
		fmt.Println("\n  Estimated Costs (approximate):")
		switch cfg.Model.Provider {
		case "openai":
			if strings.Contains(cfg.Model.Name, "gpt-4") {
				fmt.Printf("    ~$0.01-0.10 per run (depends on complexity)\n")
			} else {
				fmt.Printf("    ~$0.001-0.01 per run (depends on complexity)\n")
			}
		case "anthropic":
			fmt.Printf("    ~$0.01-0.08 per run (depends on complexity)\n")
		case "ollama":
			fmt.Printf("    Free (local model)\n")
		}
	}

	fmt.Printf("\n%s Agent config is ready to use\n", green("✓"))
	fmt.Printf("\nTo run: unagnt run --config %s --goal \"<your goal>\"\n", configFile)

	return nil
}
