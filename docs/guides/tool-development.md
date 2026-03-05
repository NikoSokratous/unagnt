# Tool Development Guide

Complete guide to creating custom tools for Agent Runtime.

## Table of Contents

1. [Overview](#overview)
2. [Tool Interface](#tool-interface)
3. [Creating Your First Tool](#creating-your-first-tool)
4. [JSON Schema](#json-schema)
5. [Permissions](#permissions)
6. [Testing](#testing)
7. [Advanced Patterns](#advanced-patterns)
8. [Best Practices](#best-practices)

## Overview

Tools are the actions an agent can take. Each tool:
- Implements a Go interface
- Declares input schema (JSON Schema)
- Specifies required permissions
- Returns structured output

### Built-in Tools

Agent Runtime includes:
- `echo` - Echo input (testing)
- `calc` - Basic arithmetic
- `http_request` - HTTP requests

Location: `pkg/tool/builtin/`

## Tool Interface

All tools implement `tool.Tool`:

```go
type Tool interface {
    Name() string                          // Tool identifier
    Version() string                       // Semantic version
    Description() string                   // Human-readable description
    InputSchema() ([]byte, error)          // JSON Schema for validation
    Permissions() []Permission             // Required permissions
    Execute(ctx context.Context, input json.RawMessage) (map[string]any, error)
}
```

## Creating Your First Tool

### Step 1: Scaffold

```bash
unagnt scaffold tool --name github_api --output tools/
```

This creates:
- `tools/github_api.go` - Implementation
- `tools/github_api_test.go` - Tests

### Step 2: Implement the Interface

```go
package tools

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/Unagnt/Unagnt/pkg/tool"
)

type GitHubAPI struct{}

func (t *GitHubAPI) Name() string {
    return "github_api"
}

func (t *GitHubAPI) Version() string {
    return "1"
}

func (t *GitHubAPI) Description() string {
    return "Interact with GitHub API (list repos, create issues, etc.)"
}

func (t *GitHubAPI) InputSchema() ([]byte, error) {
    return []byte(`{
        "type": "object",
        "properties": {
            "action": {
                "type": "string",
                "enum": ["list_repos", "create_issue", "get_pr"],
                "description": "Action to perform"
            },
            "owner": {
                "type": "string",
                "description": "Repository owner"
            },
            "repo": {
                "type": "string",
                "description": "Repository name"
            }
        },
        "required": ["action", "owner", "repo"]
    }`), nil
}

func (t *GitHubAPI) Permissions() []tool.Permission {
    return []tool.Permission{
        {Scope: "net:external", Required: true},
        {Scope: "github:read", Required: true},
    }
}

func (t *GitHubAPI) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
    var params struct {
        Action string `json:"action"`
        Owner  string `json:"owner"`
        Repo   string `json:"repo"`
    }
    
    if err := json.Unmarshal(input, &params); err != nil {
        return nil, fmt.Errorf("invalid input: %w", err)
    }
    
    // Implement action logic
    switch params.Action {
    case "list_repos":
        // Call GitHub API
        return map[string]any{
            "repos": []string{"repo1", "repo2"},
        }, nil
    
    case "create_issue":
        return nil, fmt.Errorf("not yet implemented")
    
    default:
        return nil, fmt.Errorf("unknown action: %s", params.Action)
    }
}
```

### Step 3: Register the Tool

```go
// In your main package or plugin
registry := tool.NewRegistry()
registry.Register(&GitHubAPI{})
```

## JSON Schema

### Basic Types

```json
{
  "type": "object",
  "properties": {
    "name": {"type": "string"},
    "age": {"type": "integer"},
    "active": {"type": "boolean"},
    "score": {"type": "number"}
  }
}
```

### Enums

```json
{
  "type": "string",
  "enum": ["option1", "option2", "option3"]
}
```

### Arrays

```json
{
  "type": "array",
  "items": {"type": "string"},
  "minItems": 1
}
```

### Nested Objects

```json
{
  "type": "object",
  "properties": {
    "user": {
      "type": "object",
      "properties": {
        "name": {"type": "string"},
        "email": {"type": "string", "format": "email"}
      }
    }
  }
}
```

### Validation

The runtime validates inputs against your schema before `Execute()` is called.

## Permissions

Declare what access your tool needs:

```go
func (t *MyTool) Permissions() []tool.Permission {
    return []tool.Permission{
        {Scope: "net:external", Required: true},    // Internet access
        {Scope: "fs:read", Required: true},         // Read filesystem
        {Scope: "fs:write", Required: false},       // Optional write
        {Scope: "db:write", Required: true},        // Database writes
        {Scope: "exec:shell", Required: true},      // Shell execution
    }
}
```

### Permission Scopes

- `net:external` - Make external HTTP requests
- `net:internal` - Access internal services
- `fs:read` - Read files
- `fs:write` - Write files
- `db:read` - Database queries
- `db:write` - Database modifications
- `exec:shell` - Execute shell commands
- `email:send` - Send emails

### Policy Enforcement

Permissions are checked against policies before execution.

## Testing

### Unit Test Template

```go
func TestGitHubAPI(t *testing.T) {
    tool := &GitHubAPI{}
    ctx := context.Background()
    
    // Test schema validity
    schema, err := tool.InputSchema()
    if err != nil {
        t.Fatal(err)
    }
    
    var parsed map[string]any
    if err := json.Unmarshal(schema, &parsed); err != nil {
        t.Fatalf("Invalid JSON schema: %v", err)
    }
    
    // Test execution
    input := json.RawMessage(`{"action":"list_repos","owner":"test","repo":"test"}`)
    result, err := tool.Execute(ctx, input)
    if err != nil {
        t.Fatal(err)
    }
    
    if result["repos"] == nil {
        t.Error("Expected repos in result")
    }
}
```

### Validation with CLI

```bash
unagnt tool validate --tool github_api
```

## Advanced Patterns

### 1. Stateful Tools

```go
type DatabaseTool struct {
    conn *sql.DB
}

func NewDatabaseTool(dsn string) (*DatabaseTool, error) {
    conn, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    return &DatabaseTool{conn: conn}, nil
}
```

### 2. Context Cancellation

```go
func (t *LongRunningTool) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
    // Check context periodically
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // Continue processing
    }
}
```

### 3. Progress Reporting

```go
// Use metadata in ToolResult
return map[string]any{
    "status": "in_progress",
    "progress": 45,
    "message": "Processing file 45/100",
}, nil
```

### 4. Tool Composition

```go
type CompositeTool struct {
    registry *tool.Registry
}

func (t *CompositeTool) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
    // Call other tools internally
    result1, _ := t.registry.Execute(ctx, "tool1", "1", input1)
    result2, _ := t.registry.Execute(ctx, "tool2", "1", input2)
    
    // Combine results
    return map[string]any{"combined": []any{result1, result2}}, nil
}
```

## Best Practices

### 1. Descriptive Names

Bad: `tool1`, `proc`, `do_thing`  
Good: `github_create_issue`, `sql_query`, `send_email`

### 2. Semantic Versioning

- `1` → `2` for breaking changes
- Keep old versions available for compatibility

### 3. Error Handling

```go
// Return structured errors
if unauthorized {
    return nil, fmt.Errorf("github auth failed: %w", err)
}

// Don't panic
// Don't return generic "error occurred"
```

### 4. Input Validation

JSON Schema validates structure, but add business logic validation:

```go
if params.Amount < 0 {
    return nil, fmt.Errorf("amount must be positive")
}
```

### 5. Idempotency

Design tools to be safely retryable:

```go
// Check if already done
if alreadyExists(params.ID) {
    return map[string]any{"status": "already_exists"}, nil
}
```

### 6. Timeouts

Respect context deadlines:

```go
func (t *SlowTool) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
    client := &http.Client{Timeout: 5 * time.Second}
    // Use client with ctx
}
```

## Common Pitfalls

1. **Forgetting Context**: Always use `ctx` for cancellation
2. **Blocking I/O**: Use timeouts for external calls
3. **State in Tool**: Tools should be stateless (use agent memory instead)
4. **Ignoring Errors**: Always return errors, don't log and continue
5. **Complex Schemas**: Keep schemas simple, validate in code

## Example: Complete Tool

See `pkg/tool/builtin/http_request.go` for a production-quality example.

## Distribution

### As Go Package

```go
// In your module
go get github.com/yourorg/agent-tools
```

### Tool Registry

Future: Central tool registry for discovery and sharing.

## Next Steps

- Read [Policy Writing Guide](policy-writing.md) to control tool usage
- See [examples/](../../examples/) for real-world tools
- Validate your tool: `unagnt tool validate --tool <name>`

## Questions?

- Check existing tools in `pkg/tool/builtin/`
- Open an issue on GitHub
- Join our Discord community
