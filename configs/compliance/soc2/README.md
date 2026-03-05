# SOC2 Compliance Pack

Pre-built policy bundle for SOC2-aligned controls (access control CC6.1, audit CC7.2).

## Usage

```bash
unagnt policy apply configs/compliance/soc2/policy.yaml --activate
```

Export audit for SIEM: `GET /v1/compliance/audit/export?format=cef&range=7d`

See [../README.md](../README.md) for the full audit export API.
