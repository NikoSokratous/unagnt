# Unagnt VS Code Extension

The Unagnt extension provides workflow authoring, local run, and time-travel debugging inside VS Code.

## Installation

Install from the VS Code marketplace or build from source:

```bash
cd extensions/vscode-unagnt
npm install
npm run compile
```

Then load the extension in VS Code (Run > Run Extension) or package for distribution.

## Configuration

| Setting | Description | Default |
|---------|-------------|---------|
| `unagnt.serverUrl` | Agent Runtime server URL | `http://localhost:8080` |
| `unagnt.apiKey`   | API key for authenticated requests | (empty) |

## Features

### Explorer

Open the Unagnt view in the Activity Bar to see:

- **Runs** – Recent agent runs with state
- **Snapshots** – Replay snapshots for time-travel debugging

### Commands

- **Unagnt: Validate Workflow** – Validate the active workflow YAML
- **Unagnt: Run Agent** – Start an agent run (prompts for agent name and goal)
- **Unagnt: Refresh** – Refresh the explorer

### Snippets

In YAML files, type `unagnt-workflow` for a workflow scaffold.

### Launch Config

Add to `.vscode/launch.json` for replay debugging:

```json
{
  "type": "unagnt-replay",
  "request": "launch",
  "name": "Unagnt Replay Debug",
  "snapshotId": "snap-001"
}
```

## Requirements

- Unagnt server running (`unagntd`)
- Node.js 18+
