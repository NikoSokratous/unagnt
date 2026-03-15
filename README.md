# Unagnt

<div align="center">

**Production-grade runtime for autonomous AI agents**

[![CI](https://github.com/NikoSokratous/unagnt/actions/workflows/ci.yml/badge.svg)](https://github.com/NikoSokratous/unagnt/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/NikoSokratous/unagnt)](https://goreportcard.com/report/github.com/NikoSokratous/unagnt)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/NikoSokratous/unagnt)](https://github.com/NikoSokratous/unagnt/releases)

[Features](#-features) •
[Quick Start](#-quick-start) •
[Documentation](#-documentation) •
[Examples](#-examples) •
[Architecture](#-architecture) •
[Contributing](#-contributing)

</div>

---

## What is Unagnt?

Unagnt is an **enterprise-grade orchestration platform** for autonomous AI agents. Unlike chatbots or LLM wrappers, Unagnt provides the **infrastructure** needed to run AI agents safely and reliably in production.

### Why Unagnt? (vs LangGraph, Temporal, CrewAI)

Unagnt is Go-native, ships as a single binary, and runs with zero external infrastructure by default (SQLite + in-memory). You own your stack: no SaaS lock-in, air-gapped support, and policy enforcement built-in. LangGraph and CrewAI are Python-centric; Temporal is a general workflow engine. Unagnt focuses on AI agents with policy, audit, and cost controls out of the box.

### Feature Status

| Feature | Status |
|---------|--------|
| Single-agent run (unagnt run) | Working |
| Policy engine | Working |
| DAG workflow (runtime-backed + checkpoint/resume) | Working |
| Cost tracking | Partial |
| Real workflow execution | Working |
| Visual designer | UI exists |
| Kubernetes operator | Requires make generate |

### Highlights

✅ **Production-Ready**: Policy enforcement, audit logs, and security by default  
✅ **Observable**: Full tracing, metrics, and replay capabilities  
✅ **Scalable**: Kubernetes-native with auto-scaling and multi-tenancy  
✅ **Developer-Friendly**: CLI tools, SDKs, visual workflow designer  
✅ **Provider-Agnostic**: OpenAI, Anthropic, Ollama, or any LLM  

---

## 🎯 Key Features

### 🔐 Security & Governance
- **Policy Engine**: YAML-based policies with CEL expressions
- **Risk Scoring**: Real-time risk assessment across 7 categories
- **Audit Logs**: Encrypted, tamper-proof execution logs; SIEM export (JSON/CSV/CEF)
- **RBAC**: Role-based access control; SSO/SAML/OIDC; advanced RBAC (templates, delegation)
- **Secrets**: Vault, AWS/GCP backends; `secret:ref:` in config

### 🔄 Workflow Orchestration
- **Visual Designer**: Drag-and-drop workflow builder
- **DAG Execution**: Parallel agent coordination
- **Durable Execution**: Checkpoints after each level; resume from last checkpoint after crash or restart
- **Conditional Logic**: CEL-based branching
- **Approval Steps**: Native human-in-the-loop steps in workflows (pause for approval, then continue)
- **Template Marketplace**: Pre-built workflows for common tasks

### 📊 Observability
- **Distributed Tracing**: OpenTelemetry integration
- **Real-time Metrics**: Prometheus-compatible metrics
- **Cost Attribution**: Per-agent/tenant cost tracking
- **SLA Monitoring**: Uptime, latency, and error tracking

### 🚀 Enterprise Scale
- **Kubernetes Operator**: Custom resources for agents and workflows
- **Helm Charts**: Production-ready deployments
- **Air-Gapped**: Offline bundle, local LLMs (Ollama), compliance configs included
- **Compliance Pack**: SOC2, HIPAA configs; SIEM audit export
- **Service Mesh**: Istio/Linkerd integration with mTLS
- **Auto-Scaling**: Queue depth and custom metrics-based scaling

### 🧠 Agent Capabilities
- **Memory Systems**: Working, persistent, semantic, and event log
- **Context Assembly**: Automatic retrieval of policies, memory, workflow state, and knowledge
- **RAG (Retrieval Augmented Generation)**: Ingest documents, semantic search, ground responses in your knowledge base
- **Embeddings**: OpenAI and local (sentence-transformers) for semantic memory and RAG
- **Local Storage**: SQLite for persistent memory; in-memory vector store for development
- **Tool Framework**: Schema-validated, versioned, permission-controlled; MCP support
- **Deterministic Replay**: 5 replay modes for debugging
- **Multi-Agent**: Task delegation and collaboration

---

## 🚀 Quick Start

### Installation (Zero Dependencies)

```bash
# One-liner: install and run
go install github.com/NikoSokratous/unagnt/cmd/unagnt@latest
unagnt run --config examples/cli-assistant/agent.yaml --goal "Echo hello world"
```

Zero dependencies: SQLite + in-memory by default. No Postgres, Redis, or Qdrant required.

### Your First Agent

```bash
# Set your API key
export OPENAI_API_KEY=sk-...

# Run the CLI assistant example
unagnt run --config examples/cli-assistant/agent.yaml --goal "Echo hello world"
```

### Optional: Model Routing (v3)

Use `model_routing` to let runtime pick a model by strategy (`auto`, `cost`, `latency`, `capability`):

```yaml
# examples/cli-assistant/agent.yaml
name: cli-assistant
model:
  provider: openai
  name: gpt-4o

model_routing:
  enabled: true
  strategy: auto
  candidates:
    - provider: openai
      name: gpt-4o-mini
    - provider: openai
      name: gpt-4o
```

👉 **Try the full walkthrough in 2 minutes**: [docs/E2E_EXAMPLE.md](docs/E2E_EXAMPLE.md)

### Build from Source

```bash
git clone https://github.com/NikoSokratous/unagnt.git
cd unagnt
make build
./bin/unagntd   # Start the server
```

To build the Kubernetes operator: `make build-operator` (requires controller-gen; see [k8s/operator/BUILD_NOTES.md](k8s/operator/BUILD_NOTES.md)).

### Visual Workflow Designer

```bash
# Start the web UI
cd web && npm install && npm run dev

# Open http://localhost:3000
# Navigate to Workflow Designer to build workflows visually
```

👉 **See [QUICKSTART.md](QUICKSTART.md) for detailed getting started guide**

---

## 📚 Documentation

### For Users
- **[Quick Start Guide](QUICKSTART.md)** - Get up and running in 5 minutes
- **[User Guide](docs/USER_GUIDE.md)** - Complete feature walkthrough
- **[Embeddings & RAG](docs/EMBEDDINGS.md)** - Semantic search and knowledge-base setup
- **[Context Assembly](docs/CONTEXT_ASSEMBLY.md)** - How context is built for each LLM call
- **[CLI Reference](docs/CLI_REFERENCE.md)** - All unagnt commands
- **[API Reference](docs/API_REFERENCE.md)** - REST and GraphQL APIs
- **[Workflow Guide](examples/workflows/templates/README.md)** - Building workflows

### For Developers
- **[Architecture](docs/ARCHITECTURE.md)** - System design and components
- **[Development Guide](docs/DEVELOPMENT.md)** - Setup and contribution guide
- **[Plugin Development](docs/PLUGIN_DEVELOPMENT.md)** - Creating custom tools
- **[API Integration](docs/API_INTEGRATION.md)** - SDK usage and examples

### For Operators
- **[Release Notes v3.0](RELEASE_NOTES_V3.0.md)** - v3 runtime hardening and orchestration rollout
- **[Release Notes v2.0](RELEASE_NOTES_V2.0.md)** - v1.1 through v2.0 changelog
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Kubernetes, Docker, bare metal
- **[Air-Gapped Deployment](deploy/air-gapped/README.md)** - Offline / disconnected install
- **[Enterprise SSO](docs/guides/enterprise-sso.md)** - SAML, OIDC configuration
- **[Secrets Management](docs/guides/secrets-management.md)** - Vault, external backends
- **[Configuration](docs/CONFIGURATION.md)** - All configuration options
- **[Security Guide](docs/SECURITY.md)** - Best practices and hardening
- **[Monitoring](docs/MONITORING.md)** - Observability setup

---

## 🎨 Examples

### Code Review Bot
```yaml
# examples/workflows/code-review.yaml
name: "Automated Code Review"
agents:
  - linter-agent: Static analysis
  - security-agent: Vulnerability scanning
  - complexity-agent: Maintainability check
workflow:
  - fetch-code → static-analysis → security-scan → generate-report
```

### Data Pipeline
```yaml
# examples/workflows/data-pipeline.yaml
name: "ETL Pipeline"
agents:
  - extractor: Pull from API/DB
  - validator: Schema validation
  - transformer: Data cleaning
  - loader: Write to warehouse
```

### Research Assistant
```yaml
# examples/workflows/research.yaml
name: "Multi-Source Research"
parallel:
  - arxiv-search
  - web-search
  - pubmed-search
workflow:
  - aggregate → rank → analyze → synthesize → report
```

### Support Agent with RAG
```yaml
# examples/knowledge-base/agent.yaml
name: support_agent_with_rag
context_assembly:
  enabled: true
  embeddings:
    provider: openai
    model: text-embedding-3-small
  providers:
    - type: semantic_memory
    - type: knowledge
      config:
        sources: ["./docs"]
        top_k: 3
```
Ingest docs, then run: `unagnt context ingest ./docs`

👉 **See [examples/](examples/) for 20+ ready-to-use examples**

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Unagnt Platform                   │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Web UI       │  │ CLI (unagnt)│  │ REST/GraphQL │      │
│  │ React + Flow │  │ Go            │  │ API          │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│         │                  │                  │              │
│         └──────────────────┼──────────────────┘              │
│                            │                                 │
│  ┌────────────────────────┴────────────────────────┐        │
│  │  Auth: OAuth2 / OIDC / SAML • Advanced RBAC     │        │
│  └────────────────────────┬────────────────────────┘        │
│                            │                                 │
│         ┌──────────────────┴──────────────────┐              │
│         │     Orchestration Layer             │              │
│         │  - Workflow Engine (DAG)            │              │
│         │  - Policy Enforcement               │              │
│         │  - Cost & Budget Tracking           │              │
│         │  - Compliance / SIEM Export         │              │
│         └──────────────────┬──────────────────┘              │
│                            │                                 │
│         ┌──────────────────┴──────────────────┐              │
│         │     Agent Runtime                   │              │
│         │  - State Machine                    │              │
│         │  - Tool Executor + MCP              │              │
│         │  - Memory Manager                   │              │
│         │  - LLM Integration                  │              │
│         └──────────────────┬──────────────────┘              │
│                            │                                 │
│  Default: SQLite + in-memory (zero external deps)            │
│  Production (opt-in): PostgreSQL, Redis, Qdrant, Prometheus, Jaeger  │
│  Vault / Secrets • Air-Gapped • Compliance Pack              │
└─────────────────────────────────────────────────────────────┘
```

**Key Components:**
- **Runtime Engine**: State machine-based agent execution
- **Workflow Orchestrator**: DAG-based multi-agent coordination  
- **Policy Engine**: CEL-based policy enforcement
- **Memory Layer**: Multi-tier memory architecture
- **Tool Registry**: Versioned tools, MCP support
- **Auth**: OAuth2, OIDC, SAML 2.0, advanced RBAC
- **Compliance**: SOC2/HIPAA configs, SIEM export API
- **Deployment**: Kubernetes, air-gapped, local LLMs (Ollama)
- **Observability**: Tracing, metrics, logs, and replay

👉 **See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for deep dive**

---

## 🌟 Showcase

### Workflow Designer
![Workflow Designer](docs/images/workflow-designer.png)
*Visual drag-and-drop workflow builder with real-time validation*

### Analytics Dashboard
![Analytics](docs/images/analytics.png)
*Cost tracking, performance metrics, and SLA monitoring*

### Marketplace
![Marketplace](docs/images/marketplace.png)
*Pre-built workflow templates for common use cases*

---

## 📈 Use Cases

- **🤖 Autonomous Agents**: Production agents with safety guarantees
- **📊 Data Pipelines**: ETL workflows with AI-powered transformation
- **🔍 Code Review**: Automated PR reviews with multiple checks
- **📝 Content Generation**: SEO-optimized content creation at scale
- **💬 Customer Support**: Intelligent ticket routing and responses
- **🔬 Research**: Multi-source research with synthesis
- **💡 Support with RAG**: Ground agents in docs; ingest markdown/text, semantic search, grounded responses
- **🚀 DevOps**: Deployment automation with testing and rollback

---

## 📊 Project Stats

- **Language**: Go 1.22
- **LOC**: ~50,000 lines of production code
- **Tests**: 500+ unit and integration tests
- **Coverage**: 75%+
- **Dependencies**: Carefully curated, minimal external deps
- **License**: MIT

---

## 🗺️ Roadmap

### ✅ v1.0 (Complete)
- [x] Visual workflow designer
- [x] Kubernetes operator
- [x] Cost attribution and SLA monitoring
- [x] Audit log encryption
- [x] GraphQL API
- [x] ML model registry integration
- [x] RAG with knowledge base and document ingestion
- [x] Context assembly with embeddings (OpenAI + local)
- [x] Semantic memory search and local storage (SQLite, in-memory)

### ✅ v1.1: Pitch Readiness (Complete)
- [x] Safety-First Demo (showcase/safety-first-demo/)
- [x] Showcase App: Enterprise Compliance Bot (showcase/enterprise-compliance-bot/)
- [x] Policy Playground (Web UI)
- [x] Policy denials API and Analytics UI
- [x] ROI methodology (docs/ROI_BENCHMARK.md or equivalent)

### ✅ v1.2: Ecosystem and Governance (Complete)
- [x] MCP (Model Context Protocol) Support
- [x] HITL Demo Flow
- [x] Budget Caps and Alerts (pkg/cost/budget.go)
- [x] GitOps for Policies (unagnt policy apply)
- [x] Workflow Versioning

### ✅ v1.3: Enterprise Foundations (Complete)
- [x] SSO / SAML / OIDC
- [x] Multi-Region / HA
- [x] Secrets Management (Vault, AWS/GCP Secrets Manager)
- [x] Enterprise Integrations (Slack, Teams, Jira/ServiceNow)
- [x] Advanced RBAC

### 🚀 v2.0: Enterprise Platform and Commercialization (Target: ~12–16 weeks)

**Goal:** Compliance packaging, air-gapped deployment, and clear commercial offer.

#### 1. Compliance Pack ✅
- **Deliverable:** `configs/compliance/soc2/`, `configs/compliance/hipaa/`, `GET /v1/compliance/audit/export?format=json|csv|cef` for SIEM.

#### 2. Air-Gapped Deployment ✅
- **Deliverable:** `deploy/air-gapped/`, `scripts/offline-install.sh` (bundle + install); optional Ollama/local model config.

#### 3. Open-Core Structure ✅
- **Deliverable:** `pkg/license/` feature gating (OSS core vs enterprise features); packaging for enterprise binary.

#### 4. Hosted Tier (Optional)
- **What:** Managed Unagnt for teams that prefer not to self-host.
- **Deliverable:** Hosted offering; separate pricing/sales motion.

#### 5. Agent Marketplace
- **What:** Public or private marketplace for workflows/tools; paid listings with revenue share.
- **Deliverable:** Marketplace UI, listing/purchase flow (extend `pkg/registry/workflow_marketplace.go`, `pkg/tool/marketplace.go`).

#### 6. Support and Documentation ✅
- **Deliverable:** [SUPPORT.md](SUPPORT.md), [docs/PRICING.md](docs/PRICING.md), [docs/SOW_TEMPLATE.md](docs/SOW_TEMPLATE.md).

---

### 🎯 v3.0: Production Runtime & Agentic Orchestration (Complete)

**Goal:** Enable real production workloads and differentiate with intelligent orchestration.

**Status Overview:**
- ✅ **Phase 1 complete**: Full runtime integration is implemented.
- ✅ **Phase 2 complete**: Agentic orchestration primitives are implemented.
- ✅ **Phase 3 complete**: Runtime hardening, rollout quality, and GA-readiness baseline are implemented.

#### Phase 1: Full Runtime Integration (✅ Done)
- **Included:**
  - Runtime-backed step execution via `RuntimeStepExecutor`
  - Async runner queue/service for isolated execution
  - Webhook-triggered runs wired to real runtime execution
  - Scheduled cron-based execution
  - Event-driven triggers and trigger endpoint
  - Workflow checkpoint persistence + resume flow
- **Outcome:** Production workflows now execute end-to-end; external systems can reliably trigger runs.

#### Phase 2: Agentic Orchestration (✅ Done)
- **Included:**
  - Multi-model routing for step execution
  - Guardrails layer for goal/output safety controls
  - Dynamic tool selection controls with policy awareness
  - Incremental/streaming execution signals
  - Supporting tests and docs updates
- **Outcome:** Orchestration is smarter and more cost-aware than static DAG-only execution.

#### Phase 3: Runtime Hardening and GA Readiness (✅ Done)
- **Included scope:**
  - End-to-end validation of runtime/webhook/scheduler/trigger paths under load (queue/backpressure and cancellation scenarios covered in tests)
  - Failure-mode hardening (retries, per-attempt timeouts, dead-letter capture, replay, cancellation-safe backoff)
  - Stronger observability for run lifecycle (persisted event history endpoint + runner hardening metrics)
  - Contract/integration coverage for queue behavior, dead-letter replay, retry/cancel semantics, and API validation guards
  - Final docs/runbooks and migration notes for rollout (`docs/ARCHITECTURE.md`, `docs/guides/api-integration.md`)
- **Outcome:** v3 ships with a production-ready runtime hardening baseline and operational tooling.

---

### 🔒 v3.1: Pre-v4 Hardening Gate (Required Before v4) — Complete

**Goal:** Resolve operational limitations from v3 and establish a release-quality reliability gate before v4 starts.

**Status:** All four items implemented. See [docs/MIGRATION_V3.1.md](docs/MIGRATION_V3.1.md) and [docs/AUDIT_V3.1.md](docs/AUDIT_V3.1.md).

#### 1. Dead-Letter Retention and Archival ✅
- Configurable retention via `DEAD_LETTER_RETENTION_HOURS`; optional archival via `DEAD_LETTER_ARCHIVE_DIR`
- Background pruner (hourly), optional dir archival before prune, metrics (`agentruntime_dead_letters_pruned_total`, `agentruntime_dead_letters_archived_total`)
- Runbook: [docs/runbooks/dead-letter-retention.md](docs/runbooks/dead-letter-retention.md)

#### 2. Durable Queue Backend ✅
- Pluggable backend: memory (default) and Redis. Env: `QUEUE_BACKEND`, `QUEUE_REDIS_URL`, `QUEUE_SIZE`
- Redis provides restart-resilient queuing
- Migration notes: [docs/MIGRATION_V3.1.md](docs/MIGRATION_V3.1.md)

#### 3. Kubernetes Operator Generation Automation ✅
- CI job `operator-codegen` fails when `zz_generated.deepcopy.go` is stale
- `make generate-operator-check` for local verification; controller-gen@v0.14.0 documented

#### 4. Release Readiness Gate ✅
- [docs/RELEASE_READINESS.md](docs/RELEASE_READINESS.md): required checks, SLO baseline, pre-release checklist
- [docs/runbooks/incidents.md](docs/runbooks/incidents.md): queue saturation, dead-letter spikes, replay control

---

### v4.0: Observability & Governance — Implemented
- Agent usage analytics by tenant, workflow, and model (`GET /v1/analytics/costs/workflows`, breakdown filters)
- Model drift and performance monitoring (Collector, PerformanceStore, `GET /v1/analytics/model-performance`, `model-drift`)
- Human review queues and approval flows (`/v1/approvals/*`, `unagnt approvals list|approve|deny`)
- Compliance report generation (`/v1/compliance/reports/*`, JSON/CSV/CEF export)
- Agent A/B testing framework (`/v1/ab-tests`, selector, results API)

### v5.0: Developer Experience — Implemented
- TypeScript/Node SDK (`sdk/typescript`, parity with Go/Python)
- Tool authoring test harness and mocks (`pkg/tool/testing`)
- Time-travel debugging (ReplayCursor, `unagnt replay debug`, API)
- VS Code extension (workflow authoring, local run, Explorer)
- Local-first development with cloud sync (`unagnt sync push|pull|status`)

### 📋 v6.0: Edge & Federated (Backlog)
- Edge deployment (agents closer to data/sensors)
- Federated learning patterns for cross-org models
- Hybrid orchestration (central control + edge execution)
- Resource-constrained device support (ARM, low memory)

### 📋 v7.0: Natural Language Workflows (Backlog)
- Natural language workflow definitions
- Conversational policy configuration
- No-code branching and routing
- “Explain this workflow” in plain language

### 📋 v8.0+: Hosted & Marketplace (Backlog)
- **Hosted Tier:** Managed Unagnt; multi-tenant SaaS; usage-based pricing.
- **Agent Marketplace:** Private/public marketplace; paid listings; discovery and curation.

---

## ⚠️ Known Limitations

- **Dead-letter retention**: v3.1 adds configurable retention and optional archival (env: `DEAD_LETTER_RETENTION_HOURS`, `DEAD_LETTER_ARCHIVE_DIR`). Without these set, dead letters accumulate indefinitely; enable retention for production.
- **Queue backend**: Default is in-memory (bounded); use `QUEUE_BACKEND=redis` and `QUEUE_REDIS_URL` for restart-resilient durable queuing. Monitor `agentruntime_run_queue_depth` and `agentruntime_run_queue_rejected_total`.
- **Kubernetes operator**: Use `make build-operator` to build the operator (auto-runs codegen). CRDs are generated and committed to `k8s/crds/`. CI fails if deepcopy or CRDs are stale; run `make generate-crds` and commit before merge. See [k8s/operator/BUILD_NOTES.md](k8s/operator/BUILD_NOTES.md).

### Implementation Notes

- **Workflow execution model**: Orchestration uses a runtime-backed executor and async runner queue by default. Simulated execution remains available for tests/dev via explicit `SimulatedExecutor` wiring.
- **Advanced runtime features**: Plugin artifact download and replay side effects (Replayable `http_call` GET) are implemented. Core features are production-ready.

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md#known-limitations--implementation-notes) for details.

---

## 📄 License

Unagnt is open source software licensed under the [MIT License](LICENSE).

---

## 🙏 Acknowledgments

Built with:
- [OpenTelemetry](https://opentelemetry.io/) for observability
- [CEL](https://github.com/google/cel-go) for policy expressions
- [React Flow](https://reactflow.dev/) for visual workflows
- [Kubernetes](https://kubernetes.io/) for orchestration
- And many other amazing open source projects!

---

## 💬 Community

- **GitHub Discussions**: [Ask questions and share ideas](https://github.com/NikoSokratous/unagnt/discussions)
- **Discord**: [Join our community](https://discord.gg/Unagnt)
- **Twitter**: [@Unagnt](https://twitter.com/Unagnt)
- **Blog**: [blog.Unagnt.io](https://blog.Unagnt.io)

---

## 📞 Support

- **Documentation**: [docs.Unagnt.io](https://docs.Unagnt.io)
- **Issues**: [GitHub Issues](https://github.com/NikoSokratous/unagnt/issues)
- **Commercial Support**: [contact@Unagnt.io](mailto:contact@Unagnt.io)

---

<div align="center">

**⭐ If you find Unagnt useful, please star the repo! ⭐**

Made with ❤️ by the Unagnt community

</div>
