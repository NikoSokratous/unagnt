# Unagnt User Guide

**Version**: 1.0.0 | **Last Updated**: 2026-02-27

Complete guide for using Unagnt to build and run AI agent applications.

---

## Table of Contents

1. [Introduction](#introduction)
2. [Core Concepts](#core-concepts)
3. [Working with Agents](#working-with-agents)
4. [Building Workflows](#building-workflows)
5. [Policy Management](#policy-management)
6. [Tools & Plugins](#tools--plugins)
7. [Memory Systems](#memory-systems)
8. [Web UI](#web-ui)
9. [Best Practices](#best-practices)
10. [Advanced Topics](#advanced-topics)

---

## Introduction

Unagnt is a platform for building and running autonomous AI agents with production-grade safety, observability, and scalability.

### What You Can Build

- **Autonomous Agents**: Self-directed AI that can plan and execute tasks
- **Multi-Agent Workflows**: Coordinated teams of specialized agents
- **Automation Pipelines**: ETL, code review, content generation
- **Interactive Applications**: Customer support, research assistants
- **Scheduled Jobs**: Recurring analysis, monitoring, reporting

### Key Capabilities

✅ **Multiple LLM Providers**: OpenAI, Anthropic, Ollama, or custom  
✅ **Policy Enforcement**: Control what agents can do  
✅ **Visual Workflows**: Drag-and-drop designer  
✅ **Full Observability**: Traces, metrics, replay  
✅ **Production Ready**: Scale, secure, monitor  

---

## Core Concepts

### Agents

An **agent** is an autonomous AI entity with:
- **Goal**: What it's trying to achieve
- **Tools**: What it can do (API calls, file ops, etc.)
- **Memory**: What it remembers
- **Policy**: What it's allowed to do

### Workflows

A **workflow** coordinates multiple agents:
- **Steps**: Sequential or parallel agent tasks
- **Dependencies**: Output from one agent → input to next
- **Conditions**: CEL-based branching logic
- **Templates**: Reusable workflow patterns

### Tools

**Tools** are actions agents can execute:
- **Built-in**: HTTP requests, calculations, echo
- **Custom**: Go plugins, WASM modules
- **Schema**: JSON Schema validation
- **Permissions**: Scoped access control

### Policies

**Policies** control agent behavior:
- **Rules**: CEL expressions for allow/deny
- **Risk Scoring**: Automatic risk assessment
- **Approval Gates**: Human-in-the-loop
- **Versioning**: Test before deploying

### Memory

**Memory** provides context to agents:
- **Working**: Current conversation (session)
- **Persistent**: Long-term facts (database)
- **Semantic**: Vector similarity search
- **Event Log**: Immutable execution history

---

## Working with Agents

### Creating an Agent

#### Using CLI

```bash
# Basic agent
unagnt agent create researcher \
  --goal "Research AI safety papers" \
  --llm gpt-4 \
  --max-steps 10

# With more options
unagnt agent create analyst \
  --goal "Analyze sales data and generate report" \
  --llm claude-3-opus \
  --max-steps 20 \
  --autonomy standard \
  --tools http_request,calculator,file_write
```

#### Using API

```bash
curl -X POST http://localhost:8080/api/v1/agents \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "id": "researcher",
    "goal": "Research AI safety papers",
    "llm_provider": "openai",
    "llm_model": "gpt-4",
    "max_steps": 10,
    "autonomy_level": "standard"
  }'
```

#### Using SDK (Go)

```go
package main

import (
    "github.com/Unagnt/Unagnt/pkg/client"
)

func main() {
    c := client.New("http://localhost:8080", "your-api-key")
    
    agent, err := c.CreateAgent(&client.AgentConfig{
        ID:           "researcher",
        Goal:         "Research AI safety papers",
        LLMProvider:  "openai",
        LLMModel:     "gpt-4",
        MaxSteps:     10,
        AutonomyLevel: "standard",
    })
    if err != nil {
        panic(err)
    }
    
    println("Created agent:", agent.ID)
}
```

#### Using SDK (Python)

```python
from Unagnt import UnagntClient

client = UnagntClient(
    base_url="http://localhost:8080",
    api_key="your-api-key"
)

agent = client.agents.create(
    id="researcher",
    goal="Research AI safety papers",
    llm_provider="openai",
    llm_model="gpt-4",
    max_steps=10,
    autonomy_level="standard"
)

print(f"Created agent: {agent.id}")
```

### Running an Agent

```bash
# Start execution
unagnt agent run researcher

# With custom context
unagnt agent run researcher \
  --context "Focus on papers from 2024"

# Watch live
unagnt runs watch <run-id>

# Get final result
unagnt runs get <run-id>
```

### Autonomy Levels

| Level | Description | Use Case |
|-------|-------------|----------|
| **Manual** | Require approval for every tool | Testing, learning |
| **Cautious** | Approve high-risk tools only | Sensitive operations |
| **Standard** | Balance safety and speed | Most production use |
| **Autonomous** | Minimal approvals | Trusted workflows |
| **Unrestricted** | No approvals (policy only) | Internal automation |

### Managing Agents

```bash
# List all agents
unagnt agents list

# Get agent details
unagnt agents get researcher

# Update agent
unagnt agents update researcher --max-steps 20

# Delete agent
unagnt agents delete researcher

# View agent runs
unagnt runs list --agent researcher

# Export agent config
unagnt agents export researcher > researcher.yaml
```

---

## Building Workflows

### Simple Sequential Workflow

```yaml
# research-workflow.yaml
name: "research-and-report"
description: "Research topic and create report"

steps:
  - name: research
    agent: researcher
    goal: "Research {{.topic}}"
    output_key: papers
    
  - name: analyze
    agent: analyzer
    goal: "Analyze these papers: {{.Outputs.papers}}"
    output_key: analysis
    
  - name: report
    agent: writer
    goal: "Create report from analysis: {{.Outputs.analysis}}"
    output_key: report
```

Run it:
```bash
unagnt workflow run research-workflow.yaml \
  --param topic="AI safety in robotics"
```

### Parallel Execution

```yaml
name: "multi-source-research"
description: "Research from multiple sources in parallel"

steps:
  - name: search-arxiv
    agent: arxiv-searcher
    goal: "Find papers on {{.topic}} in arXiv"
    output_key: arxiv_results
    parallel: true
    
  - name: search-web
    agent: web-searcher
    goal: "Search web for {{.topic}}"
    output_key: web_results
    parallel: true
    
  - name: search-pubmed
    agent: pubmed-searcher
    goal: "Find papers on {{.topic}} in PubMed"
    output_key: pubmed_results
    parallel: true
    
  # This runs after all parallel steps complete
  - name: aggregate
    agent: aggregator
    goal: "Combine: arXiv={{.Outputs.arxiv_results}}, Web={{.Outputs.web_results}}, PubMed={{.Outputs.pubmed_results}}"
    output_key: combined
```

### Conditional Steps

```yaml
name: "conditional-workflow"
description: "Use CEL for branching logic"

steps:
  - name: check-data
    agent: validator
    goal: "Validate data quality"
    output_key: validation
    
  - name: process-good-data
    agent: processor
    goal: "Process validated data"
    condition: |
      Outputs.validation.status == "valid"
    output_key: result
    
  - name: fix-bad-data
    agent: fixer
    goal: "Fix data issues"
    condition: |
      Outputs.validation.status == "invalid"
    output_key: fixed_data
```

### Using the Visual Designer

1. **Open Web UI**: Navigate to `http://localhost:3000/workflows`
2. **Create New**: Click "New Workflow"
3. **Drag Agents**: Drag agent nodes onto canvas
4. **Connect**: Draw edges between nodes
5. **Configure**: Click nodes to set goals, conditions
6. **Test**: Click "Run" to test execution
7. **Save**: Save as template for reuse

### Workflow Parameters

```yaml
name: "parameterized-workflow"
parameters:
  - name: repository_url
    type: string
    required: true
    description: "GitHub repository URL"
  - name: branch
    type: string
    default: "main"
  - name: severity_threshold
    type: string
    default: "medium"
    enum: ["low", "medium", "high"]

steps:
  - name: clone
    agent: git-agent
    goal: "Clone {{.repository_url}} branch {{.branch}}"
    output_key: code
```

### Workflow Templates

```bash
# List available templates
unagnt workflow templates list

# View template details
unagnt workflow templates get code-review

# Install template
unagnt workflow templates install code-review

# Run installed template
unagnt workflow run code-review \
  --param repository_url=https://github.com/user/repo \
  --param branch=main
```

---

## Policy Management

### Creating Policies

#### Simple Policy

```yaml
# policies/basic.yaml
name: safety-policy
version: "1.0.0"
description: "Basic safety rules"
environment: production

rules:
  - id: block-dangerous-commands
    description: "Block system commands"
    condition: |
      tool.name == "shell_execute" &&
      (tool.params.command.contains("rm -rf") ||
       tool.params.command.contains("shutdown"))
    action: deny
    severity: critical
    
  - id: require-approval-for-writes
    description: "Approve file writes"
    condition: |
      tool.name == "file_write"
    action: require_approval
    severity: medium
```

#### Advanced Policy with Risk Scoring

```yaml
name: enterprise-policy
version: "2.0.0"
environment: production

# Global settings
settings:
  default_action: allow
  risk_threshold: 7.0  # Deny if risk > 7.0
  
rules:
  - id: external-api-approval
    description: "Approve external API calls"
    condition: |
      tool.name == "http_request" &&
      !tool.params.url.startsWith("https://internal.company.com")
    action: require_approval
    severity: high
    approval_webhook: https://approvals.company.com/webhook
    
  - id: cost-limit
    description: "Block expensive operations"
    condition: |
      estimated_cost > 10.0  # $10
    action: deny
    severity: high
    
  - id: data-privacy
    description: "Protect PII"
    condition: |
      tool.params.contains("ssn") ||
      tool.params.contains("credit_card") ||
      tool.params.contains("password")
    action: deny
    severity: critical
    
risk_scoring:
  categories:
    - name: data_access
      weight: 1.5
      conditions:
        - "tool.permissions.contains('fs:read')"
        - "tool.permissions.contains('db:read')"
        
    - name: data_modification
      weight: 2.0
      conditions:
        - "tool.permissions.contains('fs:write')"
        - "tool.permissions.contains('db:write')"
        
    - name: external_communication
      weight: 1.2
      conditions:
        - "tool.permissions.contains('net:external')"
```

### Applying Policies

```bash
# Validate policy syntax
unagnt policy validate policies/safety-policy.yaml

# Apply policy
unagnt policy apply policies/safety-policy.yaml

# List active policies
unagnt policy list

# Get policy details
unagnt policy get safety-policy 1.0.0

# Delete policy version
unagnt policy delete safety-policy 1.0.0
```

### Testing Policies

```bash
# Simulate against past run
unagnt policy simulate enterprise-policy 2.0.0 \
  --run-id abc123

# Test with synthetic actions
unagnt policy test enterprise-policy 2.0.0 \
  --tool http_request \
  --params '{"url": "https://evil.com"}' \
  --expected deny
```

### Policy Versioning

```bash
# Create new version
unagnt policy apply policies/safety-policy-v2.yaml

# Compare versions
unagnt policy diff safety-policy 1.0.0 2.0.0

# Rollback
unagnt policy activate safety-policy 1.0.0

# View changelog
unagnt policy changelog safety-policy
```

---

## Tools & Plugins

### Built-in Tools

#### HTTP Request
```yaml
# Use in agent
tools:
  - http_request

# Params
{
  "url": "https://api.example.com/data",
  "method": "GET",
  "headers": {"Authorization": "Bearer token"},
  "body": "{\"key\": \"value\"}"
}
```

#### File Operations
```yaml
tools:
  - file_read
  - file_write
  - file_list

# Example: Read file
{
  "path": "/data/input.txt"
}
```

#### Calculator
```yaml
tools:
  - calculator

# Example
{
  "expression": "2 + 2 * 3"
}
```

### Creating Custom Tools

#### Go Plugin

```go
// tools/my_tool.go
package main

import (
    "github.com/Unagnt/Unagnt/pkg/tool"
)

type MyTool struct{}

func (t *MyTool) Name() string {
    return "my_tool"
}

func (t *MyTool) Description() string {
    return "Does something useful"
}

func (t *MyTool) Schema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "input": map[string]interface{}{
                "type": "string",
                "description": "Input data",
            },
        },
        "required": []string{"input"},
    }
}

func (t *MyTool) Permissions() []string {
    return []string{"custom:my_tool"}
}

func (t *MyTool) Execute(params map[string]interface{}) (interface{}, error) {
    input := params["input"].(string)
    // Do something
    return map[string]interface{}{
        "output": "processed: " + input,
    }, nil
}

// Export symbol
var Tool MyTool
```

Build and deploy:
```bash
# Build plugin
go build -buildmode=plugin -o my_tool.so tools/my_tool.go

# Copy to plugins directory
cp my_tool.so /app/plugins/

# Validate
unagnt tool validate --plugin my_tool.so

# Test
unagnt tool test my_tool --params '{"input": "test"}'
```

### Scaffolding Tools

```bash
# Generate boilerplate
unagnt scaffold tool \
  --name weather_lookup \
  --output tools/

# Edit generated file
# Build and deploy
```

### Managing Tools

```bash
# List all tools
unagnt tools list

# Get tool schema
unagnt tools schema http_request

# Test tool
unagnt tools test calculator \
  --params '{"expression": "2+2"}'

# Hot-reload (WASM only)
unagnt tools reload
```

---

## Memory Systems

### Working Memory

Stores current conversation context:

```go
// Add to working memory
memory.Set("user_name", "Alice")
memory.Set("last_query", "What's the weather?")

// Retrieve
name := memory.Get("user_name")
```

### Persistent Memory

Long-term storage:

```bash
# Store fact
unagnt memory set --agent researcher \
  --key "favorite_topic" \
  --value "AI safety"

# Retrieve
unagnt memory get --agent researcher --key "favorite_topic"

# List all
unagnt memory list --agent researcher

# Delete
unagnt memory delete --agent researcher --key "favorite_topic"
```

### Semantic Memory

Vector-based similarity search:

```bash
# Add memory
unagnt memory add-semantic --agent researcher \
  --text "AI safety is crucial for AGI development" \
  --metadata '{"source": "paper", "date": "2024-01-01"}'

# Search
unagnt memory search --agent researcher \
  --query "safety in artificial intelligence" \
  --limit 5
```

### Event Log

Immutable audit trail:

```bash
# Query events
unagnt events --agent researcher --run-id abc123

# Export for replay
unagnt events export --run-id abc123 > events.json

# Search events
unagnt events search --tool http_request --since "2024-01-01"
```

---

## Web UI

### Dashboard

- **Overview**: Active runs, recent completions
- **Stats**: Success rates, avg duration
- **Charts**: Runs over time, cost trends

### Workflow Designer

- **Canvas**: Drag-and-drop nodes
- **Palette**: Available agents
- **Properties**: Configure step details
- **Run**: Test execution in browser

### Marketplace

- **Browse**: Pre-built templates
- **Install**: One-click deployment
- **Customize**: Fork and modify
- **Publish**: Share your templates

### Analytics

- **Cost**: By agent, tenant, time period
- **Performance**: Latency percentiles, throughput
- **SLA**: Uptime, error rates
- **Policies**: Denials, approvals, risk distribution

### Monitoring

- **Live Runs**: Real-time status
- **Logs**: Searchable, filterable
- **Traces**: Distributed tracing view
- **Alerts**: Configure notifications

---

## Best Practices

### 1. Start Small

Begin with simple agents:
- Single goal
- Few tools
- Manual autonomy
- Test thoroughly

Then add complexity:
- More tools
- Higher autonomy
- Integrate into workflows

### 2. Use Policies

Always start with restrictive policies:
```yaml
# Start conservative
default_action: require_approval

# Gradually relax
default_action: allow
risk_threshold: 5.0
```

### 3. Monitor Everything

Enable full observability:
- Tracing for debugging
- Metrics for alerts
- Logs for audit
- Cost tracking for budget

### 4. Test Workflows

Before production:
```bash
# 1. Validate syntax
unagnt workflow validate workflow.yaml

# 2. Dry run
unagnt workflow run workflow.yaml --dry-run

# 3. Test with sample data
unagnt workflow run workflow.yaml --param test_mode=true

# 4. Load test
unagnt workflow load-test workflow.yaml --users 100
```

### 5. Version Control

Keep configs in git:
```
repo/
  agents/
    researcher.yaml
    analyzer.yaml
  workflows/
    research-pipeline.yaml
  policies/
    production.yaml
    staging.yaml
```

### 6. Use Templates

Don't reinvent the wheel:
```bash
# Start from template
unagnt workflow templates install code-review

# Customize
unagnt workflow export code-review > my-review.yaml
# Edit my-review.yaml

# Deploy
unagnt workflow apply my-review.yaml
```

### 7. Secure Secrets

Never hardcode API keys:
```yaml
# Bad
llm_api_key: "sk-..."

# Good
llm_api_key: ${OPENAI_API_KEY}
```

Use secrets management:
- Kubernetes Secrets
- Vault
- AWS Secrets Manager
- GCP Secret Manager

### 8. Handle Failures

Implement retry logic:
```yaml
steps:
  - name: flaky-api-call
    agent: api-agent
    retry:
      max_attempts: 3
      backoff: exponential
      initial_delay: 1s
```

Add fallbacks:
```yaml
steps:
  - name: primary
    agent: gpt4-agent
    on_failure: fallback
    
  - name: fallback
    agent: gpt3-agent
    goal: "Simpler version of task"
```

---

## Advanced Topics

### Deterministic Replay

Debug past runs:
```bash
# Replay exact execution
unagnt replay <run-id> --mode exact

# Replay with live LLM
unagnt replay <run-id> --mode live

# Debug with breakpoints
unagnt replay <run-id> --mode debug
```

### Multi-Tenancy

Isolate by tenant:
```bash
# Create tenant
unagnt tenants create acme-corp

# Create agent in tenant
unagnt agent create researcher \
  --tenant acme-corp \
  --goal "Research competitors"

# Query by tenant
unagnt runs list --tenant acme-corp
```

### Cost Attribution

Track spending:
```bash
# View costs
unagnt costs --by agent
unagnt costs --by tenant
unagnt costs --date-range "2024-01-01:2024-01-31"

# Set budgets
unagnt budget set --agent researcher --limit 100.00

# Alerts
unagnt budget alert --tenant acme-corp \
  --threshold 80% \
  --webhook https://alerts.example.com
```

### Agent Collaboration

Agents working together:
```yaml
name: "collaborative-workflow"
steps:
  - name: delegate-task
    agent: coordinator
    goal: "Delegate subtasks to specialist agents"
    delegation:
      enabled: true
      max_depth: 2  # How many levels deep
```

### Webhooks

Trigger on events:
```bash
# Create webhook
unagnt webhooks create \
  --name github-pr \
  --agent code-reviewer \
  --secret webhook-secret \
  --goal-template "Review PR: {{.pull_request.url}}"

# Test webhook
curl -X POST http://localhost:8080/api/v1/webhooks/github-pr \
  -H "X-Webhook-Secret: webhook-secret" \
  -d '{"pull_request": {"url": "https://github.com/..."}}'
```

---

## Getting Help

### Documentation
- **Quick Start**: [QUICKSTART.md](../QUICKSTART.md)
- **Architecture**: [ARCHITECTURE.md](ARCHITECTURE.md)
- **Deployment**: [DEPLOYMENT.md](DEPLOYMENT.md)
- **API Docs**: [API_REFERENCE.md](API_REFERENCE.md)

### Community
- **Discord**: [Join chat](https://discord.gg/Unagnt)
- **GitHub**: [Issues & discussions](https://github.com/NikoSokratous/unagnt)
- **Blog**: [blog.Unagnt.io](https://blog.Unagnt.io)

### Support
- **Email**: [support@Unagnt.io](mailto:support@Unagnt.io)
- **Commercial**: [contact@Unagnt.io](mailto:contact@Unagnt.io)

---

**Happy building!** 🚀
