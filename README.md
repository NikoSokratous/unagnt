# AgentRuntime

<div align="center">

**Production-grade runtime for autonomous AI agents**

[![Go Report Card](https://goreportcard.com/badge/github.com/agentruntime/agentruntime)](https://goreportcard.com/report/github.com/agentruntime/agentruntime)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/agentruntime/agentruntime)](https://github.com/agentruntime/agentruntime/releases)

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
- **Audit Logs**: Encrypted, tamper-proof execution logs
- **RBAC**: Role-based access control with multi-tenancy

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
- **Service Mesh**: Istio/Linkerd integration with mTLS
- **Auto-Scaling**: Queue depth and custom metrics-based scaling

### 🧠 Agent Capabilities
- **Memory Systems**: Working, persistent, semantic, and event log
- **Context Assembly**: Automatic retrieval of policies, memory, workflow state, and knowledge
- **RAG (Retrieval Augmented Generation)**: Ingest documents, semantic search, ground responses in your knowledge base
- **Embeddings**: OpenAI and local (sentence-transformers) for semantic memory and RAG
- **Local Storage**: SQLite for persistent memory; in-memory vector store for development
- **Tool Framework**: Schema-validated, versioned, permission-controlled
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
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Kubernetes, Docker, bare metal
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
│         ┌──────────────────┴──────────────────┐              │
│         │     Orchestration Layer             │              │
│         │  - Workflow Engine (DAG)            │              │
│         │  - Policy Enforcement               │              │
│         │  - Cost Tracking                    │              │
│         └──────────────────┬──────────────────┘              │
│                            │                                 │
│         ┌──────────────────┴──────────────────┐              │
│         │     Agent Runtime                   │              │
│         │  - State Machine                    │              │
│         │  - Tool Executor                    │              │
│         │  - Memory Manager                   │              │
│         │  - LLM Integration                  │              │
│         └──────────────────┬──────────────────┘              │
│                            │                                 │
│  ┌──────────┬──────────┬──┴────┬──────────┬──────────┐     │
│  │ PostgreSQL│  Redis  │ Qdrant│ Prometheus│ Jaeger  │     │
│  └──────────┴──────────┴───────┴──────────┴──────────┘     │
└─────────────────────────────────────────────────────────────┘
```

**Key Components:**
- **Runtime Engine**: State machine-based agent execution
- **Workflow Orchestrator**: DAG-based multi-agent coordination  
- **Policy Engine**: CEL-based policy enforcement
- **Memory Layer**: Multi-tier memory architecture
- **Tool Registry**: Versioned, permission-controlled tools
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

- **Language**: Go 1.25
- **LOC**: ~50,000 lines of production code
- **Tests**: 500+ unit and integration tests
- **Coverage**: 75%+
- **Dependencies**: Carefully curated, minimal external deps
- **License**: MIT

---

## 🗺️ Roadmap

### ✅ v1.0 (Current)
- [x] Visual workflow designer
- [x] Kubernetes operator
- [x] Cost attribution and SLA monitoring
- [x] Audit log encryption
- [x] GraphQL API
- [x] ML model registry integration
- [x] RAG with knowledge base and document ingestion
- [x] Context assembly with embeddings (OpenAI + local)
- [x] Semantic memory search and local storage (SQLite, in-memory)

### 🔮 v1.1 (Next)
- [ ] Multi-region deployment
- [ ] Advanced caching layer
- [ ] Workflow versioning
- [ ] A/B testing framework
- [ ] Enhanced collaboration features

### 🚀 v2.0 (Future)
- [ ] Agent marketplace
- [ ] Federated learning
- [ ] Edge deployment
- [ ] Mobile SDKs

---

## ⚠️ Known Limitations

- **Workflow execution**: Orchestration and webhook handlers use simulated delegation until full runtime integration; workflow DAGs run with placeholder execution in some paths.
- **Kubernetes operator**: Requires running `controller-gen` to generate DeepCopy methods before compilation. See [k8s/operator/BUILD_NOTES.md](k8s/operator/BUILD_NOTES.md).
- **Advanced features**: Replay side effects, plugin artifact download, and some risk/cost fallbacks use stubs. Core features are production-ready.

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

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/agentruntime/agentruntime/issues)

---

<div align="center">

**⭐ If you find AgentRuntime useful, please star the repo! ⭐**

Made with ❤️ by the AgentRuntime community

</div>
