# API Integration Guide

Complete guide to integrating with the Agent Runtime REST API.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Authentication](#authentication)
3. [Endpoints](#endpoints)
4. [Error Handling](#error-handling)
5. [Client Libraries](#client-libraries)
6. [Examples](#examples)

## Getting Started

The Agent Runtime API (`unagntd`) provides HTTP endpoints for managing agent runs.

### Starting the Server

```bash
# Basic
unagntd --addr :8080

# With authentication
export AGENT_RUNTIME_API_KEYS="key1,key2"
unagntd --addr :8080 --store agent.db

# With custom logging
unagntd --addr :8080 --log-level debug
```

### Base URL

Local: `http://localhost:8080`  
Production: `https://api.yourdomain.com`

## Authentication

### API Keys

Set via environment variable:

```bash
export AGENT_RUNTIME_API_KEYS="secret-key-1,secret-key-2"
```

### Making Authenticated Requests

Include in `Authorization` header:

```bash
curl -H "Authorization: Bearer secret-key-1" \
  http://localhost:8080/v1/runs
```

### Public Endpoints (No Auth Required)

- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics

## Endpoints

### POST /v1/runs

Create a new agent run.

**Request:**
```json
{
  "agent_name": "demo-agent",
  "goal": "Calculate 5 + 5"
}
```

**Response (200 OK):**
```json
{
  "run_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/v1/runs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-key" \
  -d '{"agent_name":"demo-agent","goal":"Calculate 5+5"}'
```

---

### GET /v1/runs/{id}

Get details of a specific run.

**Response (200 OK):**
```json
{
  "run_id": "550e8400-e29b-41d4-a716-446655440000",
  "agent_name": "demo-agent",
  "goal": "Calculate 5 + 5",
  "state": "completed",
  "step_count": 3,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:15Z"
}
```

**Example:**
```bash
curl -H "Authorization: Bearer your-key" \
  http://localhost:8080/v1/runs/550e8400-e29b-41d4-a716-446655440000
```

---

### GET /v1/runs

List recent runs.

**Query Parameters:**
- `limit` (int) - Max runs to return (default: 100)

**Response (200 OK):**
```json
{
  "run_ids": [
    "550e8400-e29b-41d4-a716-446655440000",
    "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
  ]
}
```

**Example:**
```bash
curl -H "Authorization: Bearer your-key" \
  "http://localhost:8080/v1/runs?limit=10"
```

---

### POST /v1/runs/{id}/cancel

Cancel a running agent.

**Response (200 OK):**
```json
{
  "status": "cancelled"
}
```

**Example:**
```bash
curl -X POST \
  -H "Authorization: Bearer your-key" \
  http://localhost:8080/v1/runs/550e8400-e29b-41d4-a716-446655440000/cancel
```

---

### GET /health

Health check (no auth required).

**Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

---

### GET /ready

Readiness check (no auth required).

**Response (200 OK):**
```json
{
  "status": "ready",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Response (503 Service Unavailable):**
```json
{
  "status": "not ready",
  "error": "store unavailable"
}
```

---

### GET /metrics

Prometheus metrics (no auth required).

**Response:** Prometheus text format

```
# HELP Unagnt_runs_created_total Total number of runs created
# TYPE Unagnt_runs_created_total counter
Unagnt_runs_created_total 42
```

## Error Handling

### HTTP Status Codes

- `200` - Success
- `400` - Bad Request (invalid JSON, missing fields)
- `401` - Unauthorized (missing or invalid API key)
- `404` - Not Found (run doesn't exist)
- `500` - Internal Server Error
- `503` - Service Unavailable (not ready)

### Error Response Format

```json
{
  "error": "description of what went wrong"
}
```

### Retry Strategy

```javascript
async function createRunWithRetry(agentName, goal, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await fetch('/v1/runs', {
        method: 'POST',
        headers: {
          'Authorization': 'Bearer ' + apiKey,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({agent_name: agentName, goal: goal})
      });
      
      if (response.status === 429) {
        // Rate limited, exponential backoff
        await sleep(Math.pow(2, i) * 1000);
        continue;
      }
      
      if (!response.ok) {
        throw new Error(`API error: ${response.status}`);
      }
      
      return await response.json();
    } catch (err) {
      if (i === maxRetries - 1) throw err;
      await sleep(1000 * (i + 1));
    }
  }
}
```

## Client Libraries

### Go

```go
import "github.com/Unagnt/Unagnt/sdk/go/client"

c := client.New("http://localhost:8080", "your-api-key")
resp, err := c.CreateRun(ctx, "demo-agent", "test goal")
```

See: [sdk/go/README.md](../../sdk/go/README.md)

### Python

```python
from Unagnt import Unagnt

client = Unagnt(base_url="http://localhost:8080", api_key="key")
run_id = client.create_run("demo-agent", "test goal")
```

See: [sdk/python/README.md](../../sdk/python/README.md)

### JavaScript/TypeScript (Coming Soon)

## Examples

### Full Workflow (JavaScript)

```javascript
const API_KEY = process.env.AGENT_RUNTIME_API_KEY;
const BASE_URL = 'http://localhost:8080';

async function runAgent(agentName, goal) {
  // 1. Create run
  const createResp = await fetch(`${BASE_URL}/v1/runs`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${API_KEY}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({agent_name: agentName, goal: goal})
  });
  
  const {run_id} = await createResp.json();
  console.log(`Created run: ${run_id}`);
  
  // 2. Poll for completion
  while (true) {
    const statusResp = await fetch(`${BASE_URL}/v1/runs/${run_id}`, {
      headers: {'Authorization': `Bearer ${API_KEY}`}
    });
    
    const run = await statusResp.json();
    console.log(`State: ${run.state}, Steps: ${run.step_count}`);
    
    if (['completed', 'failed', 'cancelled'].includes(run.state)) {
      return run;
    }
    
    await new Promise(resolve => setTimeout(resolve, 2000));
  }
}

// Usage
runAgent('demo-agent', 'Calculate 15 + 27')
  .then(run => console.log('Final state:', run.state))
  .catch(err => console.error('Error:', err));
```

### Python with Error Handling

```python
from Unagnt import Unagnt
from Unagnt.errors import APIError, TimeoutError

client = Unagnt(
    base_url="http://localhost:8080",
    api_key="your-key"
)

try:
    # Create and wait
    run_id = client.create_run("demo-agent", "test goal")
    run = client.wait_for_run(run_id, timeout=60)
    
    if run.state == "completed":
        print(f"Success! Steps: {run.step_count}")
    else:
        print(f"Failed with state: {run.state}")
        
except TimeoutError:
    print("Run timed out")
    client.cancel_run(run_id)
    
except APIError as e:
    print(f"API error {e.status_code}: {e.message}")
```

## Webhook Integration (Planned)

Future support for webhooks to trigger agents:

```yaml
# agent-webhook.yaml
webhooks:
  - path: /webhooks/github
    agent: code-review-bot
    goal_template: "Review PR #{body.pull_request.number}"
```

## Rate Limiting (Planned)

Future API will include rate limiting headers:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642348800
```

## OpenAPI Specification

Full API spec: [api/openapi.yaml](../../api/openapi.yaml)

Import into Postman, Insomnia, or generate client code.

## Monitoring

### Prometheus Metrics

Scrape `/metrics` for observability:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'Unagnt'
    static_configs:
      - targets: ['localhost:8080']
```

### Health Checks in Kubernetes

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 30

readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
```

## Security

### Best Practices

1. **Use HTTPS** in production
2. **Rotate API keys** regularly
3. **Set strong keys** (min 32 characters)
4. **Limit key scope** (per-agent keys coming soon)
5. **Monitor `/metrics`** for anomalies

### Example: Key Rotation

```bash
# Generate new keys
export NEW_KEYS="$(openssl rand -hex 32),$(openssl rand -hex 32)"

# Update unagntd
export AGENT_RUNTIME_API_KEYS="$OLD_KEYS,$NEW_KEYS"

# Graceful migration period
# Then remove old keys after client migration
```

## Next Steps

- Try the [Go client](../../sdk/go/)
- Use the [Python SDK](../../sdk/python/)
- See [examples/](../../examples/) for complete apps

## Support

- GitHub Issues: https://github.com/NikoSokratous/unagnt/issues
- Discord: (coming soon)
