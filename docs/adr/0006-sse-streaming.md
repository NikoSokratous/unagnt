# ADR 0006: Server-Sent Events for Real-Time Streaming

**Status**: Accepted  
**Date**: 2026-02-26  
**Decision Makers**: Unagnt Core Team

## Context

The runtime needs to provide real-time visibility into agent execution. Users want to:
- Monitor agent progress live
- Build interactive dashboards
- Debug execution issues as they occur
- Create CLI tools with live progress indicators

## Decision

We will implement Server-Sent Events (SSE) for real-time event streaming, backed by an in-memory pub/sub `EventHub`.

### Architecture

1. **EventHub**: Central pub/sub system for runtime events
   - In-memory, buffered channels per subscriber
   - Non-blocking publish to prevent slowdowns
   - Automatic cleanup on disconnect

2. **SSE Endpoint**: `GET /v1/runs/{id}/stream`
   - HTTP streaming with `text/event-stream`
   - 30-second heartbeat for connection keep-alive
   - Graceful disconnect on completion/cancellation

3. **Client Libraries**: Streaming support in Go and Python SDKs
   - Go: Channel-based async streaming
   - Python: Async generator pattern

## Alternatives Considered

### WebSockets

**Pros:**
- Bidirectional communication
- Lower latency for two-way messaging
- Better for interactive control

**Cons:**
- More complex to implement and debug
- Requires connection upgrade
- Not needed for unidirectional streaming
- Browser compatibility issues with auth headers

**Decision**: Rejected. SSE is simpler and sufficient for our use case.

### Polling

**Pros:**
- Simplest to implement
- Works everywhere
- No persistent connections

**Cons:**
- High latency (1-5s between updates)
- Increased server load
- Wasteful for long-running agents
- Poor user experience

**Decision**: Rejected. Not suitable for real-time monitoring.

### gRPC Streaming

**Pros:**
- Built-in bidirectional streaming
- Efficient binary protocol
- Type safety with protobufs

**Cons:**
- Not browser-friendly (requires gRPC-Web)
- More complex client setup
- Overkill for unidirectional events

**Decision**: Rejected. SSE provides better browser compatibility.

## Consequences

### Positive

- Simple implementation with standard HTTP
- Works in browsers without additional libraries
- Low latency (< 100ms for events)
- Automatic reconnection in browsers
- Easy to test with `curl`

### Negative

- Unidirectional only (server → client)
- Limited browser connection pool (6 per domain)
- Text-based (slightly less efficient than binary)
- No built-in reconnection in Go/Python clients

### Neutral

- Requires persistent HTTP connections
- Buffer size tuning needed for high-frequency events

## Implementation Notes

### Buffer Management

```go
// Non-blocking publish prevents slow subscribers from blocking the runtime
select {
case ch <- event:
default: // Drop event if buffer full
}
```

### Connection Lifecycle

1. Client connects to `/v1/runs/{id}/stream`
2. Server subscribes to EventHub for that run
3. Events are marshaled to JSON and sent
4. Heartbeats keep connection alive
5. On completion/error, stream closes
6. Server unsubscribes from EventHub

### Scaling Considerations

For distributed deployments:
- Use Redis Pub/Sub to share events across `unagntd` instances
- Implement sticky sessions for load balancer
- Consider message queue for event persistence

## Examples

### CLI Monitoring

```bash
curl -N http://localhost:8080/v1/runs/{id}/stream
```

### Browser Dashboard

```javascript
const source = new EventSource(`/v1/runs/${runId}/stream`);
source.onmessage = (e) => {
  const event = JSON.parse(e.data);
  updateDashboard(event);
};
```

## References

- [SSE Specification](https://html.spec.whatwg.org/multipage/server-sent-events.html)
- [EventSource API](https://developer.mozilla.org/en-US/docs/Web/API/EventSource)
- Implementation: `pkg/observe/eventhub.go`, `pkg/orchestrate/stream.go`
