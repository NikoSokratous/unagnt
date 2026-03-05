# ADR 0005: Deterministic Replay

## Status

Accepted

## Context

Debugging agent failures is critical. When an agent makes a wrong decision or crashes, developers need to:
- Reproduce the exact execution
- Compare different model versions
- Test policy changes against past runs
- Audit decisions for compliance

Traditional approaches (re-running with same inputs) are non-deterministic due to:
- LLM API variance
- External API changes
- Time-dependent operations
- Random number generation

## Decision

We implement **deterministic replay** by recording all non-deterministic inputs:

### What We Record

1. **StepRecord** for each execution step:
   - State before step
   - Tool call (name, version, input)
   - Tool result (output, error, duration)
   - LLM reasoning
   - Timestamps
   - Metadata

2. **Event Log** for observability:
   - Init, planning, tool calls, completions
   - Model metadata (provider, name, version)
   - Token usage

### Replay Implementation

**ReplayPlanner** (`pkg/runtime/replay.go`):
- Returns pre-recorded actions instead of calling LLM
- Guarantees identical tool selections

**ReplayExecutor** (`pkg/runtime/replay.go`):
- Returns pre-recorded results instead of executing tools
- Maintains exact timing and outputs

### Storage

- SQLite `history` table: `internal/store/sqlite.go`
- PostgreSQL support: `internal/store/postgres.go`

## Consequences

### Positive

- **Perfect Reproduction**: Replay matches original run exactly
- **Regression Testing**: Compare runs across model versions
- **Cost Savings**: Replay doesn't call LLM APIs
- **Fast Debugging**: Instant replay without waiting for tools
- **Audit Compliance**: Immutable record of all decisions

### Negative

- Storage overhead (every run is recorded)
- Cannot replay with different tools (would diverge)
- Sensitive data in recordings (PII concerns)

## Usage

```bash
# Record a run (automatic)
unagnt run --config agent.yaml --goal "..." --store agent.db

# Replay the run
unagnt replay --run-id abc123 --store agent.db

# Compare two runs
unagnt diff run-1 run-2 --store agent.db
```

## Security Considerations

### Sensitive Data

Recordings may contain:
- API keys in tool inputs
- User PII
- Proprietary information

**Mitigation**:
- Store access control via filesystem permissions
- GDPR delete: `unagnt memory delete --agent-id X`
- Future: Selective redaction of sensitive fields

### Tampering

Recordings are append-only but not cryptographically signed.

**Future Enhancement**: Add HMAC signatures to prevent tampering.

## Alternatives Considered

### 1. Snapshot-Based Replay
- Save full agent state at each step
- Cons: Huge storage, harder to diff

### 2. Log-Based Replay (What We Chose)
- Record inputs/outputs only
- Pros: Compact, easy to implement

### 3. No Replay
- Cons: Debugging nightmare, no regression testing

### 4. Time-Travel Debugging
- Record all intermediate states
- Cons: Massive overhead, complex implementation

## Future Enhancements

- Partial replay (from step N)
- Replay with modified inputs
- Replay visualizer (web UI)
- Differential replay (show divergence points)

## Related

- ADR 0001: State Machine (enables clean recording points)
- ADR 0003: Memory (execution log is separate tier)
- Implementation: `pkg/runtime/replay.go`, `pkg/observe/tracer.go`
