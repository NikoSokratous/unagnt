package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/registry"
	"github.com/NikoSokratous/unagnt/pkg/workflow"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	_ "modernc.org/sqlite"
)

func newWorkflowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Workflow management commands",
		Long:  "Execute, manage, and visualize DAG-based workflows",
	}

	cmd.AddCommand(
		newWorkflowRunCmd(),
		newWorkflowStatusCmd(),
		newWorkflowListCmd(),
		newWorkflowVisualizeCmd(),
		newWorkflowValidateCmd(),
		newWorkflowTemplateCmd(),
	)

	return cmd
}

func newWorkflowRunCmd() *cobra.Command {
	var resumeID string

	cmd := &cobra.Command{
		Use:   "run <workflow-file>",
		Short: "Execute a workflow",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowFile := args[0]

			// Load workflow
			data, err := os.ReadFile(workflowFile)
			if err != nil {
				return fmt.Errorf("read workflow file: %w", err)
			}

			var workflowDef map[string]interface{}
			if err := yaml.Unmarshal(data, &workflowDef); err != nil {
				return fmt.Errorf("parse workflow: %w", err)
			}

			// Build DAG
			dag, err := buildDAGFromWorkflow(workflowDef)
			if err != nil {
				return fmt.Errorf("build DAG: %w", err)
			}

			// Create executor (without state store for now)
			executor := workflow.NewExecutor(nil, nil)

			ctx := context.Background()
			workflowID := fmt.Sprintf("wf-%d", time.Now().Unix())

			if resumeID != "" {
				// Resume from checkpoint
				fmt.Printf("Resuming workflow: %s\n", resumeID)
				// result, err := executor.Resume(ctx, resumeID, dag)
				return fmt.Errorf("resume not yet implemented")
			} else {
				// Execute new workflow
				fmt.Printf("Starting workflow: %s\n", workflowID)
				result, err := executor.Execute(ctx, dag, workflowID)
				if err != nil {
					return fmt.Errorf("execute workflow: %w", err)
				}

				// Print results
				printWorkflowResult(result)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&resumeID, "resume", "", "Resume workflow from checkpoint ID")

	return cmd
}

func newWorkflowStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <workflow-id>",
		Short: "Show workflow status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowID := args[0]

			// Placeholder: would load from state store
			fmt.Printf("Workflow ID: %s\n", workflowID)
			fmt.Println("Status: completed")
			fmt.Println("Progress: 5/5 steps completed")

			return nil
		},
	}
}

func newWorkflowListCmd() *cobra.Command {
	var status string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workflows",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Placeholder: would load from state store
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tSTATUS\tSTARTED\tDURATION")
			fmt.Fprintln(w, "wf-123\tdata-pipeline\tcompleted\t2m ago\t45s")
			fmt.Fprintln(w, "wf-124\tcode-review\trunning\t30s ago\t-")
			w.Flush()

			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum workflows to show")

	return cmd
}

func newWorkflowVisualizeCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "visualize <workflow-file>",
		Short: "Generate DAG visualization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowFile := args[0]

			// Load workflow
			data, err := os.ReadFile(workflowFile)
			if err != nil {
				return fmt.Errorf("read workflow file: %w", err)
			}

			var workflowDef map[string]interface{}
			if err := yaml.Unmarshal(data, &workflowDef); err != nil {
				return fmt.Errorf("parse workflow: %w", err)
			}

			// Build DAG
			dag, err := buildDAGFromWorkflow(workflowDef)
			if err != nil {
				return fmt.Errorf("build DAG: %w", err)
			}

			// Generate visualization
			if outputFormat == "dot" {
				fmt.Println(dag.ToDOT())
			} else {
				// ASCII visualization
				printDAGAscii(dag)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&outputFormat, "format", "ascii", "Output format (ascii, dot)")

	return cmd
}

func newWorkflowValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <workflow-file>",
		Short: "Validate workflow DAG and conditions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowFile := args[0]

			// Load workflow
			data, err := os.ReadFile(workflowFile)
			if err != nil {
				return fmt.Errorf("read workflow file: %w", err)
			}

			var workflowDef map[string]interface{}
			if err := yaml.Unmarshal(data, &workflowDef); err != nil {
				return fmt.Errorf("parse workflow: %w", err)
			}

			// Build and validate DAG
			dag, err := buildDAGFromWorkflow(workflowDef)
			if err != nil {
				return fmt.Errorf("build DAG: %w", err)
			}

			if err := dag.Validate(); err != nil {
				fmt.Printf("❌ Validation failed: %v\n", err)
				return err
			}

			fmt.Println("✓ Workflow is valid")
			fmt.Printf("  Nodes: %d\n", len(dag.Nodes))
			fmt.Printf("  Roots: %v\n", dag.GetRoots())
			fmt.Printf("  Leaves: %v\n", dag.GetLeaves())

			// Show execution order
			levels, _ := dag.GetExecutionLevels()
			fmt.Println("\nExecution plan:")
			for i, level := range levels {
				fmt.Printf("  Level %d (parallel): %v\n", i+1, level)
			}

			return nil
		},
	}
}

