package builtin

import (
	"context"
	"encoding/json"

	"github.com/NikoSokratous/unagnt/pkg/tool"
)

// Echo is a built-in tool that echoes back input (useful for testing).
type Echo struct{}

func (Echo) Name() string        { return "echo" }
func (Echo) Version() string     { return "1" }
func (Echo) Description() string { return "Echo back the input (for testing)" }
func (Echo) Permissions() []tool.Permission {
	return nil
}

func (Echo) InputSchema() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"message": map[string]string{"type": "string", "description": "Message to echo"},
		},
	})
}

func (Echo) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
	var v map[string]any
	if err := json.Unmarshal(input, &v); err != nil {
		v = map[string]any{"raw": string(input)}
	}
	return map[string]any{"echoed": v}, nil
}
