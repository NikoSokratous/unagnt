# Unagnt v1.0 - Release Notes

**Release Date**: February 26, 2026  
**Status**: Production Ready  
**GitHub**: https://github.com/NikoSokratous/unagnt

---

## 🎉 What's New in v1.0

Unagnt v1.0 is a major release that transforms the platform from a foundational agent framework into a comprehensive, production-grade orchestration system with enterprise features.

---

## 🚀 Headline Features

### 1. Visual Workflow Designer
Design complex multi-agent workflows with an intuitive drag-and-drop interface. No code required!

**Key Capabilities**:
- Drag-and-drop node creation
- Real-time validation
- YAML import/export
- Cycle detection
- Dependency visualization

### 2. Interactive Debugger
Debug workflows in real-time with breakpoints, step-through execution, and variable inspection.

**Key Capabilities**:
- Conditional breakpoints (CEL expressions)
- Step-over execution
- Variable inspection
- Execution history
- Real-time state updates

### 3. Cost Attribution
Track costs in real-time per agent, tenant, and user with provider-specific pricing.

**Key Capabilities**:
- Real-time tracking
- OpenAI & Anthropic pricing
- Budget management
- Cost breakdown analytics

### 4. Kubernetes Native
Deploy and manage agents declaratively with Kubernetes CRDs and operators.

**Key Capabilities**:
- 3 Custom Resource Definitions
- Automatic lifecycle management
- HPA integration
- Service mesh support

### 5. RAG & Context Assembly with Embeddings
Ground agents in your knowledge base and memory with semantic search and embeddings.

**Key Capabilities**:
- **RAG (Retrieval Augmented Generation)**: Ingest documents (markdown, text), chunk with overlap, semantic search at query time
- **Context Assembly Engine**: Automatic retrieval of policies, workflow state, memory, tool outputs, and knowledge before each LLM call
- **Embeddings**: OpenAI text-embedding-3-small or local sentence-transformers (all-MiniLM-L6-v2)
- **Semantic Memory**: Similar past interactions retrieved by meaning, not keywords
- **Local Storage**: SQLite for persistent key-value memory; in-memory vector store for development
- **CLI**: `unagnt context ingest`, `unagnt context knowledge list`, `unagnt context search`

See [examples/knowledge-base/](examples/knowledge-base/) and [docs/EMBEDDINGS.md](docs/EMBEDDINGS.md).

---

## ✨ New Features by Category

### Visual Tools & Workflow Enhancement

#### Workflow Designer
- **React Flow Canvas**: Drag-and-drop workflow design
- **Node Types**: Agent, Tool, Condition, Parallel, Join
- **Real-Time Validation**: Cycle detection, disconnected nodes
- **YAML Import/Export**: Seamless conversion
- **Mini-Map**: Quick navigation for large workflows

#### Workflow Debugger
- **Breakpoints**: Unconditional and conditional (CEL)
- **Execution Control**: Play, Pause, Resume, Step Over
- **Variable Inspection**: Real-time state viewing
- **History Tracking**: Complete execution trace
- **WebSocket Support**: Real-time updates

#### Plugin Hot-Reloading
- **File System Watching**: Automatic detection of changes
- **Graceful Reload**: Zero-downtime updates
- **Version Tracking**: File modification timestamps
- **Multi-Directory Support**: Watch multiple locations

#### Workflow Marketplace
- **Template Sharing**: Publish and discover workflows
- **Ratings & Reviews**: 5-star rating system
- **Category Organization**: Data Processing, Code Review, Research
- **Download Tracking**: Usage analytics
- **Parameterized Templates**: Reusable workflows

### Operations & Infrastructure

#### Kubernetes Operator
- **Agent CRD**: Declarative agent management
- **Workflow CRD**: Workflow as code
- **Policy CRD**: Governance rules
- **Reconciliation**: Automatic state management
- **Scale Subresource**: HPA integration

#### Helm Charts
- **Multi-Environment**: Dev, staging, production
- **Dependencies**: PostgreSQL, Redis via Bitnami
- **Autoscaling**: HPA with CPU/memory targets
- **Security**: Pod security policies, network policies
- **Observability**: Prometheus, Grafana, Jaeger

#### Service Mesh Integration
- **Istio Support**: Full Istio integration
- **Linkerd Support**: Optional Linkerd integration
- **mTLS**: Mutual TLS encryption (STRICT mode)
- **Traffic Management**: Circuit breaking, retries, timeouts
- **Observability**: Distributed tracing

### Cost, Performance & Analytics

