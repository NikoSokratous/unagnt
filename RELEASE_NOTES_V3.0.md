# Unagnt v3.0 - Release Notes

**Release Date**: March 2026  
**Status**: Production Ready  
**GitHub**: https://github.com/NikoSokratous/unagnt

---

This document covers all v3.0 changes. For earlier milestones, see [RELEASE_NOTES_V2.0.md](RELEASE_NOTES_V2.0.md) and [RELEASE_NOTES_V1.0.md](RELEASE_NOTES_V1.0.md).

---

## 📋 Summary

| Release | Theme | Status |
|---------|-------|--------|
| **v3.0 Phase 1** | Full Runtime Integration | ✅ Complete |
| **v3.0 Phase 2** | Agentic Orchestration | ✅ Complete |
| **v3.0 Phase 3** | Runtime Hardening and GA Readiness | ✅ Complete |

---

## 🚀 v3.0 Phase 1: Full Runtime Integration

Moved orchestration from simulated execution to real runtime-backed execution paths.

### Runtime Step Execution
- **Location**: `pkg/orchestrate/runtime_executor.go`
- Added `RuntimeStepExecutor` to execute workflow steps through the real runtime engine
- Integrated LLM provider wiring (OpenAI/Anthropic/Ollama), tool registry, MCP sources, and policy executor

### Async Runner Service
- **Location**: `pkg/orchestrate/runner.go`
- Added queue + worker runner for isolated asynchronous execution of runs from API, webhook, scheduler, and events
- Added run lifecycle persistence and cancellation tracking

### Webhook / Scheduler / Event Trigger Wiring
- **Locations**:
  - `pkg/orchestrate/webhook.go`
  - `pkg/orchestrate/scheduler.go`
  - `pkg/orchestrate/triggers.go`
  - `pkg/orchestrate/server.go`
- Webhook-triggered runs execute through runner (with callback payloads)
- Scheduler upgraded to cron semantics via `robfig/cron/v3`
- Added event-trigger endpoint: `POST /v1/triggers/events`

### Checkpoint and Resume
- **Locations**:
  - `pkg/workflow/executor.go`
  - `pkg/workflow/state.go`
  - `cmd/unagnt/workflow.go`
- Added workflow + node checkpoint persistence and resume flow
- Added CLI support for `--resume` with workflow state DB initialization

---

## 🧠 v3.0 Phase 2: Agentic Orchestration

Added smarter orchestration behavior beyond static DAG execution.

### Multi-Model Routing
- **Locations**:
  - `internal/config/agent.go`
  - `pkg/orchestrate/model_routing.go`
  - `pkg/orchestrate/runtime_executor.go`
- Added optional `model_routing` config (`auto`, `cost`, `latency`, `capability`)
- Runtime now selects model candidates dynamically per step goal

### Guardrails and Policy-Aware Tooling
- **Location**: `pkg/orchestrate/runtime_executor.go`
- Runtime executor integrates policy enforcement, risk scoring, and approval gates through policy executor wiring

### Incremental Streaming Signals
- **Locations**:
  - `pkg/orchestrate/stream.go`
  - `pkg/orchestrate/runner.go`
  - `pkg/observe/eventhub.go`
- Added/extended run lifecycle event emission and SSE delivery for incremental execution visibility

---

## 🛡️ v3.0 Phase 3: Runtime Hardening and GA Readiness

Hardened failure handling, observability, and operator troubleshooting workflows.

### Retry / Timeout Hardening
- **Location**: `pkg/orchestrate/runner.go`
- Added per-run controls:
  - `max_retries`
  - `retry_backoff_ms`
  - `timeout_ms`
- Added cancellation-aware backoff and improved terminal-state persistence safety

### Dead-Letter Capture and Replay
- **Locations**:
  - `internal/store/store.go`
  - `internal/store/sqlite.go`
  - `pkg/orchestrate/server.go`
- Added dead-letter persistence for terminal failures
- Added endpoints:
  - `GET /v1/runs/dead-letters`
  - `POST /v1/runs/dead-letters/{id}/replay`
- Replay supports optional overrides (`goal`, retries/backoff/timeout, payload)

### Lifecycle Troubleshooting API
- **Location**: `pkg/orchestrate/server.go`
- Added `GET /v1/runs/{id}/events` for persisted event history

### Observability Metrics
- **Location**: `pkg/orchestrate/metrics.go`
- Added hardening-focused metrics:
  - `agentruntime_run_retries_total`
  - `agentruntime_run_dead_letters_total`
  - `agentruntime_run_queue_depth`
  - `agentruntime_run_queue_rejected_total`
  - `agentruntime_run_failures_total{reason,source}`

---

## ✅ Test and Validation Coverage

Expanded tests for v3 runtime behavior, including:
- Runner queue execution and lifecycle events
- Retry/timeout/cancellation semantics
- Dead-letter persistence and replay
- Event history endpoint and API validation guards
- Workflow checkpoint/resume correctness
- Model routing behavior

Key test files include:
- `pkg/orchestrate/runner_test.go`
- `pkg/orchestrate/server_deadletter_test.go`
- `pkg/orchestrate/server_events_test.go`
- `pkg/orchestrate/server_run_validation_test.go`
- `pkg/orchestrate/model_routing_test.go`
- `pkg/workflow/executor_resume_test.go`
- `pkg/workflow/state_test.go`

---

## 📦 Upgrade Notes (v2.x -> v3.0)

1. Pull latest code and rebuild:
   ```bash
   git pull
   make build
   ```

2. Review API usage updates:
   - Optional run hardening fields on `POST /v1/runs`
   - New dead-letter and run events endpoints

3. Update observability dashboards/alerts for new runtime hardening metrics.

4. For webhook deployments, verify callback and trigger flows with the updated runner-based execution model.

---

## ⚠️ Known Limitations

- Dead-letter retention/archival is operator-managed (no built-in TTL/pruner yet).
- Runner queue is in-memory and bounded; tune worker/queue settings and monitor queue metrics.
- Operator code generation still requires `make generate-operator` before first build.

---

## 🔗 Related Docs

- [README.md](README.md) - Roadmap and known limitations
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - Runtime hardening runbook and migration notes
- [docs/guides/api-integration.md](docs/guides/api-integration.md) - API usage and examples
- [docs/guides/webhooks.md](docs/guides/webhooks.md) - Webhook execution and callback behavior

---

**Version**: 3.0.0  
**Release Date**: March 2026  
**Status**: Production Ready ✅
