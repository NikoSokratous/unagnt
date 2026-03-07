# Time-Travel Debugging Guide

Navigate backward and forward through recorded runs for deterministic debugging.

## Overview

Time-travel debugging lets you:

- **Step forward/back** through tool and model calls in a snapshot
- **Seek** to any sequence position
- **Inspect state** at each point
- Use **CLI** or **API** for integration with UIs

## CLI Usage

### Interactive Debug

```bash
# With demo snapshot (built-in)
unagnt replay debug snap-001

# Load from JSON file
unagnt replay debug my-snap --file ./snapshot.json
```

**Commands:**

| Command | Description |
|---------|-------------|
| `s`, `step` | Step forward one action |
| `b`, `back` | Step backward one action |
| `g <seq>`, `goto <seq>` | Seek to sequence number |
| `p`, `print` | Print current state |
| `q`, `quit` | Exit |

### Replay with Range

```bash
unagnt replay run snap-001 --mode exact --start 3 --stop 7
```

## API

| Endpoint | Description |
|----------|-------------|
| `GET /v1/replay/snapshots` | List snapshots (query: `run_id`, `limit`) |
| `GET /v1/replay/snapshots/{id}` | Get full snapshot |
| `POST /v1/replay/snapshots/{id}/seek` | Get state at sequence (body: `{"sequence": 3}`) |

### Example: Seek and Inspect

```bash
# Get state at sequence 5
curl -X POST http://localhost:8080/v1/replay/snapshots/snap-001/seek \
  -H "Content-Type: application/json" \
  -d '{"sequence": 5}'
```

Response includes `position`, `tool_calls`, `model_calls`, `current_action`, `can_step_forward`, `can_step_back`.

## Programmatic Usage

```go
import "github.com/NikoSokratous/unagnt/pkg/replay"

replayer := replay.NewReplayer(snapshot)
cursor := replayer.Cursor()

// Step through
cursor.StepForward()
st := cursor.GetStateAt(cursor.Position())
fmt.Println(st.CurrentAction.ToolName, string(st.CurrentAction.Input))

cursor.StepBack()
cursor.SeekToSequence(5)
```

## Snapshot Storage

Snapshots persist in `run_snapshots` (migration 006). Use `replay.NewSQLiteSnapshotStore(db)` for SQLite persistence.

## See Also

- [REPLAY_QUICKREF.md](../REPLAY_QUICKREF.md) – Replay modes and CLI reference
- [Deterministic Replay](deterministic-replay.md) – Recording and replay concepts
