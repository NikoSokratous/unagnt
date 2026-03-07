# Unagnt VS Code Extension

Workflow authoring, local run, and time-travel debugging for the [Unagnt Agent Runtime](https://github.com/NikoSokratous/unagnt).

## Features

- **Explorer**: Tree view of runs and replay snapshots
- **Workflow validation**: Validate YAML workflows via `Unagnt: Validate Workflow`
- **Local run**: Start agent runs via `Unagnt: Run Agent`
- **Snippets**: Workflow YAML snippets (prefix: `unagnt-workflow`)
- **Launch config**: "Unagnt Replay Debug" for time-travel debugging

## Configuration

| Setting | Description | Default |
|---------|-------------|---------|
| `unagnt.serverUrl` | Agent Runtime server URL | `http://localhost:8080` |
| `unagnt.apiKey`   | API key for authenticated requests | (empty) |

## Usage

1. Start the Unagnt server: `unagntd --addr :8080`
2. Open the Unagnt view in the Activity Bar
3. Use commands from the Command Palette (`Ctrl+Shift+P`): "Unagnt: Validate Workflow", "Unagnt: Run Agent"

## Development

```bash
npm install
npm run compile
```

Press F5 in VS Code to launch the Extension Development Host.
