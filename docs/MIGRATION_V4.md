# Migration to v4

This document describes v4 changes and how to migrate.

## Agent Usage Analytics by Workflow & Model

### Schema

Migration 015 adds `workflow_id` and `workflow_name` to `cost_entries`:

```bash
cat migrations/015_cost_workflow_analytics.sql | sqlite3 your.db
```

### API

- `GET /v1/analytics/costs/workflows?tenant_id=&range=` – costs by workflow
- `GET /v1/analytics/costs/breakdown?workflow_id=&model=` – optional filters

### TrackLLMCall

`TrackLLMCall` now accepts optional `workflowID` and `workflowName`. Pass `""` when not in a workflow.

## Model Drift and Performance

### Schema

Migration 016 creates `model_performance_snapshots`:

```bash
cat migrations/016_model_performance.sql | sqlite3 your.db
```

### API

- `GET /v1/analytics/model-performance?provider=&model_id=&limit=`
- `GET /v1/analytics/model-drift?provider=&model_id=`

Requires `AnalyticsAPI.SetModelPerformanceStore(mlops.NewPerformanceStore(db))` and `mlops.Collector` wiring for LLM calls.

### Runbook

See [docs/runbooks/model-drift.md](runbooks/model-drift.md).

## Human Review Queues

### Schema

Migration 017 creates `approval_requests`:

```bash
cat migrations/017_approval_requests.sql | sqlite3 your.db
```

### API

- `GET /v1/approvals/pending`
- `GET /v1/approvals/{id}`
- `POST /v1/approvals/{id}/approve`
- `POST /v1/approvals/{id}/deny`

### CLI

```bash
unagnt approvals list [--url http://localhost:8080]
unagnt approvals approve <id>
unagnt approvals deny <id>
```

### Runbook

See [docs/runbooks/approval-queues.md](runbooks/approval-queues.md).

## Compliance Report Generation

### API

- `POST /v1/compliance/reports/generate` – body: `{"type":"daily"|"weekly"|"monthly"|"custom", "date"?, "start"?, "end"?}`
- `GET /v1/compliance/reports?type=&limit=`
- `GET /v1/compliance/reports/{id}`
- `GET /v1/compliance/reports/{id}/export?format=json|csv|cef`

Requires `risk.ReportGenerator` and `compliance_reports` table (migration 007).

## A/B Testing

### Schema

Migration 018 creates `ab_test_assignments`:

```bash
cat migrations/018_ab_test_assignments.sql | sqlite3 your.db
```

### API

- `POST /v1/ab-tests` – create
- `GET /v1/ab-tests`
- `PATCH /v1/ab-tests/{id}` – `{"active": true|false}`
- `GET /v1/analytics/ab-tests/{id}/results`

### Guide

See [docs/guides/ab-testing.md](guides/ab-testing.md).
