# Unagnt v2.0 - Release Notes

**Release Date**: March 2026  
**Status**: Production Ready  
**GitHub**: https://github.com/NikoSokratous/unagnt

---

This document covers all changes from v1.1 through v2.0. For v1.0 features, see [RELEASE_NOTES_V1.0.md](RELEASE_NOTES_V1.0.md).

---

## 📋 Summary

| Release | Theme | Status |
|---------|-------|--------|
| **v1.1** | Pitch Readiness | ✅ Complete |
| **v1.2** | Ecosystem and Governance | ✅ Complete |
| **v1.3** | Enterprise Foundations | ✅ Complete |
| **v2.0** | Enterprise Platform and Commercialization | ✅ Complete |

---

## 🚀 v1.1: Pitch Readiness

Focused on demos, policy tooling, and ROI storytelling for sales and investor pitches.

### Safety-First Demo
- **Location**: `showcase/safety-first-demo/`
- End-to-end demo of policy enforcement, risk scoring, and deterministic replay
- Scripted walkthrough for presentations

### Enterprise Compliance Bot
- **Location**: `showcase/enterprise-compliance-bot/`
- Showcase app demonstrating compliance workflows with policy and audit

### Policy Playground (Web UI)
- Test policies interactively in the browser
- Visual feedback for policy evaluation

### Policy Denials API and Analytics UI
- API for querying policy denials
- Analytics dashboard for policy enforcement metrics

### ROI Methodology
- **Docs**: [docs/ROI_BENCHMARK.md](docs/ROI_BENCHMARK.md)
- Framework for quantifying value and cost savings

---

## 🔧 v1.2: Ecosystem and Governance

Expanded ecosystem integration, cost controls, and policy-as-code workflows.

### MCP (Model Context Protocol) Support
- **Location**: `pkg/mcp/`
- Connect agents to MCP-compatible tools and data sources
- Standard protocol for model context exchange

### HITL (Human-in-the-Loop) Demo Flow
- **Location**: `showcase/hitl-demo/`
- Human-in-the-loop workflow demonstrations
- Approval and escalation patterns

### Budget Caps and Alerts
- **Location**: `pkg/cost/budget.go`
- Per-agent, per-tenant, and per-user budget limits
- Alerts when costs approach or exceed thresholds

### GitOps for Policies
- `unagnt policy apply` for declarative policy deployment
- Policy-as-code workflows with version control

### Workflow Versioning
- Version tracking for workflow definitions
- Support for versioned workflow runs and rollbacks
- See [docs/guides/policy-versioning.md](docs/guides/policy-versioning.md)

---

## 🏢 v1.3: Enterprise Foundations

Enterprise identity, resilience, secrets, and integrations.

### SSO / SAML / OIDC
- **Guide**: [docs/guides/enterprise-sso.md](docs/guides/enterprise-sso.md)
- OAuth2, OpenID Connect, and SAML 2.0 support
- Enterprise IdP integration (Okta, Azure AD, OneLogin, Auth0, Keycloak)
- Claim/attribute mappings for groups and roles
- Session management and secure cookies

### Advanced RBAC
- **Guide**: [docs/guides/advanced-rbac.md](docs/guides/advanced-rbac.md)
- Role templates, org hierarchy, delegation
- Fine-grained permissions and scoped access

### Secrets Management
- **Guide**: [docs/guides/secrets-management.md](docs/guides/secrets-management.md)
- HashiCorp Vault (KV v2) integration
- AWS Secrets Manager and GCP Secret Manager stubs
- Secret references in config: `secret:ref:path/to/secret`

### Multi-Region / HA
- High-availability deployment patterns
- Cross-region replication and failover support
- See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)

### Enterprise Integrations
- Slack, Microsoft Teams integrations
- Jira, ServiceNow integration examples
- See [examples/integrations/slack/](examples/integrations/slack/) and related

---

## 🚀 v2.0: Enterprise Platform and Commercialization

Compliance packaging, air-gapped deployment, open-core licensing, and commercial support.

### 1. Compliance Pack ✅
- **Location**: `configs/compliance/`
- **SOC2**: Ready-to-use configs and controls
- **HIPAA**: Healthcare-focused policies and documentation
- **SIEM Export API**: `GET /v1/compliance/audit/export?format=json|csv|cef`
- See [configs/compliance/README.md](configs/compliance/README.md)

### 2. Air-Gapped Deployment ✅
- **Location**: `deploy/air-gapped/`, `scripts/offline-install.sh`
- Bundle creation: `./scripts/offline-install.sh bundle`
- Transfer tarball and run `install.sh` in air-gapped environment
- Local LLM support (Ollama) for fully disconnected operation
- See [deploy/air-gapped/README.md](deploy/air-gapped/README.md)

### 3. Open-Core Structure ✅
- **Location**: `pkg/license/`
- Feature gating: OSS core vs enterprise features
- Licensing module for enterprise binary packaging
- Clean separation for commercial offerings

### 4. Support and Documentation ✅
- **SUPPORT.md**: Tier definitions (Community, Pro, Enterprise)
- **docs/PRICING.md**: Pricing and packaging
- **docs/SOW_TEMPLATE.md**: Statement of Work templates for implementation and training
- **docs/ROI_BENCHMARK.md**: Value methodology (from v1.1)

---

## 🐛 Bug Fixes and Improvements

### Concurrency
- **Plugin discovery**: Fixed `sync: Unlock of unlocked RWMutex` in watcher goroutine (`pkg/tool/discovery.go`)

### Kubernetes Operator
- DeepCopy generation for Message types
- Scheme registration per [k8s/operator/BUILD_NOTES.md](k8s/operator/BUILD_NOTES.md)
- Makefile targets for operator generation

### Risk Engine
- Improved `containsPII` regex patterns for privacy detection

### Registry
- Plugin artifact download wired to registry (no longer placeholder)

### Workflow Engine
- `StepExecutor` interface for pluggable execution
- `SimulatedExecutor` for testing and design-time validation
- See Known Limitations for execution model details

### Replay
- `executeSideEffect` for safe HTTP GET replay
- Replayable `http_call` side effects in deterministic replay

---

## 📦 Installation

### From v1.0 / v1.x

1. **Backup**
   ```bash
   pg_dump Unagnt > backup.sql
   ```

2. **Pull and migrate**
   ```bash
   git pull
   unagnt migrate
   ```

3. **Air-gapped (optional)**
   ```bash
   ./scripts/offline-install.sh bundle
   # Transfer tarball, then:
   tar -xzf Unagnt-air-gapped-*.tar.gz
   cd Unagnt-air-gapped && ./install.sh
   ```

4. **Kubernetes**
   ```bash
   helm upgrade Unagnt ./k8s/helm -f values.yaml
   ```

---

## ⚠️ Known Limitations

- **Workflow execution**: Default `StepExecutor` is simulated. Wire a custom executor for real agent runs. Webhook-triggered runs complete without execution until runner integration (Phase 2–3).
- **Operator build**: Run `make generate-operator` before first build. See [k8s/operator/BUILD_NOTES.md](k8s/operator/BUILD_NOTES.md).
- **Policy tests**: SQLite tests require CGO; may fail on Windows with CGO disabled.

---

## 🔗 Related Docs

- [RELEASE_NOTES_V1.0.md](RELEASE_NOTES_V1.0.md) – v1.0 features
- [README.md](README.md) – Roadmap and quick start
- [SUPPORT.md](SUPPORT.md) – Support tiers and contact
- [docs/PRICING.md](docs/PRICING.md) – Commercial pricing

---

**Version**: 2.0.0  
**Release Date**: March 2026  
**Status**: Production Ready ✅
