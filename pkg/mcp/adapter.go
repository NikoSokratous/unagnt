package mcp

import (
	"context"
	"encoding/json"

	"github.com/NikoSokratous/unagnt/pkg/tool"
	"github.com/mark3labs/mcp-go/mcp"
)

// MCPToolAdapter adapts an MCP tool to the tool.Tool interface.
type MCPToolAdapter struct {
	mcpTool  mcp.Tool
	client   *MCPClient
	prefixedName string
}

// NewMCPToolAdapter creates an adapter for an MCP tool.
func NewMCPToolAdapter(mcpTool mcp.Tool, client *MCPClient, prefixedName string) *MCPToolAdapter {
	return &MCPToolAdapter{
		mcpTool:      mcpTool,
		client:       client,
		prefixedName: prefixedName,
	}
}

// Name returns the tool name (with optional prefix).
func (a *MCPToolAdapter) Name() string {
	if a.prefixedName != "" {
		return a.prefixedName
	}
	return a.mcpTool.Name
}

// Version returns the tool version (MCP tools use "1" by default).
func (a *MCPToolAdapter) Version() string {
	return "1"
}

// Description returns the tool description.
func (a *MCPToolAdapter) Description() string {
	return a.mcpTool.Description
}

// Permissions returns required permissions (MCP tools default to external).
func (a *MCPToolAdapter) Permissions() []tool.Permission {
	return []tool.Permission{{Scope: "mcp:external", Required: true}}
}

// InputSchema returns the JSON schema for tool input.
func (a *MCPToolAdapter) InputSchema() ([]byte, error) {
	return ToolToInputSchema(a.mcpTool)
}

// Execute runs the MCP tool.
func (a *MCPToolAdapter) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
	var args map[string]interface{}
	if err := json.Unmarshal(input, &args); err != nil {
		args = make(map[string]interface{})
	}
	if args == nil {
		args = make(map[string]interface{})
	}

	result, err := a.client.CallTool(ctx, a.Name(), args)
	if err != nil {
		return nil, err
	}

	if result.IsError {
		msg := "MCP tool error"
		if len(result.Content) > 0 {
			msg = mcp.GetTextFromContent(result.Content[0])
			if msg == "" {
				msg = "MCP tool error"
			}
		}
		return map[string]any{"error": msg}, nil
	}

	out := make(map[string]any)
	if len(result.Content) > 0 {
		textParts := make([]string, 0)
		for _, c := range result.Content {
			if t := mcp.GetTextFromContent(c); t != "" {
				textParts = append(textParts, t)
			}
		}
		if len(textParts) > 0 {
			out["text"] = textParts
			if len(textParts) == 1 {
				out["output"] = textParts[0]
			}
		}
	}
	if result.StructuredContent != nil {
		out["structured"] = result.StructuredContent
	}
	if len(out) == 0 {
		out["ok"] = true
	}
	return out, nil
}
