package llm

import "encoding/json"

// Role identifies the speaker in a conversation turn.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message is a unified message type for any LLM provider.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a completion request.
type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Tools       []ToolDef `json:"tools,omitempty"`
}

// ToolDef describes a tool for function calling.
type ToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  Parameters      `json:"parameters,omitempty"`
	Schema      json.RawMessage `json:"-"` // Optional: raw JSON Schema for parameters (used when non-nil)
}

// Parameters describes JSON Schema for tool input.
type Parameters struct {
	Type       string           `json:"type,omitempty"`
	Properties map[string]Param `json:"properties,omitempty"`
	Required   []string         `json:"required,omitempty"`
}

// Param describes a single parameter.
type Param struct {
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}

// ChatResponse is the unified completion response.
type ChatResponse struct {
	Content      string        `json:"content,omitempty"`
	ToolCalls    []ToolCallRef `json:"tool_calls,omitempty"`
	FinishReason string        `json:"finish_reason,omitempty"`
	Model        string        `json:"model,omitempty"`
	Usage        Usage         `json:"usage,omitempty"`
}

// ToolCallRef references a tool invocation in the response.
type ToolCallRef struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Usage tracks token consumption.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
