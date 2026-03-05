package context

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/runtime"
)

// ToolOutputProvider consolidates recent tool outputs.
type ToolOutputProvider struct {
	priority int
	Window   int // Number of recent tool calls to include
}

// NewToolOutputProvider creates a new tool output provider.
func NewToolOutputProvider(priority int, window int) *ToolOutputProvider {
	if window == 0 {
		window = 10
	}
	return &ToolOutputProvider{
		priority: priority,
		Window:   window,
	}
}

// Name returns the provider name.
func (p *ToolOutputProvider) Name() string {
	return "tool_outputs"
}

// Priority returns the provider priority.
func (p *ToolOutputProvider) Priority() int {
	return p.priority
}

// Fetch retrieves recent tool outputs from history.
func (p *ToolOutputProvider) Fetch(ctx context.Context, input ContextInput) (*ContextFragment, error) {
	history := input.StepInput.History

	if len(history) == 0 {
		return nil, nil
	}

	// Get recent tool calls (up to window size)
	start := len(history) - p.Window
	if start < 0 {
		start = 0
	}

	recentHistory := history[start:]
	content := p.formatToolOutputs(recentHistory)

	if content == "" {
		return nil, nil
	}

	tokenCount := len(content) / 4 // Simple estimate

	return &ContextFragment{
		ProviderName: p.Name(),
		Type:         FragmentTypeToolOutput,
		Content:      content,
		Priority:     p.priority,
		TokenCount:   tokenCount,
		Metadata: map[string]any{
			"window":       p.Window,
			"history_size": len(recentHistory),
		},
		Timestamp: time.Now(),
	}, nil
}

// formatToolOutputs formats tool execution history.
func (p *ToolOutputProvider) formatToolOutputs(history []runtime.StepRecord) string {
	if len(history) == 0 {
		return ""
	}

	parts := []string{"Execution history:"}

	for i, record := range history {
		stepNum := i + 1

		if record.Action != nil {
			// Format tool call
			toolCall := fmt.Sprintf("Step %d: Called %s@%s",
				stepNum, record.Action.Tool, record.Action.Version)

			// Add input if available
			if len(record.Action.Input) > 0 {
				inputJSON, _ := json.Marshal(record.Action.Input)
				toolCall += fmt.Sprintf(" with input: %s", string(inputJSON))
			}

			parts = append(parts, toolCall)

			// Add result if available
			if record.Result != nil {
				if record.Result.Error != "" {
					parts = append(parts, fmt.Sprintf("  Error: %s", record.Result.Error))
				} else if len(record.Result.Output) > 0 {
					outputJSON, _ := json.Marshal(record.Result.Output)
					outputStr := string(outputJSON)
					// Truncate long outputs
					if len(outputStr) > 200 {
						outputStr = outputStr[:200] + "..."
					}
					parts = append(parts, fmt.Sprintf("  Result: %s", outputStr))
				}
			}

			// Add reasoning if available
			if record.Reasoning != "" {
				reasoning := record.Reasoning
				if len(reasoning) > 150 {
					reasoning = reasoning[:150] + "..."
				}
				parts = append(parts, fmt.Sprintf("  Reasoning: %s", reasoning))
			}
		}
	}

	return strings.Join(parts, "\n")
}
