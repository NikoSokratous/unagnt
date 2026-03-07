# Model Drift Monitoring

This runbook covers model performance monitoring and drift detection for LLM models (v4).

## Overview

Model drift occurs when model performance degrades over time (e.g., increased latency, error rate). The system collects LLM call outcomes, persists performance snapshots, and compares recent vs baseline to flag drift.

## Metrics

- **latency_p50/p95/p99_ms**: Latency percentiles per model
- **error_rate**: Fraction of failed calls
- **throughput**: Approximate calls per minute

## Endpoints

| Endpoint | Purpose |
|----------|---------|
| `GET /v1/analytics/model-performance?provider=&model_id=&limit=` | Recent performance snapshots |
| `GET /v1/analytics/model-drift?provider=&model_id=` | Drift detection result |

## Thresholds

- **Latency**: Drift when recent p99 > 1.5× baseline p99
- **Error rate**: Drift when recent error rate exceeds baseline by 0.05 (5%)

Override via query params (when supported) or configuration.

## Escalation

1. **Check model-performance** – confirm latency/error trend
2. **Review recent deployments** – model or provider changes
3. **Check provider status** – OpenAI, Anthropic, etc.
4. **Consider fallback** – switch to backup model or provider
5. **Trigger retrain** – for custom models, use `Monitor.TriggerRetrain` (if wired)

## Data Retention

Snapshots are stored in `model_performance_snapshots`. Baseline uses data from the last 7 days. Prune old snapshots if storage is a concern (e.g., retain 30 days).
