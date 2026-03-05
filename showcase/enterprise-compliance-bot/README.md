# Enterprise Compliance Bot - Showcase

End-to-end demo deployable via one command. Demonstrates policy governance (block data exfil, require approval), cost tracking, and policy denials.

## Quick Start

### Option 1: unagnt (no K8s)

```bash
cd showcase/enterprise-compliance-bot
unagnt run --config agent.yaml --goal "Calculate 15% of 240 and echo the result"
```

### Option 2: Kubernetes

```bash
# Deploy Unagnt (see project root)
helm install Unagnt ./k8s/helm -f k8s/helm/values.yaml

# Apply showcase resources
kubectl apply -f k8s/agent.yaml
kubectl apply -f k8s/policy.yaml
```

### Option 3: Make (if available)

```bash
make showcase-deploy
# Or: helm install Unagnt ./k8s/helm && kubectl apply -f showcase/enterprise-compliance-bot/k8s/
```

## Policy

- **Block**: External HTTP requests (data exfiltration)
- **Block**: Requests to known exfil domains (evil.com, pastebin.com, etc.)
- **Require approval**: Non-GET HTTP requests to internal URLs

## Cost + Denials Dashboard

Start the web UI and navigate to **Analytics**:

```bash
# From project root
make run   # or: go run ./cmd/unagnt web
# Open http://localhost:8080, go to /analytics
```

The Analytics page shows:
- Total cost and cost by agent
- Policy denials (when audit logging is enabled)

## Files

| File | Purpose |
|------|---------|
| `agent.yaml` | Agent config (tools, model, policy) |
| `policy.yaml` | CEL policy rules |
| `k8s/agent.yaml` | K8s Agent CR |
| `k8s/policy.yaml` | K8s Policy CR |
