package mcp

import (
	"context"

	"github.com/NikoSokratous/unagnt/pkg/tool"
)

// LoadMCPSource connects to an MCP server and registers its tools into the registry.
func LoadMCPSource(ctx context.Context, reg *tool.Registry, cfg MCPSourceConfig) (client *MCPClient, err error) {
	client, err = NewMCPClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	tools, err := client.ListTools(ctx)
	if err != nil {
		client.Close()
		return nil, err
	}

	for _, t := range tools {
		prefixedName := client.PrefixedName(t.Name)
		adapter := NewMCPToolAdapter(t, client, prefixedName)
		reg.Register(adapter)
	}

	return client, nil
}
