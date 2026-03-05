# HITL (Human-in-the-Loop) Demo

End-to-end demo of the approval gate: agent pauses, approval server receives request, human approves via HTTP, execution continues.

## Flow

1. Agent runs with a policy that has `require_approval` for certain tools (e.g. non-GET HTTP)
2. When the agent tries to use such a tool, the PolicyExecutor invokes the ApprovalGate
3. ApprovalGate POSTs the request to the approval server and blocks
4. Human visits approval server, sees pending request, approves or denies
5. Approval server responds; agent continues or fails

## Quick Start

### Terminal 1: Start approval server

```bash
cd showcase/hitl-demo/approval-server
go run .
# Listens on :9090
```

### Terminal 2: Run agent with approval webhook

```bash
unagnt run -c showcase/hitl-demo/agent.yaml -g "Make a POST request to http://localhost:8080/health with body 'test'" --approval-webhook http://localhost:9090/request
```

When the agent tries the POST, it will block. In another terminal:

```bash
# List pending
curl http://localhost:9090/pending

# Approve (use the id from pending)
curl -X POST http://localhost:9090/approve/<id>
```

The agent will then continue.

### CLI approval (default)

Without `--approval-webhook`, unagnt uses stdin: you'll be prompted "Approve? [y/N]:" in the same terminal.

## Slack Integration

To forward approval requests to Slack:

1. Create a Slack Incoming Webhook
2. Modify the approval server to POST to Slack when a request arrives
3. Use Slack's "Approve" / "Deny" buttons that call back to your approval server

Example (add to approval server):

```go
// On new request:
slack.Post(webhookURL, map[string]string{
    "text": fmt.Sprintf("Approval needed: %s with input %v", pr.Tool, pr.Input),
    "attachments": [{"actions": [
        {"type":"button","text":"Approve","url":"http://yourserver:9090/approve/"+id},
        {"type":"button","text":"Deny","url":"http://yourserver:9090/deny/"+id},
    ]}],
})
```

## Recording a Video

1. Start approval server in background
2. Run agent with `--approval-webhook`
3. Show "blocked" state in agent terminal
4. curl /pending, then /approve
5. Show agent completing

## Files

| File | Purpose |
|------|---------|
| `agent.yaml` | Agent config |
| `policy.yaml` | CEL policy with require_approval |
| `approval-server/main.go` | Reference approval webhook server |
