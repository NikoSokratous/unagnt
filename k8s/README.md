# Unagnt Kubernetes Operator

Complete Kubernetes operator for managing Unagnt resources with production-grade features.

## Overview

The Unagnt operator enables declarative management of:
- **Agents**: AI agents with LLM configurations, tools, and autoscaling
- **Workflows**: DAG-based multi-agent workflows with scheduling
- **Policies**: Governance rules for safety and compliance

## Quick Start

### Install CRDs

CRDs in `k8s/crds/` are generated from `api/v1/types.go`. After changing types, run `make generate-crds` and commit the updated YAMLs.

```bash
kubectl apply -f k8s/crds/
```

### Deploy Operator

```bash
# Using kubectl
kubectl apply -f k8s/operator/deploy/

# Using Helm
helm install Unagnt ./k8s/helm \
  --namespace Unagnt \
  --create-namespace
```

### Create an Agent

```yaml
apiVersion: Unagnt.io/v1
kind: Agent
metadata:
  name: coder-agent
spec:
  role: coder
  llm:
    provider: openai
    model: gpt-4
  replicas: 2
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 10
```

```bash
kubectl apply -f agent.yaml
```

### Check Agent Status

```bash
kubectl get agents
kubectl describe agent coder-agent
```

## Custom Resources

### Agent

Manages AI agent deployments with:
- LLM configuration (provider, model, temperature)
- Tool integration
- Memory backends (Qdrant, Weaviate)
- Autoscaling (HPA integration)
- Resource limits

### Workflow

Manages multi-agent workflows with:
- DAG-based execution
- Conditional steps (CEL expressions)
- Scheduled execution (cron)
- Timeout and retry policies
- Parallel execution

### Policy

Enforces governance rules with:
- CEL-based conditions
- Actions: allow, deny, warn, require_approval
- Severity levels
- Approval workflows

## Helm Chart

### Installation

```bash
helm repo add Unagnt https://charts.Unagnt.io
helm install Unagnt Unagnt/Unagnt
```

### Configuration

Key values:

```yaml
replicaCount: 3
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10

postgresql:
  enabled: true
  
redis:
  enabled: true

serviceMesh:
  enabled: true
  type: istio
  mtls:
    mode: STRICT

observability:
  prometheus:
    enabled: true
  grafana:
    enabled: true
  jaeger:
    enabled: true
```

### Multi-Environment Setup

**Development**:
```bash
helm install Unagnt ./k8s/helm \
  -f k8s/helm/values-dev.yaml
```

**Staging**:
```bash
helm install Unagnt ./k8s/helm \
  -f k8s/helm/values-staging.yaml
```

**Production**:
```bash
helm install Unagnt ./k8s/helm \
  -f k8s/helm/values-prod.yaml
```

## Service Mesh Integration

### Istio

The operator integrates with Istio for:
- **mTLS**: Mutual TLS between services
- **Traffic Management**: Circuit breaking, retries, timeouts
- **Observability**: Distributed tracing with Jaeger
- **Security**: Authorization policies

Example configuration:

```yaml
serviceMesh:
  enabled: true
  type: istio
  mtls:
    mode: STRICT
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 100
      http:
        http1MaxPendingRequests: 100
    outlierDetection:
      consecutiveErrors: 5
      interval: 30s
```

### Linkerd

```yaml
serviceMesh:
  enabled: true
  type: linkerd
```

Linkerd automatically provides:
- mTLS encryption
- Load balancing
- Request retries
- Timeouts

## Autoscaling

### Horizontal Pod Autoscaler

Automatically scales based on:
- CPU utilization
- Memory utilization
- Custom metrics (queue depth, request rate)

```yaml
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
```

### Agent-Specific Autoscaling

```yaml
apiVersion: Unagnt.io/v1
kind: Agent
metadata:
  name: my-agent
spec:
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 20
    targetCPUUtilization: 60
```

## Observability

### Prometheus

Metrics exposed:
- Agent execution time
- Workflow success/failure rates
- Policy violations
- Resource utilization
- LLM API latency

```yaml
observability:
  prometheus:
    enabled: true
    serviceMonitor:
      enabled: true
      interval: 30s
```

