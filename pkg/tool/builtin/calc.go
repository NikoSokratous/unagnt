package builtin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/NikoSokratous/unagnt/pkg/tool"
)

// Calc is a built-in tool for simple arithmetic.
type Calc struct{}

func (Calc) Name() string    { return "calc" }
func (Calc) Version() string { return "1" }
func (Calc) Description() string {
	return "Perform simple arithmetic (add, subtract, multiply, divide)"
}
func (Calc) Permissions() []tool.Permission {
	return nil
}

func (Calc) InputSchema() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":     "object",
		"required": []string{"op", "a", "b"},
		"properties": map[string]any{
			"op": map[string]any{
				"type":        "string",
				"description": "Operation: add, sub, mul, div",
				"enum":        []string{"add", "sub", "mul", "div"},
			},
			"a": map[string]string{"type": "number", "description": "First operand"},
			"b": map[string]string{"type": "number", "description": "Second operand"},
		},
	})
}

func (Calc) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
	var req struct {
		Op string  `json:"op"`
		A  float64 `json:"a"`
		B  float64 `json:"b"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	var result float64
	switch req.Op {
	case "add":
		result = req.A + req.B
	case "sub":
		result = req.A - req.B
	case "mul":
		result = req.A * req.B
	case "div":
		if req.B == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = req.A / req.B
	default:
		return nil, fmt.Errorf("unknown op: %s", req.Op)
	}

	return map[string]any{"result": result}, nil
}
