package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/NikoSokratous/unagnt/pkg/runtime"
)

// ToolInfo describes an available tool for the planner.
type ToolInfo struct {
	Name        string
	Version     string
	Description string
	InputSchema []byte // JSON Schema for tool parameters (passed to LLM for function calling)
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
	systemContent := `You are an autonomous agent. Given a goal and prior steps, choose ONE action:

1. Call a tool: Use the appropriate tool(s) for the goal. Provide valid JSON arguments matching each tool's schema.
   - For echoing/outputting text: use echo with {"message": "..."}
   - For calculations: use calc
   - For HTTP: use http_request
   - Use whatever tool fits the goal.

2. Complete: Respond with text only (no tool call) when the goal is achieved.

RULES:
- You MUST use tools to accomplish the goal—do not complete without calling a tool when the goal requires tool output.
- Complete ONLY when ALL parts of the goal are achieved. If the goal says "fetch and echo" or "calculate and output", you MUST call each required tool. Stating "I will echo" in reasoning is NOT enough—you must actually call the echo tool.
- For multi-step goals: call tools in sequence (e.g. http_request first, then echo with the result). One tool per step.
- Once every part of the goal is satisfied (e.g. both fetch and echo have run), then complete. Do not complete if any requested action is missing from the history.`

	userContent := "Goal: " + input.Goal + "\n\nHistory:\n" + formatHistory(input.History)
	if len(input.History) > 0 {
		userContent += "\n\nIf the goal is satisfied by the history above, respond with completion (text only, no tool call)."
	}

	msgs := []Message{
		{Role: RoleSystem, Content: systemContent},
		{Role: RoleUser, Content: userContent},
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
			if r.Result != nil && r.Result.Output != nil {
				// Show echoed output clearly so model can see goal was achieved
				if echoed, ok := r.Result.Output["echoed"]; ok {
					s += " -> output: " + stringify(echoed)
				} else {
					s += " -> " + stringify(r.Result.Output)
				}
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
		// OpenAI requires tool names to match ^[a-zA-Z0-9_-]+$ (no @)
		name := t.Name + "_v" + t.Version
		def := ToolDef{Name: name, Description: desc}
		if len(t.InputSchema) > 0 {
			def.Schema = t.InputSchema
		}
		defs = append(defs, def)
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
		// parse name_vX format (OpenAI-safe; @ not allowed in tool names)
		if idx := strings.LastIndex(name, "_v"); idx >= 0 {
			name, ver = name[:idx], name[idx+2:]
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
