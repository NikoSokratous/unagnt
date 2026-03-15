package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/NikoSokratous/unagnt/pkg/llm"
)

// Client is the OpenAI API client.
type Client struct {
	APIKey  string
	BaseURL string // defaults to https://api.openai.com/v1
	Model   string
	HTTP    *http.Client
}

// NewClient creates an OpenAI client.
func NewClient(apiKey, model string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: "https://api.openai.com/v1",
		Model:   model,
		HTTP:    http.DefaultClient,
	}
}

// Name implements llm.Provider.
func (c *Client) Name() string {
	return "openai"
}

// Chat implements llm.Provider.
func (c *Client) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	body := map[string]any{
		"model":       c.Model,
		"messages":    toOpenAIMessages(req.Messages),
		"temperature": req.Temperature,
	}
	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	}
	if len(req.Tools) > 0 {
		body["tools"] = toOpenAITools(req.Tools)
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/chat/completions", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai api error %d: %s", resp.StatusCode, string(msg))
	}

	var out openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return toLLMResponse(&out), nil
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content   string `json:"content"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func toOpenAIMessages(msgs []llm.Message) []openAIMessage {
	out := make([]openAIMessage, len(msgs))
	for i, m := range msgs {
		out[i] = openAIMessage{Role: string(m.Role), Content: m.Content}
	}
	return out
}

func toOpenAITools(tools []llm.ToolDef) []map[string]any {
	out := make([]map[string]any, len(tools))
	for i, t := range tools {
		fn := map[string]any{
			"name":        t.Name,
			"description": t.Description,
		}
		if len(t.Schema) > 0 {
			var params map[string]any
			if err := json.Unmarshal(t.Schema, &params); err == nil {
				fn["parameters"] = params
			} else {
				fn["parameters"] = t.Parameters
			}
		} else {
			fn["parameters"] = t.Parameters
		}
		out[i] = map[string]any{
			"type":     "function",
			"function": fn,
		}
	}
	return out
}

func toLLMResponse(r *openAIResponse) *llm.ChatResponse {
	res := &llm.ChatResponse{
		Model: r.Model,
		Usage: llm.Usage{
			PromptTokens:     r.Usage.PromptTokens,
			CompletionTokens: r.Usage.CompletionTokens,
			TotalTokens:      r.Usage.TotalTokens,
		},
	}
	if len(r.Choices) > 0 {
		c := &r.Choices[0]
		res.Content = c.Message.Content
		res.FinishReason = c.FinishReason
		for _, tc := range c.Message.ToolCalls {
			res.ToolCalls = append(res.ToolCalls, llm.ToolCallRef{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			})
		}
	}
	return res
}
