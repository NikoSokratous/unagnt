package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Event represents a streaming event from the agent runtime.
type Event struct {
	RunID     string         `json:"run_id"`
	StepID    string         `json:"step_id,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Type      string         `json:"type"`
	Agent     string         `json:"agent"`
	Data      map[string]any `json:"data,omitempty"`
	Model     ModelMeta      `json:"model,omitempty"`
}

// ModelMeta captures model metadata.
type ModelMeta struct {
	Provider string `json:"provider,omitempty"`
	Name     string `json:"name,omitempty"`
	Version  string `json:"version,omitempty"`
}

// StreamEvents opens a Server-Sent Events connection to stream run events.
func (c *Client) StreamEvents(ctx context.Context, runID string) (<-chan Event, <-chan error, error) {
	url := fmt.Sprintf("%s/v1/runs/%s/stream", c.baseURL, runID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, nil, fmt.Errorf("stream failed: status %d", resp.StatusCode)
	}

	eventChan := make(chan Event, 10)
	errChan := make(chan error, 1)

	go func() {
		defer close(eventChan)
		defer close(errChan)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)

		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines and comments (heartbeats)
			if line == "" || line[0] == ':' {
				continue
			}

			// SSE format: "data: {...}"
			if len(line) > 6 && line[:6] == "data: " {
				data := line[6:]

				var event Event
				if err := json.Unmarshal([]byte(data), &event); err != nil {
					errChan <- fmt.Errorf("parse event: %w", err)
					return
				}

				select {
				case eventChan <- event:
				case <-ctx.Done():
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			select {
			case errChan <- fmt.Errorf("scan error: %w", err):
			case <-ctx.Done():
			}
		}
	}()

	return eventChan, errChan, nil
}
