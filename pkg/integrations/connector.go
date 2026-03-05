package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Connector sends notifications to external systems (Slack, Teams, Jira, etc.).
type Connector interface {
	Send(ctx context.Context, msg *Message) error
}

// Message represents an outbound integration message.
type Message struct {
	Type     string            `json:"type"` // approval_request, alert, notification
	Title    string            `json:"title"`
	Body     string            `json:"body"`
	Actions  []Action          `json:"actions,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	TenantID string            `json:"tenant_id,omitempty"`
}

// Action represents an actionable button (e.g. Approve/Deny).
type Action struct {
	Label string `json:"label"`
	URL   string `json:"url"`
	Value string `json:"value"`
}

// Config holds connector configuration per tenant.
type Config struct {
	Type       string            `yaml:"type"` // slack, teams, webhook
	WebhookURL string            `yaml:"webhook_url"`
	Token      string            `yaml:"token,omitempty"`
	Extra      map[string]string `yaml:"extra,omitempty"`
}

// NewConnector creates a connector from config.
func NewConnector(cfg *Config) (Connector, error) {
	if cfg == nil {
		return nil, fmt.Errorf("integrations: config required")
	}
	switch cfg.Type {
	case "slack":
		return NewSlackConnector(cfg), nil
	case "teams":
		return NewTeamsConnector(cfg), nil
	case "webhook":
		return NewWebhookConnector(cfg), nil
	default:
		return nil, fmt.Errorf("integrations: unsupported type %q", cfg.Type)
	}
}

// WebhookConnector posts JSON to a URL (generic).
type WebhookConnector struct {
	url    string
	client *http.Client
}

// NewWebhookConnector creates a generic webhook connector.
func NewWebhookConnector(cfg *Config) *WebhookConnector {
	return &WebhookConnector{
		url:    cfg.WebhookURL,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send posts the message as JSON to the webhook URL.
func (w *WebhookConnector) Send(ctx context.Context, msg *Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", w.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned %d", resp.StatusCode)
	}
	return nil
}
