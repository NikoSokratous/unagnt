package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/NikoSokratous/unagnt/pkg/llm"
)

// Client is the Anthropic API client.
type Client struct {
	APIKey     string
	Model      string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates an Anthropic client.
func NewClient(apiKey, model string) *Client {
	return &Client{
		APIKey:     apiKey,
		Model:      model,
		BaseURL:    "https://api.anthropic.com/v1",
		HTTPClient: http.DefaultClient,
	}
}

// Name implements llm.Provider.
func (c *Client) Name() string {
	return "anthropic"
}

// Chat implements llm.Provider.
func (c *Client) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	messages := toAnthropicMessages(req.Messages)

	body := map[string]any{
		"model":      c.Model,
		"messages":   messages,
		"max_tokens": 4096,
	}

	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	}

	// Add system message if present
	var systemMsg string
	for _, msg := range req.Messages {
		if msg.Role == llm.RoleSystem {
			systemMsg = msg.Content
			break
		}
	}
	if systemMsg != "" {
		body["system"] = systemMsg
	}

	// Add tools if present
	if len(req.Tools) > 0 {
		body["tools"] = toAnthropicTools(req.Tools)
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/messages", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic api error %d: %s", resp.StatusCode, string(msg))
	}

	var out anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return toLLMResponse(&out), nil
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type  string `json:"type"`
		Text  string `json:"text,omitempty"`
		ID    string `json:"id,omitempty"`
		Name  string `json:"name,omitempty"`
		Input any    `json:"input,omitempty"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func toAnthropicMessages(msgs []llm.Message) []anthropicMessage {
	var out []anthropicMessage
	for _, m := range msgs {
		// Skip system messages (handled separately)
		if m.Role == llm.RoleSystem {
			continue
		}
		out = append(out, anthropicMessage{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}
	return out
}

func toAnthropicTools(tools []llm.ToolDef) []map[string]any {
	out := make([]map[string]any, len(tools))
	for i, t := range tools {
		out[i] = map[string]any{
			"name":        t.Name,
			"description": t.Description,
			"input_schema": map[string]any{
				"type":       t.Parameters.Type,
				"properties": t.Parameters.Properties,
				"required":   t.Parameters.Required,
			},
		}
	}
	return out
}

func toLLMResponse(r *anthropicResponse) *llm.ChatResponse {
	res := &llm.ChatResponse{
		Model: r.Model,
		Usage: llm.Usage{
			PromptTokens:     r.Usage.InputTokens,
			CompletionTokens: r.Usage.OutputTokens,
			TotalTokens:      r.Usage.InputTokens + r.Usage.OutputTokens,
		},
		FinishReason: r.StopReason,
	}

	// Extract text and tool calls from content blocks
	for _, block := range r.Content {
		switch block.Type {
		case "text":
			res.Content += block.Text
		case "tool_use":
			argsBytes, _ := json.Marshal(block.Input)
			res.ToolCalls = append(res.ToolCalls, llm.ToolCallRef{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: string(argsBytes),
			})
		}
	}

	return res
}
