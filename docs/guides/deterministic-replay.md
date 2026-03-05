# Deterministic Replay Guide

Deterministic replay enables you to capture, store, and replay agent executions for debugging, testing, and validation.

## Table of Contents

1. [Overview](#overview)
2. [Recording Executions](#recording-executions)
3. [Replay Modes](#replay-modes)
4. [CLI Commands](#cli-commands)
5. [API Reference](#api-reference)
6. [Use Cases](#use-cases)
7. [Best Practices](#best-practices)

## Overview

The replay system provides:

- **Execution Recording**: Capture complete agent runs including model calls, tool executions, and side effects
- **Multiple Replay Modes**: Exact, live, mixed, debug, and validation modes
- **Snapshot Storage**: Compress and store execution snapshots
- **Divergence Detection**: Identify differences between original and replayed executions
- **Debug Mode**: Step through executions with breakpoints

## Recording Executions

### Automatic Recording

Enable recording in your agent configuration:

```yaml
recording:
  enabled: true
  capture_model: true
  capture_tools: true
  capture_side_effects: true
  capture_environment: true
  compress_data: true
  max_snapshot_size: 104857600  # 100MB
```

### Programmatic Recording

```go
import "github.com/Unagnt/Unagnt/pkg/replay"

config := replay.RecordingConfig{
    Enabled:            true,
    CaptureModel:       true,
    CaptureTools:       true,
    CaptureSideEffects: true,
    CompressData:       true,
}

recorder := replay.NewRecorder(config)

// Start recording
recorder.StartRecording(runID, agentName, goal, agentConfig)

// Record model call
recorder.RecordModelCall(model, provider, prompt, response, tokens, metadata)

// Record tool execution
recorder.RecordToolCall(toolName, input, output, err, duration)

// Stop and save
snapshot, err := recorder.StopRecording("completed")
```

## Replay Modes

### 1. Exact Mode

Uses recorded responses for deterministic replay.

```bash
unagnt replay run snap-123 --mode exact
```

**Use for:**
- Debugging exact execution paths
- Regression testing
- Demo/presentations

**Characteristics:**
- 100% deterministic
- Fast (no API calls)
- Perfect reproducibility

### 2. Live Mode

Re-executes with live APIs.

```bash
unagnt replay run snap-123 --mode live
```

**Use for:**
- Testing API changes
- Validating current behavior
- Checking for drift

**Characteristics:**
- Non-deterministic
- Slower (real API calls)
- May diverge from original

### 3. Mixed Mode

Recorded models, live tools.

```bash
unagnt replay run snap-123 --mode mixed
```

**Use for:**
- Testing tool changes
- Keeping model responses consistent
- Hybrid debugging

### 4. Debug Mode

Step-through debugging with breakpoints.

```bash
unagnt replay run snap-123 --mode debug --breakpoint 5 --breakpoint 10
```

**Use for:**
- Interactive debugging
- Understanding execution flow
- Inspecting state at specific points

### 5. Validation Mode

Verifies snapshot integrity without execution.

```bash
unagnt replay validate snap-123
```

**Use for:**
- Checking data corruption
- Verifying checksums
- Quick integrity checks

## CLI Commands

### List Snapshots

```bash
# List all snapshots
unagnt replay list

# List snapshots for a specific run
unagnt replay list --run run-123

# Limit results
unagnt replay list --limit 10
```

Output:
```
SNAPSHOT ID  RUN ID    AGENT      MODEL CALLS  TOOL CALLS  STATE      SIZE
snap-001     run-123   agent-1    5            10          completed  100KB
snap-002     run-124   agent-2    8            15          completed  250KB
```

### Replay a Run

```bash
# Exact replay
unagnt replay run snap-123 --mode exact

# Partial replay (sequences 5-10)
unagnt replay run snap-123 --mode exact --start 5 --stop 10

# With breakpoints
unagnt replay run snap-123 --mode debug --breakpoint 5 --breakpoint 10

# Execute side effects
unagnt replay run snap-123 --mode exact --verify-side-effects
```

Output:
```
Loading snapshot snap-123...
Replaying in exact mode...

=== Replay Results ===
Snapshot: snap-123
Mode: exact
Duration: 123ms
Success: true

Actions Rerun: 10
Matches: 10
Divergences: 0

=== Metrics ===
Accuracy: 100.0%
Original Duration: 5.2s
Replay Duration: 123ms
Speedup Factor: 42.28x
```

### Debug Replay

```bash
unagnt replay debug snap-123
```

Interactive commands:
- `c` - Continue to next breakpoint
- `s` - Step to next action
- `i` - Inspect current state
- `q` - Quit debug session

### Validate Snapshot

```bash
unagnt replay validate snap-123
```

Output:
```
Validating snapshot snap-123...

✓ Checksums valid
✓ Sequence integrity verified
✓ Timestamps monotonic
✓ No data corruption detected

Snapshot is valid!
```

### Compare Snapshots

```bash
unagnt replay diff snap-123 snap-124
```

Output:
```
Comparing snap-123 and snap-124...

Differences:
  Model calls: 5 vs 6 (+1)
  Tool calls: 10 vs 10 (same)
  Duration: 5.2s vs 5.5s (+0.3s)
  Final state: completed vs completed (same)
```

## API Reference

### RunSnapshot

```go
type RunSnapshot struct {
    ID            string
    RunID         string
    Version       string
    CreatedAt     time.Time
    AgentName     string
    Goal          string
    AgentConfig   map[string]interface{}
    ModelCalls    []ModelCall
    ToolCalls     []ToolExecution
    Environment   map[string]string
    StartTime     time.Time
    EndTime       time.Time
    FinalState    string
    Checksums     map[string]string
    Compressed    bool
    Encrypted     bool
    SizeBytes     int64
}
```

### ModelCall

```go
type ModelCall struct {
    Sequence      int
    Timestamp     time.Time
    Model         string
    Provider      string
    Prompt        string
    Response      string
    TokensUsed    int
    Temperature   float64
    Seed          *int64  // For deterministic generation
    FinishReason  string
}
```

### ToolExecution

```go
type ToolExecution struct {
    Sequence       int
    Timestamp      time.Time
    ToolName       string
    Input          json.RawMessage
    Output         json.RawMessage
    Error          string
    Duration       time.Duration
    SideEffects    []SideEffect
    PolicyDecision string
}
```

### ReplayOptions

```go
type ReplayOptions struct {
    Mode               ReplayMode  // exact, live, mixed, debug, validation
    SnapshotID         string
    StartFromSequence  int
    StopAtSequence     int
    Breakpoints        []int
    ExecuteSideEffects bool
    VerifyOutputs      bool
}
```

### ReplayResult

```go
type ReplayResult struct {
    Success       bool
    SnapshotID    string
    Mode          ReplayMode
    StartedAt     time.Time
    CompletedAt   time.Time
    Duration      time.Duration
    ActionsRerun  int
    Matches       int
    Divergences   []Divergence
    Metrics       ReplayMetrics
}
```

## Use Cases

### 1. Regression Testing

Ensure agent behavior remains consistent across code changes:

```bash
# Record baseline execution
unagnt run my-agent --goal "process data" --record

# After code changes, replay
unagnt replay run snap-baseline --mode live --verify-outputs

# Check for divergences
```

### 2. Debugging Failures

Investigate failed runs by replaying with breakpoints:

```bash
# List recent runs
unagnt logs list --state failed

# Get snapshot for failed run
unagnt replay list --run run-failed-123

# Debug replay with breakpoints
unagnt replay debug snap-xyz --breakpoint 5 --breakpoint 12
```

### 3. Performance Testing

Compare execution times across versions:

```bash
# Replay with exact mode (fast)
unagnt replay run snap-123 --mode exact

# Check speedup factor in metrics
```

### 4. Compliance Audits

Validate historical executions:

```bash
# Validate snapshot integrity
unagnt replay validate snap-audit-2026-02

# Replay with verification
unagnt replay run snap-audit-2026-02 --mode validation
```

### 5. Training & Documentation

Replay successful runs for training:

```bash
# Debug mode for step-by-step walkthrough
unagnt replay debug snap-success-example
```

## Best Practices

### Recording

1. **Enable Compression**: Reduce storage costs
   ```yaml
   compress_data: true
   ```

2. **Set Size Limits**: Prevent unbounded growth
   ```yaml
   max_snapshot_size: 104857600  # 100MB
   ```

3. **Encrypt PII**: Protect sensitive data
   ```yaml
   encrypt_pii: true
   ```

4. **Selective Recording**: Don't record everything
   ```yaml
   capture_environment: false  # Skip env vars in production
   ```

### Replaying

1. **Start with Validation**: Always validate before replaying
   ```bash
   unagnt replay validate snap-123
   unagnt replay run snap-123
   ```

2. **Use Exact Mode for Speed**: For quick checks
   ```bash
   unagnt replay run snap-123 --mode exact
   ```

3. **Use Live Mode for Testing**: To catch API changes
   ```bash
   unagnt replay run snap-123 --mode live
   ```

4. **Don't Execute Side Effects in Production**: Unless explicitly intended
   ```bash
   # Safe - no side effects
   unagnt replay run snap-123 --mode exact
   
   # Dangerous - executes writes, API calls, etc.
   unagnt replay run snap-123 --verify-side-effects
   ```

### Storage

1. **Regular Cleanup**: Archive old snapshots
   ```bash
   # Archive snapshots older than 30 days
   unagnt replay list --older-than 30d | xargs unagnt replay archive
   ```

2. **Retention Policy**: Set automatic cleanup rules
   ```yaml
   retention:
     keep_successful: 30d
     keep_failed: 90d
     keep_archived: 1y
   ```

### Debugging

1. **Use Breakpoints**: Target specific sequences
   ```bash
   unagnt replay debug snap-123 --breakpoint 10 --breakpoint 15
   ```

2. **Partial Replay**: Test specific sections
   ```bash
   unagnt replay run snap-123 --start 5 --stop 10
   ```

3. **Compare Snapshots**: Find differences
   ```bash
   unagnt replay diff snap-before snap-after
   ```

## Divergence Analysis

### Impact Levels

- **Minor**: Negligible (timing, whitespace)
- **Major**: Significant (output changed)
- **Critical**: Breaking (error, crash)

### Common Divergence Types

| Type | Cause | Resolution |
|------|-------|------------|
| `output` | API response changed | Update tests or investigate API |
| `decision` | Policy changed | Use policy versioning |
| `error` | External failure | Retry or mock external service |
| `timing` | Performance change | Acceptable if within bounds |
| `sequence` | Data corruption | Re-record snapshot |

### Example Divergence Report

```
Divergence Details:
1. Seq 5 (output): API response format changed [major]
   Expected: {"status":"ok","data":[1,2,3]}
   Actual:   {"success":true,"items":[1,2,3]}

2. Seq 12 (timing): Execution slower [minor]
   Expected: 100ms
   Actual:   150ms
```

## Snapshot Format

Snapshots are stored as JSON with the following structure:

```json
{
  "id": "snap-123",
  "run_id": "run-456",
  "version": "1.0.0",
  "created_at": "2026-02-26T10:00:00Z",
  "agent_name": "my-agent",
  "goal": "process data",
  "model_calls": [...],
  "tool_calls": [...],
  "checksums": {
    "model_calls": "abc123...",
    "tool_calls": "def456..."
  },
  "compressed": true,
  "size_bytes": 102400
}
```

## Troubleshooting

### Snapshot Not Found

```bash
# List available snapshots
unagnt replay list

# Check run ID mapping
unagnt logs list | grep run-123
```

### Replay Divergence

```bash
# Validate snapshot first
unagnt replay validate snap-123

# Compare with live execution
unagnt replay run snap-123 --mode live
```

### Checksum Mismatch

```bash
# Re-validate
unagnt replay validate snap-123

# If corrupted, re-record
unagnt run my-agent --record
```

### Large Snapshots

```bash
# Check size
unagnt replay list

# Enable compression for future runs
# Add to agent config:
recording:
  compress_data: true
```

## Advanced Features

### Freeze Points

Execution freeze points allow manual intervention:

```go
freezePoint := &replay.FreezePoint{
    Sequence: 10,
    Reason:   "High-risk operation detected",
    Options: []replay.FreezeOption{
        {ID: "continue", Label: "Continue", RiskImpact: "high"},
        {ID: "skip", Label: "Skip", RiskImpact: "low"},
        {ID: "abort", Label: "Abort", RiskImpact: "none"},
    },
}
```

### Side Effect Rollback

Revert state changes:

```go
sideEffect := replay.SideEffect{
    Type:       replay.SideEffectFileWrite,
    Target:     "/path/to/file",
    Reversible: true,
    RevertData: json.RawMessage(`{"original_content":"..."}`),
}
```

### Model Seeding

For truly deterministic model responses:

```go
metadata := map[string]interface{}{
    "seed":        int64(42),
    "temperature": 0.0,
}
recorder.RecordModelCall(model, provider, prompt, response, tokens, metadata)
```

## Integration with Policy Engine

Replay integrates with the policy engine for compliance:

```bash
# Replay with policy validation
unagnt replay run snap-123 --mode validation --check-policy
```

Policy decisions are recorded in each tool execution:

```json
{
  "sequence": 5,
  "tool_name": "delete_file",
  "policy_decision": "denied",
  "side_effects": []
}
```

## Performance

### Snapshot Sizes

| Type | Uncompressed | Compressed | Ratio |
|------|--------------|------------|-------|
| Small (5 actions) | 50KB | 15KB | 3.3x |
| Medium (50 actions) | 500KB | 120KB | 4.2x |
| Large (500 actions) | 5MB | 1.2MB | 4.2x |

### Replay Speed

| Mode | Speedup vs Original |
|------|---------------------|
| Exact | 20-50x |
| Mixed | 10-20x |
| Live | 1x (same speed) |
| Debug | N/A (interactive) |
| Validation | 100-200x |

## Database Schema

Snapshots are stored in SQLite:

```sql
CREATE TABLE run_snapshots (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL,
    version TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    agent_name TEXT NOT NULL,
    goal TEXT,
    model_calls JSON NOT NULL,
    tool_calls JSON NOT NULL,
    final_state TEXT NOT NULL,
    compressed BOOLEAN DEFAULT false,
    size_bytes INTEGER
);
```

See `migrations/006_deterministic_replay.sql` for full schema.

## Examples

### Example 1: Record and Replay

```bash
# Run agent with recording
unagnt run my-agent --goal "analyze logs" --record

# List snapshots
unagnt replay list

# Replay exactly
unagnt replay run snap-abc123 --mode exact
```

### Example 2: Debug Failed Run

```bash
# Find failed run
unagnt logs list --state failed

# Get snapshot
unagnt replay list --run run-failed-xyz

# Debug with breakpoints
unagnt replay debug snap-failed --breakpoint 8
```

### Example 3: Regression Testing

```bash
# Record baseline
unagnt run my-agent --goal "process data" --record
# Output: Snapshot saved as snap-baseline

# Make code changes
# ...

# Test for regressions
unagnt replay run snap-baseline --mode live

# Check divergences
# If divergences found, investigate
```

### Example 4: Performance Comparison

```bash
# Replay old snapshot
unagnt replay run snap-old --mode exact

# Record new run
unagnt run my-agent --goal "same task" --record

# Compare
unagnt replay diff snap-old snap-new
```

## Security Considerations

1. **PII Encryption**: Enable for sensitive data
   ```yaml
   encrypt_pii: true
   ```

2. **Storage Access**: Restrict snapshot access
   ```bash
   chmod 600 snapshots/*.json
   ```

3. **Side Effects**: Never execute in production
   ```bash
   # Don't do this in production!
   unagnt replay run snap-123 --verify-side-effects
   ```

4. **Secrets**: Snapshots may contain API keys
   - Store securely
   - Rotate keys regularly
   - Use vault integration

## Next Steps

- [Policy Engine Guide](policy-versioning.md)
- [Policy Testing](policy-testing.md)
- [API Documentation](api-reference.md)
