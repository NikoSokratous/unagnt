package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newScaffoldCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scaffold",
		Short: "Generate tool boilerplate code",
	}
	cmd.AddCommand(newToolScaffoldCmd())
	return cmd
}

func newToolScaffoldCmd() *cobra.Command {
	var toolName string
	var output string

	cmd := &cobra.Command{
		Use:   "tool",
		Short: "Scaffold a new tool",
		Long:  "Generates boilerplate for a new tool implementation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runToolScaffold(toolName, output)
		},
	}

	cmd.Flags().StringVarP(&toolName, "name", "n", "", "Tool name (required)")
	cmd.Flags().StringVarP(&output, "output", "o", "tools", "Output directory")
	cmd.MarkFlagRequired("name")

	return cmd
}

func runToolScaffold(toolName, outputDir string) error {
	green := color.New(color.FgGreen).SprintFunc()

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	// Convert tool name to proper case
	structName := strings.Title(strings.ReplaceAll(toolName, "_", " "))
	structName = strings.ReplaceAll(structName, " ", "")

	// Generate tool implementation
	toolCode := fmt.Sprintf(`package tools

import (
	"context"
	"encoding/json"

	"github.com/NikoSokratous/unagnt/pkg/tool"
)

// %s is a tool that...
type %s struct{}

// Name returns the tool name.
func (t *%s) Name() string {
	return "%s"
}

// Version returns the tool version.
func (t *%s) Version() string {
	return "1"
}

// Description returns the tool description.
func (t *%s) Description() string {
	return "%s"
}

// InputSchema returns the JSON schema for tool input validation.
func (t *%s) InputSchema() ([]byte, error) {
	return []byte(`+"`"+`{
		"type": "object",
		"properties": {
			"param1": {
				"type": "string",
				"description": "%s",
			}
		},
		"required": ["param1"]
	}`+"`"+`), nil
}

// Permissions returns the required permissions.
func (t *%s) Permissions() []tool.Permission {
	return []tool.Permission{
		// Add permissions as needed, e.g.:
		// {Scope: "net:external", Required: true},
	}
}

// Execute runs the tool with the given input.
func (t *%s) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
	var params struct {
		Param1 string `+"`json:\"param1\"`"+`
	}
	
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}
	
	// TODO: Implement tool logic here
	
	return map[string]any{
		"result": "success",
		"data":   params.Param1,
	}, nil
}
`, structName, structName, structName, toolName, structName, structName, generateDescription(toolName), structName, generateParamDescription("param1"), structName, structName)

	// Generate test file
	testCode := fmt.Sprintf(`package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func Test%s(t *testing.T) {
	tool := &%s{}
	
	// Test basic properties
	if tool.Name() != "%s" {
		t.Errorf("Name() = %%v, want %s", tool.Name())
	}
	
	// Test schema
	schema, err := tool.InputSchema()
	if err != nil {
		t.Fatalf("InputSchema() error = %%v", err)
	}
	if len(schema) == 0 {
		t.Error("InputSchema() returned empty schema")
	}
	
	// Test execution
	ctx := context.Background()
	input := json.RawMessage(`+"`"+`{"param1":"test"}`+"`"+`)
	
	result, err := tool.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %%v", err)
	}
	
	if result["result"] != "success" {
		t.Errorf("Execute() result = %%v, want success", result["result"])
	}
}
`, structName, structName, toolName, toolName)

	// Write files
	toolFile := filepath.Join(outputDir, toolName+".go")
	testFile := filepath.Join(outputDir, toolName+"_test.go")

	if err := os.WriteFile(toolFile, []byte(toolCode), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(testFile, []byte(testCode), 0644); err != nil {
		return err
	}

	fmt.Printf("%s Created tool scaffolding:\n", green("✓"))
	fmt.Printf("  - %s\n", toolFile)
	fmt.Printf("  - %s\n", testFile)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Edit %s and implement the Execute() method\n", toolFile)
	fmt.Printf("  2. Update InputSchema() with actual parameters\n")
	fmt.Printf("  3. Add required permissions\n")
	fmt.Printf("  4. Run tests: go test %s\n", outputDir)

	return nil
}

// generateDescription generates a meaningful description based on tool name.
func generateDescription(toolName string) string {
	// Generate meaningful description based on tool name
	words := strings.Split(toolName, "_")
	if len(words) == 0 {
		return fmt.Sprintf("A tool named %s", toolName)
	}

	// Convert snake_case to readable text
	readable := make([]string, len(words))
	for i, word := range words {
		if i == 0 {
			readable[i] = strings.Title(word)
		} else {
			readable[i] = word
		}
	}

	return fmt.Sprintf("%s - Add your description here", strings.Join(readable, " "))
}

// generateParamDescription generates a description for a parameter.
func generateParamDescription(paramName string) string {
	// Convert snake_case or camelCase to readable text
	words := strings.FieldsFunc(paramName, func(r rune) bool {
		return r == '_' || (r >= 'A' && r <= 'Z')
	})

	if len(words) == 0 {
		return fmt.Sprintf("The %s parameter", paramName)
	}

	readable := strings.Join(words, " ")
	return fmt.Sprintf("The %s to use", readable)
}
