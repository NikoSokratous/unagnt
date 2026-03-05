# Context Assembly Engine

## Overview

The Context Assembly Engine is a sophisticated system that automatically retrieves and assembles memory, policies, workflow state, and tool outputs before each LLM call. This infrastructure-level feature eliminates manual prompt engineering and ensures agents have the right context at the right time.

## Why Context Assembly?

Traditional agent frameworks require developers to manually build prompts by:
- Writing code to retrieve from memory stores
- Manually concatenating policy constraints
- Hand-coding workflow state tracking
- Managing token budgets manually
- Debugging black-box prompt construction

Unagnt's Context Assembly Engine handles all of this automatically through a declarative, configurable pipeline.

## Architecture

```
User Input + Goal
       ↓
Context Assembly Engine
       ↓
┌──────────────────────────┐
│   Context Providers      │
│  - Memory Provider       │
│  - Policy Provider       │
│  - Workflow Provider     │
│  - Tool Output Provider  │
│  - Knowledge Provider    │
└──────────────────────────┘
       ↓
   Context Fragments
       ↓
   Context Assembler
   (Token Budgeting)
       ↓
  Assembled Messages
       ↓
   LLM Chat API
```

## Key Concepts

### Context Providers

Providers retrieve specific types of context:

**Memory Provider**
- Retrieves from working memory (current session)
- Fetches from persistent memory (facts, preferences)
- Searches semantic memory (similar past interactions) with embeddings
- Configurable similarity thresholds and top-K
- Supports both OpenAI and local embedding models

**Policy Provider**
- Injects active policy constraints
- Formats as readable system context
- Only includes policies relevant to available tools
- Example: "You must not delete files. Operations > $5 require approval."

**Workflow Provider**
- Adds workflow execution state
- Shows current step position
- Summarizes previous steps
- Includes parallel branch status

**Tool Output Provider**
- Consolidates recent tool execution results
- Configurable window (e.g., last 10 calls)
- Includes both successes and errors
- Formats as structured history

**Knowledge Provider**
- RAG (Retrieval Augmented Generation) integration
- Semantic search on domain knowledge with embeddings
- Automatic document chunking and ingestion
- Configurable knowledge sources (directories)
- Supports markdown and text files

### Context Fragments

Each provider returns a `ContextFragment`:

```go
type ContextFragment struct {
    ProviderName string       // e.g., "memory"
    Type         FragmentType // e.g., FragmentTypeMemory
    Content      string       // Formatted context text
    Priority     int          // Lower = higher priority
    TokenCount   int          // Estimated tokens
    Metadata     map[string]any
    Timestamp    time.Time
}
```

### Context Assembler

The assembler merges fragments with token budgeting:

**Token Budget Allocation:**
- System prompt: 500 tokens
- Policies: 1000 tokens
- Workflow state: 500 tokens
- Memory: 3000 tokens (largest allocation)
- Tool outputs: 2000 tokens
- User goal: 1000 tokens
- Knowledge: 1000 tokens

**Total: 8000 tokens** (leaves 4000 for response in 12k context window)

**Priority System:**
1. Policies (always include)
2. Workflow state (always include)
3. Semantic memory (up to budget)
4. Tool outputs (up to budget)
5. Knowledge retrieval (if tokens remain)

## Configuration

### Agent YAML

Enable context assembly in your agent configuration:

```yaml
agent:
  name: support_agent
  model:
    provider: anthropic
    name: claude-3-5-sonnet
  
  # Context assembly configuration
  context_assembly:
    enabled: true
    max_context_tokens: 8000
    parallel: false  # Fetch providers in parallel
    
    providers:
      - type: policy
        priority: 1
        enabled: true
        
      - type: workflow_state
        priority: 2
        enabled: true
        
      - type: semantic_memory
        priority: 3
        enabled: true
        config:
          top_k: 5
          similarity_threshold: 0.7
          use_embeddings: true  # NEW: Enable semantic search
          
      - type: tool_outputs
        priority: 4
        enabled: true
        config:
          window: 10
          
      - type: knowledge
        priority: 5
        enabled: true  # NEW: RAG enabled in v1.0
        config:
          sources: ["./docs", "./knowledge"]
          top_k: 3
          chunk_size: 500
          chunk_overlap: 50
    
    assembly:
      token_budget:
        system_prompt: 500
        policies: 1000
        workflow_state: 500
        memory: 3000
        tool_outputs: 2000
        user_goal: 1000
        knowledge: 1500  # Budget for RAG results
```

### Programmatic Configuration

