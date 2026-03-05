# Multi-Agent Orchestration Guide

## Overview

The Unagnt supports complex multi-agent workflows where multiple agents collaborate to accomplish sophisticated tasks. This includes sequential pipelines, parallel execution, and agent-to-agent delegation.

## Workflow Engine

### Sequential Workflows

Execute agents one after another, passing outputs between steps:

```yaml
name: research-and-summarize
description: Research a topic and generate a summary

steps:
  - name: research
    agent: research-assistant
    goal: "Research {{.topic}} and gather key facts"
    output_key: research_data
    timeout: 5m
  
  - name: summarize
    agent: summarizer
    goal: "Summarize the following research: {{.research_data}}"
    output_key: summary
    timeout: 2m
  
  - name: review
    agent: quality-reviewer
    goal: "Review this summary for accuracy: {{.summary}}"
    output_key: final_summary

timeout: 15m
on_error: stop
```

**Execution:**

```bash
unagnt workflow run --config research-workflow.yaml
```

### Parallel Workflows

Run multiple agents simultaneously for faster processing:

```yaml
name: multi-source-analysis
description: Analyze data from multiple sources concurrently

parallel:
  - name: twitter-analysis
    agent: twitter-analyzer
    goal: "Analyze Twitter sentiment for {{.topic}}"
    output_key: twitter_data
  
  - name: reddit-analysis
    agent: reddit-analyzer
    goal: "Analyze Reddit discussions for {{.topic}}"
    output_key: reddit_data
  
  - name: news-analysis
    agent: news-analyzer
    goal: "Analyze news articles for {{.topic}}"
    output_key: news_data

aggregate: true
timeout: 10m
on_error: continue
```

**Key differences:**
- All steps run concurrently
- `on_error: continue` allows partial success
- `aggregate: true` combines all outputs

## Agent Delegation

Agents can delegate sub-tasks to other specialized agents:

### Programmatic Delegation

```go
import "github.com/Unagnt/Unagnt/pkg/runtime"

delegator := runtime.NewAgentDelegator()

// Synchronous delegation (wait for completion)
result, err := delegator.DelegateToAgent(
    ctx,
    "specialist-agent",
    "Analyze this data: ...",
    map[string]interface{}{
        "data": data,
        "priority": "high",
    },
)

fmt.Printf("Delegate completed: %s\n", result.State)
fmt.Printf("Output: %v\n", result.Output)
```

### Asynchronous Delegation

```go
// Fire and forget
runID, err := delegator.DelegateToAgentAsync(
    ctx,
    "background-processor",
    "Process large dataset",
    context,
)

// Later, check result
result, err := delegator.GetDelegationResult(runID)
```

## Workflow Patterns

### Sequential Pipeline

**Use case**: Data processing pipeline

```yaml
steps:
  - name: extract
    agent: data-extractor
    goal: "Extract data from {{.source}}"
    output_key: raw_data
  
  - name: transform
    agent: data-transformer
    goal: "Transform: {{.raw_data}}"
    output_key: clean_data
  
  - name: load
    agent: data-loader
    goal: "Load data to {{.destination}}: {{.clean_data}}"
```

### Fan-Out/Fan-In

**Use case**: Parallel processing with aggregation

```yaml
name: document-analysis

parallel:
  - name: sentiment
    agent: sentiment-analyzer
    output_key: sentiment
  
  - name: entities
    agent: entity-extractor
    output_key: entities
  
  - name: topics
    agent: topic-classifier
    output_key: topics

# Aggregation step (planned for future)
# aggregate_with: aggregator-agent
```

### Conditional Execution

**Use case**: Decision-based routing

```yaml
steps:
  - name: classify
    agent: classifier
    goal: "Classify request: {{.request}}"
    output_key: category
  
  - name: process-urgent
    agent: urgent-handler
    goal: "Handle urgent: {{.request}}"
    condition: "{{eq .category 'urgent'}}"
  
  - name: process-normal
    agent: normal-handler
    goal: "Handle normal: {{.request}}"
    condition: "{{ne .category 'urgent'}}"
```

## CLI Commands

### Validate Workflow

```bash
unagnt workflow validate --config workflow.yaml
```

### Run Workflow

```bash
unagnt workflow run --config workflow.yaml --timeout 10m
```

### Dry Run

```bash
unagnt workflow run --config workflow.yaml --dry-run
```

## Monitoring

Track workflow execution:

```go
result, err := engine.Execute(ctx, workflow)

fmt.Printf("Workflow: %s\n", result.WorkflowName)
fmt.Printf("Status: %s\n", result.Status)
fmt.Printf("Duration: %v\n", result.Duration)

for _, step := range result.Steps {
    fmt.Printf("  Step: %s (%s)\n", step.Name, step.Status)
    fmt.Printf("    Agent: %s\n", step.Agent)
    fmt.Printf("    Duration: %v\n", step.Duration)
}
```

## Error Handling

### Stop on Error (Default)

```yaml
on_error: stop
```

First failure terminates the workflow.

### Continue on Error

```yaml
on_error: continue
```

All steps attempt execution regardless of failures. Useful for:
- Data collection from multiple sources
- Best-effort processing
- Audit/compliance checks

### Retry Logic

```yaml
steps:
  - name: flaky-api
    agent: api-caller
    goal: "Call external API"
    retry: 3
    timeout: 30s
```

## Best Practices

1. **Granular agents**: Each agent should have a single, well-defined responsibility
2. **Clear outputs**: Use descriptive `output_key` names
3. **Timeouts**: Set realistic timeouts per step and overall
4. **Idempotency**: Design agents to handle retries safely
5. **Error context**: Include enough info in outputs for debugging
6. **Testing**: Validate workflows before production deployment

## Advanced Patterns

### Human-in-the-Loop Workflows

```yaml
steps:
  - name: draft
    agent: content-generator
    goal: "Generate draft for {{.topic}}"
    output_key: draft
  
  - name: review
    agent: human-reviewer  # Has autonomy_level: supervised
    goal: "Review draft: {{.draft}}"
    output_key: approved_draft
  
  - name: publish
    agent: publisher
    goal: "Publish: {{.approved_draft}}"
```

### Recursive Workflows

Agent can trigger sub-workflows:

```go
func (a *OrchestratorAgent) Execute(ctx context.Context, goal string) error {
    // Analyze goal complexity
    if isComplex(goal) {
        // Delegate to workflow engine
        workflow := generateWorkflow(goal)
        return engine.Execute(ctx, workflow)
    }
    
    // Handle directly
    return a.handleSimpleGoal(ctx, goal)
}
```

## Performance

- **Parallel execution**: Steps run concurrently, reducing total time
- **Resource limits**: Configure max concurrent agents
- **Connection pooling**: Reuse agent instances when possible
- **Streaming**: Monitor progress in real-time via SSE

## Examples

See `examples/multi-agent-workflow/` for complete examples:
- Research and summarization pipeline
- Multi-source data analysis
- CI/CD automation workflows
- Content moderation pipeline
