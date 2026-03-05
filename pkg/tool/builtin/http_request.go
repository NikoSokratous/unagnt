package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/tool"
)

// HTTPRequest is a built-in tool for making HTTP requests.
type HTTPRequest struct{}

func (HTTPRequest) Name() string        { return "http_request" }
func (HTTPRequest) Version() string     { return "1" }
func (HTTPRequest) Description() string { return "Make an HTTP request to a URL" }
func (HTTPRequest) Permissions() []tool.Permission {
	return []tool.Permission{{Scope: "net:external", Required: true}}
}

func (HTTPRequest) InputSchema() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":     "object",
		"required": []string{"url", "method"},
		"properties": map[string]any{
			"url":    map[string]string{"type": "string", "description": "URL to request"},
			"method": map[string]string{"type": "string", "description": "HTTP method (GET, POST, etc)"},
			"body":   map[string]string{"type": "string", "description": "Request body (optional)"},
		},
	})
}

func (h HTTPRequest) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
	var req struct {
		URL    string `json:"url"`
		Method string `json:"method"`
		Body   string `json:"body"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if req.URL == "" {
		return nil, fmt.Errorf("url is required")
	}
	if req.Method == "" {
		req.Method = "GET"
	}

	var body io.Reader
	if req.Body != "" {
		body = bytes.NewReader([]byte(req.Body))
	}
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, body)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"body":        string(respBody),
	}, nil
}
