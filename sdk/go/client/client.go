package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the agentd API client.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New creates a new API client.
func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// WithHTTPClient sets a custom HTTP client.
func (c *Client) WithHTTPClient(client *http.Client) *Client {
	c.httpClient = client
	return c
}

// Run represents an agent run.
type Run struct {
	RunID     string    `json:"run_id"`
	AgentName string    `json:"agent_name"`
	Goal      string    `json:"goal"`
	State     string    `json:"state"`
	StepCount int       `json:"step_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateRunRequest is the request for creating a run.
type CreateRunRequest struct {
	AgentName string `json:"agent_name"`
	Goal      string `json:"goal"`
}

// CreateRunResponse is the response from creating a run.
type CreateRunResponse struct {
	RunID string `json:"run_id"`
}

// ListRunsResponse is the response from listing runs.
type ListRunsResponse struct {
	RunIDs []string `json:"run_ids"`
}

// CreateRun creates a new agent run.
func (c *Client) CreateRun(ctx context.Context, agentName, goal string) (*CreateRunResponse, error) {
	req := CreateRunRequest{
		AgentName: agentName,
		Goal:      goal,
	}

	var resp CreateRunResponse
	if err := c.doRequest(ctx, "POST", "/v1/runs", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetRun retrieves details of a specific run.
func (c *Client) GetRun(ctx context.Context, runID string) (*Run, error) {
	var run Run
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/v1/runs/%s", runID), nil, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

// ListRuns lists recent runs.
func (c *Client) ListRuns(ctx context.Context, limit int) (*ListRunsResponse, error) {
	if limit <= 0 {
		limit = 100
	}

	var resp ListRunsResponse
	url := fmt.Sprintf("/v1/runs?limit=%d", limit)
	if err := c.doRequest(ctx, "GET", url, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// CancelRun cancels an ongoing run.
func (c *Client) CancelRun(ctx context.Context, runID string) error {
	var resp map[string]string
	return c.doRequest(ctx, "POST", fmt.Sprintf("/v1/runs/%s/cancel", runID), nil, &resp)
}

// WaitForRun polls until the run completes or fails.
func (c *Client) WaitForRun(ctx context.Context, runID string, pollInterval time.Duration) (*Run, error) {
	if pollInterval == 0 {
		pollInterval = 2 * time.Second
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			run, err := c.GetRun(ctx, runID)
			if err != nil {
				return nil, err
			}

			// Check if terminal state
			if run.State == "completed" || run.State == "failed" || run.State == "cancelled" {
				return run, nil
			}
		}
	}
}

// HealthCheck checks if the service is healthy.
func (c *Client) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %d", resp.StatusCode)
	}

	return nil
}

// doRequest performs an HTTP request with automatic auth and JSON encoding.
func (c *Client) doRequest(ctx context.Context, method, path string, body, result any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return err
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(msg))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}
