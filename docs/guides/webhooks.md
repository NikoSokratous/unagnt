# Webhook Integration Guide

## Overview

Webhooks allow external systems to trigger agent executions. This enables event-driven automation, integration with CI/CD pipelines, GitHub actions, payment processors, and more.

## Configuration

Webhooks are configured in a YAML file (`webhooks.yaml`):

```yaml
webhooks:
  - path: /webhook/github
    agent: code-review-bot
    goal_template: "Review PR: {{.pull_request.html_url}}"
    auth_secret: $GITHUB_WEBHOOK_SECRET
    callback_url: "{{.pull_request.statuses_url}}"
    headers:
      Content-Type: application/json

  - path: /webhook/deploy
    agent: deployment-agent
    goal_template: "Deploy {{.repository}} to {{.environment}}"
    auth_secret: ${DEPLOY_SECRET}
```

### Configuration Fields

- **path**: Webhook endpoint (must start with `/`)
- **agent**: Agent config file to execute (e.g., `code-review-bot.yaml`)
- **goal_template**: Go template for dynamic goal generation
- **auth_secret**: HMAC secret for signature verification (supports `$ENV_VAR`)
- **callback_url**: Optional URL to POST results (supports templates)
- **headers**: Custom headers for callback requests

## Security

### Signature Verification

All webhooks support HMAC-SHA256 signature verification. The server checks these headers (in order):

1. `X-Hub-Signature-256` (GitHub format: `sha256=<hex>`)
2. `X-Signature` (Generic format: `<hex>`)
3. `X-Webhook-Signature` (Alternative format: `<hex>`)

**Example verification (GitHub):**

```bash
signature=$(echo -n "$payload" | openssl dgst -sha256 -hmac "$secret" | sed 's/^.* //')
curl -X POST http://localhost:8080/webhook/github \
  -H "X-Hub-Signature-256: sha256=$signature" \
  -H "Content-Type: application/json" \
  -d "$payload"
```

### Secret Management

Secrets can be:
- Hardcoded in config (not recommended)
- Environment variables: `$SECRET` or `${SECRET}`
- External secret managers (via env vars)

```yaml
auth_secret: ${WEBHOOK_SECRET}  # Resolves from environment
```

## Goal Templates

Templates use Go's `text/template` syntax with access to the webhook payload:

```yaml
goal_template: |
  Review pull request #{{.number}} by {{.user.login}}
  Repository: {{.repository.full_name}}
  Changes: {{.pull_request.changed_files}} files
```

**Available context:**
- All fields from the JSON payload
- Nested access: `{{.user.profile.email}}`
- Conditional logic: `{{if .urgent}}High priority{{end}}`

## CLI Management

### Add Webhook

```bash
unagnt webhook add \
  --path /webhook/ci \
  --agent ci-agent \
  --goal "Run tests for {{.commit}}" \
  --secret $CI_SECRET \
  --callback "{{.callback_url}}" \
  --file webhooks.yaml
```

### List Webhooks

```bash
unagnt webhook list --file webhooks.yaml
```

### Test Template

Test goal rendering with a sample payload:

```bash
# Create test payload
echo '{"commit":"abc123","branch":"main"}' > payload.json

# Test rendering
unagnt webhook test \
  --path /webhook/ci \
  --payload payload.json \
  --file webhooks.yaml
```

## Execution Flow

1. **Receive**: Webhook endpoint receives POST request
2. **Verify**: HMAC signature is validated (if configured)
3. **Parse**: JSON payload is parsed
4. **Render**: Goal template is rendered with payload data
5. **Execute**: Agent run is launched asynchronously
6. **Respond**: HTTP 202 Accepted returned immediately
7. **Callback**: Results are POSTed to callback URL (if configured)

## Callback Mechanism

When a callback URL is configured, the runtime POSTs results after execution:

```json
{
  "run_id": "550e8400-e29b-41d4-a716-446655440000",
  "state": "completed",
  "original_payload": { /* original webhook payload */ },
  "output": { /* agent output */ }
}
```

