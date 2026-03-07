# A/B Testing for Agent Models

v4 adds support for A/B testing different LLM models.

## Overview

Create an A/B test with two models and a traffic split. The runtime (when wired) selects model A or B per run based on the split. Results can be queried for comparison.

## Creating an A/B Test

```bash
curl -X POST http://localhost:8080/v1/ab-tests \
  -H "Content-Type: application/json" \
  -d '{"name":"gpt4-vs-mini","model_a":"gpt-4","model_b":"gpt-4-mini","traffic_split":0.5}'
```

- `model_a`, `model_b`: Model identifiers (e.g. `gpt-4`, `gpt-4-mini`, `openai:gpt-4`)
- `traffic_split`: 0.0–1.0, fraction of traffic to model A

## Listing and Updating

```bash
curl http://localhost:8080/v1/ab-tests
curl -X PATCH http://localhost:8080/v1/ab-tests/{id} -d '{"active":false}'
```

## Results

```bash
curl http://localhost:8080/v1/analytics/ab-tests/{id}/results
```

Returns per-model metrics (requests, latency, error rate) when assignment data is available.

## Runtime Integration

To use A/B tests in runs, the planner/LLM layer must:

1. Look up active A/B tests for the agent/tenant
2. Call `abtest.Selector.SelectModel(ctx, test, runID)`
3. Use the returned model for the LLM call
4. Optionally record the assignment in `ab_test_assignments` for analytics