```go
import (
    "github.com/Unagnt/Unagnt/pkg/context"
    "github.com/Unagnt/Unagnt/pkg/llm"
)

// Create context engine
config := context.EngineConfig{
    MaxTokens:     8000,
    EnableCache:   true,
    CacheDuration: 30 * time.Second,
    Parallel:      true,
}

assembler := context.NewDefaultAssembler()
engine := context.NewEngine(config, assembler)

// Add providers
engine.AddProvider(context.NewPolicyProvider(policyEngine, 1))
engine.AddProvider(context.NewWorkflowProvider(2))
engine.AddProvider(context.NewMemoryProvider(memoryManager, 3))
engine.AddProvider(context.NewToolOutputProvider(4, 10))
engine.AddProvider(context.NewKnowledgeProvider(5, []string{"docs"}))

// Use with planner
planner := &llm.PlannerAdapter{
    Provider:      llmProvider,
    Tools:         tools,
    ContextEngine: engine,
}
```

## CLI Commands

### Inspect Context

View assembled context for a run:

```bash
unagnt context inspect <run-id>

# Inspect specific step
unagnt context inspect <run-id> --step 5

# JSON output
unagnt context inspect <run-id> --format json
```

Output shows:
- Total token usage
- Fragments included/excluded
- Provider execution times
- Token counts per section

### Explain Context

Understand why each piece was included:

```bash
unagnt context explain <run-id> --step 5
```

Shows:
- Why each fragment was included/excluded
- Token budget allocation
- Similarity scores for memory retrieval
- Truncation decisions

### Knowledge Management (NEW in v1.0)

```bash
# Ingest documents into knowledge base
unagnt context ingest ./docs --source "documentation"

# List ingested documents
unagnt context knowledge list

# Search knowledge base
unagnt context search "how do I configure policies?" --top-k 5

# Clear knowledge base
unagnt context knowledge clear --yes
```

### Compare Context

Diff context between runs:

```bash
unagnt context diff <run-1> <run-2>
```

Shows:
- Token usage differences
- Fragment changes
- Performance differences

### Context Statistics

View performance metrics:

```bash
unagnt context stats <run-id>
```

Shows:
- Assembly duration
- Provider fetch times
- Cache hit rates
- Truncation events

### Validate Configuration

Check context assembly config:

```bash
unagnt context validate agent.yaml
```

Validates:
- Token budgets sum correctly
- Provider priorities are unique
- All types are recognized
- Configuration is well-formed

## Debugging

### Interactive Debugger

Enhanced `unagnt debug` with context inspection:

```bash
unagnt debug --config agent.yaml --goal "Research AI papers"
```

Commands:
- `context` - Show full context assembly
- `context memory` - Show only memory fragments
- `context policy` - Show only policy context
- `context tokens` - Show token budget breakdown

Example session:

```
debug> context
Context Assembly Information:
  Total tokens: ~7850 / 8000
  Fragments:
    ✓ Policy context (850 tokens)
    ✓ Workflow state (420 tokens)
    ✓ Memory (2800 tokens)
    ✓ Tool outputs (1900 tokens)

debug> context tokens
Token Budget Breakdown:
  System prompt: 500 / 500 (100%)
  Policy: 850 / 1000 (85%)
  Workflow: 420 / 500 (84%)
  Memory: 2800 / 3000 (93%)
  Tools: 1900 / 2000 (95%)
  User goal: 880 / 1000 (88%)
  ─────────────────────────
  Total: 7850 / 8000 (98%)
```

## Observability

### Metrics

Context assembly exposes metrics:

```
context_assembly_duration_ms{provider="memory"} 32
context_assembly_duration_ms{provider="policy"} 2
context_fragment_count{type="memory"} 5
context_tokens_used{section="memory"} 2800
context_truncation_events 0
```

### Tracing

Distributed tracing spans:

```
context.assemble (45ms)
├── context.provider.fetch (memory) (32ms)
├── context.provider.fetch (policy) (2ms)
├── context.provider.fetch (workflow) (1ms)
├── context.provider.fetch (tools) (8ms)
└── context.token_budget (2ms)
```

### Logs

Structured logging with context metadata:

```json
{
  "level": "info",
  "msg": "context_assembled",
  "run_id": "abc123",
  "step": 5,
  "total_tokens": 7850,
  "duration_ms": 45,
  "fragments": 4,
  "truncated": 0
}
```

## Performance

### Caching Strategy

Context fragments are cached for 30 seconds:
- Policies rarely change between steps
- Workflow state updates on step completion
- Memory retrieval can be cached briefly
- **40-60% latency reduction** from caching

### Parallel Fetching

Enable parallel provider execution:

```yaml
context_assembly:
  parallel: true
```

- Fetches all providers concurrently
- **2-3x faster** than sequential
- Trade-off: slightly higher resource usage

### Token Budget Optimization

