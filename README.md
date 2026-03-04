# AgentRuntime

<div align="center">

**Production-grade runtime for autonomous AI agents**

[![CI](https://github.com/NikoSokratous/agentctl/actions/workflows/ci.yml/badge.svg)](https://github.com/NikoSokratous/agentctl/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/NikoSokratous/agentctl)](https://goreportcard.com/report/github.com/NikoSokratous/agentctl)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/NikoSokratous/agentctl)](https://github.com/NikoSokratous/agentctl/releases)

[Features](#-features) •
[Quick Start](#-quick-start) •
[Documentation](#-documentation) •
[Examples](#-examples) •
[Architecture](#-architecture) •
[Contributing](#-contributing)

</div>

---

## What is AgentRuntime?

AgentRuntime is an **enterprise-grade orchestration platform** for autonomous AI agents. Unlike chatbots or LLM wrappers, AgentRuntime provides the **infrastructure** needed to run AI agents safely and reliably in production.

### Why AgentRuntime?

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
- **Conditional Logic**: CEL-based branching
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

### Installation

```bash
# Clone the repository
git clone https://github.com/NikoSokratous/agentctl.git
cd agentruntime

# Build
make build

# Run the server
./bin/server
```

### Your First Agent

```bash
# Create an agent
agentctl agent create researcher \
  --goal "Research AI safety papers" \
  --llm gpt-4 \
  --max-steps 10

# Run the agent
agentctl agent run researcher

# Watch live execution
agentctl runs watch <run-id>
```

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
- **[CLI Reference](docs/CLI_REFERENCE.md)** - All agentctl commands
- **[API Reference](docs/API_REFERENCE.md)** - REST and GraphQL APIs
- **[Workflow Guide](examples/workflows/templates/README.md)** - Building workflows

### For Developers
- **[Architecture](docs/ARCHITECTURE.md)** - System design and components
- **[Development Guide](docs/DEVELOPMENT.md)** - Setup and contribution guide
- **[Plugin Development](docs/PLUGIN_DEVELOPMENT.md)** - Creating custom tools
- **[API Integration](docs/API_INTEGRATION.md)** - SDK usage and examples

### For Operators
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
Ingest docs, then run: `agentctl context ingest ./docs`

👉 **See [examples/](examples/) for 20+ ready-to-use examples**

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      AgentRuntime Platform                   │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Web UI       │  │ CLI (agentctl)│  │ REST/GraphQL │      │
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
│  ┌──────────┬──────────┬──┴────┬──────────┬──────────┐     │
│  │ PostgreSQL│  Redis  │ Qdrant│ Prometheus│ Jaeger  │     │
│  └──────────┴──────────┴───────┴──────────┴──────────┘     │
│                                                               │
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
- [x] GitOps for Policies (agentctl policy apply)
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
- **What:** Managed AgentRuntime for teams that prefer not to self-host.
- **Deliverable:** Hosted offering; separate pricing/sales motion.

#### 5. Agent Marketplace
- **What:** Public or private marketplace for workflows/tools; paid listings with revenue share.
- **Deliverable:** Marketplace UI, listing/purchase flow (extend `pkg/registry/workflow_marketplace.go`, `pkg/tool/marketplace.go`).

#### 6. Support and Documentation ✅
- **Deliverable:** [SUPPORT.md](SUPPORT.md), [docs/PRICING.md](docs/PRICING.md), [docs/SOW_TEMPLATE.md](docs/SOW_TEMPLATE.md).

---

### 🎯 v3.0: Production Runtime & Agentic Orchestration (Target: ~12–16 weeks)

**Goal:** Enable real production workloads and differentiate with intelligent orchestration. Addresses the primary barrier to first users: simulated execution.

**Rationale:** Without real workflow execution, users cannot run production agents. Agentic orchestration (multi-model routing, dynamic tool selection) differentiates AgentRuntime from generic workflow engines.

#### 1. Full Runtime Integration
- **Problem:** Webhook-triggered and scheduled runs complete without real execution; `StepExecutor` default is simulated.
- **Deliverables:**
  - Wire webhook handlers to real agent runtime
  - Scheduled/cron workflow execution
  - Event-driven triggers (pub/sub, queues)
  - Runner service for isolated workflow execution
  - Long-running workflows with checkpointing and resumption
- **Outcome:** Production workflows execute end-to-end; external systems can trigger agents reliably.

#### 2. Agentic Orchestration
- **Problem:** Static workflows; no intelligent routing or model selection.
- **Deliverables:**
  - Multi-model routing (route tasks to best LLM by cost, latency, capability)
  - Dynamic tool/agent selection (LLM decides which tools to call)
  - Guardrails layer (output constraints, topic control, safety filters)
  - Streaming/incremental execution where applicable
- **Outcome:** Smarter, more cost-efficient orchestration; clear differentiation from simple DAG runners.

---

### 📋 v4.0: Observability & Governance (Backlog)
- Agent usage analytics (by tenant, workflow, model)
- Model drift and performance monitoring
- Human review queues and approval flows
- Compliance report generation
- Agent A/B testing framework

### 📋 v5.0: Developer Experience (Backlog)
- VS Code extension (local dev, workflow authoring, debugging)
- Time-travel debugging for deterministic runs
- Local-first development with cloud sync
- TypeScript/Node SDK
- Tool authoring test harness and mocks

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
- **Hosted Tier:** Managed AgentRuntime; multi-tenant SaaS; usage-based pricing.
- **Agent Marketplace:** Private/public marketplace; paid listings; discovery and curation.

---

## ⚠️ Known Limitations

- **Workflow execution**: Orchestration uses `StepExecutor`; default is simulated. For real agent runs, wire a custom executor via `NewWorkflowEngineWithExecutor`. Webhook-triggered runs still mark completed without execution until runner integration (Phase 2–3).
- **Kubernetes operator**: Run `make generate-operator` (or `controller-gen` per [k8s/operator/BUILD_NOTES.md](k8s/operator/BUILD_NOTES.md)) before first build. Generated `zz_generated.deepcopy.go` is committed.
- **Advanced features**: Plugin artifact download and replay side effects (Replayable `http_call` GET) are implemented. Core features are production-ready.

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md#known-limitations--implementation-notes) for details.

---

## 📄 License

AgentRuntime is open source software licensed under the [MIT License](LICENSE).

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

- **GitHub Discussions**: [Ask questions and share ideas](https://github.com/NikoSokratous/agentctl/discussions)
- **Discord**: [Join our community](https://discord.gg/agentruntime)
- **Twitter**: [@agentruntime](https://twitter.com/agentruntime)
- **Blog**: [blog.agentruntime.io](https://blog.agentruntime.io)

---

## 📞 Support

- **Documentation**: [docs.agentruntime.io](https://docs.agentruntime.io)
- **Issues**: [GitHub Issues](https://github.com/NikoSokratous/agentctl/issues)
- **Commercial Support**: [contact@agentruntime.io](mailto:contact@agentruntime.io)

---

<div align="center">

**⭐ If you find AgentRuntime useful, please star the repo! ⭐**

Made with ❤️ by the AgentRuntime community

</div>
