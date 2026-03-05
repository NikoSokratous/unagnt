# Compliance Packs (v2.0)

Pre-built policy bundles and audit export for common compliance frameworks. Use with Unagnt's policy engine and audit log export API.

## Packs

| Pack | Description | Policy path |
|------|-------------|-------------|
| **SOC2** | Access control, audit logging, change management | [soc2/policy.yaml](soc2/policy.yaml) |
| **HIPAA** | Access, audit controls, integrity, transmission security | [hipaa/policy.yaml](hipaa/policy.yaml) |

## Audit Export API (SIEM)

Export policy audit logs for SIEM or compliance reporting:

```
GET /v1/compliance/audit/export?format={json|csv|cef}&range={1h|24h|7d|30d}&limit=...
```

- **format**: `json`, `csv`, or `cef` (Common Event Format for ArcSight, Splunk, etc.).
- **range**: `1h`, `24h`, `7d`, `30d`.
- **limit**: Optional max number of records (default 10000).
- **agent_name**, **policy_name**, **decision**: Optional filters.

Requires the server to be started with audit logging enabled (policy audit DB).

## Quick Start

1. Apply a bundle: `unagnt policy apply configs/compliance/soc2/policy.yaml --activate`
2. Run agents; policy decisions are logged.
3. Export for SIEM: `curl "http://localhost:8080/v1/compliance/audit/export?format=cef&range=7d" -o audit.cef`

See each pack's README for framework-specific notes and customization.
