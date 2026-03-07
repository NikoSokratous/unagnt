# Unagnt TypeScript/Node SDK

TypeScript and Node.js client for the [Unagnt Agent Runtime](https://github.com/NikoSokratous/unagnt) API.

**Requires Node.js 18+** (uses native `fetch`).

## Installation

```bash
npm install @unagnt/client
```

Or from source:

```bash
cd sdk/typescript
npm install
npm run build
```

## Quick Start

```typescript
import { AgentRuntime } from '@unagnt/client';

const client = new AgentRuntime({
  baseUrl: 'http://localhost:8080',
  apiKey: process.env.AGENT_RUNTIME_API_KEY,
});

// Create a run
const runId = await client.createRun('demo-agent', 'List files in current directory');
console.log('Run ID:', runId);

// Wait for completion
const run = await client.waitForRun(runId);
console.log('State:', run.state, 'Steps:', run.step_count);

// Stream events (optional)
await client.streamRun(runId, {
  onEvent: (chunk) => console.log('Event:', chunk.type, chunk.data),
});
```

## API Reference

### AgentRuntime

| Method | Description |
|--------|-------------|
| `createRun(agentName, goal, opts?)` | Create a new run. Returns `run_id`. |
| `getRun(runId)` | Get run details. |
| `listRuns(limit?)` | List recent run IDs. |
| `cancelRun(runId)` | Cancel an ongoing run. |
| `getRunEvents(runId)` | Get persisted event history. |
| `waitForRun(runId, pollIntervalMs?, timeoutMs?)` | Poll until run completes. |
| `streamRun(runId, { onEvent, onError })` | Stream events via SSE. |
| `healthCheck()` | Check if server is healthy. |

### Options

```typescript
new AgentRuntime({
  baseUrl: 'http://localhost:8080',  // default
  apiKey: 'your-api-key',             // optional
  timeout: 30000,                     // request timeout (ms)
});
```

### Types

- `Run` - Run details
- `CreateRunRequest` - Optional fields: `max_retries`, `retry_backoff_ms`, `timeout_ms`
- `RunEvent`, `StreamChunk` - Event payloads

### Errors

- `APIError` - HTTP errors (statusCode, message)
- `NotFoundError` - 404
- `UnauthorizedError` - 401

## Development

```bash
npm install
npm run build
npm test
```
