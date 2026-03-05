package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/NikoSokratous/unagnt/pkg/llm"
)

// Client is the Ollama API client for local models.
type Client struct {
	BaseURL string
	Model   string
	HTTP    *http.Client
}

// NewClient creates an Ollama client.
func NewClient(model string) *Client {
	return &Client{
		BaseURL: "http://localhost:11434",
		Model:   model,
		HTTP:    http.DefaultClient,
	}
}

// Name implements llm.Provider.
func (c *Client) Name() string {
	return "ollama"
}

// Chat implements llm.Provider.
func (c *Client) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	body := map[string]any{
		"model":    c.Model,
		"messages": toOllamaMessages(req.Messages),
		"stream":   false,
	}
	if req.Temperature > 0 {
		body["options"] = map[string]float64{"temperature": req.Temperature}
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/api/chat", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama api error %d: %s", resp.StatusCode, string(msg))
	}

	var out ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return toLLMResponse(&out), nil
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Model string `json:"model"`
}

func toOllamaMessages(msgs []llm.Message) []ollamaMessage {
	out := make([]ollamaMessage, len(msgs))
	for i, m := range msgs {
		out[i] = ollamaMessage{Role: string(m.Role), Content: m.Content}
	}
	return out
}

func toLLMResponse(r *ollamaResponse) *llm.ChatResponse {
	return &llm.ChatResponse{
		Content: r.Message.Content,
		Model:   r.Model,
	}
}
