# Migration to v3.1

This document describes changes in v3.1 and how to migrate.

## Durable Queue Backend

v3.1 adds pluggable queue backends. The default remains in-memory; you can switch to Redis for restart-resilient queuing.

### Switching from Memory to Redis

1. **Ensure Redis is available** and note the connection URL (e.g. `redis://localhost:6379`).

2. **Set environment variables** when starting unagntd:
   ```bash
   export QUEUE_BACKEND=redis
   export QUEUE_REDIS_URL=redis://localhost:6379
   unagntd --addr :8080 --store agent.db
   ```

3. **Behavior change**:
   - With memory: queue is lost on restart; in-flight runs are lost.
   - With Redis: queued runs survive restarts; workers will pick them up after restart.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `QUEUE_BACKEND` | `memory` or `redis` | `memory` |
| `QUEUE_REDIS_URL` | Redis URL (required when backend is `redis`) | - |
| `QUEUE_SIZE` | Queue capacity for memory backend | 256 |

## Dead-Letter Retention and Archival

v3.1 adds configurable dead-letter retention and optional archival.

### Enabling Retention

1. **Set retention window** (hours to keep before prune):
   ```bash
   export DEAD_LETTER_RETENTION_HOURS=168
   unagntd --addr :8080
   ```

2. **Optional archival** (archive to directory before prune):
   ```bash
   export DEAD_LETTER_RETENTION_HOURS=168
   export DEAD_LETTER_ARCHIVE_DIR=/var/data/dead-letter-archive
   unagntd --addr :8080
   ```

See [docs/runbooks/dead-letter-retention.md](runbooks/dead-letter-retention.md) for details.

## Metrics

New metrics in v3.1:

| Metric | Description |
|--------|-------------|
| `agentruntime_dead_letters_pruned_total` | Dead letters pruned by retention |
| `agentruntime_dead_letters_archived_total` | Dead letters archived before prune |

Existing queue metrics (`agentruntime_run_queue_depth`, `agentruntime_run_queue_rejected_total`) are unchanged and work with both memory and Redis backends.
