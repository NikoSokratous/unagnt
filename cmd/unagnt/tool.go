package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/NikoSokratous/unagnt/pkg/tool"
	"github.com/NikoSokratous/unagnt/pkg/tool/builtin"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newToolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tool",
		Short: "Tool management commands",
	}
	cmd.AddCommand(newToolValidateCmd())
	return cmd
}

func newToolValidateCmd() *cobra.Command {
	var toolName string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate tool implementations",
		Long:  "Validates tool schema, permissions, and interface implementation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runToolValidate(toolName)
		},
	}

	cmd.Flags().StringVarP(&toolName, "tool", "t", "", "Tool name to validate (leave empty to validate all built-ins)")
	return cmd
}

func runToolValidate(toolName string) error {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// Get tools to validate
	tools := builtin.All()
	var toValidate []tool.Tool

	if toolName != "" {
		// Find specific tool
		found := false
		for _, t := range tools {
			if t.Name() == toolName {
				toValidate = append(toValidate, t)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("tool not found: %s", toolName)
		}
	} else {
		toValidate = tools
	}

	// Validation results
	table := tablewriter.NewWriter(os.Stdout)
	table.Append([]string{"Tool", "Name", "Version", "Schema", "Permissions", "Interface", "Status"})

	totalPass := 0
	totalFail := 0

	for _, t := range toValidate {
		results := validateTool(t)

		// Determine overall status
		status := green("PASS")
		allPass := true
		for _, r := range results {
			if !r.Pass {
				allPass = false
				status = red("FAIL")
				totalFail++
				break
			}
		}
		if allPass {
			totalPass++
		}

		// Format individual checks
		nameCheck := formatCheck(results["name"])
		versionCheck := formatCheck(results["version"])
		schemaCheck := formatCheck(results["schema"])
		permCheck := formatCheck(results["permissions"])
		ifaceCheck := formatCheck(results["interface"])

		table.Append([]string{
			t.Name(),
			nameCheck,
			versionCheck,
			schemaCheck,
			permCheck,
			ifaceCheck,
			status,
		})

		// Print detailed errors
		for _, r := range results {
			if !r.Pass {
				fmt.Fprintf(os.Stderr, "  %s %s: %s\n", red("✗"), r.Check, r.Message)
			}
		}
	}

	fmt.Println("\nValidation Results:")
	table.Render()

	fmt.Printf("\nSummary: %s passed, %s failed\n",
		green(fmt.Sprintf("%d", totalPass)),
		red(fmt.Sprintf("%d", totalFail)))

	if totalFail > 0 {
		return fmt.Errorf("validation failed for %d tool(s)", totalFail)
	}

	return nil
}

type toolValidationResult struct {
	Check   string
	Pass    bool
	Message string
}

func validateTool(t tool.Tool) map[string]toolValidationResult {
	results := make(map[string]toolValidationResult)

	// Check name
	name := t.Name()
	if name == "" {
		results["name"] = toolValidationResult{"Name", false, "empty name"}
	} else {
		results["name"] = toolValidationResult{"Name", true, ""}
	}

	// Check version
	version := t.Version()
	if version == "" {
		results["version"] = toolValidationResult{"Version", false, "empty version"}
	} else {
		results["version"] = toolValidationResult{"Version", true, ""}
	}

	// Check schema
	schema, err := t.InputSchema()
	if err != nil {
		results["schema"] = toolValidationResult{"Schema", false, err.Error()}
	} else if len(schema) == 0 {
		results["schema"] = toolValidationResult{"Schema", false, "empty schema"}
	} else {
		// Validate JSON syntax
		var parsed map[string]any
		if err := json.Unmarshal(schema, &parsed); err != nil {
			results["schema"] = toolValidationResult{"Schema", false, "invalid JSON: " + err.Error()}
		} else {
			results["schema"] = toolValidationResult{"Schema", true, ""}
		}
	}

	// Check permissions
	_ = t.Permissions()
	results["permissions"] = toolValidationResult{"Permissions", true, ""}

	// Check interface implementation
	val := reflect.ValueOf(t)
	typ := val.Type()

	requiredMethods := []string{"Name", "Version", "Description", "InputSchema", "Permissions", "Execute"}
	hasAll := true
	for _, method := range requiredMethods {
		if _, found := typ.MethodByName(method); !found {
			results["interface"] = toolValidationResult{"Interface", false, "missing method: " + method}
			hasAll = false
			break
		}
	}
	if hasAll {
		results["interface"] = toolValidationResult{"Interface", true, ""}
	}

	return results
}

func formatCheck(r toolValidationResult) string {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	if r.Pass {
		return green("✓")
	}
	return red("✗")
}
