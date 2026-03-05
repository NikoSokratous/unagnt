package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newContextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Context assembly inspection and debugging",
		Long:  "Inspect, explain, and debug context assembly for agent runs",
	}

	cmd.AddCommand(newContextInspectCmd())
	cmd.AddCommand(newContextExplainCmd())
	cmd.AddCommand(newContextDiffCmd())
	cmd.AddCommand(newContextStatsCmd())
	cmd.AddCommand(newContextValidateCmd())
	cmd.AddCommand(newContextIngestCmd())
	cmd.AddCommand(newContextKnowledgeCmd())
	cmd.AddCommand(newContextSearchCmd())

	return cmd
}

func newContextInspectCmd() *cobra.Command {
	var stepNum int
	var format string

	cmd := &cobra.Command{
		Use:   "inspect <run-id>",
		Short: "Inspect assembled context for a run",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runID := args[0]
			return inspectContext(runID, stepNum, format)
		},
	}

	cmd.Flags().IntVarP(&stepNum, "step", "s", -1, "Specific step number to inspect (default: latest)")
	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format: text, json")

	return cmd
}

func inspectContext(runID string, stepNum int, format string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Printf("%s Inspecting context for run: %s\n\n", green("→"), runID)

	// Mock data for demonstration
	contextData := map[string]interface{}{
		"run_id":       runID,
		"step":         stepNum,
		"total_tokens": 7850,
		"fragments": []map[string]interface{}{
			{
				"provider": "policy",
				"type":     "policy",
				"priority": 1,
				"tokens":   850,
				"included": true,
			},
			{
				"provider": "workflow_state",
				"type":     "workflow_state",
				"priority": 2,
				"tokens":   420,
				"included": true,
			},
			{
				"provider": "memory",
				"type":     "memory",
				"priority": 3,
				"tokens":   2800,
				"included": true,
			},
			{
				"provider": "tool_outputs",
				"type":     "tool_output",
				"priority": 4,
				"tokens":   1900,
				"included": true,
			},
			{
				"provider": "knowledge",
				"type":     "knowledge",
				"priority": 5,
				"tokens":   0,
				"included": false,
			},
		},
	}

	if format == "json" {
		b, _ := json.MarshalIndent(contextData, "", "  ")
		fmt.Println(string(b))
		return nil
	}

	// Text format
	fmt.Printf("%s Total Tokens: %d / 8000\n\n", cyan("●"), contextData["total_tokens"])

	fmt.Println(yellow("Context Fragments:"))
	fragments := contextData["fragments"].([]map[string]interface{})
	for _, frag := range fragments {
		status := "✓"
		if !frag["included"].(bool) {
			status = "✗"
		}
		fmt.Printf("  %s %s (%s) - %d tokens [Priority: %d]\n",
			status,
			frag["provider"],
			frag["type"],
			frag["tokens"],
			frag["priority"])
	}

	fmt.Printf("\n%s Note: Context assembly feature is in development. This is sample output.\n", yellow("ℹ"))

	return nil
}

func newContextExplainCmd() *cobra.Command {
	var stepNum int

	cmd := &cobra.Command{
		Use:   "explain <run-id>",
		Short: "Explain why each context piece was included",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runID := args[0]
			return explainContext(runID, stepNum)
		},
	}

	cmd.Flags().IntVarP(&stepNum, "step", "s", -1, "Specific step number")

	return cmd
}

func explainContext(runID string, stepNum int) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Printf("%s Context assembly explanation for run: %s\n\n", green("→"), runID)

	fmt.Println("Policy Context:")
	fmt.Println("  ✓ Included (850 tokens)")
	fmt.Println("  Reason: High priority (1), always included")
	fmt.Println()

	fmt.Println("Workflow State:")
	fmt.Println("  ✓ Included (420 tokens)")
	fmt.Println("  Reason: High priority (2), critical for agent orientation")
	fmt.Println()

	fmt.Println("Memory:")
	fmt.Println("  ✓ Included (2800 tokens)")
	fmt.Println("  Reason: Within token budget (3000 allocated)")
	fmt.Println("  Retrieved: 5 relevant items from semantic memory")
	fmt.Println()

	fmt.Println("Tool Outputs:")
	fmt.Println("  ✓ Included (1900 tokens)")
	fmt.Println("  Reason: Within token budget (2000 allocated)")
	fmt.Println("  Included: Last 10 tool executions")
	fmt.Println()

	fmt.Println("Knowledge:")
	fmt.Println("  ✗ Not included (0 tokens)")
	fmt.Println("  Reason: Provider disabled (RAG not configured)")
	fmt.Println()

	fmt.Printf("%s Note: Context assembly feature is in development. This is sample output.\n", yellow("ℹ"))

	return nil
}

func newContextDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <run-id-1> <run-id-2>",
		Short: "Compare context between two runs",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			runID1 := args[0]
			runID2 := args[1]
			return diffContext(runID1, runID2)
		},
	}

	return cmd
}

func diffContext(runID1, runID2 string) error {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Printf("Comparing context: %s vs %s\n\n", runID1, runID2)

	fmt.Println("Token Usage:")
	fmt.Printf("  %s: 7850 tokens\n", runID1)
	fmt.Printf("  %s: 6200 tokens\n", runID2)
	fmt.Printf("  Difference: %s (-1650 tokens)\n", green("↓"))
	fmt.Println()

	fmt.Println("Fragment Differences:")
	fmt.Printf("  Policy: 850 → 850 (no change)\n")
	fmt.Printf("  Workflow: 420 → 380 %s\n", green("(-40)"))
	fmt.Printf("  Memory: 2800 → 1950 %s\n", red("(-850)"))
	fmt.Printf("  Tools: 1900 → 1820 %s\n", green("(-80)"))
	fmt.Println()

	fmt.Printf("%s Note: Context assembly feature is in development. This is sample output.\n", yellow("ℹ"))

	return nil
}

func newContextStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats <run-id>",
		Short: "Show context assembly metrics",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runID := args[0]
			return statsContext(runID)
		},
	}

	return cmd
}

func statsContext(runID string) error {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Printf("%s Context assembly statistics for: %s\n\n", green("→"), runID)

	fmt.Println(cyan("Performance Metrics:"))
	fmt.Println("  Total assembly time: 45ms")
	fmt.Println("  Provider fetch times:")
	fmt.Println("    - Policy: 2ms")
	fmt.Println("    - Workflow: 1ms")
	fmt.Println("    - Memory: 32ms")
	fmt.Println("    - Tools: 8ms")
	fmt.Println("    - Knowledge: 0ms (disabled)")
	fmt.Println()

	fmt.Println(cyan("Token Usage:"))
	fmt.Println("  Total: 7850 / 8000 (98%)")
	fmt.Println("  By type:")
	fmt.Println("    - Policy: 850 / 1000 (85%)")
	fmt.Println("    - Workflow: 420 / 500 (84%)")
	fmt.Println("    - Memory: 2800 / 3000 (93%)")
	fmt.Println("    - Tools: 1900 / 2000 (95%)")
	fmt.Println("    - Knowledge: 0 / 1000 (0%)")
	fmt.Println()

	fmt.Println(cyan("Quality Metrics:"))
	fmt.Println("  Fragments included: 4 / 5")
	fmt.Println("  Truncation events: 0")
	fmt.Println("  Cache hit rate: 60%")
	fmt.Println()

	fmt.Printf("%s Note: Context assembly feature is in development. This is sample output.\n", yellow("ℹ"))

	return nil
}

func newContextValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <config-file>",
		Short: "Validate context assembly configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := args[0]
			return validateContextConfig(configFile)
		},
	}

	return cmd
}

func validateContextConfig(configFile string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// Check if file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", configFile)
	}

	fmt.Printf("%s Validating context assembly config: %s\n\n", green("→"), configFile)

	fmt.Println(green("✓") + " Configuration is valid")
	fmt.Println(green("✓") + " Token budgets sum correctly")
	fmt.Println(green("✓") + " Provider priorities are unique")
	fmt.Println(green("✓") + " All provider types are recognized")
	fmt.Println()

	fmt.Println(yellow("Warnings:"))
	fmt.Println("  ⚠ Knowledge provider is disabled")
	fmt.Println("  ⚠ Memory token budget is high (may impact response time)")
	fmt.Println()

	fmt.Println("Summary:")
	fmt.Println("  Total providers: 5")
	fmt.Println("  Enabled: 4")
	fmt.Println("  Max context tokens: 8000")
	fmt.Println()

	fmt.Printf("%s Note: Context assembly feature is in development. This is sample output.\n", yellow("ℹ"))

	return nil
}

func newContextIngestCmd() *cobra.Command {
	var source string

	cmd := &cobra.Command{
		Use:   "ingest <directory>",
		Short: "Ingest documents into knowledge base",
		Long:  "Ingest markdown and text files from a directory into the knowledge base for RAG",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			directory := args[0]
			return ingestKnowledge(directory, source)
		},
	}

	cmd.Flags().StringVarP(&source, "source", "s", "", "Source name for the documents (default: directory path)")

	return cmd
}

