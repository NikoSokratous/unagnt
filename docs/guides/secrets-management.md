# Secrets Management

Unagnt supports pluggable secret backends so production deployments can avoid environment variables and plaintext config.

## Supported Backends

| Backend | Status | Use Case |
|---------|--------|----------|
| `env` | Implemented | Development; secrets from env vars |
| `vault` | Implemented | HashiCorp Vault KV v2 |
| `aws` | Stub | AWS Secrets Manager |
| `gcp` | Stub | GCP Secret Manager |

## Configuration

```yaml
secrets:
  backend: vault  # vault, aws, gcp, env
  vault:
    address: https://vault.example.com
    token: ""     # Use VAULT_TOKEN env in production
    mount: secret
  aws:
    region: us-east-1
  gcp:
    project: my-project
```

## Secret References in Config

Reference secrets in YAML using:

- `secret:ref:path/to/secret` – Vault path (returns default "value" key)
- `secret:ref:path/to/secret#key` – Vault path and field
- `env:VAR_NAME` – Environment variable
- `$VAR` or `${VAR}` – Environment variable (legacy)

Example (auth config):

```yaml
auth:
  providers:
    - id: okta
      type: oidc
      oidc:
        client_secret: secret:ref:Unagnt/okta#client_secret
```

## Vault Setup

1. Enable KV v2: `vault secrets enable -path=secret kv-v2`
2. Store a secret: `vault kv put secret/Unagnt/okta client_secret=xxx`
3. Set `VAULT_ADDR` and `VAULT_TOKEN` (or token in config for dev only).
4. Config: `secret:ref:Unagnt/okta#client_secret`

## AWS Secrets Manager

Add `github.com/aws/aws-sdk-go-v2/service/secretsmanager` and implement `GetSecret` in `pkg/secrets/aws.go`. Use IAM roles for auth.

## GCP Secret Manager

Add `cloud.google.com/go/secretmanager` and implement `GetSecret` in `pkg/secrets/gcp.go`. Use workload identity for auth.
