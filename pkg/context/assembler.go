package context

import (
	"fmt"
	"strings"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/llm"
)

// DefaultAssembler implements ContextAssembler with token budgeting.
type DefaultAssembler struct {
	TokenCounter TokenCounter
}

// NewDefaultAssembler creates a new default assembler.
func NewDefaultAssembler() *DefaultAssembler {
	return &DefaultAssembler{
		TokenCounter: &SimpleTokenCounter{},
	}
}

// Assemble merges fragments with token budgeting.
func (a *DefaultAssembler) Assemble(fragments []ContextFragment, config AssemblyConfig) ([]llm.Message, error) {
	startTime := time.Now()

	if config.MaxTokens == 0 {
		config.MaxTokens = 8000
	}

	// Initialize token budget if not provided
	if config.TokenBudget == nil {
		config.TokenBudget = a.defaultTokenBudget()
	}

	// Group fragments by type
	fragmentsByType := make(map[FragmentType][]ContextFragment)
	for _, fragment := range fragments {
		fragmentsByType[fragment.Type] = append(fragmentsByType[fragment.Type], fragment)
	}

	// Build messages with token budget
	messages := []llm.Message{}
	totalTokens := 0
	included := []ContextFragment{}
	truncated := 0

	// 1. System prompt (always include)
	if systemFragments, ok := fragmentsByType[FragmentTypeSystemPrompt]; ok && len(systemFragments) > 0 {
		content := a.mergeFragments(systemFragments)
		messages = append(messages, llm.Message{
			Role:    llm.RoleSystem,
			Content: content,
		})
		totalTokens += a.TokenCounter.Count(content)
		included = append(included, systemFragments...)
	} else {
		// Default system prompt
		defaultPrompt := "You are an autonomous agent. Given a goal and prior steps, choose the next tool to call or indicate completion."
		messages = append(messages, llm.Message{
			Role:    llm.RoleSystem,
			Content: defaultPrompt,
		})
		totalTokens += a.TokenCounter.Count(defaultPrompt)
	}

	// 2. Policy context (high priority)
	if policyFragments, ok := fragmentsByType[FragmentTypePolicy]; ok {
		budget := config.TokenBudget[FragmentTypePolicy]
		content, used := a.applyBudget(policyFragments, budget)
		if content != "" {
			// Append to system message
			messages[0].Content += "\n\n" + content
			totalTokens += used
			included = append(included, policyFragments...)
		}
	}

	// 3. Workflow state
	if workflowFragments, ok := fragmentsByType[FragmentTypeWorkflowState]; ok {
		budget := config.TokenBudget[FragmentTypeWorkflowState]
		content, used := a.applyBudget(workflowFragments, budget)
		if content != "" {
			messages[0].Content += "\n\n" + content
			totalTokens += used
			included = append(included, workflowFragments...)
		}
	}

	// 4. Build user message with remaining content
	userContent := strings.Builder{}

	// User goal (always include)
	if goalFragments, ok := fragmentsByType[FragmentTypeUserGoal]; ok {
		content := a.mergeFragments(goalFragments)
		userContent.WriteString(content)
		totalTokens += a.TokenCounter.Count(content)
		included = append(included, goalFragments...)
	}

	// Memory (up to budget)
	if memoryFragments, ok := fragmentsByType[FragmentTypeMemory]; ok {
		budget := config.TokenBudget[FragmentTypeMemory]
		remainingBudget := config.MaxTokens - totalTokens
		if budget > remainingBudget {
			budget = remainingBudget
		}
		content, used := a.applyBudget(memoryFragments, budget)
		if content != "" {
			if userContent.Len() > 0 {
				userContent.WriteString("\n\n")
			}
			userContent.WriteString("Relevant context from memory:\n")
			userContent.WriteString(content)
			totalTokens += used
			included = append(included, memoryFragments...)
		}
	}

	// Tool outputs (up to budget)
	if toolFragments, ok := fragmentsByType[FragmentTypeToolOutput]; ok {
		budget := config.TokenBudget[FragmentTypeToolOutput]
		remainingBudget := config.MaxTokens - totalTokens
		if budget > remainingBudget {
			budget = remainingBudget
		}
		content, used := a.applyBudget(toolFragments, budget)
		if content != "" {
			if userContent.Len() > 0 {
				userContent.WriteString("\n\n")
			}
			userContent.WriteString(content)
			totalTokens += used
			included = append(included, toolFragments...)
		}
	}

	// Knowledge (if space remains)
	if knowledgeFragments, ok := fragmentsByType[FragmentTypeKnowledge]; ok {
		budget := config.TokenBudget[FragmentTypeKnowledge]
		remainingBudget := config.MaxTokens - totalTokens
		if budget > remainingBudget {
			budget = remainingBudget
		}
		content, used := a.applyBudget(knowledgeFragments, budget)
		if content != "" {
			if userContent.Len() > 0 {
				userContent.WriteString("\n\n")
			}
			userContent.WriteString("Retrieved knowledge:\n")
			userContent.WriteString(content)
			totalTokens += used
			included = append(included, knowledgeFragments...)
		}
	}

	if userContent.Len() > 0 {
		messages = append(messages, llm.Message{
			Role:    llm.RoleUser,
			Content: userContent.String(),
		})
	}

	// Add debug metadata if requested
	if config.IncludeDebug {
		debugInfo := fmt.Sprintf("\n\n[DEBUG] Total tokens: %d, Assembly time: %v, Fragments included: %d, Truncated: %d",
			totalTokens, time.Since(startTime), len(included), truncated)
		if len(messages) > 0 {
			messages[len(messages)-1].Content += debugInfo
		}
	}

	return messages, nil
}

// mergeFragments combines multiple fragments into a single string.
func (a *DefaultAssembler) mergeFragments(fragments []ContextFragment) string {
	parts := make([]string, 0, len(fragments))
	for _, fragment := range fragments {
		if fragment.Content != "" {
			parts = append(parts, fragment.Content)
		}
	}
	return strings.Join(parts, "\n\n")
}

// applyBudget applies token budget to fragments, returning content and tokens used.
func (a *DefaultAssembler) applyBudget(fragments []ContextFragment, budget int) (string, int) {
	if budget <= 0 {
		return "", 0
	}

	content := a.mergeFragments(fragments)
	tokens := a.TokenCounter.Count(content)

	if tokens <= budget {
		return content, tokens
	}

	// Truncate to fit budget
	// Simple truncation: keep first N characters
	maxChars := budget * 4 // Rough conversion
	if maxChars < len(content) {
		content = content[:maxChars] + "... [truncated]"
	}

	return content, budget
}

// defaultTokenBudget returns default token allocations.
func (a *DefaultAssembler) defaultTokenBudget() map[FragmentType]int {
	return map[FragmentType]int{
		FragmentTypeSystemPrompt:  500,
		FragmentTypePolicy:        1000,
		FragmentTypeWorkflowState: 500,
		FragmentTypeMemory:        3000,
		FragmentTypeToolOutput:    2000,
		FragmentTypeUserGoal:      1000,
		FragmentTypeKnowledge:     1000,
	}
}
