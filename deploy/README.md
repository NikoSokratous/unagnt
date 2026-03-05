# Unagnt Deployment

## Docker Compose

### Default (SQLite only)

Zero external dependencies. Good for local development and testing.

```bash
docker-compose up
```

This starts `unagntd` with SQLite storage. No Postgres, Redis, or Qdrant required.

### Production profile (full stack)

For production with PostgreSQL, Redis, Qdrant, Prometheus, and Jaeger:

```bash
docker-compose -f docker-compose.yml -f docker-compose.production.yml --profile production up
```

## Docker (single image)

```bash
docker build -f deploy/Dockerfile -t unagnt .
docker run -p 8080:8080 unagnt
```
