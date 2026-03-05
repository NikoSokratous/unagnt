# Streaming Guide

## Overview

The Unagnt provides real-time streaming of agent execution events using Server-Sent Events (SSE). This allows clients to monitor agent runs as they happen, enabling live dashboards and interactive debugging tools.

## Architecture

### Event Hub

The `EventHub` is a pub/sub system that manages event distribution:

```go
eventHub := observe.NewEventHub(100) // buffer size per subscriber
eventChan := eventHub.Subscribe(runID)
defer eventHub.Unsubscribe(runID, eventChan)
```

### Event Types

All events follow the `observe.Event` structure:

- `EventInit`: Agent execution started
- `EventToolCall`: Tool being invoked
- `EventToolResult`: Tool execution completed
- `EventApprovalRequired`: Human approval needed
- `EventCompleted`: Agent finished successfully
- `EventError`: Agent encountered an error
- `EventInterrupted`: Agent was cancelled

## Server-Side Implementation

### SSE Endpoint

The `/v1/runs/{id}/stream` endpoint provides real-time event streaming:

```
GET /v1/runs/{run-id}/stream
Accept: text/event-stream
```

**Response format:**

```
data: {"timestamp":"2026-02-26T10:00:00Z","type":"init","agent":"my-agent","data":{...}}

data: {"timestamp":"2026-02-26T10:00:01Z","type":"tool_call","agent":"my-agent","data":{...}}

: heartbeat
```

Heartbeats are sent every 30 seconds to keep the connection alive.

## Client-Side Usage

### Go SDK

```go
import "github.com/Unagnt/Unagnt/sdk/go/client"

client := client.New("http://localhost:8080", "your-api-key")

// Stream events
eventChan, errChan := client.StreamEvents(ctx, runID)

for {
    select {
    case event := <-eventChan:
        fmt.Printf("Event: %s - %v\n", event.Type, event.Data)
        
        // Check for completion
        if event.Type == "completed" || event.Type == "error" {
            return
        }
    
    case err := <-errChan:
        fmt.Printf("Stream error: %v\n", err)
        return
    
    case <-ctx.Done():
        return
    }
}
```

### Python SDK

```python
from Unagnt import UnagntClient, stream_events

client = UnagntClient(
    base_url="http://localhost:8080",
    api_key="your-api-key"
)

# Async streaming
async for event in stream_events(client, run_id):
    print(f"Event: {event.type} - {event.data}")
    
    if event.type in ["completed", "error"]:
        break
```

### JavaScript/Browser

```javascript
const eventSource = new EventSource(`/v1/runs/${runId}/stream`);

eventSource.onmessage = (event) => {
  if (event.data === ': heartbeat') return;
  
  const eventData = JSON.parse(event.data);
  console.log('Event:', eventData.type, eventData.data);
  
  // Close on completion
  if (['completed', 'error', 'interrupted'].includes(eventData.type)) {
    eventSource.close();
  }
};

eventSource.onerror = (error) => {
  console.error('Stream error:', error);
  eventSource.close();
};
```

## Production Considerations

### Connection Management

- **Timeout**: Connections timeout after 5 minutes of inactivity
- **Heartbeat**: Server sends heartbeat every 30 seconds
- **Buffer**: Event channels have a 100-event buffer per subscriber
- **Backpressure**: If a subscriber's buffer is full, events are dropped (non-blocking publish)

### Scalability

For high-scale deployments:

1. **Use Redis Pub/Sub**: Distribute events across multiple `unagntd` instances
2. **Load Balancing**: Use sticky sessions to route clients to the same backend
3. **Event Replay**: Store events in a time-series database for late joiners

### Error Handling

Always handle disconnections gracefully:

```go
for {
    select {
    case event, ok := <-eventChan:
        if !ok {
            // Channel closed, reconnect or exit
            return
        }
        // Process event
    }
}
```

## Examples

### Live Dashboard

```typescript
// React component
const [events, setEvents] = useState<Event[]>([]);

useEffect(() => {
  const source = new EventSource(`/v1/runs/${runId}/stream`);
  
  source.onmessage = (e) => {
    const event = JSON.parse(e.data);
    setEvents(prev => [...prev, event]);
  };
  
  return () => source.close();
}, [runId]);
```

### CLI Progress Monitor

```go
func monitorRun(ctx context.Context, client *client.Client, runID string) {
    eventChan, errChan := client.StreamEvents(ctx, runID)
    
    for {
        select {
        case event := <-eventChan:
            switch event.Type {
            case "tool_call":
                fmt.Printf("→ Calling tool: %s\n", event.Data["tool"])
            case "tool_result":
                fmt.Printf("✓ Tool completed: %s\n", event.Data["tool"])
            case "completed":
                fmt.Println("✓ Agent completed successfully")
                return
            case "error":
                fmt.Printf("✗ Error: %s\n", event.Data["error"])
                return
            }
        case err := <-errChan:
            fmt.Printf("Stream error: %v\n", err)
            return
        }
    }
}
```

## Best Practices

1. **Always close streams**: Use `defer` or cleanup functions
2. **Handle reconnections**: Implement exponential backoff for retries
3. **Buffer management**: Don't block on event processing
4. **Timeout handling**: Set reasonable timeouts on contexts
5. **Error recovery**: Gracefully handle stream interruptions

## Debugging

Enable verbose logging to see event flow:

```bash
export LOG_LEVEL=debug
unagntd --log-level=debug
```

Monitor active connections:

```bash
curl http://localhost:8080/metrics | grep stream_connections
```