const workflowVersionsSchema = `
CREATE TABLE IF NOT EXISTS workflow_versions (
    id TEXT PRIMARY KEY,
    workflow_name TEXT NOT NULL,
    version TEXT NOT NULL,
    template_yaml TEXT NOT NULL,
    parameters TEXT,
    author TEXT,
    changelog TEXT,
    effective_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    metadata TEXT,
    UNIQUE(workflow_name, version)
);
CREATE INDEX IF NOT EXISTS idx_workflow_versions_name ON workflow_versions(workflow_name);
`

func openWorkflowDB(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = "agent.db"
	}
	abs, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", abs)
	if err != nil {
		return nil, err
	}
	if _, err := db.ExecContext(context.Background(), workflowVersionsSchema); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func newWorkflowTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Workflow template commands",
	}

	cmd.PersistentFlags().String("db", "agent.db", "Database path for workflow version store")

	cmd.AddCommand(
		newWorkflowTemplateListCmd(),
		newWorkflowTemplateCreateCmd(),
		newWorkflowTemplateValidateCmd(),
		newWorkflowTemplateVersionsCmd(),
		newWorkflowTemplatePublishCmd(),
	)

	return cmd
}

func newWorkflowTemplateListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available workflow templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := workflow.NewTemplateRegistry()
			if err := registry.LoadEmbeddedTemplates(); err != nil {
				return err
			}

			templates := registry.List()

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\tCATEGORY\tVERSION\tDESCRIPTION")

			for _, tmpl := range templates {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					tmpl.Name,
					tmpl.Category,
					tmpl.Version,
					tmpl.Description,
				)
			}

			w.Flush()
			return nil
		},
	}
}

func newWorkflowTemplateCreateCmd() *cobra.Command {
	var outputFile string
	var paramsJSON string

	cmd := &cobra.Command{
		Use:   "create <template-name>",
		Short: "Create workflow from template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]

			// Parse parameters
			params := make(map[string]interface{})
			if paramsJSON != "" {
				if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
					return fmt.Errorf("parse parameters: %w", err)
				}
			}

			// Load registry
			registry := workflow.NewTemplateRegistry()
			if err := registry.LoadEmbeddedTemplates(); err != nil {
				return err
			}

			// Instantiate template
			workflowData, err := registry.Instantiate(templateName, params)
			if err != nil {
				return fmt.Errorf("instantiate template: %w", err)
			}

			// Write output
			if outputFile != "" {
				if err := os.WriteFile(outputFile, workflowData, 0644); err != nil {
					return fmt.Errorf("write output: %w", err)
				}
				fmt.Printf("Workflow written to: %s\n", outputFile)
			} else {
				fmt.Println(string(workflowData))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&outputFile, "output", "", "Output file path")
	cmd.Flags().StringVar(&paramsJSON, "params", "", "Template parameters (JSON)")

	return cmd
}

func newWorkflowTemplateVersionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "versions <workflow-name>",
		Short: "List workflow versions (from version store)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath, _ := cmd.Parent().PersistentFlags().GetString("db")
			db, err := openWorkflowDB(dbPath)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer db.Close()

			mp := registry.NewWorkflowMarketplace(db)
			versions, err := mp.ListWorkflowVersions(context.Background(), args[0])
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "VERSION\tAUTHOR\tCREATED\tCHANGELOG")
			for _, v := range versions {
				changelog := v.Changelog
				if len(changelog) > 40 {
					changelog = changelog[:37] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					v.Version,
					v.Author,
					v.CreatedAt.Format("2006-01-02 15:04"),
					changelog,
				)
			}
			w.Flush()
			return nil
		},
	}
}

