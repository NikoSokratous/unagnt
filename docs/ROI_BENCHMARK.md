# ROI Benchmark: Governance Reduces Cost

This document provides a methodology and benchmark scenario demonstrating how policy governance in Unagnt reduces operational cost.

## Executive Summary

- **Scenario**: 1,000 tool calls per day, mixed safe and risky actions
- **Without governance**: All calls execute; ~30% are wasteful or high-cost
- **With governance**: Policy blocks 30% of wasteful calls; route 20% to cheaper models
- **Estimated savings**: 25–35% cost reduction

---

## Benchmark Setup

### Assumptions

| Parameter | Value |
|-----------|-------|
| Daily tool calls | 1,000 |
| Cost per allowed call (premium model) | $0.002 |
| Cost per allowed call (cheaper model) | $0.0005 |
| Wasteful/risky call share | 30% |
| Policy denial rate (on risky subset) | 90% |
| Routing to cheaper model (on safe subset) | 20% |

### Call Distribution

- **Safe calls (70%)**: Routine operations (echo, calc, internal API GET)
- **Risky calls (30%)**: External HTTP, data exfil attempts, high-cost mutations

---

## Scenarios

### Scenario A: No Governance

| Metric | Value |
|--------|-------|
| Total calls | 1,000 |
| Allowed | 1,000 |
| Denied | 0 |
| Model used | Premium for all |
| Daily cost | 1,000 × $0.002 = **$2.00** |

### Scenario B: Policy Governance (Block Risky)

| Metric | Value |
|--------|-------|
| Total calls | 1,000 |
| Risky (blocked) | 270 (90% of 300) |
| Allowed | 730 |
| Model used | Premium for allowed |
| Daily cost | 730 × $0.002 = **$1.46** |
| **Savings** | **27%** |

### Scenario C: Policy + Model Routing

| Metric | Value |
|--------|-------|
| Total calls | 1,000 |
| Denied | 270 |
| Allowed (premium) | 584 (80% of 730) |
| Allowed (cheaper) | 146 (20% of 730) |
| Daily cost | (584 × $0.002) + (146 × $0.0005) = $1.17 + $0.07 = **$1.24** |
| **Savings** | **38%** |

---

## Methodology

1. **Instrument**: Use `pkg/cost/tracker.go` and `pkg/policy/audit.go` to record costs and denials.
2. **Benchmark run**: Execute 1,000 representative tool calls; record total cost, denials, and model usage.
3. **Compare**: Run the same workload with governance off vs on; compute % cost reduction.
4. **Extrapolate**: Scale to monthly/yearly for pitch and planning.

### Formula

```
savings = (cost_no_governance - cost_with_governance) / cost_no_governance
```

---

## How to Run a Benchmark

1. Deploy Unagnt with cost tracking and policy enforcement.
2. Use a workload generator (or `unagnt run` in a loop) to trigger 1,000 tool calls.
3. Ensure a mix of safe and risky actions (e.g., `http_request` to internal vs external URLs).
4. Query analytics: `GET /v1/analytics/costs`, `GET /v1/analytics/denials/stats`.
5. Compare runs with policy enabled vs disabled.

---

## References

- Cost tracking: `pkg/cost/tracker.go`
- Policy engine: `pkg/policy/engine.go`
- Analytics API: `pkg/api/analytics.go`
