# ADR 0004: LLM Provider Abstraction

## Status

Accepted

## Context

The runtime must support multiple LLM providers (OpenAI, Anthropic, Ollama, future providers). Each has different:
- API formats
- Authentication methods
- Tool/function calling conventions
- Message structures
- Rate limits and pricing

## Decision

We implement a **unified `Provider` interface** with adapters for each LLM API.

### Core Interface

```go
type Provider interface {
    Name() string
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}
```

### Unified Message Format

```go
type Message struct {
    Role    Role   // system, user, assistant, tool
    Content string
}

type ToolDef struct {
    Name        string
    Description string
    Parameters  Parameters // JSON Schema
}

type ChatRequest struct {
    Messages    []Message
    Temperature float64
    MaxTokens   int
    Tools       []ToolDef
}

type ChatResponse struct {
    Content      string
    ToolCalls    []ToolCallRef
    FinishReason string
    Usage        Usage
    Model        string
}
```

### Provider Implementations

- **OpenAI**: `pkg/llm/openai/openai.go`
- **Anthropic**: `pkg/llm/anthropic/anthropic.go`
- **Ollama**: `pkg/llm/ollama/ollama.go`

## Consequences

### Positive

- **Model Agnostic**: Switch providers without changing agent code
- **Testing**: Mock providers for unit tests
- **Flexibility**: Add new providers easily
- **Optimization**: Provider-specific optimizations hidden behind interface
- **Cost Control**: Easy to swap for cheaper models

### Negative

- Lowest common denominator (can't use provider-specific features)
- Conversion overhead for each provider
- Tool calling format differences require mapping

## Alternatives Considered

### 1. LangChain/LlamaIndex
- Pros: Existing ecosystem, many integrations
- Cons: Heavy dependencies, opinionated abstractions, not infrastructure-focused

### 2. Direct API Calls
- Pros: No abstraction overhead, full API access
- Cons: Tight coupling, can't swap providers, hard to test

### 3. Provider-Specific Agents
- Pros: Full API access, optimized per provider
- Cons: Code duplication, vendor lock-in

## Implementation Details

### Tool Calling Mapping

**OpenAI Format**:
```json
{
  "tools": [{
    "type": "function",
    "function": {
      "name": "calc",
      "parameters": {...}
    }
  }]
}
```

**Anthropic Format**:
```json
{
  "tools": [{
    "name": "calc",
    "input_schema": {...}
  }]
}
```

Both map to our `ToolDef` structure.

### PlannerAdapter

`pkg/llm/planner.go` wraps providers to implement `runtime.LLMPlanner`:

```go
type PlannerAdapter struct {
    Provider Provider
    Tools    []ToolInfo
}

func (p *PlannerAdapter) Plan(ctx, input) (*PlannedAction, error)
```

## Trade-offs

### Supporting Provider-Specific Features

For features only some providers support (e.g., Claude's "thinking" tokens):
- Add optional fields to `ChatResponse`
- Providers return what they can
- Runtime ignores unsupported fields

### Future: Streaming

Will add:
```go
type StreamingProvider interface {
    Provider
    ChatStream(ctx, req) (<-chan Delta, error)
}
```

## Related

- Tool Registry: `pkg/tool/registry.go`
- Runtime Integration: `cmd/unagnt/run.go`