#### Cost Attribution
- **Provider Pricing**: OpenAI, Anthropic
- **Multi-Level Tracking**: Agent, tenant, user
- **Token-Based Billing**: Input/output tokens
- **Budget Management**: Limits and alerts
- **Real-Time Tracking**: 30s flush interval

#### SLA Monitoring
- **Uptime Tracking**: Service availability
- **Latency Metrics**: P50, P95, P99
- **Error Rate**: Success/failure tracking
- **SLA Targets**: Configurable thresholds
- **Violation Alerts**: Automatic notifications

#### Auto-Scaling
- **Custom Metrics**: 7 metric types
- **Weighted Scoring**: Multi-metric decisions
- **Intelligent Scaling**: Aggressive up, conservative down
- **Cooldown Periods**: Prevent thrashing
- **Predictive Scaling**: Historical data analysis

#### Analytics Dashboard
- **Summary Cards**: Cost, uptime, requests, error rate
- **Cost Visualization**: Pie charts by agent
- **Performance Timeline**: CPU, memory, throughput
- **SLA Metrics**: Bar charts by service
- **Time Ranges**: 1h, 24h, 7d, 30d

### Security & Reliability

#### Audit Log Encryption
- **AES-256-GCM**: Industry-standard encryption
- **Key Rotation**: Automatic 90-day rotation
- **Key Versioning**: Multiple key versions
- **KMS Integration**: AWS, GCP, Vault support
- **Tamper-Proof**: Authenticated encryption

#### Disaster Recovery
- **Automated Backups**: Scheduled full/incremental
- **Multi-Region Replication**: Async replication
- **Point-in-Time Recovery**: Restore to any time
- **Cross-Region Failover**: Automatic failover
- **WAL Archiving**: Transaction log backup

#### GraphQL API
- **Complete Schema**: Agents, Workflows, Runs, Costs
- **Queries**: Flexible data fetching
- **Mutations**: Create and execute operations
- **Subscriptions**: Real-time updates via WebSocket
- **Field-Level Permissions**: Granular access control

#### ML Model Registry
- **Multi-Provider**: MLflow, Seldon, KServe
- **Model Versioning**: Track versions
- **Model Deployment**: Deploy to environments
- **A/B Testing**: Traffic splitting
- **Drift Detection**: Monitor model performance

### Context Assembly, RAG & Embeddings

#### Context Assembly Engine
- **5 Built-in Providers**: Policy, workflow state, semantic memory, tool outputs, knowledge
- **Token Budgeting**: Per-section limits and smart truncation
- **Parallel Fetching**: Configurable concurrency for faster assembly
- **CLI Commands**: `unagnt context inspect`, `unagnt context explain`, `unagnt context validate`
- **Observability**: Metrics, tracing, and structured logs

#### RAG & Knowledge Base
- **Document Ingestion**: Ingest markdown and text from directories
- **Chunking**: Configurable chunk size (default 500 tokens) and overlap (default 50)
- **Semantic Retrieval**: Embed query, search vector store, inject top-K chunks into context
- **CLI**: `unagnt context ingest ./docs`, `unagnt context knowledge list`, `unagnt context search "query"`
- **Example**: [examples/knowledge-base/](examples/knowledge-base/) - support agent with RAG

#### Embeddings
- **OpenAI**: text-embedding-3-small (1536 dims, ~$0.02/1M tokens)
- **Local**: sentence-transformers (all-MiniLM-L6-v2, 384 dims, free)
- **Config**: `context_assembly.embeddings.provider`, `model`, `api_key_env`

#### Memory & Local Storage
- **Persistent Memory**: SQLite and PostgreSQL key-value stores
- **Semantic Store**: In-memory (development), Qdrant, Weaviate
- **In-memory Semantic Store**: Cosine similarity search, no external deps for local dev

---

## 🔧 Improvements

### Performance
- **Query Optimization**: 50% faster dashboard queries
- **Connection Pooling**: Reduced database load
- **Caching**: GraphQL query caching
- **Batch Operations**: Reduced API calls

### Reliability
- **Health Checks**: Liveness and readiness probes
- **Graceful Shutdown**: Proper cleanup
- **Retry Logic**: Automatic retries with backoff
- **Circuit Breakers**: Prevent cascade failures

### Security
- **Pod Security Contexts**: Non-root containers
- **Read-Only Filesystem**: Immutable containers
- **Network Policies**: Restrict traffic
- **Secrets Management**: External secrets support

### Observability
- **Structured Logging**: JSON logs
- **Distributed Tracing**: Full trace propagation
- **Custom Metrics**: Business metrics
- **Alert Rules**: Prometheus alerting

---

## 📊 Statistics

### Code
- **15,000+** lines of code
- **50+** new files
- **13** database migrations
- **40+** database tables

