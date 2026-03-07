# Dead-Letter Retention and Archival Runbook

This runbook describes how to configure and operate dead-letter retention, pruning, and archival.

## Configuration

Dead-letter retention is enabled via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `DEAD_LETTER_RETENTION_HOURS` | Hours to keep dead letters before pruning. Must be > 0 to enable. | Disabled (no pruning) |
| `DEAD_LETTER_ARCHIVE_DIR` | Directory to archive dead letters before prune. If set, enables archival. | None |

**Example:**

```bash
export DEAD_LETTER_RETENTION_HOURS=168   # 7 days
export DEAD_LETTER_ARCHIVE_DIR=/var/data/dead-letter-archive
unagntd --addr :8080 --store agent.db
```

- If only `DEAD_LETTER_RETENTION_HOURS` is set: dead letters older than that window are pruned (deleted) on a hourly schedule.
- If both are set: dead letters are first archived to JSON files in the directory, then pruned.

## Pruning Schedule

The pruner runs hourly by default. Each run:

1. Computes `olderThan = now - retention_hours`
2. If archival enabled: fetches dead letters with `failed_at < olderThan`, writes each to `{dir}/dead_letter_{run_id}_{failed_at}.json`, then deletes all older rows
3. If archival disabled: deletes all rows with `failed_at < olderThan`

## Archived File Format

Each archived file is JSON with the dead-letter structure:

```json
{
  "run_id": "...",
  "agent_name": "...",
  "goal": "...",
  "source": "...",
  "error": "...",
  "payload": "...",
  "attempt": 1,
  "max_retries": 2,
  "failed_at": "2025-03-07T12:00:00Z"
}
```

Filename pattern: `dead_letter_{run_id}_{failed_at}.json`. The run_id is sanitized for filesystem safety.

## Querying Archived Files

To inspect archived dead letters:

```bash
ls -la /var/data/dead-letter-archive/
cat /var/data/dead-letter-archive/dead_letter_abc123_20250307T120000Z0700.json | jq .
```

## Replay from Archive

To replay an archived dead letter:

1. Inspect the JSON for `run_id`, `agent_name`, `goal`, `payload`
2. Insert the dead letter back into the store (or use the replay API with equivalent parameters)
3. Or create a new run via `POST /v1/runs` with the same agent, goal, and payload

The replay API (`POST /v1/runs/dead-letters/{run_id}/replay`) only works for dead letters still in the store; archived items must be recreated manually or via script if needed.

## Metrics

| Metric | Description |
|--------|-------------|
| `agentruntime_dead_letters_pruned_total` | Total dead letters deleted by retention pruner |
| `agentruntime_dead_letters_archived_total` | Total dead letters written to archive before prune |
| `agentruntime_run_dead_letters_total` | Total dead letters created (existing) |

## Alerting Suggestions

- **Dead-letter growth**: Alert when `rate(agentruntime_run_dead_letters_total[5m])` spikes above baseline
- **Pruning activity**: Monitor `agentruntime_dead_letters_pruned_total` to ensure retention is running
- **Archive failures**: No direct metric; check logs if archival is failing silently