**Retry logic:**
- 3 attempts with exponential backoff (1s, 2s, 4s)
- Custom headers from config are included
- Failures are logged but don't block the agent

## Integrations

### GitHub

```yaml
- path: /webhook/github/pr
  agent: code-reviewer
  goal_template: |
    Review PR #{{.number}}: {{.pull_request.title}}
    Files changed: {{.pull_request.changed_files}}
  auth_secret: $GITHUB_WEBHOOK_SECRET
  callback_url: "{{.pull_request.comments_url}}"
```

Configure in GitHub:
- Webhook URL: `https://your-domain.com/webhook/github/pr`
- Content type: `application/json`
- Secret: Same as `GITHUB_WEBHOOK_SECRET`
- Events: Pull requests

### GitLab

```yaml
- path: /webhook/gitlab/merge
  agent: ci-agent
  goal_template: "Test merge request {{.object_attributes.iid}}"
  auth_secret: $GITLAB_TOKEN
```

### Stripe

```yaml
- path: /webhook/stripe/payment
  agent: payment-processor
  goal_template: "Process payment {{.data.object.id}}"
  auth_secret: $STRIPE_WEBHOOK_SECRET
```

### Custom Services

```yaml
- path: /webhook/custom/deploy
  agent: deployer
  goal_template: |
    Deploy {{.service}} to {{.environment}}
    Version: {{.version}}
    Requested by: {{.user}}
  auth_secret: $CUSTOM_SECRET
  callback_url: "https://api.example.com/callbacks/{{.request_id}}"
  headers:
    Authorization: "Bearer {{.api_token}}"
```

## Testing

### Local Testing

```bash
# Start unagntd with webhooks
unagntd --webhooks webhooks.yaml

# Send test webhook
curl -X POST http://localhost:8080/webhook/test \
  -H "Content-Type: application/json" \
  -d '{"test":"data","value":123}'
```

### Integration Testing

```go
func TestWebhookHandler(t *testing.T) {
    payload := `{"repo":"test","branch":"main"}`
    req := httptest.NewRequest("POST", "/webhook/test", strings.NewReader(payload))
    
    // Add signature
    mac := hmac.New(sha256.New, []byte("secret"))
    mac.Write([]byte(payload))
    signature := hex.EncodeToString(mac.Sum(nil))
    req.Header.Set("X-Signature", signature)
    
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusAccepted, w.Code)
}
```

## Monitoring

Monitor webhook activity:

```bash
# View webhook runs
unagnt logs --source webhook

# Check webhook metrics
curl http://localhost:8080/metrics | grep webhook
```

## Best Practices

1. **Always use secrets**: Never deploy without signature verification
2. **Template validation**: Test templates locally before deploying
3. **Idempotency**: Design agents to handle duplicate webhooks
4. **Timeout configuration**: Set reasonable timeouts in agent configs
5. **Error handling**: Implement proper error recovery in agents
6. **Monitoring**: Set up alerts for webhook failures

## Troubleshooting

### Signature Verification Failed

- Check that `auth_secret` matches the sending service
- Verify the correct header is being used
- Ensure payload is sent as raw body (not URL-encoded)

### Template Rendering Errors

```bash
# Test template locally
unagnt webhook test --path /webhook/test --payload test.json
```

### Agent Not Executing

- Check agent config path is correct
- Verify agent has required permissions
- Review logs: `unagnt logs --run-id <id>`

## Advanced Usage

### Dynamic Agent Selection

```yaml
goal_template: |
  {{if eq .type "urgent"}}
  Agent: emergency-responder
  {{else}}
  Agent: standard-processor
  {{end}}
  Task: {{.description}}
```

### Payload Transformation

```yaml
goal_template: |
  Process order:
  - ID: {{.order_id}}
  - Items: {{range .items}}{{.name}} ({{.quantity}}){{end}}
  - Total: ${{.total}}
```

### Conditional Execution

Use CEL in policy to conditionally allow webhook-triggered runs:

```yaml
rules:
  - name: webhook-hours
    condition: |
      run.source == "webhook" && 
      time.getHours() >= 9 && 
      time.getHours() <= 17
    allow: true
```