func newWorkflowTemplatePublishCmd() *cobra.Command {
	var name, version, author, changelog string
	cmd := &cobra.Command{
		Use:   "publish <template-file>",
		Short: "Publish a workflow template version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read template: %w", err)
			}

			var tmpl workflow.Template
			if err := yaml.Unmarshal(data, &tmpl); err != nil {
				return fmt.Errorf("parse template: %w", err)
			}
			if err := workflow.ValidateTemplate(&tmpl); err != nil {
				return fmt.Errorf("validate template: %w", err)
			}

			workflowName := name
			if workflowName == "" {
				workflowName = tmpl.Name
			}
			ver := version
			if ver == "" {
				ver = tmpl.Version
			}
			if workflowName == "" || ver == "" {
				return fmt.Errorf("workflow name and version required (use --name and --version or set in template)")
			}

			dbPath, _ := cmd.Parent().PersistentFlags().GetString("db")
			db, err := openWorkflowDB(dbPath)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer db.Close()

			paramsJSON, _ := json.Marshal(tmpl.Parameters)
			mp := registry.NewWorkflowMarketplace(db)
			v := &registry.WorkflowVersion{
				WorkflowName: workflowName,
				Version:      ver,
				TemplateYAML: string(data),
				Parameters:   string(paramsJSON),
				Author:       author,
				Changelog:    changelog,
			}
			if err := mp.PublishWorkflowVersion(context.Background(), v); err != nil {
				return err
			}
			fmt.Printf("✓ Published %s@%s\n", workflowName, ver)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Workflow name (default: from template)")
	cmd.Flags().StringVar(&version, "version", "", "Version (default: from template)")
	cmd.Flags().StringVar(&author, "author", "", "Author")
	cmd.Flags().StringVar(&changelog, "changelog", "", "Changelog")
	return cmd
}

func newWorkflowTemplateValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <template-file>",
		Short: "Validate workflow template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateFile := args[0]

			data, err := os.ReadFile(templateFile)
			if err != nil {
				return fmt.Errorf("read template: %w", err)
			}

			var tmpl workflow.Template
			if err := yaml.Unmarshal(data, &tmpl); err != nil {
				return fmt.Errorf("parse template: %w", err)
			}

			if err := workflow.ValidateTemplate(&tmpl); err != nil {
				fmt.Printf("❌ Template validation failed: %v\n", err)
				return err
			}

			fmt.Println("✓ Template is valid")
			fmt.Printf("  Name: %s\n", tmpl.Name)
			fmt.Printf("  Version: %s\n", tmpl.Version)
			fmt.Printf("  Parameters: %d\n", len(tmpl.Parameters))

			return nil
		},
	}
}

// Helper functions

func buildDAGFromWorkflow(workflowDef map[string]interface{}) (*workflow.DAG, error) {
	dag := workflow.NewDAG()

	steps, ok := workflowDef["steps"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("workflow must have 'steps' field")
	}

	// Add nodes
	for _, stepData := range steps {
		stepMap, ok := stepData.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := stepMap["name"].(string)
		if name == "" {
			continue
		}

		node := &workflow.Node{
			ID:        name,
			Name:      name,
			Agent:     getString(stepMap, "agent"),
			Goal:      getString(stepMap, "goal"),
			Condition: getString(stepMap, "condition"),
			OutputKey: getString(stepMap, "output_key"),
			Timeout:   getString(stepMap, "timeout"),
			Retry:     getInt(stepMap, "retry"),
		}

		// Get dependencies
		if deps, ok := stepMap["depends_on"].([]interface{}); ok {
			for _, dep := range deps {
				if depStr, ok := dep.(string); ok {
					node.Dependencies = append(node.Dependencies, depStr)
				}
			}
		}

		if err := dag.AddNode(node); err != nil {
			return nil, err
		}
	}

	// Add edges based on dependencies
	for _, node := range dag.Nodes {
		for _, dep := range node.Dependencies {
			if err := dag.AddEdge(dep, node.ID); err != nil {
				return nil, err
			}
		}
	}

	return dag, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(int); ok {
		return v
	}
	return 0
}

func printWorkflowResult(result *workflow.ExecutionResult) {
	fmt.Printf("\nWorkflow: %s\n", result.WorkflowID)
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Duration: %v\n\n", result.Duration)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NODE\tSTATUS\tDURATION\tERROR")

	for nodeID, nodeResult := range result.NodeResults {
		errMsg := ""
		if nodeResult.Error != "" {
			errMsg = nodeResult.Error
		}

		fmt.Fprintf(w, "%s\t%s\t%v\t%s\n",
			nodeID,
			nodeResult.Status,
			nodeResult.Duration,
			errMsg,
		)
	}

	w.Flush()
}

func printDAGAscii(dag *workflow.DAG) {
	fmt.Println("Workflow DAG:")
	fmt.Println()

	levels, _ := dag.GetExecutionLevels()

	for i, level := range levels {
		fmt.Printf("Level %d:\n", i+1)
		for _, nodeID := range level {
			node := dag.Nodes[nodeID]
			fmt.Printf("  ├─ %s (%s)\n", node.Name, node.Agent)

			if len(node.Dependencies) > 0 {
				fmt.Printf("     Dependencies: %v\n", node.Dependencies)
			}
		}
		if i < len(levels)-1 {
			fmt.Println("  │")
			fmt.Println("  ↓")
			fmt.Println()
		}
	}
}
