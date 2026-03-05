# Unagnt Python SDK

Python client library for the Unagnt API.

## Installation

```bash
pip install unagnt
```

Or install from source:

```bash
cd sdk/python
pip install -e .
```

## Quick Start

### Synchronous Client

```python
from unagnt import Unagnt

# Create client
client = Unagnt(
    base_url="http://localhost:8080",
    api_key="your-api-key"
)

# Create a run
run_id = client.create_run(
    agent_name="demo-agent",
    goal="Calculate 15 + 27"
)

print(f"Created run: {run_id}")

# Wait for completion
run = client.wait_for_run(run_id, timeout=60)

print(f"State: {run.state}")
print(f"Steps: {run.step_count}")
```

### Async Client

```python
import asyncio
from unagnt import AsyncUnagnt

async def main():
    async with AsyncUnagnt(
        base_url="http://localhost:8080",
        api_key="your-api-key"
    ) as client:
        # Create run
        run_id = await client.create_run("demo-agent", "test goal")
        
        # Wait for completion
        run = await client.wait_for_run(run_id)
        print(f"Completed: {run.state}")

asyncio.run(main())
```

## API Reference

### Unagnt

Synchronous client for the API.

**Methods:**
- `create_run(agent_name, goal)` - Create a new run
- `get_run(run_id)` - Get run details
- `list_runs(limit=100)` - List recent runs
- `cancel_run(run_id)` - Cancel a run
- `wait_for_run(run_id, poll_interval=2.0, timeout=None)` - Wait for completion
- `health_check()` - Check service health

### AsyncUnagnt

Asynchronous client (all methods are async/await).

### Types

- `Run` - Run details with state, step count, timestamps
- `CreateRunRequest` - Request parameters for creating a run
- `CreateRunResponse` - Response with run_id

### Errors

- `UnagntError` - Base exception
- `APIError` - API request failed
- `NotFoundError` - Resource not found (404)
- `UnauthorizedError` - Auth failed (401)
- `TimeoutError` - Operation timed out

## Development

### Setup

```bash
pip install -e ".[dev]"
```

### Run Tests

```bash
pytest
```

### Type Checking

```bash
mypy unagnt/
```

## License

MIT