Smart truncation preserves most important context:
- Always include high-priority fragments
- Truncate low-priority content first
- Keep recent + relevant items
- Log truncation events for visibility

## Best Practices

### 1. Tune Token Budgets

Adjust based on your use case:

```yaml
assembly:
  token_budget:
    memory: 4000  # Increase if memory-heavy
    tool_outputs: 1500  # Decrease if simple tools
```

### 2. Set Appropriate Priorities

Lower number = higher priority:
- Critical context: 1-2
- Important context: 3-4
- Optional context: 5+

### 3. Configure Memory Retrieval

Tune semantic search parameters:

```yaml
providers:
  - type: semantic_memory
    config:
      top_k: 10  # More results
      similarity_threshold: 0.6  # Lower threshold
```

### 4. Monitor Performance

Watch for:
- High assembly duration (> 100ms)
- Frequent truncation events
- Low cache hit rates
- Provider timeouts

### 5. Use Validation

Always validate configuration:

```bash
unagnt context validate agent.yaml
```

## Comparison with Other Frameworks

### LangGraph

**Manual prompt building:**
```python
# LangGraph: Manual context assembly
def build_context(state):
    messages = [SystemMessage(content="You are an agent")]
    
    # Manually fetch memory
    if state.get("memory"):
        memory_str = retrieve_memory(state["query"])
        messages.append(HumanMessage(content=f"Memory: {memory_str}"))
    
    # Manually add policies
    if state.get("policies"):
        policy_str = format_policies(state["policies"])
        messages[0].content += f"\n\nPolicies: {policy_str}"
    
    # Manually format history
    for step in state.get("history", []):
        # ... manual formatting ...
    
    return messages
```

**Unagnt:**
```yaml
# Automatic context assembly
context_assembly:
  enabled: true
  providers:
    - type: memory
    - type: policy
    - type: tool_outputs
```

### CrewAI

CrewAI has no built-in context assembly - all manual.

### AutoGen

AutoGen focuses on multi-agent conversations, not context assembly.

## Roadmap

### v1.0 (Current Release)

- ✅ Core context engine
- ✅ 5 built-in providers
- ✅ Token budgeting
- ✅ CLI commands
- ✅ Observability
- ✅ Semantic search with embeddings (OpenAI + local)
- ✅ RAG with knowledge base ingestion
- ✅ Document chunking and embedding generation

### v1.1 (Next Release)

- [ ] Context optimization recommendations
- [ ] A/B testing different assemblies
- [ ] Visual context flow in Web UI
- [ ] Context caching layer improvements
- [ ] Multi-source knowledge aggregation

### v2.0 (Long-term)

- [ ] ML-based token budget optimization
- [ ] Adaptive context assembly
- [ ] Multi-modal context (images, audio)
- [ ] Context versioning and rollback

## Troubleshooting

### High Token Usage

**Problem:** Context always near max tokens

**Solutions:**
- Reduce memory top_k
- Decrease tool output window
- Lower token budgets for low-priority sections
- Enable truncation for optional content

### Slow Assembly

**Problem:** Assembly takes > 100ms

**Solutions:**
- Enable parallel fetching
- Enable caching
- Optimize memory queries
- Use faster providers

### Missing Context

**Problem:** Important context not included

**Solutions:**
- Increase token budget for that section
- Raise provider priority
- Check provider is enabled
- Verify fetch doesn't error

### Cache Issues

**Problem:** Stale context

**Solutions:**
- Reduce cache duration
- Disable caching for dynamic content
- Clear cache on state changes

## Contributing

To add a custom provider:

1. Implement `ContextProvider` interface
2. Add to engine configuration
3. Register in provider registry
4. Add tests
5. Update documentation

Example:

```go
type CustomProvider struct {
    Priority int
}

func (p *CustomProvider) Name() string {
    return "custom"
}

func (p *CustomProvider) Priority() int {
    return p.Priority
}

func (p *CustomProvider) Fetch(ctx context.Context, input ContextInput) (*ContextFragment, error) {
    // Your custom logic
    return &ContextFragment{
        ProviderName: p.Name(),
        Type:         FragmentType("custom"),
        Content:      "Custom context",
        Priority:     p.Priority,
        TokenCount:   100,
    }, nil
}
```

## References

- [Embeddings and Semantic Search Guide](EMBEDDINGS.md)
- [Architecture Documentation](ARCHITECTURE.md)
- [User Guide](USER_GUIDE.md)
- [API Reference](API_REFERENCE.md)
- [Examples](../examples/)

## Support

- GitHub Issues: [Report bugs](https://github.com/NikoSokratous/unagnt/issues)
- Discussions: [Ask questions](https://github.com/NikoSokratous/unagnt/discussions)
- Discord: [Join community](https://discord.gg/Unagnt)