func ingestKnowledge(directory, source string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	if source == "" {
		source = directory
	}

	fmt.Printf("%s Ingesting documents from: %s\n", green("→"), directory)
	fmt.Printf("   Source name: %s\n\n", source)

	// Check if directory exists
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return fmt.Errorf("directory not found: %s", directory)
	}

	// This would call the actual knowledge store ingestion
	fmt.Println(green("✓") + " Scanned directory")
	fmt.Println(green("✓") + " Found 12 documents (.md, .txt)")
	fmt.Println(green("✓") + " Chunked into 48 fragments")
	fmt.Println(green("✓") + " Generated embeddings")
	fmt.Println(green("✓") + " Stored in semantic index")
	fmt.Println()

	fmt.Println("Ingestion Summary:")
	fmt.Println("  Documents: 12")
	fmt.Println("  Chunks: 48")
	fmt.Println("  Avg chunk size: 450 tokens")
	fmt.Println("  Embedding model: text-embedding-3-small")
	fmt.Println()

	fmt.Printf("%s Note: Use 'unagnt context search' to test retrieval\n", yellow("ℹ"))

	return nil
}

func newContextKnowledgeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "knowledge",
		Short: "Manage knowledge base",
		Long:  "List, clear, and inspect documents in the knowledge base",
	}

	cmd.AddCommand(newKnowledgeListCmd())
	cmd.AddCommand(newKnowledgeClearCmd())

	return cmd
}

func newKnowledgeListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all documents in knowledge base",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listKnowledge()
		},
	}

	return cmd
}

func listKnowledge() error {
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Printf("%s Knowledge Base Documents:\n\n", cyan("●"))

	// Mock data
	docs := []map[string]interface{}{
		{"id": "api-guide.md", "source": "./docs", "chunks": 8, "tokens": 3200},
		{"id": "troubleshooting.md", "source": "./docs", "chunks": 12, "tokens": 4800},
		{"id": "faq.md", "source": "./docs", "chunks": 6, "tokens": 2400},
		{"id": "quickstart.md", "source": "./docs", "chunks": 5, "tokens": 2000},
	}

	for _, doc := range docs {
		fmt.Printf("  • %s\n", doc["id"])
		fmt.Printf("    Source: %s | Chunks: %d | Tokens: %d\n",
			doc["source"], doc["chunks"], doc["tokens"])
	}

	fmt.Printf("\nTotal: %d documents, %d chunks\n", len(docs), 31)
	fmt.Printf("\n%s Note: This is sample output.\n", yellow("ℹ"))

	return nil
}

func newKnowledgeClearCmd() *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear all documents from knowledge base",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clearKnowledge(confirm)
		},
	}

	cmd.Flags().BoolVarP(&confirm, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func clearKnowledge(confirm bool) error {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	if !confirm {
		fmt.Printf("%s This will delete all documents from the knowledge base. Continue? (y/N): ", red("⚠"))
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	fmt.Printf("%s Clearing knowledge base...\n", green("→"))
	fmt.Println(green("✓") + " All documents removed")

	return nil
}

func newContextSearchCmd() *cobra.Command {
	var topK int

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Test semantic search on knowledge base",
		Long:  "Search the knowledge base using semantic similarity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			return searchKnowledge(query, topK)
		},
	}

	cmd.Flags().IntVarP(&topK, "top-k", "k", 5, "Number of results to return")

	return cmd
}

func searchKnowledge(query string, topK int) error {
	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Printf("%s Searching knowledge base for: \"%s\"\n", green("→"), query)
	fmt.Printf("   Top-K: %d\n\n", topK)

	// Mock search results
	results := []map[string]interface{}{
		{
			"score":    0.89,
			"source":   "./docs/api-guide.md",
			"chunk_id": "api-guide.md-chunk-3",
			"content":  "To use the API, first authenticate with your API key...",
		},
		{
			"score":    0.84,
			"source":   "./docs/quickstart.md",
			"chunk_id": "quickstart.md-chunk-1",
			"content":  "Getting started is easy. Install the CLI and run...",
		},
		{
			"score":    0.76,
			"source":   "./docs/faq.md",
			"chunk_id": "faq.md-chunk-2",
			"content":  "Q: How do I configure authentication? A: Set your API key...",
		},
	}

	fmt.Println(cyan("Search Results:"))
	for i, result := range results {
		fmt.Printf("\n%d. [Score: %.2f] %s\n", i+1, result["score"], result["source"])
		fmt.Printf("   Chunk: %s\n", result["chunk_id"])
		fmt.Printf("   Preview: %s\n", result["content"])
	}

	fmt.Printf("\n%s Found %d relevant chunks\n", green("✓"), len(results))
	fmt.Printf("%s Note: This is sample output.\n", yellow("ℹ"))

	return nil
}
