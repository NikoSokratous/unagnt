# Agent Runtime Go Client

Go client library for the Agent Runtime API.

## Installation

```bash
go get github.com/NikoSokratous/unagnt/sdk/go
```

## Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/NikoSokratous/unagnt/sdk/go/client"
)

func main() {
    // Create client
    c := client.New("http://localhost:8080", "your-api-key")
    
    ctx := context.Background()
    
    // Create a run
    resp, err := c.CreateRun(ctx, "demo-agent", "Calculate 15 + 27")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Created run: %s\n", resp.RunID)
    
    // Wait for completion
    run, err := c.WaitForRun(ctx, resp.RunID, 2*time.Second)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Run completed with state: %s\n", run.State)
    fmt.Printf("Steps: %d\n", run.StepCount)
}
```

## API

### Creating a Client

```go
client := client.New(baseURL, apiKey)
```

### Methods

- `CreateRun(ctx, agentName, goal)` - Create a new agent run
- `GetRun(ctx, runID)` - Get run details
- `ListRuns(ctx, limit)` - List recent runs
- `CancelRun(ctx, runID)` - Cancel a running agent
- `WaitForRun(ctx, runID, pollInterval)` - Wait for run completion
- `HealthCheck(ctx)` - Check service health

### Custom HTTP Client

```go
import "net/http"

httpClient := &http.Client{
    Timeout: 60 * time.Second,
}

client := client.New(baseURL, apiKey).WithHTTPClient(httpClient)
```

## Testing

```bash
go test ./...
```

## License

MIT