### APIs
- **50+** REST endpoints
- **10+** GraphQL types
- **2** subscription types
- **30+** CLI commands

### Testing
- **30+** unit tests
- **10+** integration tests
- **~70%** code coverage

### Documentation
- **5,000+** lines of documentation
- **10+** comprehensive guides
- **4** phase completion docs

---

## 🎯 Use Cases

### 1. Multi-Agent Code Review
```yaml
workflow:
  - Linting agent checks syntax
  - Security agent scans vulnerabilities
  - Multiple reviewer agents provide feedback
  - Aggregator creates final summary
  - Auto-approve if all checks pass
```

### 2. Research & Synthesis
```yaml
workflow:
  - Query expansion agent generates search terms
  - Multiple search agents gather information
  - Content extraction agents process results
  - Fact-checking agent verifies information
  - Report generation agent creates summary
```

### 3. Data Pipeline
```yaml
workflow:
  - Extraction agent fetches data
  - Validation agent checks quality
  - Transformation agents process data
  - Conditional routing based on results
  - Load agent writes to destination
```

### 4. Support Agent with RAG
```yaml
# Ingest docs, run agent
context_assembly:
  enabled: true
  embeddings:
    provider: openai
    model: text-embedding-3-small
  providers:
    - type: knowledge
      config:
        sources: ["./docs"]
        top_k: 3
```
```bash
unagnt context ingest ./docs
unagnt run --agent agent.yaml --goal "Help me with deployment"
```

---

## 🔄 Breaking Changes

### From v0.6 to v1.0

1. **Workflow Format**
   - Added `depends_on` field to steps
   - Changed from sequential to DAG-based

2. **Database Schema**
   - New tables for cost tracking
   - New tables for SLA monitoring
   - New tables for encrypted audit logs

3. **API Endpoints**
   - Some endpoints reorganized under `/v1/`
   - New GraphQL endpoint at `/graphql`

4. **Configuration**
   - New Helm values for Phase 1-4 features
   - Updated environment variables

---

## 📦 Installation

### Quick Start
```bash
# Clone repository
git clone https://github.com/NikoSokratous/unagnt.git
cd Unagnt

# Install dependencies
go mod download
cd web && npm install

# Run migrations
unagnt migrate

# Start services
docker-compose up -d

# Access UI
open http://localhost:3000
```

### Kubernetes
```bash
# Install CRDs
kubectl apply -f k8s/crds/

# Install with Helm
helm install Unagnt ./k8s/helm \
  --namespace Unagnt \
  --create-namespace
```

---

## ⬆️ Upgrade Guide

### From v0.6

1. **Backup your database**
   ```bash
   pg_dump Unagnt > backup.sql
   ```

2. **Run migrations**
   ```bash
   unagnt migrate --to 013
   ```

3. **Update configuration**
   ```bash
   # Update Helm values
   helm upgrade Unagnt ./k8s/helm -f values.yaml
   ```

4. **Verify**
   ```bash
   unagnt version
   kubectl get agents
   ```

---

## 🐛 Bug Fixes

- Fixed memory leak in workflow execution
- Resolved race condition in cost tracking
- Corrected timezone handling in SLA reports
- Fixed GraphQL subscription memory issues
- Resolved key rotation edge cases

---

## 📝 Known Issues

1. **GraphQL Subscriptions**
   - Require WebSocket connection
   - Not supported in all GraphQL clients

2. **Hot-Reloading**
   - Requires file system access
   - Not compatible with read-only filesystems

3. **ML Model Registry**
   - KMS integration requires additional setup
   - Some providers need API keys

---

## 🔮 Future Roadmap

### v1.1 (Q2 2026)
- Time-travel debugging
- VS Code extension
- Advanced analytics

### v1.2 (Q3 2026)
- Multi-cloud abstraction
- Agent marketplace
- Workflow IDE

### v2.0 (Q4 2026)
- Federated learning
- Edge deployment
- Natural language workflows

---

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

---

## 📄 License

MIT License - see [LICENSE](./LICENSE)

---

## 🙏 Acknowledgments

Special thanks to:
- React Flow team
- Kubernetes controller-runtime maintainers
- Istio community
- MLflow project
- All open-source contributors

---

## 📞 Support

- **Documentation**: https://docs.Unagnt.io
- **GitHub Issues**: https://github.com/NikoSokratous/unagnt/issues
- **Discord**: https://discord.gg/Unagnt
- **Email**: support@Unagnt.io

---

## 🎉 Thank You!

Thank you for using Unagnt! We're excited to see what you build with v1.0.

---

**Version**: 1.0.0  
**Release Date**: February 26, 2026  
**Status**: Production Ready ✅