### Grafana Dashboards

Pre-built dashboards for:
- Agent performance
- Workflow execution
- Policy compliance
- Cost tracking

### Jaeger Tracing

Distributed tracing for:
- Workflow execution paths
- Agent interactions
- Tool invocations
- LLM API calls

## Security

### Network Policies

```yaml
security:
  networkPolicy:
    enabled: true
```

Restricts traffic to:
- Same namespace only
- Specific ports
- Known services

### Pod Security

```yaml
security:
  podSecurityPolicy:
    enabled: true

securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  readOnlyRootFilesystem: true
```

### Secrets Management

Integration with:
- Kubernetes Secrets
- AWS Secrets Manager
- HashiCorp Vault
- Azure Key Vault

```yaml
security:
  secrets:
    externalSecrets:
      enabled: true
      backend: aws-secrets-manager
```

## Backup & Recovery

### Automated Backups

```yaml
backup:
  enabled: true
  schedule: "0 2 * * *"  # Daily at 2 AM
  retention: 7
  storage:
    type: s3
    bucket: Unagnt-backups
```

### Point-in-Time Recovery

```bash
# Restore from backup
helm upgrade Unagnt ./k8s/helm \
  --set backup.restore.enabled=true \
  --set backup.restore.timestamp="2026-02-26T14:00:00Z"
```

## Multi-Tenancy

### Namespace Isolation

```yaml
tenancy:
  enabled: true
  isolation:
    level: namespace
```

Each tenant gets:
- Dedicated namespace
- Resource quotas
- Network policies
- RBAC rules

### Cluster-Level Isolation

```yaml
tenancy:
  isolation:
    level: cluster
```

## Examples

See `k8s/examples/` for:
- `agent-example.yaml` - Agent configurations
- `workflow-example.yaml` - Workflow definitions
- `policy-example.yaml` - Policy rules

## Troubleshooting

### Check Operator Logs

```bash
kubectl logs -n Unagnt-system \
  -l control-plane=controller-manager \
  -f
```

### Debug Agent

```bash
kubectl describe agent my-agent
kubectl logs -l Unagnt.io/agent=my-agent
```

### Debug Workflow

```bash
kubectl describe workflow my-workflow
kubectl get jobs -l Unagnt.io/workflow=my-workflow
```

## Development

### Build Operator

```bash
cd k8s/operator
go build -o bin/manager main.go
```

### Run Locally

```bash
go run main.go
```

### Generate CRDs

```bash
make manifests
```

## Architecture

```
┌─────────────────────────────────────────────┐
│           Kubernetes Cluster                │
│                                             │
│  ┌─────────────────────────────────────┐  │
│  │   Unagnt Operator             │  │
│  │                                     │  │
│  │  ┌─────────┐  ┌─────────┐         │  │
│  │  │ Agent   │  │Workflow │         │  │
│  │  │Reconcile│  │Reconcile│         │  │
│  │  └────┬────┘  └────┬────┘         │  │
│  │       │            │              │  │
│  │       v            v              │  │
│  │  ┌──────────────────────┐        │  │
│  │  │   Resource Manager   │        │  │
│  │  └──────────────────────┘        │  │
│  └─────────────────────────────────────┘  │
│                                             │
│  ┌─────────────────────────────────────┐  │
│  │         Agent Deployments           │  │
│  │  ┌──────┐  ┌──────┐  ┌──────┐     │  │
│  │  │Agent │  │Agent │  │Agent │     │  │
│  │  │ Pod  │  │ Pod  │  │ Pod  │     │  │
│  │  └──────┘  └──────┘  └──────┘     │  │
│  └─────────────────────────────────────┘  │
│                                             │
│  ┌─────────────────────────────────────┐  │
│  │      Service Mesh (Istio)           │  │
│  │  - mTLS                             │  │
│  │  - Traffic Management               │  │
│  │  - Observability                    │  │
│  └─────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

## Resources

- [CRD Reference](./crds/README.md)
- [Helm Values](./helm/values.yaml)
- [Examples](./examples/)
- [API Documentation](https://docs.Unagnt.io/api)

## License

MIT License - see LICENSE file for details
