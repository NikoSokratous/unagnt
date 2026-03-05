# Deployment Guide

## Overview

Unagnt agents can be deployed in multiple ways:
- Local development
- Docker containers
- Kubernetes clusters
- Cloud platforms (AWS, GCP, Azure)

## Local Deployment

### Prerequisites

- Go 1.21+
- Unagnt CLI installed

### Steps

1. Build your agent configuration:
   ```bash
   unagnt init my-agent
   ```

2. Test locally:
   ```bash
   unagnt run --agent agent.yaml --goal "test goal"
   ```

3. Run in daemon mode:
   ```bash
   unagnt daemon --agent agent.yaml
   ```

## Docker Deployment

### Build Container

Create a `Dockerfile`:

```dockerfile
FROM golang:1.21 as builder
WORKDIR /app
COPY . .
RUN go build -o Unagnt

FROM debian:bookworm-slim
COPY --from=builder /app/Unagnt /usr/local/bin/
COPY agent.yaml /etc/Unagnt/
CMD ["Unagnt", "daemon", "--config", "/etc/Unagnt/agent.yaml"]
```

Build and run:

```bash
docker build -t my-agent:latest .
docker run -d -p 8080:8080 my-agent:latest
```

## Kubernetes Deployment

### Create Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: support-agent
spec:
  replicas: 3
  selector:
    matchLabels:
      app: support-agent
  template:
    metadata:
      labels:
        app: support-agent
    spec:
      containers:
      - name: agent
        image: my-agent:latest
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: agent-secrets
              key: openai-key
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

Apply:

```bash
kubectl apply -f deployment.yaml
```

## Best Practices

### 1. Environment Configuration

Use environment variables for sensitive data:

```yaml
model:
  provider: openai
  api_key_env: OPENAI_API_KEY
```

### 2. Health Checks

Implement health endpoints:

```go
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
})
```

### 3. Logging

Configure structured logging:

```yaml
logging:
  level: info
  format: json
  output: stdout
```

### 4. Metrics

Enable Prometheus metrics:

```yaml
observability:
  metrics:
    enabled: true
    port: 9090
```

### 5. Scaling

Configure horizontal pod autoscaling:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: support-agent-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: support-agent
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Cloud Platform Deployment

### AWS ECS

Use AWS Fargate for serverless containers:

```bash
aws ecs create-cluster --cluster-name agent-cluster
aws ecs register-task-definition --cli-input-json file://task-def.json
aws ecs create-service --cluster agent-cluster --service-name support-agent
```

### Google Cloud Run

Deploy directly from source:

```bash
gcloud run deploy support-agent \
  --source . \
  --region us-central1 \
  --allow-unauthenticated
```

### Azure Container Instances

```bash
az container create \
  --resource-group myResourceGroup \
  --name support-agent \
  --image my-agent:latest \
  --cpu 1 --memory 1
```

## Monitoring

Set up monitoring and alerting:

1. **Logs**: Use centralized logging (ELK, Datadog)
2. **Metrics**: Export to Prometheus/Grafana
3. **Tracing**: Use OpenTelemetry for distributed tracing
4. **Alerts**: Configure alerts for failures and performance issues

## Security

### API Keys

- Store in secrets management (AWS Secrets Manager, Vault)
- Rotate regularly
- Use least privilege access

### Network

- Use VPC/private networks
- Enable TLS/SSL
- Configure firewalls and security groups

### Updates

- Keep dependencies updated
- Monitor for security vulnerabilities
- Have a rollback plan

## Troubleshooting

### Container won't start

Check logs:
```bash
docker logs <container-id>
kubectl logs <pod-name>
```

### High memory usage

- Reduce context token limits
- Enable memory cleanup
- Add resource limits

### Slow performance

- Enable parallel context fetching
- Use caching
- Scale horizontally

## Next Steps

- Set up CI/CD pipelines
- Implement blue-green deployments
- Configure auto-scaling policies
- Set up disaster recovery
