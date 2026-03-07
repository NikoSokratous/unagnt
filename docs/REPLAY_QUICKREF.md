# Deterministic Replay - Quick Reference

## Recording

### Enable in Agent Config
```yaml
recording:
  enabled: true
  capture_model: true
  capture_tools: true
  compress_data: true
```

### Programmatic
```go
recorder := replay.NewRecorder(config)
recorder.StartRecording(runID, agentName, goal, config)
recorder.RecordModelCall(...)
recorder.RecordToolCall(...)
snapshot, _ := recorder.StopRecording("completed")
```

## Replay Modes

| Mode | Speed | Use Case |
|------|-------|----------|
| `exact` | 20-50x | Deterministic replay |
| `live` | 1x | Test API changes |
| `mixed` | 10-20x | Test tools only |
| `debug` | Interactive | Step-through |
| `validation` | 100-200x | Integrity check |

## CLI Commands

```bash
# List snapshots
unagnt replay list
unagnt replay list --run run-123

# Replay
unagnt replay run snap-123 --mode exact
unagnt replay run snap-123 --mode live
unagnt replay run snap-123 --start 5 --stop 10

# Debug with breakpoints
unagnt replay debug snap-123

# Validate
unagnt replay validate snap-123

# Compare
unagnt replay diff snap-old snap-new
```

## Replay Options

```bash
--mode <exact|live|mixed|debug|validation>
--start <sequence>             # Start from sequence N
--stop <sequence>              # Stop at sequence N
--breakpoint <sequence>        # Add breakpoint (repeatable)
--verify-side-effects          # Execute side effects (dangerous!)
```

## Time-Travel Debugging

```bash
unagnt replay debug <snapshot-id>          # Interactive step-through
unagnt replay debug <snapshot-id> --file snapshot.json   # Load from file
```

Commands: `(s)tep` forward, `(b)ack`, `(g)oto <seq>`, `(p)rint`, `(q)uit`.

## Example Workflows

### Regression Test
```bash
# 1. Record baseline
unagnt run my-agent --goal "task" --record

# 2. Make code changes
# ...

# 3. Test for regressions
unagnt replay run snap-baseline --mode live
```

### Debug Failed Run
```bash
# 1. Find failed run
unagnt logs list --state failed

# 2. Get snapshot
unagnt replay list --run run-failed-xyz

# 3. Debug
unagnt replay debug snap-failed
```

### Performance Comparison
```bash
unagnt replay run snap-old --mode exact
unagnt run my-agent --goal "same task" --record
unagnt replay diff snap-old snap-new
```

## Data Structures

### RunSnapshot
- Complete execution capture
- Model + tool calls
- Side effects
- Environment state
- Checksums for integrity

### ModelCall
- Prompt + response
- Model parameters
- Token usage
- Optional seed for determinism

### ToolExecution
- Input + output
- Timing
- Side effects
- Policy decisions

### ReplayResult
- Success status
- Matches + divergences
- Performance metrics
- Accuracy %

## Safety

⚠️ **Never use `--verify-side-effects` in production!**

Side effects include:
- File writes/deletes
- HTTP calls
- Database modifications
- External API calls

These should only be re-executed in test environments.

## Performance

| Snapshot Size | Compressed | Savings |
|---------------|------------|---------|
| 50KB | 15KB | 70% |
| 500KB | 120KB | 76% |
| 5MB | 1.2MB | 76% |

| Mode | Speedup |
|------|---------|
| Exact | 20-50x |
| Mixed | 10-20x |
| Live | 1x |
| Validation | 100-200x |

## Database Schema

```sql
-- Snapshots
run_snapshots (id, run_id, agent_name, model_calls, tool_calls, ...)

-- Freeze points (manual intervention)
freeze_points (id, run_id, snapshot_id, sequence, options, decision, ...)

-- Replay history
replay_results (id, snapshot_id, mode, success, divergences, metrics, ...)
```

## Integration

### With Policy Engine
- Policy decisions recorded in tool calls
- Replay validates policy consistency
- Audit trail linkage

### With API Server
- Auto-record via `--record` flag
- SSE for replay progress
- Future: REST API for snapshots

### With Testing
- Regression test framework
- CI/CD integration ready
- Performance benchmarking

## Next Steps

Ready for **Phase 3**: Risk Scoring & Audit Framework
