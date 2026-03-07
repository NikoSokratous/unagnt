# Approval Queues

Human-in-the-loop approval flow for high-risk tool executions (v4).

## Overview

When a policy rule has `action: require_approval`, the runtime enqueues a request instead of blocking. Approvers can list pending requests and approve or deny via API or CLI.

## API

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/v1/approvals/pending` | GET | List pending requests |
| `/v1/approvals/{id}` | GET | Get request details |
| `/v1/approvals/{id}/approve` | POST | Approve |
| `/v1/approvals/{id}/deny` | POST | Deny |

## CLI

```bash
unagnt approvals list [--url http://localhost:8080]
unagnt approvals approve <id> [--url http://localhost:8080]
unagnt approvals deny <id> [--url http://localhost:8080]
```

## Queue Backends

- **In-memory**: Default for development; lost on restart
- **SQLite**: Persistent; requires migration 017 (`approval_requests` table)

## Troubleshooting

- **Empty pending list**: No runs have hit `require_approval` rules, or approvals were already processed
- **404 on approve/deny**: Request ID invalid or already resolved
- **Server not found**: Ensure unagntd is running with approvals API mounted
