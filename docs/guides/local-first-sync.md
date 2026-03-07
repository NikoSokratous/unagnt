# Local-First Development with Cloud Sync

Work offline with workflows and runs, then sync when online.

## Overview

- **Local store**: SQLite (`agent.db`) holds runs locally
- **Push**: Upload local runs to the server
- **Pull**: Download server runs to local
- **Conflict resolution**: Last-write-wins (per run_id)

## Configuration

| Env | Description | Default |
|-----|-------------|---------|
| `UNAGNT_SERVER_URL` | Server URL | `http://localhost:8080` |
| `AGENT_RUNTIME_API_KEY` | API key | (empty) |

## CLI

```bash
# Push local runs to server
unagnt sync push [--url http://localhost:8080] [--db agent.db]

# Pull server runs to local
unagnt sync pull [--url http://localhost:8080] [--db agent.db]

# Show sync status
unagnt sync status [--db agent.db]
```

## API

- `POST /v1/sync/push` – Receive bundle (body: `{ runs: [...], timestamp }`)
- `POST /v1/sync/pull?since=<RFC3339>` – Return server bundle (optional `since` for delta)

## Workflow

1. Work locally: create runs with `unagntd` or CLI
2. Push: `unagnt sync push` to upload to remote
3. On another machine: `unagnt sync pull` to get remote runs
4. Conflicts: newer `updated_at` wins per run

## See Also

- [API Integration](api-integration.md) – REST API
- [Deployment](../DEPLOYMENT.md) – Server setup
