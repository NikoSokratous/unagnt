# Release Readiness Gate

This document defines the baseline checks and SLOs for v4-ready status. Use this as a release gate checklist before cutting v4.0 or any production release.

## Required Checks

Before release, all of the following CI jobs must pass:

| Job | Purpose |
|-----|---------|
| Test Go | Unit and package tests |
| Lint | golangci-lint |
| Build | unagnt, unagntd binaries; CLI validation |
| Integration Tests | End-to-end and integration coverage |
| Operator Codegen | Generated operator files up to date |
| Docker Build | Container builds successfully |

Run the full CI pipeline and ensure no failures. For local verification:

```bash
make test
make lint
make build
make generate-operator-check
make generate-crds-check
make build-operator
```

## SLO Baseline

### Queue Behavior

- **Rejection rate**: `agentruntime_run_queue_rejected_total` should not grow unbounded under normal load. Monitor for sustained increases; investigate if rejections spike.
- **Queue depth**: `agentruntime_run_queue_depth` should stay within capacity. Alert if depth consistently exceeds 80% of configured queue size.
- **Latency**: Consider p99 run duration (`agentruntime_run_duration_seconds`) for baseline; tune workers and queue size per environment.

### Dead-Letter Growth

- **Rate**: `agentruntime_run_dead_letters_total` should grow at a rate consistent with expected failure modes. Alert on sudden spikes.
- **Retention**: With dead-letter retention enabled, pruning should keep table size bounded. Monitor `agentruntime_dead_letters_pruned_total` for pruning activity.

### API Health

- `/health` and `/ready` should return 200 under normal conditions.
- API request duration (`agentruntime_api_request_duration_seconds`) can be used for baseline latency SLOs.

## Pre-Release Checklist

- [ ] All CI jobs pass
- [ ] Operator codegen is up to date (`make generate-operator-check`, `make generate-crds-check`)
- [ ] Release notes updated
- [ ] Docs and migration notes reviewed
- [ ] Incident runbooks ([docs/runbooks/incidents.md](runbooks/incidents.md)) available and reviewed
