# Unagnt Deployment Guide

**Last Updated**: v1.0.0 | 2026-02-27

Complete guide for deploying Unagnt to production environments.

---

## Table of Contents

1. [Deployment Options](#deployment-options)
2. [Prerequisites](#prerequisites)
3. [Docker Deployment](#docker-deployment)
4. [Kubernetes Deployment](#kubernetes-deployment)
5. [Configuration](#configuration)
6. [Monitoring Setup](#monitoring-setup)
7. [Backup & Recovery](#backup--recovery)
8. [Scaling](#scaling)
9. [Security Hardening](#security-hardening)
10. [Troubleshooting](#troubleshooting)

---

## Deployment Options

| Option | Best For | Complexity | Scalability |
|--------|----------|------------|-------------|
| **Local (Binary)** | Development, testing | Low | Single machine |
| **Docker** | Quick deploy, testing | Medium | Single server |
| **Docker Compose** | Multi-service local | Medium | Single server |
| **Kubernetes** | Production, enterprise | High | Unlimited |
| **Managed K8s** | Production (simplified) | Medium | Unlimited |

---

## Prerequisites

### All Deployments
- **OS**: Linux (Ubuntu 22.04+), macOS, Windows with WSL2
- **Go**: 1.22+ (for building from source)
- **Database**: PostgreSQL 14+ or SQLite 3
- **LLM API Key**: OpenAI, Anthropic, or Ollama

### Docker Deployment
- **Docker**: 24.0+
- **Docker Compose**: 2.20+ (optional)

### Kubernetes Deployment
- **Kubernetes**: 1.25+
- **kubectl**: Matching cluster version
- **Helm**: 3.12+
- **Storage Class**: For persistent volumes

### Recommended Resources

**Minimum** (testing):
- CPU: 2 cores
- RAM: 4GB
- Disk: 20GB

**Production** (Kubernetes):
- CPU: 8+ cores (across nodes)
- RAM: 16GB+ (across nodes)
- Disk: 100GB+ SSD

---

## Docker Deployment

### Option 1: Single Container

#### 1. Build Image

```bash
# Clone repository
git clone https://github.com/NikoSokratous/unagnt.git
cd Unagnt

# Build Docker image
docker build -t Unagnt:latest .
```

#### 2. Run Container

```bash
# Run with SQLite (simple)
docker run -d \
  --name Unagnt \
  -p 8080:8080 \
  -p 3000:3000 \
  -e OPENAI_API_KEY="sk-..." \
  -e DATABASE_URL="sqlite:///data/Unagnt.db" \
  -v $(pwd)/data:/data \
  Unagnt:latest

# Run with PostgreSQL (recommended)
docker run -d \
  --name Unagnt \
  -p 8080:8080 \
  -p 3000:3000 \
  -e OPENAI_API_KEY="sk-..." \
  -e DATABASE_URL="postgresql://user:pass@host:5432/Unagnt" \
  -e REDIS_URL="redis://host:6379" \
  Unagnt:latest
```

#### 3. Verify

```bash
# Check health
curl http://localhost:8080/health

# Check logs
docker logs -f Unagnt
```

### Option 2: Docker Compose

#### 1. Create `docker-compose.yml`

```yaml
version: '3.8'

services:
  Unagnt:
    build: .
    ports:
      - "8080:8080"
      - "3000:3000"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - DATABASE_URL=postgresql://Unagnt:password@postgres:5432/Unagnt
      - REDIS_URL=redis://redis:6379
      - QDRANT_URL=http://qdrant:6333
      - LOG_LEVEL=info
    depends_on:
      - postgres
      - redis
      - qdrant
    volumes:
      - ./policies:/app/policies
      - ./tools:/app/tools
    restart: unless-stopped

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_USER=Unagnt
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=Unagnt
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    restart: unless-stopped

  qdrant:
    image: qdrant/qdrant:latest
    volumes:
      - qdrant_data:/qdrant/storage
    ports:
      - "6333:6333"
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    ports:
      - "9090:9090"
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
    ports:
      - "3001:3000"
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  qdrant_data:
  prometheus_data:
  grafana_data:
```

#### 2. Create `.env`

```bash
# .env
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
LOG_LEVEL=info
```

#### 3. Deploy

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f Unagnt

# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

#### 4. Initialize Database

```bash
# Run migrations
docker-compose exec Unagnt ./bin/unagnt init

# Seed templates
docker-compose exec Unagnt ./bin/seed-templates
```

---

## Kubernetes Deployment

### Option 1: Helm Charts (Recommended)

#### 1. Add Helm Repository

```bash
# Add Unagnt Helm repo
helm repo add Unagnt https://helm.Unagnt.io
helm repo update

# Or use local charts
cd k8s/helm
```

#### 2. Create `values.yaml`

```yaml
# values-production.yaml
replicaCount: 3

image:
  repository: Unagnt/Unagnt
  tag: "1.0.0"
  pullPolicy: IfNotPresent

resources:
  limits:
    cpu: 2000m
    memory: 4Gi
  requests:
    cpu: 500m
    memory: 1Gi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

service:
  type: LoadBalancer
  port: 80
  targetPort: 8080

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: Unagnt.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: Unagnt-tls
      hosts:
        - Unagnt.example.com

# External secrets (if using external-secrets operator)
externalSecrets:
  enabled: true
  backend: gcpSecretsManager
  projectId: "my-project"

# Or use standard secrets
secrets:
  openaiApiKey: "sk-..."
  anthropicApiKey: "sk-ant-..."

postgresql:
  enabled: true
  auth:
    username: Unagnt
    password: ""  # Auto-generated
    database: Unagnt
  primary:
    persistence:
      enabled: true
      size: 50Gi
      storageClass: fast-ssd
  resources:
    limits:
      cpu: 2000m
      memory: 4Gi
    requests:
      cpu: 500m
      memory: 2Gi

redis:
  enabled: true
  auth:
    enabled: true
    password: ""  # Auto-generated
  master:
    persistence:
      enabled: true
      size: 10Gi
  resources:
    limits:
      cpu: 500m
      memory: 1Gi

qdrant:
  enabled: true
  persistence:
    enabled: true
    size: 100Gi
  resources:
    limits:
      cpu: 2000m
      memory: 8Gi

# Monitoring
prometheus:
  enabled: true
  serviceMonitor:
    enabled: true

grafana:
  enabled: true
  adminPassword: ""  # Auto-generated

# Service Mesh (optional)
istio:
  enabled: false
  mtls:
    mode: STRICT
```

#### 3. Deploy

```bash
# Create namespace
kubectl create namespace Unagnt

# Install with Helm
helm install Unagnt Unagnt/Unagnt \
  --namespace Unagnt \
  --values values-production.yaml \
  --wait

# Or from local charts
helm install Unagnt ./k8s/helm \
  --namespace Unagnt \
  --values values-production.yaml \
  --wait
```

#### 4. Verify Deployment

```bash
# Check pods
kubectl get pods -n Unagnt

# Check services
kubectl get svc -n Unagnt

# Check ingress
kubectl get ingress -n Unagnt

# View logs
kubectl logs -n Unagnt -l app=Unagnt -f

# Port forward for testing
kubectl port-forward -n Unagnt svc/Unagnt 8080:80
```

### Option 2: Manual Kubernetes Manifests

#### 1. Create Namespace

```bash
kubectl create namespace Unagnt
```

#### 2. Create Secrets

```bash
# Create API key secret
kubectl create secret generic Unagnt-secrets \
  --namespace Unagnt \
  --from-literal=openai-api-key='sk-...' \
  --from-literal=anthropic-api-key='sk-ant-...' \
  --from-literal=database-password='secure-password'
```

#### 3. Apply Manifests

```bash
# PostgreSQL
kubectl apply -f k8s/manifests/postgres.yaml

# Redis
kubectl apply -f k8s/manifests/redis.yaml

# Qdrant
kubectl apply -f k8s/manifests/qdrant.yaml

# Unagnt
kubectl apply -f k8s/manifests/Unagnt-deployment.yaml
kubectl apply -f k8s/manifests/Unagnt-service.yaml
kubectl apply -f k8s/manifests/Unagnt-ingress.yaml

# Operator (for CRDs)
kubectl apply -f k8s/manifests/operator.yaml
kubectl apply -f k8s/crds/
```

#### 4. Run Migrations

```bash
# Run migration job
kubectl apply -f k8s/manifests/migrations-job.yaml

# Check migration status
kubectl logs -n Unagnt job/Unagnt-migrations
```

---

## Configuration

### Environment Variables

#### Core Settings

```bash
# Server
PORT=8080                    # API server port
WEB_PORT=3000               # Web UI port (if bundled)
LOG_LEVEL=info              # debug, info, warn, error
ENVIRONMENT=production      # dev, staging, production

# Database
DATABASE_URL=postgresql://user:pass@host:5432/Unagnt
DATABASE_MAX_CONNS=25
DATABASE_MAX_IDLE_CONNS=5

# Redis
REDIS_URL=redis://host:6379
REDIS_PASSWORD=
REDIS_DB=0

# Vector Database
QDRANT_URL=http://localhost:6333
QDRANT_API_KEY=

# LLM Providers
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
OLLAMA_URL=http://localhost:11434

# Default LLM
LLM_PROVIDER=openai         # openai, anthropic, ollama
LLM_MODEL=gpt-4            # Model to use by default
```

#### Security

```bash
# Authentication
AUTH_ENABLED=true
API_KEY_HEADER=X-API-Key
BEARER_TOKEN_SECRET=your-secret-key

# OAuth2/OIDC
OAUTH_ENABLED=true
OAUTH_GOOGLE_CLIENT_ID=...
OAUTH_GOOGLE_CLIENT_SECRET=...
OAUTH_GITHUB_CLIENT_ID=...
OAUTH_GITHUB_CLIENT_SECRET=...

# Encryption
AUDIT_LOG_ENCRYPTION=true
KMS_PROVIDER=aws            # aws, gcp, vault
KMS_KEY_ID=...
```

#### Observability

```bash
# Tracing
TRACING_ENABLED=true
OTLP_ENDPOINT=localhost:4317
OTLP_INSECURE=false
TRACE_SAMPLER=0.1           # Sample 10% of traces

# Metrics
METRICS_ENABLED=true
METRICS_PORT=9090
```

#### Policies

```bash
# Policy Engine
POLICY_DIR=/app/policies
POLICY_DEFAULT_ENVIRONMENT=production
POLICY_CACHE_TTL=300        # seconds
```

### Configuration File

Create `config.yaml`:

```yaml
server:
  port: 8080
  timeout: 30s
  max_request_size: 10MB

database:
  url: ${DATABASE_URL}
  max_connections: 25
  max_idle_connections: 5
  connection_lifetime: 5m

redis:
  url: ${REDIS_URL}
  pool_size: 10

llm:
  provider: openai
  model: gpt-4
  max_tokens: 4000
  temperature: 0.7
  timeout: 60s

security:
  auth_enabled: true
  cors_origins:
    - https://example.com
  rate_limits:
    per_ip: 100/minute
    per_tenant: 1000/hour

observability:
  tracing:
    enabled: true
    endpoint: ${OTLP_ENDPOINT}
    sample_rate: 0.1
  metrics:
    enabled: true
    port: 9090
  logging:
    level: info
    format: json

policies:
  directory: /app/policies
  default_environment: production
  cache_ttl: 5m
```

Load with:
```bash
./bin/server --config config.yaml
```

---

## Monitoring Setup

### Prometheus Configuration

Create `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'Unagnt'
    static_configs:
      - targets: ['Unagnt:9090']
    metrics_path: '/metrics'

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
```

### Grafana Dashboards

Import pre-built dashboards:

1. **Unagnt Overview** (ID: TBD)
   - Active runs
   - Success/failure rates
   - P50/P95/P99 latencies

2. **Cost Analytics** (ID: TBD)
   - Cost by agent
   - Cost by tenant
   - Cost trends

3. **Policy Analytics** (ID: TBD)
   - Policy denials
   - Risk score distribution
   - Human-in-the-loop metrics

### Jaeger Tracing

```bash
# Deploy Jaeger
kubectl apply -f k8s/manifests/jaeger.yaml

# Or with Docker
docker run -d \
  --name jaeger \
  -p 16686:16686 \
  -p 4317:4317 \
  jaegertracing/all-in-one:latest
```

Configure Unagnt:
```bash
export TRACING_ENABLED=true
export OTLP_ENDPOINT=jaeger:4317
```

---

## Backup & Recovery

### Database Backups

#### PostgreSQL

```bash
# Manual backup
kubectl exec -n Unagnt postgres-0 -- \
  pg_dump -U Unagnt Unagnt | gzip > backup.sql.gz

# Restore
gunzip < backup.sql.gz | \
  kubectl exec -i -n Unagnt postgres-0 -- \
  psql -U Unagnt Unagnt

# Automated with CronJob
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:15
            command:
            - /bin/sh
            - -c
            - pg_dump -U Unagnt Unagnt | gzip | aws s3 cp - s3://backups/$(date +%Y%m%d).sql.gz
```

### Volume Backups

Using Velero:

```bash
# Install Velero
velero install \
  --provider aws \
  --bucket Unagnt-backups \
  --secret-file ./credentials-velero

# Backup namespace
velero backup create Unagnt-$(date +%Y%m%d) \
  --include-namespaces Unagnt

# Schedule daily backups
velero schedule create Unagnt-daily \
  --schedule="@daily" \
  --include-namespaces Unagnt

# Restore
velero restore create --from-backup Unagnt-20240101
```

---

## Scaling

### Horizontal Pod Autoscaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: Unagnt-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: Unagnt
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  - type: Pods
    pods:
      metric:
        name: active_runs
      target:
        type: AverageValue
        averageValue: "50"
```

### Database Scaling

#### Read Replicas

```yaml
# Add read replicas
postgresql:
  replication:
    enabled: true
    readReplicas: 2
```

Configure connection pooling:
```go
// Use pgbouncer or connection pool
DATABASE_URL=postgresql://user:pass@pgbouncer:5432/Unagnt?pool_size=25
```

### Caching Strategy

```bash
# Enable aggressive caching
POLICY_CACHE_TTL=600        # 10 minutes
TOOL_SCHEMA_CACHE_TTL=3600  # 1 hour
LLM_RESPONSE_CACHE=true
```

---

## Security Hardening

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: Unagnt-network-policy
spec:
  podSelector:
    matchLabels:
      app: Unagnt
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
```

### Pod Security

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: Unagnt
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    fsGroup: 1000
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: Unagnt
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
      readOnlyRootFilesystem: true
```

### Secrets Management

Using External Secrets Operator:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: gcpsm-secret-store
spec:
  provider:
    gcpsm:
      projectID: "my-project"

---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: Unagnt-secrets
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: gcpsm-secret-store
    kind: SecretStore
  target:
    name: Unagnt-secrets
  data:
  - secretKey: openai-api-key
    remoteRef:
      key: openai-api-key
```

---

## Troubleshooting

### Common Issues

#### 1. Pods Not Starting

```bash
# Check pod status
kubectl describe pod -n Unagnt <pod-name>

# Check logs
kubectl logs -n Unagnt <pod-name>

# Common causes:
# - ImagePullBackOff: Check image name/tag
# - CrashLoopBackOff: Check application logs
# - Pending: Check resource requests/limits
```

#### 2. Database Connection Failures

```bash
# Test connection
kubectl exec -n Unagnt <pod-name> -- \
  psql -h postgres -U Unagnt -d Unagnt -c "SELECT 1"

# Check service
kubectl get svc -n Unagnt postgres

# Check network policy
kubectl get networkpolicy -n Unagnt
```

#### 3. High Memory Usage

```bash
# Check metrics
kubectl top pods -n Unagnt

# Increase limits
kubectl set resources deployment Unagnt \
  --limits=memory=8Gi

# Check for memory leaks
kubectl exec -n Unagnt <pod-name> -- \
  curl http://localhost:8080/debug/pprof/heap
```

#### 4. Slow Response Times

```bash
# Check traces in Jaeger
open http://localhost:16686

# Check Prometheus metrics
curl http://localhost:9090/api/v1/query?query=http_request_duration_seconds

# Increase replicas
kubectl scale deployment Unagnt --replicas=10
```

### Debug Mode

```bash
# Enable debug logging
kubectl set env deployment/Unagnt LOG_LEVEL=debug

# View debug logs
kubectl logs -n Unagnt -l app=Unagnt -f --tail=100
```

---

## Production Checklist

- [ ] TLS/SSL configured
- [ ] Secrets in KMS/Vault (not ConfigMaps)
- [ ] Resource limits set
- [ ] HPA configured
- [ ] Monitoring enabled (Prometheus + Grafana)
- [ ] Tracing enabled (Jaeger)
- [ ] Backups scheduled (Velero)
- [ ] Network policies applied
- [ ] Pod security policies enforced
- [ ] Health checks configured
- [ ] Load testing completed
- [ ] Disaster recovery plan documented
- [ ] Incident response plan ready
- [ ] On-call rotation established

---

## Support

- **Documentation**: [docs.Unagnt.io](https://docs.Unagnt.io)
- **Issues**: [GitHub](https://github.com/NikoSokratous/unagnt/issues)
- **Discord**: [Community](https://discord.gg/Unagnt)
- **Commercial**: [contact@Unagnt.io](mailto:contact@Unagnt.io)
