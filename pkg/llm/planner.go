package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/NikoSokratous/unagnt/pkg/runtime"
)

// ToolInfo describes an available tool for the planner.
type ToolInfo struct {
	Name        string
	Version     string
	Description string
}

// PlannerAdapter bridges an LLM Provider to runtime.LLMPlanner.
type PlannerAdapter struct {
	Provider      Provider
	Tools         []ToolInfo    // available tools for function calling
	ContextEngine ContextEngine // NEW: optional context engine
}

// ContextEngine interface for context assembly (to avoid import cycle).
type ContextEngine interface {
	FetchAll(ctx context.Context, input interface{}) ([]interface{}, error)
	Assemble(fragments []interface{}, config interface{}) ([]Message, error)
}

// Plan implements runtime.LLMPlanner.
func (p *PlannerAdapter) Plan(ctx context.Context, input runtime.StepInput) (*runtime.PlannedAction, error) {
	var messages []Message

	// Use context engine if available
	if p.ContextEngine != nil {
		var err error
		messages, err = p.buildMessagesWithContext(ctx, input)
		if err != nil {
			// Fallback to simple mode on error
			messages = p.buildMessagesSimple(input)
		}
	} else {
		// Simple mode (backward compatible)
		messages = p.buildMessagesSimple(input)
	}

	tools := p.buildToolDefs(input)
	req := &ChatRequest{
		Messages:    messages,
		Temperature: 0.2,
		MaxTokens:   4096,
		Tools:       tools,
	}
	resp, err := p.Provider.Chat(ctx, req)
	if err != nil {
		return nil, err
	}
	return p.toPlannedAction(resp)
}

// buildMessagesWithContext uses the context engine (new mode).
func (p *PlannerAdapter) buildMessagesWithContext(ctx context.Context, input runtime.StepInput) ([]Message, error) {
	// This would use the context engine, but to avoid import cycles,
	// the actual integration will be done at the engine construction level
	// For now, fallback to simple mode
	return p.buildMessagesSimple(input), nil
}

// buildMessagesSimple builds messages without context engine (backward compatible).
func (p *PlannerAdapter) buildMessagesSimple(input runtime.StepInput) []Message {
	msgs := []Message{
		{Role: RoleSystem, Content: "You are an autonomous agent. Given a goal and prior steps, choose the next tool to call or indicate completion."},
		{Role: RoleUser, Content: "Goal: " + input.Goal + "\n\nHistory:\n" + formatHistory(input.History)},
	}
	return msgs
}

func formatHistory(h []runtime.StepRecord) string {
	if len(h) == 0 {
		return "(none)"
	}
	var s string
	for i, r := range h {
		s += fmt.Sprintf("Step %d: ", i+1)
		if r.Action != nil {
			s += "Called " + r.Action.Tool + " with " + stringify(r.Action.Input)
			if r.Result != nil {
				s += " -> " + stringify(r.Result.Output)
			}
		}
		if r.Reasoning != "" {
			s += " Reasoning: " + r.Reasoning
		}
		s += "\n"
	}
	return s
}

func stringify(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func (p *PlannerAdapter) buildToolDefs(input runtime.StepInput) []ToolDef {
	if len(p.Tools) == 0 {
		return nil
	}
	var defs []ToolDef
	for _, t := range p.Tools {
		desc := t.Description
		if desc == "" {
			desc = "Tool: " + t.Name
		}
		defs = append(defs, ToolDef{
			Name:        t.Name + "@" + t.Version,
			Description: desc,
		})
	}
	return defs
}

func (p *PlannerAdapter) toPlannedAction(r *ChatResponse) (*runtime.PlannedAction, error) {
	if len(r.ToolCalls) > 0 {
		tc := r.ToolCalls[0]
		var input map[string]any
		_ = json.Unmarshal([]byte(tc.Arguments), &input)
		name := tc.Name
		ver := "1"
		// parse name@version if present
		for i := 0; i < len(name); i++ {
			if name[i] == '@' {
				name, ver = name[:i], name[i+1:]
				break
			}
		}
		return &runtime.PlannedAction{
			Type:      "tool_call",
			Tool:      name,
			Version:   ver,
			Input:     input,
			Reasoning: r.Content,
		}, nil
	}
	return &runtime.PlannedAction{
		Type:      "complete",
		Reasoning: r.Content,
	}, nil
}
