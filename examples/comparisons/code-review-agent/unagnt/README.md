# Unagnt Implementation

Production-ready code review agent with built-in governance, observability, and cost tracking.

## Features Included

✅ Policy enforcement (security, cost limits)  
✅ Cost tracking per execution  
✅ Human approval gates  
✅ Automatic retries with backoff  
✅ Distributed tracing  
✅ Deterministic replay for debugging  
✅ Multi-step workflow with error handling  

## Files

- `code-review.yaml` - Main workflow definition (80 lines)
- `policy.yaml` - Security and cost policies (40 lines)
- `custom-tools.go` - Custom GitHub integration tool (optional, 30 lines)

**Total**: ~150 lines for full production system

## Running

```bash
# Apply policy
unagnt policy apply policy.yaml

# Run workflow
unagnt workflow run code-review.yaml \
  --param pr_url=https://github.com/user/repo/pull/123 \
  --param severity_threshold=medium \
  --param auto_comment=false

# Watch execution in real-time
unagnt runs watch <run-id>

# View cost breakdown
unagnt costs --run <run-id>

# Replay for debugging
unagnt replay <run-id> --mode debug
```

## What You Get Out-of-the-Box

1. **Policy Enforcement**: Automatically blocks dangerous operations
2. **Cost Tracking**: See exactly what each step costs
3. **Approval Gates**: Human review before posting comments
4. **Retry Logic**: Automatic retry with exponential backoff
5. **Observability**: Full tracing, metrics, structured logs
6. **Debugging**: Replay any execution deterministically
7. **Security**: Permission-gated tool execution

## Key Highlights

### Declarative Configuration
```yaml
# No code needed - just declare what you want
name: "code-review-agent"
human_approval:
  required_for: ["post_comment"]
retry:
  max_attempts: 3
  backoff: exponential
cost_tracking: true
```

### Built-in Governance
```yaml
# Policies prevent problems before they happen
rules:
  - id: limit-api-calls
    condition: |
      tool.name == "github_api" && 
      estimated_cost > 5.0
    action: deny
```

### Automatic Observability
```bash
# Every execution is traced
unagnt runs get <run-id>
# Shows: duration, cost, tool calls, state transitions

# Replay any run
unagnt replay <run-id>
# Reproduce exact behavior for debugging
```

## Architecture

```
GitHub PR Event
    ↓
Workflow Engine
    ↓
┌─────────────────────────┐
│ Step 1: Fetch PR        │ → Policy Check → Execute → Log
├─────────────────────────┤
│ Step 2: Static Analysis │ → Policy Check → Execute → Log
├─────────────────────────┤
│ Step 3: Code Review     │ → Policy Check → Execute → Log
├─────────────────────────┤
│ Step 4: Check Breaking  │ → Policy Check → Execute → Log
├─────────────────────────┤
│ Step 5: Generate Report │ → Policy Check → Execute → Log
├─────────────────────────┤
│ Step 6: Post Comment    │ → Approval Gate → Execute → Log
└─────────────────────────┘
    ↓
Cost Report + Trace
```

All steps automatically:
- ✅ Enforced by policy
- ✅ Tracked for cost
- ✅ Traced for observability
- ✅ Retried on failure
- ✅ Logged for audit

## Comparison to Manual Implementation

| Feature | Unagnt | Manual Code |
|---------|-------------|-------------|
| Workflow definition | 80 lines YAML | 200+ lines code |
| Policy enforcement | 40 lines YAML | 100+ lines code |
| Cost tracking | Built-in | 50+ lines code |
| Retries | 3 lines config | 30+ lines code |
| Tracing | Automatic | 50+ lines + vendor |
| Approval gates | Declarative | 40+ lines code |
| Replay/debug | Built-in | Not feasible |
| **Total effort** | **150 lines config** | **470+ lines code** |

## Production Deployment

```bash
# Deploy to Kubernetes
helm install code-review-agent Unagnt/Unagnt \
  --set workflows.codeReview.enabled=true \
  --set policies.production=true

# Auto-scale based on queue depth
kubectl autoscale deployment code-review-agent \
  --min=2 --max=10 \
  --metric=custom.googleapis.com/workflow-queue-depth
```

## Next Steps

1. Customize the workflow for your needs
2. Add custom tools if needed
3. Configure webhooks for auto-triggering
4. Set up cost alerts
5. Deploy to production

That's it! No massive codebase to maintain. 🚀
