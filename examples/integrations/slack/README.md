# Slack Integration for Approvals and Alerts

Use Slack Incoming Webhooks to send approval requests and alerts.

## Setup

1. Create a Slack Incoming Webhook: https://api.slack.com/messaging/webhooks
2. Configure the webhook URL in your integration config.

## Config

```yaml
integrations:
  - type: slack
    webhook_url: https://hooks.slack.com/services/xxx/yyy/zzz
    # Per-tenant: add tenant_id in config
```

## Approval Request Example

```go
import "github.com/Unagnt/Unagnt/pkg/integrations"

conn, _ := integrations.NewConnector(&integrations.Config{
    Type:       "slack",
    WebhookURL: "https://hooks.slack.com/services/...",
})
conn.Send(ctx, &integrations.Message{
    Type:  "approval_request",
    Title: "Tool approval required",
    Body:  "Agent wants to run `http_request`",
    Actions: []integrations.Action{
        {Label: "Approve", URL: "https://approvals.example.com/approve/123"},
        {Label: "Deny", URL: "https://approvals.example.com/deny/123"},
    },
})
```
