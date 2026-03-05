# Integration Tests

This directory contains end-to-end integration tests for the Unagnt.

## Running Tests

```bash
# Run all integration tests
go test ./tests/integration -v

# Run specific test
go test ./tests/integration -v -run TestStreamingIntegration

# Run with timeout
go test ./tests/integration -v -timeout 5m
```

## Test Coverage

- **End-to-end run**: Complete agent execution lifecycle
- **Streaming**: SSE event streaming
- **Webhooks**: External webhook triggers
- **Workflows**: Multi-agent orchestration
- **Rate limiting**: API rate limiting
- **Plugin discovery**: Plugin loading and execution
- **Authentication**: API key validation
- **Memory**: Persistence and retrieval
- **Observability**: Metrics and health checks

## Requirements

Some tests require external services:
- **Vector DB tests**: Qdrant or Weaviate running locally
- **Redis tests**: Redis instance for distributed rate limiting
- **Webhook tests**: Mock webhook sender

Tests that require external services are skipped by default.

## Configuration

Set environment variables for integration tests:

```bash
export AGENTD_URL=http://localhost:8080
export AGENTD_API_KEY=test-key
export QDRANT_URL=http://localhost:6333
export REDIS_URL=redis://localhost:6379
```

## CI/CD

Integration tests run in GitHub Actions with Docker Compose:

```yaml
services:
  qdrant:
    image: qdrant/qdrant
  redis:
    image: redis:alpine
  unagntd:
    build: .
    depends_on:
      - qdrant
      - redis
```
