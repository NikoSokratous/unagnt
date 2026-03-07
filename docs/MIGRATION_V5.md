# Migration to v5

v5.0 Developer Experience changes.

## TypeScript/Node SDK

New `sdk/typescript` package:

```bash
npm install @unagnt/client
# or from source: cd sdk/typescript && npm install && npm run build
```

See [sdk/typescript/README.md](../sdk/typescript/README.md).

## Tool Testing Harness

New `pkg/tool/testing` package:

- `ToolHarness`, `MockTool`, `MockExecutor`
- Assertion helpers: `AssertNoError`, `AssertOutputContains`, `AssertOutputEqual`

See [docs/guides/tool-testing.md](guides/tool-testing.md).

## Time-Travel Debugging

- `replay.ReplayCursor` for step forward/back
- `replay.Replayer.GetStateAt(seq)`, `ReplayRange`
- API: `GET /v1/replay/snapshots`, `GET /v1/replay/snapshots/{id}`, `POST /v1/replay/snapshots/{id}/seek`
- CLI: `unagnt replay debug <id>` with `s`, `b`, `g`, `p`, `q` commands

See [docs/guides/time-travel-debugging.md](guides/time-travel-debugging.md).

## VS Code Extension

Install from `extensions/vscode-unagnt`. Config: `unagnt.serverUrl`, `unagnt.apiKey`.

See [docs/guides/vscode-extension.md](guides/vscode-extension.md).

## Local-First Sync

- API: `POST /v1/sync/push`, `POST /v1/sync/pull`
- CLI: `unagnt sync push`, `unagnt sync pull`, `unagnt sync status`

See [docs/guides/local-first-sync.md](guides/local-first-sync.md).
