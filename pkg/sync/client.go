package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client is the sync API client.
type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewClient creates a sync client.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Push sends a bundle to the server.
func (c *Client) Push(ctx context.Context, bundle *DeltaBundle) error {
	body, err := json.Marshal(bundle)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/sync/push", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("push failed: %d", resp.StatusCode)
	}
	return nil
}

// Pull fetches a bundle from the server (optionally since a timestamp).
func (c *Client) Pull(ctx context.Context, sinceTime time.Time) (*DeltaBundle, error) {
	url := c.baseURL + "/v1/sync/pull"
	if !sinceTime.IsZero() {
		url += "?since=" + sinceTime.Format(time.RFC3339)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pull failed: %d", resp.StatusCode)
	}
	var bundle DeltaBundle
	if err := json.NewDecoder(resp.Body).Decode(&bundle); err != nil {
		return nil, err
	}
	return &bundle, nil
}
