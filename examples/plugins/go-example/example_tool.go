package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/NikoSokratous/unagnt/pkg/tool"
)

// ExampleTool is a sample Go plugin tool.
type ExampleTool struct{}

// NewTool is the required plugin entry point.
func NewTool() tool.Tool {
	return &ExampleTool{}
}

func (t *ExampleTool) Name() string {
	return "example_plugin"
}

func (t *ExampleTool) Version() string {
	return "1.0.0"
}

func (t *ExampleTool) Description() string {
	return "Example Go plugin tool for demonstration"
}

func (t *ExampleTool) InputSchema() ([]byte, error) {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Message to process",
			},
			"uppercase": map[string]interface{}{
				"type":        "boolean",
				"description": "Convert to uppercase",
				"default":     false,
			},
		},
		"required": []string{"message"},
	}
	return json.Marshal(schema)
}

func (t *ExampleTool) Permissions() []tool.Permission {
	return []tool.Permission{
		{
			Scope:    "compute:cpu",
			Required: true,
		},
	}
}

func (t *ExampleTool) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
	var params struct {
		Message   string `json:"message"`
		Uppercase bool   `json:"uppercase"`
	}
	
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	
	// Validate required fields
	if params.Message == "" {
		return nil, fmt.Errorf("message is required")
	}
	
	result := params.Message
	if params.Uppercase {
		result = strings.ToUpper(result)
	}
	
	return map[string]any{
		"result":     result,
		"length":     len(result),
		"uppercase":  params.Uppercase,
		"processed":  true,
	}, nil
}

// main is required so this package builds with go build ./...
// When built as a plugin (go build -buildmode=plugin), the plugin loader uses NewTool.
func main() {}
