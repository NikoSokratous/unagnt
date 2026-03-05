# ADR 0008: Multi-Agent Orchestration

**Status**: Accepted  
**Date**: 2026-02-26  
**Decision Makers**: Unagnt Core Team

## Context

Complex tasks often require multiple specialized agents working together. Users need:
- Sequential workflows (pipeline pattern)
- Parallel execution (fan-out pattern)
- Agent-to-agent delegation
- Output passing between agents
- Conditional execution
- Error handling strategies

## Decision

Implement a workflow engine supporting:

1. **Sequential workflows**: Steps execute one after another
2. **Parallel workflows**: Steps execute concurrently
3. **Agent delegation**: Direct agent-to-agent task passing
4. **Output context**: Share data between workflow steps

### Workflow Configuration

YAML-based workflow definitions:

```yaml
name: workflow-name
description: Workflow description

# Sequential execution
steps:
  - name: step1
    agent: agent-1
    goal: "Do task 1"
    output_key: result1
  
  - name: step2
    agent: agent-2
    goal: "Do task 2 with {{.result1}}"

# OR parallel execution
parallel:
  - name: task-a
    agent: agent-a
    goal: "Task A"
  
  - name: task-b
    agent: agent-b
    goal: "Task B"

timeout: 10m
on_error: stop  # or "continue"
```

### Execution Model

- **Isolated runs**: Each step gets its own `RunMeta` and `RunID`
- **Shared context**: Output from step N available to step N+1
- **Timeout cascade**: Parent timeout applies to all children
- **Cancellation propagation**: Cancelling workflow cancels all active steps

## Alternatives Considered

### 1. Embedded Workflows (Agent-Defined)

**Approach**: Let agents internally coordinate with other agents

**Pros:**
- More flexible, agents control flow
- No separate workflow configuration

**Cons:**
- Less observable (coordination hidden in code)
- Harder to audit and replay
- No declarative workflow definition
- Difficult to enforce policies

**Decision**: Rejected. Declarative workflows provide better governance.

### 2. DAG-Based Workflows (Airflow-style)

**Approach**: Define workflows as directed acyclic graphs with dependencies

**Pros:**
- Very powerful and expressive
- Supports complex branching
- Well-understood pattern (Airflow, Argo)

**Cons:**
- More complex to configure
- Overkill for most use cases
- Harder to reason about
- Requires graph visualization

**Decision**: Deferred. Start simple with sequential/parallel, add DAG later if needed.

### 3. State Machine Workflows (Step Functions)

**Approach**: AWS Step Functions-style state machines

**Pros:**
- Very powerful
- Built-in error handling and retries
- Standard JSON schema

**Cons:**
- Steep learning curve
- Verbose configuration
- Complex error handling logic

**Decision**: Rejected. Too complex for initial implementation.

### 4. Code-Based Composition

**Approach**: Define workflows in code (Go/Python)

```go
workflow := NewWorkflow().
    AddStep("research", researchAgent).
    AddStep("summarize", summaryAgent).
    Execute()
```

**Pros:**
- Type-safe
- IDE support
- Programmatic control

**Cons:**
- Requires recompilation
- Not declarative
- Harder to version and audit

**Decision**: Rejected. YAML is more accessible and auditable.

## Consequences

### Positive

- **Observability**: Each step is a tracked run with full event history
- **Reusability**: Workflows are portable YAML files
- **Flexibility**: Sequential and parallel modes cover most patterns
- **Governance**: Policies apply to all workflow steps
- **Debugging**: Can replay entire workflows or individual steps

### Negative

- **Overhead**: Each step creates a new run (storage cost)
- **Latency**: Inter-agent communication adds delay
- **Complexity**: New abstraction layer to maintain
- **Learning curve**: Users must learn workflow syntax

### Neutral

- **Delegation vs Workflows**: Two ways to compose agents (may confuse users)
- **Template complexity**: Go templates powerful but error-prone

## Implementation Details

### Workflow Engine

```go
engine := orchestrate.NewWorkflowEngine(server)
result, err := engine.Execute(ctx, workflowConfig)

for _, step := range result.Steps {
    fmt.Printf("Step %s: %s\n", step.Name, step.Status)
}
```

### Agent Delegation

```go
delegator := runtime.NewAgentDelegator()

// Sync delegation
result, err := delegator.DelegateToAgent(ctx, "specialist", "Task", context)

// Async delegation
runID, err := delegator.DelegateToAgentAsync(ctx, "worker", "Background task", nil)
```

### Output Passing

```yaml
steps:
  - name: fetch
    output_key: data
  
  - name: process
    goal: "Process {{.data}}"  # Access previous output
```

## Migration and Adoption

### Phase 1 (v0.4)

- Basic sequential and parallel workflows
- Agent delegation primitives
- CLI commands for validation and execution

### Phase 2 (v0.5)

- Conditional step execution (CEL expressions)
- DAG-based workflows
- Workflow templates and composition

### Phase 3 (v0.6)

- Visual workflow editor
- Workflow marketplace
- Advanced error recovery

## Testing Strategy

- **Unit tests**: Test workflow validation and execution logic
- **Integration tests**: Test multi-agent coordination
- **Performance tests**: Measure overhead of orchestration layer
- **Example workflows**: Provide reference implementations

## Monitoring

Workflow-level metrics:
- `workflow_execution_duration_seconds`
- `workflow_step_count`
- `workflow_failures_total`
- `active_workflows`

## References

- Implementation: `pkg/orchestrate/workflow.go`
- Delegation: `pkg/runtime/delegation.go`
- Examples: `examples/multi-agent-workflow/`
