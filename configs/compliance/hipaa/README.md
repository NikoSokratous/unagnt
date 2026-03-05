# HIPAA Compliance Pack

Pre-built policy bundle for HIPAA-aligned controls (access, audit, integrity, transmission).

## Usage

```bash
unagnt policy apply configs/compliance/hipaa/policy.yaml --activate
```

Export audit for SIEM: `GET /v1/compliance/audit/export?format=cef&range=7d`

See [../README.md](../README.md) for the full audit export API.
