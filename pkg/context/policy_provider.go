package context

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/llm"
	"github.com/NikoSokratous/unagnt/pkg/policy"
)

// PolicyProvider injects active policies as system context.
type PolicyProvider struct {
	Engine   *policy.Engine
	priority int
}

// NewPolicyProvider creates a new policy provider.
func NewPolicyProvider(engine *policy.Engine, priority int) *PolicyProvider {
	return &PolicyProvider{
		Engine:   engine,
		priority: priority,
	}
}

// Name returns the provider name.
func (p *PolicyProvider) Name() string {
	return "policy"
}

// Priority returns the provider priority.
func (p *PolicyProvider) Priority() int {
	return p.priority
}

// Fetch retrieves policy context.
func (p *PolicyProvider) Fetch(ctx context.Context, input ContextInput) (*ContextFragment, error) {
	// Format policies as context for the agent
	// Even if engine is nil, we provide generic policy guidance
	policyText := p.formatPolicies(input.Tools)

	if policyText == "" {
		return nil, nil
	}

	tokenCount := len(policyText) / 4 // Simple estimate

	source := "generic"
	if p.Engine != nil {
		source = "policy_engine"
	}

	return &ContextFragment{
		ProviderName: p.Name(),
		Type:         FragmentTypePolicy,
		Content:      policyText,
		Priority:     p.priority,
		TokenCount:   tokenCount,
		Metadata:     map[string]any{"source": source},
		Timestamp:    time.Now(),
	}, nil
}

// formatPolicies formats policies as readable context.
func (p *PolicyProvider) formatPolicies(tools []llm.ToolInfo) string {
	parts := []string{"Policy constraints:"}

	// Get policy rules from engine if available
	// Otherwise, provide generic policy guidance
	parts = append(parts, p.formatGenericPolicies(tools))

	return strings.Join(parts, "\n")
}

// formatGenericPolicies formats generic policy guidance.
func (p *PolicyProvider) formatGenericPolicies(tools []llm.ToolInfo) string {
	// Generic policy statements that apply to most agents
	policies := []string{
		"- You must follow all policy rules and constraints",
		"- High-risk operations may require approval before execution",
		"- All actions are logged and audited for compliance",
	}

	// Add tool-specific guidance if available
	if len(tools) > 0 {
		toolNames := make([]string, 0, len(tools))
		for _, tool := range tools {
			toolNames = append(toolNames, tool.Name)
		}
		policies = append(policies, fmt.Sprintf("- Available tools: %s", strings.Join(toolNames, ", ")))
	}

	return strings.Join(policies, "\n")
}
