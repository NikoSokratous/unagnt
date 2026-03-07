# Unagnt Architecture

**Last Updated**: v2.0.0 | March 2026

This document provides a comprehensive overview of Unagnt's architecture, design principles, and implementation details.

---

## Table of Contents

1. [Overview](#overview)
2. [Design Principles](#design-principles)
3. [System Architecture](#system-architecture)
4. [Core Components](#core-components)
5. [Data Flow](#data-flow)
6. [Deployment Architecture](#deployment-architecture)
7. [Security Architecture](#security-architecture)
8. [Scalability](#scalability)

---

## Overview

Unagnt is a **layered, microservices-ready platform** for orchestrating autonomous AI agents at scale. It follows clean architecture principles with clear separation of concerns.

### Key Characteristics

- **State-Machine Based**: Deterministic execution with explicit state transitions
- **Policy-Driven**: Declarative governance with runtime enforcement
- **Observable**: Comprehensive tracing, metrics, and audit logs
- **Extensible**: Plugin architecture for tools, LLMs, and memory backends
- **Cloud-Native**: Kubernetes-native with Helm charts and operators

---

## Design Principles

### 1. Security by Default
- All tool executions require explicit permissions
- Policy enforcement at runtime
- Encrypted audit logs
- No implicit trust

### 2. Observability First
- Every action is traced and logged
- Deterministic replay for debugging
- Cost attribution per agent/tenant
- Real-time metrics

### 3. Fail-Safe Design
- Human-in-the-loop for high-risk actions
- Graceful degradation
- Automatic rollback on policy violations
- Circuit breakers for external services

### 4. Developer Experience
- CLI-first design
- Visual workflow designer
- Comprehensive SDKs
- Hot-reloadable plugins

### 5. Production-Ready
- Multi-tenancy with namespace isolation
- RBAC and OAuth2/OIDC
- Auto-scaling and load balancing
- High availability

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Client Layer                              │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Web UI     │  │ CLI (unagnt)│  │  SDKs       │          │
│  │   (React)    │  │    (Go)      │  │ (Go/Python) │          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
│         │                  │                  │                  │
│         └──────────────────┼──────────────────┘                  │
└─────────────────────────────┼────────────────────────────────────┘
                              │
┌─────────────────────────────┼────────────────────────────────────┐
│                        API Gateway                                │
├─────────────────────────────┼────────────────────────────────────┤
│  ┌──────────────────────────┴───────────────────────────┐        │
│  │  REST API          GraphQL API        SSE Streaming  │        │
│  │  (Gin)             (graphql-go)       (Real-time)    │        │
│  └──────────────────────────┬───────────────────────────┘        │
│                              │                                    │
│  ┌──────────────────────────┴───────────────────────────┐        │
│  │         Authentication & Authorization                │        │
│  │   - Bearer Token / API Key                           │        │
│  │   - OAuth2 / OIDC / SAML 2.0 (Okta, Azure AD, etc.)  │        │
│  │   - Advanced RBAC (templates, org hierarchy)         │        │
│  └──────────────────────────┬───────────────────────────┘        │
└─────────────────────────────┼────────────────────────────────────┘
                              │
┌─────────────────────────────┼────────────────────────────────────┐
│                    Orchestration Layer                            │
├─────────────────────────────┼────────────────────────────────────┤
│  ┌──────────────────────────┴───────────────────────────┐        │
│  │              Workflow Engine                          │        │
│  │  - DAG Execution (Sequential, Parallel, Conditional) │        │
│  │  - CEL Expression Evaluation                         │        │
│  │  - Template Rendering (Go templates)                 │        │
│  │  - Agent Delegation & Coordination                   │        │
│  └──────────────────────────┬───────────────────────────┘        │
│                              │                                    │
│  ┌──────────────────────────┴───────────────────────────┐        │
│  │              Policy Engine                            │        │
│  │  - CEL-based rule evaluation                         │        │
│  │  - Risk scoring (7 categories)                       │        │
│  │  - Human-in-the-loop gates                           │        │
│  │  - Policy versioning & simulation                    │        │
│  └──────────────────────────┬───────────────────────────┘        │
│                              │                                    │
│  ┌──────────────────────────┴───────────────────────────┐        │
│  │         Cost & SLA Tracking                           │        │
│  │  - Provider pricing integration                       │        │
│  │  - Per-agent/tenant attribution                       │        │
│  │  - Latency, uptime, error tracking                   │        │
│  └──────────────────────────┬───────────────────────────┘        │
└─────────────────────────────┼────────────────────────────────────┘
                              │
┌─────────────────────────────┼────────────────────────────────────┐
│                      Agent Runtime Core                           │
├─────────────────────────────┼────────────────────────────────────┤
│  ┌──────────────────────────┴───────────────────────────┐        │
│  │              State Machine                            │        │
│  │  States: INIT → PLANNING → EXECUTING → EVAL →       │        │
│  │          REPLANNING → COMPLETE / FAILED / INTERRUPTED│        │
│  │  - Deterministic transitions                         │        │
│  │  - Interrupt handling                                │        │
│  │  - State persistence                                 │        │
│  └──────────────────────────┬───────────────────────────┘        │
│                              │                                    │
│  ┌──────────────────────────┴───────────────────────────┐        │
│  │              Planner                                  │        │
│  │  - Goal decomposition                                │        │
│  │  - Context preparation                               │        │
│  │  - LLM prompt construction                           │        │
│  └──────────────────────────┬───────────────────────────┘        │
│                              │                                    │
│  ┌──────────────────────────┴───────────────────────────┐        │
│  │              Executor                                 │        │
│  │  - Tool validation & execution                       │        │
│  │  - Permission checks                                 │        │
│  │  - Result extraction                                 │        │
│  │  - Error handling                                    │        │
│  └──────────────────────────┬───────────────────────────┘        │
│                              │                                    │
│  ┌──────────────────────────┴───────────────────────────┐        │
│  │              Memory Manager                           │        │
│  │  - Working Memory (in-memory)                        │        │
│  │  - Persistent Memory (K/V store)                     │        │
│  │  - Semantic Memory (vector search)                   │        │
│  │  - Event Log (immutable)                             │        │
│  └──────────────────────────┬───────────────────────────┘        │
└─────────────────────────────┼────────────────────────────────────┘
                              │
┌─────────────────────────────┼────────────────────────────────────┐
│                    Integration Layer                              │
├─────────────────────────────┼────────────────────────────────────┤
│  ┌──────────────┬───────────┴─────────┬──────────────┐          │
│  │ LLM Provider │   Tool Registry     │  Memory      │          │
│  │  Adapters    │   + MCP support     │  Backends    │          │
│  ├──────────────┤──────────────────────┼──────────────┤          │
│  │ - OpenAI     │ - Versioned         │ - SQLite     │          │
│  │ - Anthropic  │ - Schema-validated  │ - PostgreSQL │          │
│  │ - Ollama     │ - Permission-gated  │ - Redis      │          │
│  │ - Custom     │ - Go Plugins (.so)  │ - Qdrant     │          │
│  │              │ - WASM, MCP         │ - Weaviate   │          │
│  └──────────────┴─────────────────────┴──────────────┘          │
└─────────────────────────────┼────────────────────────────────────┘
                              │
┌─────────────────────────────┼────────────────────────────────────┐
│                    Observability Layer                            │
├─────────────────────────────┼────────────────────────────────────┤
│  ┌──────────────────────────┴───────────────────────────┐        │
│  │  Tracing (OpenTelemetry)                             │        │
│  │  - OTLP, Zipkin, Jaeger exporters                    │        │
│  │  - Distributed trace propagation                     │        │
│  └──────────────────────────┬───────────────────────────┘        │
│                              │                                    │
│  ┌──────────────────────────┴───────────────────────────┐        │
│  │  Metrics (Prometheus)                                │        │
│  │  - Run counts, durations, errors                     │        │
│  │  - Tool execution metrics                            │        │
│  │  - Policy denial rates                               │        │
│  └──────────────────────────┬───────────────────────────┘        │
│                              │                                    │
│  ┌──────────────────────────┴───────────────────────────┐        │
│  │  Logging (Zerolog)                                   │        │
│  │  - Structured JSON logs                              │        │
│  │  - Encrypted audit logs                              │        │
│  │  - Deterministic replay recording                    │        │
│  └──────────────────────────────────────────────────────┘        │
└───────────────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. Runtime Engine

**Location**: `pkg/runtime/`

**Responsibilities**:
- State machine execution
- Agent lifecycle management
- Interruption handling
- Replay orchestration

**Key Types**:
```go
type Engine struct {
    config    *Config
    planner   Planner
    executor  Executor
    state     *State
}

type State struct {
    Current      StateType
    Steps        int
    Outputs      map[string]interface{}
    Memory       Memory
    ExecutionLog []Event
}
```

### 2. Workflow Orchestrator

**Location**: `pkg/orchestrate/`

**Responsibilities**:
- DAG workflow execution
- Agent coordination
- Conditional branching (CEL)
- Template rendering

**Execution Modes**:
- **Sequential**: Steps run in order
- **Parallel**: Steps run concurrently
- **Conditional**: CEL-based branching

**Example**:
```go
workflow := &Workflow{
    Steps: []Step{
        {Name: "research", Agent: "researcher", Parallel: false},
        {Name: "analyze", Agent: "analyzer", Condition: "step.research.success"},
        {Name: "report", Agent: "writer"},
    },
}
```

### 3. Policy Engine

**Location**: `pkg/policy/`

**Responsibilities**:
- Rule evaluation (CEL)
- Risk scoring
- Human-in-the-loop gates
- Policy versioning

**Policy Evaluation Flow**:
```
Tool Execution Request
    ↓
Policy Lookup (by environment)
    ↓
CEL Evaluation (all rules)
    ↓
Risk Scoring (7 categories)
    ↓
Action: ALLOW / DENY / REQUIRE_APPROVAL
    ↓
If REQUIRE_APPROVAL → Webhook
    ↓
Result: Execute or Block
```

### 4. Tool Registry

**Location**: `pkg/tool/`

**Responsibilities**:
- Tool discovery and loading
- Schema validation
- Permission enforcement
- Versioning

**Tool Types**:
- **Built-in**: Echo, Calculator, HTTP Request
- **Go Plugins**: `.so` files with hot-reload
- **WASM Modules**: Portable, sandboxed
- **External**: HTTP/gRPC services

### 5. Memory Manager

**Location**: `pkg/memory/`

**Memory Tiers**:

| Tier | Backend | Use Case | Persistence |
|------|---------|----------|-------------|
| **Working** | In-memory map | Current conversation | Session only |
| **Persistent** | SQLite/PostgreSQL | Facts, preferences | Permanent |
| **Semantic** | HNSW / Qdrant | Similarity search | Permanent |
| **Event Log** | Append-only DB | Audit, replay | Immutable |

### 6. MCP (Model Context Protocol)

**Location**: `pkg/mcp/`

**Responsibilities**:
- Connect agents to MCP-compatible tools and data sources
- Standard protocol for model context exchange
- See [pkg/mcp/README.md](../pkg/mcp/README.md)

### 7. LLM Integration

**Location**: `pkg/llm/`

**Supported Providers**:
- **OpenAI**: GPT-3.5, GPT-4, GPT-4o
- **Anthropic**: Claude 3 (Sonnet, Opus, Haiku)
- **Ollama**: Llama, Mistral, etc.
- **Custom**: Implement `LLMProvider` interface

**Unified Interface**:
```go
type LLMProvider interface {
    Complete(ctx context.Context, messages []Message, tools []Tool) (*Response, error)
    Stream(ctx context.Context, messages []Message, tools []Tool) (<-chan Response, error)
}
```

---

## Data Flow

### Agent Execution Flow

```
1. Client Request
   ├─ HTTP POST /api/v1/agents/:id/run
   └─ unagnt agent run <name>
       ↓
2. Authentication & Authorization
   ├─ Validate Bearer token / API key
   ├─ Check RBAC permissions
   └─ Resolve tenant namespace
       ↓
3. Runtime Initialization
   ├─ Load agent config
   ├─ Load policies
   ├─ Initialize memory
   └─ Create run record
       ↓
4. State Machine Loop
   ┌──────────────────────────┐
   │ INIT                     │
   │ - Load context           │
   │ - Prepare tools          │
   └──────────┬───────────────┘
              ↓
   ┌──────────────────────────┐
   │ PLANNING                 │
   │ - Call LLM               │
   │ - Parse tool calls       │
   └──────────┬───────────────┘
              ↓
   ┌──────────────────────────┐
   │ EXECUTING                │
   │ - Validate tool schema   │
   │ - Check permissions      │
   │ - Evaluate policy        │
   │ - Execute tool           │
   └──────────┬───────────────┘
              ↓
   ┌──────────────────────────┐
   │ EVAL                     │
   │ - Check goal completion  │
   │ - Update memory          │
   │ - Increment step counter │
   └──────────┬───────────────┘
              ↓
   Decision: Complete? Max steps?
   ├─ Yes → COMPLETE
   └─ No  → REPLANNING → PLANNING
       ↓
5. Finalization
   ├─ Save final state
   ├─ Emit metrics
   ├─ Write audit log
   └─ Return response
```

### Workflow Execution Flow

```
1. Workflow Definition (YAML/JSON)
   ↓
2. Parse & Validate
   ├─ Schema validation
   ├─ DAG cycle detection
   └─ Parameter validation
       ↓
3. Topological Sort (for dependencies)
   ↓
4. Execute Steps
   ├─ Sequential: One at a time
   ├─ Parallel: Concurrently with sync
   └─ Conditional: CEL evaluation
       ↓
5. For Each Step:
   ├─ Render goal template
   ├─ Create Agent Engine
   ├─ Execute agent
   ├─ Collect outputs
   └─ Pass to next step
       ↓
6. Aggregate Results
   └─ Return workflow result
```

---

## Deployment Architecture

### Kubernetes Deployment

```
┌─────────────────────────────────────────────────────────────┐
│                       Kubernetes Cluster                     │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Ingress (Nginx / Istio Gateway)                       │ │
│  │  - TLS termination                                     │ │
│  │  - Rate limiting                                       │ │
│  └────────────────────────┬───────────────────────────────┘ │
│                            │                                 │
│  ┌────────────────────────┴───────────────────────────────┐ │
│  │  Unagnt Pods (Deployment)                        │ │
│  │  - API Server                                          │ │
│  │  - Workflow Engine                                     │ │
│  │  - Policy Engine                                       │ │
│  │  - HPA: 2-10 replicas                                  │ │
│  └────────────────────────┬───────────────────────────────┘ │
│                            │                                 │
│  ┌────────────────────────┴───────────────────────────────┐ │
│  │  Unagnt Operator (StatefulSet)                   │ │
│  │  - Watches Agent/Workflow CRDs                         │ │
│  │  - Reconciliation loop                                 │ │
│  │  - Scheduled workflow execution                        │ │
│  └────────────────────────┬───────────────────────────────┘ │
│                            │                                 │
│  ┌────────────────────────┴───────────────────────────────┐ │
│  │  Web UI (Deployment)                                   │ │
│  │  - React SPA                                           │ │
│  │  - Nginx server                                        │ │
│  │  - HPA: 2-5 replicas                                   │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  PostgreSQL (StatefulSet)                            │   │
│  │  - PVC: 50Gi                                         │   │
│  │  - Backups: Velero / pg_dump                         │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Redis (StatefulSet)                                 │   │
│  │  - Rate limiting, caching                            │   │
│  │  - PVC: 10Gi                                         │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Qdrant (StatefulSet)                                │   │
│  │  - Vector storage                                    │   │
│  │  - PVC: 100Gi                                        │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Monitoring Stack                                    │   │
│  │  - Prometheus (scraping)                             │   │
│  │  - Grafana (dashboards)                              │   │
│  │  - Jaeger (tracing)                                  │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

**Key Resources**:
- **Custom Resources**: `Agent`, `Workflow`, `Policy`
- **ConfigMaps**: Policies, LLM configs
- **Secrets**: API keys, DB credentials (or Vault/external backends)
- **Services**: ClusterIP, LoadBalancer
- **PVCs**: PostgreSQL, Redis, Qdrant

### Air-Gapped Deployment

For fully disconnected environments, use `scripts/offline-install.sh bundle` to create a tarball with binaries and configs (including `configs/compliance/`). Transfer to the air-gapped environment and run `install.sh`. Local LLMs (e.g. Ollama) supported. See [deploy/air-gapped/README.md](../deploy/air-gapped/README.md).

### Compliance and SIEM

- **Compliance Pack**: SOC2 and HIPAA configs in `configs/compliance/`
- **SIEM Export**: `GET /v1/compliance/audit/export?format=json|csv|cef` for audit log export to SIEM systems

---

## Security Architecture

### Defense in Depth

```
┌─────────────────────────────────────────────────────────────┐
│  Layer 1: Network                                            │
│  - Ingress with TLS                                         │
│  - Service Mesh (mTLS between services)                     │
│  - Network Policies (pod isolation)                         │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Layer 2: Authentication                                     │
│  - Bearer Token / API Key                                   │
│  - OAuth2 / OIDC / SAML 2.0 (Okta, Azure AD, OneLogin)      │
│  - JWT validation                                           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Layer 3: Authorization                                      │
│  - RBAC (role templates, org hierarchy, delegation)         │
│  - Tenant namespace isolation                               │
│  - Resource ownership validation                            │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Layer 4: Policy Enforcement                                 │
│  - CEL-based rules                                          │
│  - Risk scoring                                             │
│  - Human-in-the-loop gates                                  │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Layer 5: Tool Permissions                                   │
│  - Schema validation                                        │
│  - Permission scopes (fs:read, net:external, etc.)          │
│  - Sandboxed execution (WASM)                               │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Layer 6: Audit & Monitoring                                 │
│  - Encrypted audit logs (KMS)                               │
│  - Anomaly detection                                        │
│  - Alert on suspicious activity                             │
└─────────────────────────────────────────────────────────────┘
```

### Secrets Management

- **Kubernetes Secrets**: Encrypted at rest
- **HashiCorp Vault**: KV v2 backend, secret references (`secret:ref:path/to/secret`)
- **AWS/GCP Secrets Manager**: Stubs for future integration
- **External KMS**: AWS KMS, GCP KMS for key encryption
- **Key Rotation**: Automated with zero downtime
- **Audit Logs**: Encrypted with separate keys

See [Secrets Management](guides/secrets-management.md).

---

## Scalability

### Horizontal Scaling

**Stateless Components** (can scale freely):
- API servers
- Web UI
- Workflow workers

**Stateful Components** (careful scaling):
- PostgreSQL: Read replicas
- Redis: Cluster mode
- Qdrant: Sharding

### Performance Optimizations

1. **Caching**
   - Policy cache (in-memory)
   - Tool schema cache
   - LLM response cache (Redis)

2. **Connection Pooling**
   - Database connections
   - HTTP clients
   - Redis connections

3. **Async Processing**
   - Background workflows
   - Webhook handlers
   - Event processing

4. **Rate Limiting**
   - Per-tenant quotas
   - LLM provider rate limits
   - API rate limiting

### Load Testing Results

| Metric | Single Instance | 3 Instances | 10 Instances |
|--------|----------------|-------------|--------------|
| **Req/sec** | 100 | 280 | 900 |
| **P95 Latency** | 250ms | 180ms | 120ms |
| **Concurrent Agents** | 50 | 150 | 500 |

---

## Technology Stack

### Backend
- **Language**: Go 1.22+
- **Web Framework**: Gin
- **GraphQL**: graphql-go
- **ORM**: database/sql (stdlib)
- **Observability**: OpenTelemetry, Zerolog, Prometheus

### Frontend
- **Framework**: React 18
- **Workflow Designer**: React Flow
- **State Management**: Zustand
- **GraphQL Client**: Apollo Client
- **Charts**: D3.js

### Data Stores
- **Relational**: PostgreSQL 14+, SQLite 3
- **Cache**: Redis 7+
- **Vector**: Qdrant, Weaviate
- **Tracing**: Jaeger, Zipkin

### Infrastructure
- **Container**: Docker
- **Orchestration**: Kubernetes 1.25+
- **Service Mesh**: Istio, Linkerd
- **CI/CD**: GitHub Actions

---

## Known Limitations & Implementation Notes

### Workflow Execution

The workflow engine uses a pluggable `StepExecutor` interface and now defaults to a runtime-backed executor. A background runner queue executes API/webhook/scheduled/event-triggered runs asynchronously. `SimulatedExecutor` remains available when explicitly selected (design-time/testing). Workflow checkpoints and resume are backed by `workflow_states` and `workflow_step_states`.

### Phase 1 Verification Checklist (v3)

Use this checklist before release or when changing runtime/orchestration internals.

- **Runtime-backed execution default**
  - Confirm `WorkflowEngine` defaults to `RuntimeStepExecutor` (no simulated default path in production).
- **Async runner path active**
  - Submit `POST /v1/runs` and verify run lifecycle transitions (`pending -> running -> completed/failed`) through store and API.
- **Webhook execution wiring**
  - Trigger a configured webhook endpoint and verify it queues a `RunRequest`, executes via runner, and emits callback payload with output/error.
- **Scheduler wiring**
  - Register a cron job and verify scheduled runs are submitted through runner, not direct simulated execution.
- **Event trigger wiring**
  - Publish `POST /v1/triggers/events` and verify accepted events produce queued runs.
- **Checkpoint + resume correctness**
  - Execute a workflow with `--db`, confirm checkpoint rows in `workflow_states` and `workflow_step_states`, then resume with `unagnt workflow run <file> --resume <workflow-id> --db <path>`.
  - Ensure node state rows are workflow-scoped (`workflowID:nodeID`) so workflows with identical node names do not collide.
- **Regression tests**
  - Run `go test ./pkg/workflow ./pkg/orchestrate`.
  - Run full suite with integration-safe env: `OPENAI_API_KEY="" go test ./...`.

### Phase 3 Runtime Hardening Runbook (v3)

Use this checklist for production hardening validation and incident response drills.

- **Retry/timeout controls**
  - Submit `POST /v1/runs` with `max_retries`, `retry_backoff_ms`, and `timeout_ms`.
  - Verify expected lifecycle outcomes (`completed`, `failed`, `interrupted`) and retry behavior under transient failures.
- **Backpressure behavior**
  - Saturate runner queue and verify rejected submissions return `503` and increment queue rejection metrics.
  - Confirm queue depth trends using `agentruntime_run_queue_depth`.
- **Dead-letter capture**
  - Force terminal failures and verify entries are persisted in dead-letter storage and exposed via `GET /v1/runs/dead-letters`.
- **Dead-letter replay**
  - Replay a failed run using `POST /v1/runs/dead-letters/{id}/replay`.
  - Validate optional overrides (`goal`, `payload`, `max_retries`, `retry_backoff_ms`, `timeout_ms`) are applied.
- **Ops metrics baseline**
  - Track and alert on:
    - `agentruntime_run_retries_total`
    - `agentruntime_run_dead_letters_total`
    - `agentruntime_run_queue_depth`
    - `agentruntime_run_queue_rejected_total`
    - `agentruntime_run_failures_total{reason,source}`

### Phase 3 Migration Notes (v3)

- **API request updates**
  - `POST /v1/runs` accepts optional hardening controls:
    - `max_retries`
    - `retry_backoff_ms`
    - `timeout_ms`
  - Values must be non-negative; invalid values return `400`.
- **New diagnostics endpoints**
  - `GET /v1/runs/{id}/events` for persisted run event history.
  - `GET /v1/runs/dead-letters` with optional `limit` and `source` query filters.
  - `POST /v1/runs/dead-letters/{id}/replay` to requeue failed runs with optional overrides.
- **Operational behavior changes**
  - Runner now supports bounded retries/backoff and per-attempt timeout.
  - Terminal failures are persisted to dead-letter storage.
  - Retry backoff is cancellation-aware (cancel during backoff interrupts run quickly).
- **Metrics additions**
  - `agentruntime_run_retries_total`
  - `agentruntime_run_dead_letters_total`
  - `agentruntime_run_queue_depth`
  - `agentruntime_run_queue_rejected_total`
  - `agentruntime_run_failures_total{reason,source}`

### Kubernetes Operator

Use `make build-operator` to build the operator; it runs `generate-operator` and `generate-crds` first. CRDs are generated and committed to `k8s/operator/config/crd/bases/` and `k8s/crds/`. See [k8s/operator/BUILD_NOTES.md](../k8s/operator/BUILD_NOTES.md).

### Remaining Stubs

- **Policy tests**: SQLite-backed policy tests require CGO; may fail on Windows with CGO disabled.
- **RBAC/Cost**: Some cost-exceeded checks and RBAC fallbacks use simplified implementations for non-enterprise paths.

---

## Further Reading

- **[Deployment Guide](DEPLOYMENT.md)** - Production deployment patterns
- **[Air-Gapped Deployment](../deploy/air-gapped/README.md)** - Offline / disconnected deployment
- **[Security Guide](SECURITY.md)** - Hardening and best practices
- **[Enterprise SSO](guides/enterprise-sso.md)** - SAML, OIDC configuration
- **[Secrets Management](guides/secrets-management.md)** - Vault, AWS, GCP backends
- **[Advanced RBAC](guides/advanced-rbac.md)** - Role templates, delegation
- **[Compliance Pack](../configs/compliance/README.md)** - SOC2, HIPAA, SIEM export
- **[Plugin Development](PLUGIN_DEVELOPMENT.md)** - Creating custom tools
- **[API Reference](API_REFERENCE.md)** - REST and GraphQL APIs
- **[SUPPORT.md](../SUPPORT.md)** - Support tiers and contact

---

**Questions?** Open an issue or join our [Discord](https://discord.gg/Unagnt)!
