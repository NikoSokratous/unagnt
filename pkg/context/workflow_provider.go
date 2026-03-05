package context

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/runtime"
)

// WorkflowProvider adds workflow state context.
type WorkflowProvider struct {
	priority int
}

// NewWorkflowProvider creates a new workflow provider.
func NewWorkflowProvider(priority int) *WorkflowProvider {
	return &WorkflowProvider{
		priority: priority,
	}
}

// Name returns the provider name.
func (p *WorkflowProvider) Name() string {
	return "workflow_state"
}

// Priority returns the provider priority.
func (p *WorkflowProvider) Priority() int {
	return p.priority
}

// Fetch retrieves workflow state context.
func (p *WorkflowProvider) Fetch(ctx context.Context, input ContextInput) (*ContextFragment, error) {
	stepInput := input.StepInput

	// Format workflow state information
	parts := []string{}

	// Current step information
	if stepInput.StepNum > 0 {
		parts = append(parts, fmt.Sprintf("Current execution step: %d/%d",
			stepInput.StepNum, stepInput.StepNum+10)) // Approximate remaining
	}

	// History summary
	if len(stepInput.History) > 0 {
		parts = append(parts, fmt.Sprintf("Previous steps completed: %d", len(stepInput.History)))

		// Recent step summary
		recentSteps := p.formatRecentSteps(stepInput.History, 3)
		if recentSteps != "" {
			parts = append(parts, "Recent steps:\n"+recentSteps)
		}
	}

	// Agent information
	if stepInput.AgentName != "" {
		parts = append(parts, fmt.Sprintf("Agent: %s", stepInput.AgentName))
	}

	if len(parts) == 0 {
		return nil, nil
	}

	content := "Workflow state:\n" + strings.Join(parts, "\n")
	tokenCount := len(content) / 4 // Simple estimate

	return &ContextFragment{
		ProviderName: p.Name(),
		Type:         FragmentTypeWorkflowState,
		Content:      content,
		Priority:     p.priority,
		TokenCount:   tokenCount,
		Metadata: map[string]any{
			"step_num":     stepInput.StepNum,
			"agent_name":   stepInput.AgentName,
			"history_size": len(stepInput.History),
		},
		Timestamp: time.Now(),
	}, nil
}

// formatRecentSteps formats recent step history.
func (p *WorkflowProvider) formatRecentSteps(history []runtime.StepRecord, limit int) string {
	if len(history) == 0 {
		return ""
	}

	start := len(history) - limit
	if start < 0 {
		start = 0
	}

	parts := []string{}
	for i := start; i < len(history); i++ {
		parts = append(parts, fmt.Sprintf("  %d. Step completed", i+1))
	}

	return strings.Join(parts, "\n")
}
