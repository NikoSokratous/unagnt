# Policy Versioning Guide

## Overview

Policy versioning in Unagnt v0.5 provides a complete audit trail of all policy changes, enabling safe policy evolution, rollback capabilities, and compliance tracking.

## Key Concepts

### Policy Versions

Each policy can have multiple versions, identified by semantic version numbers (e.g., 1.0.0, 1.1.0, 2.0.0).

### Active Version

Only one version of a policy can be active at a time. The active version is used for runtime policy evaluation.

### Changelog

Every version includes a changelog describing what changed, who made the change, and when.

## Usage

### Creating a Policy Version

```go
package main

import (
    "context"
    "github.com/Unagnt/Unagnt/pkg/policy"
)

func main() {
    ctx := context.Background()
    store, _ := policy.NewVersionStore(db, "policies")
    
    version := &policy.PolicyVersion{
        PolicyName:  "security-policy",
        Version:     "1.0.0",
        Content:     policyYAML,
        Format:      "yaml",
        Author:      "security@company.com",
        Changelog:   "Initial security policy with file access controls",
        Active:      true,
    }
    
    store.SaveVersion(ctx, version)
}
```

### Using the CLI

#### List All Policies

```bash
$ unagnt policy list

POLICY NAME         ACTIVE VERSION  LAST UPDATED
security-policy     1.2.0          2026-02-26 14:30
data-access         2.0.1          2026-02-25 09:15
compliance          1.5.3          2026-02-24 16:45
```

#### View Policy History

```bash
$ unagnt policy versions security-policy

VERSION  AUTHOR                   CREATED           ACTIVE  CHANGELOG
1.2.0    security@company.com    2026-02-26 14:30  ✓      Added PII protection rules
1.1.0    security@company.com    2026-02-20 10:00         Updated risk thresholds
1.0.0    security@company.com    2026-02-15 09:00         Initial version
```

#### Activate a Specific Version

```bash
$ unagnt policy activate security-policy 1.1.0
✓ Activated security-policy@1.1.0
```

This immediately switches the runtime to use version 1.1.0.

### Rollback

To rollback to a previous version:

```bash
# View history
$ unagnt policy versions security-policy

# Rollback to previous version
$ unagnt policy activate security-policy 1.1.0
```

## Best Practices

### 1. Semantic Versioning

Follow semantic versioning principles:

- **Major (X.0.0)**: Breaking changes that require agent updates
- **Minor (1.X.0)**: New rules or features, backward compatible
- **Patch (1.0.X)**: Bug fixes, clarifications, no functional changes

Example:
```
1.0.0 → 1.1.0  Adding new allow rule (minor)
1.1.0 → 2.0.0  Changing default from allow to deny (major)
2.0.0 → 2.0.1  Fixing typo in deny reason (patch)
```

### 2. Meaningful Changelogs

Write clear, actionable changelogs:

**Good:**
```yaml
version: 1.2.0
changelog: "Added PII protection: blocks read_file on paths containing 'pii' or 'customers'"
```

**Bad:**
```yaml
version: 1.2.0
changelog: "Updated policy"
```

### 3. Test Before Activating

Always test a new policy version before activating:

```bash
# Test the policy
$ unagnt policy test security-policy --file tests.yaml

# Simulate against historical runs
$ unagnt policy simulate security-policy 2.0.0 --run run-123

# If tests pass, activate
$ unagnt policy activate security-policy 2.0.0
```

### 4. Gradual Rollout

For critical policies, use shadow mode:

```go
simulator.Simulate(ctx, policy.SimulationRequest{
    PolicyName:    "security-policy",
    PolicyVersion: "2.0.0",
    Mode:          policy.SimulationModeShadow,
    // Runs alongside production without affecting it
})
```

Monitor the shadow results before full activation.

### 5. Maintain History

Never delete policy versions. They provide:
- Audit trail for compliance
- Rollback capability
- Historical analysis

## Policy Format

Policies are stored in YAML or JSON:

```yaml
apiVersion: policy/v2
kind: Policy
metadata:
  name: security-policy
  version: 1.2.0
  changelog:
    - version: 1.2.0
      date: 2026-02-26
      author: security@company.com
      changes: "Added PII protection rules"
    - version: 1.1.0
      date: 2026-02-20
      changes: "Updated risk thresholds"
      
spec:
  defaultEffect: allow  # or deny
  mode: enforcing       # or permissive, audit
  
rules:
  - id: block-pii-access
    tool: read_file
    effect: deny
    condition: "input.path.contains('pii')"
    reason: "PII access requires approval"
    riskScore: 0.9
    requireApproval: true
```

## Version Storage

Versions are stored in two places:

1. **Database**: Metadata (version, author, timestamps, active flag)
2. **Filesystem**: Full policy content (in `policies/` directory)

This dual storage provides:
- Fast metadata queries
- Version control friendly policy content
- Easy backup and restore

## Compliance & Audit

All policy changes are tracked:

```sql
SELECT * FROM policy_versions 
WHERE policy_name = 'security-policy'
ORDER BY created_at DESC;
```

For compliance audits:

```bash
# Export all versions
$ unagnt policy versions security-policy --format json > audit.json

# Show who changed what when
$ unagnt policy versions security-policy
```

## API Reference

### Store API

```go
// Save a new version
SaveVersion(ctx, version) error

// Get specific version
GetVersion(ctx, name, version) (*PolicyVersion, error)

// Get active version
GetActiveVersion(ctx, name) (*PolicyVersion, error)

// List all versions
ListVersions(ctx, name) ([]PolicyVersionMetadata, error)

// Activate version
SetActiveVersion(ctx, name, version) error

// List all policies
ListPolicies(ctx) ([]string, error)
```

### Version Structure

```go
type PolicyVersion struct {
    ID          string
    PolicyName  string
    Version     string
    Content     []byte
    Format      string  // "yaml" or "json"
    Author      string
    Changelog   string
    EffectiveAt time.Time
    CreatedAt   time.Time
    Supersedes  string  // Previous version ID
    Active      bool
    Metadata    map[string]string
}
```

## Troubleshooting

### Version Not Found

```bash
$ unagnt policy activate security-policy 1.5.0
Error: policy version not found: security-policy@1.5.0
```

Check available versions:
```bash
$ unagnt policy versions security-policy
```

### Cannot Activate

Ensure the version exists:
```bash
$ unagnt policy versions security-policy | grep 1.5.0
```

### Rollback Issues

If rollback fails, check policy audit logs:
```bash
$ unagnt policy audit query --policy security-policy --limit 100
```

## Examples

### Example 1: Emergency Rollback

```bash
# Production issue detected
$ unagnt policy activate security-policy 1.1.0

# Verify rollback
$ unagnt policy list | grep security-policy
security-policy     1.1.0          2026-02-26 15:45

# Check impact
$ unagnt policy audit stats --days 1
```

### Example 2: Gradual Migration

```bash
# Step 1: Test new version
$ unagnt policy test security-policy-v2 --file tests.yaml

# Step 2: Shadow test
$ unagnt policy simulate security-policy 2.0.0 \
    --mode shadow --duration 1h

# Step 3: Analyze shadow results
$ unagnt policy audit query --policy security-policy \
    --decision deny --limit 100

# Step 4: Activate if safe
$ unagnt policy activate security-policy 2.0.0
```

### Example 3: Compliance Audit

```bash
# Generate compliance report
$ unagnt policy versions security-policy --format json \
    > compliance/policy-history-$(date +%Y%m%d).json

# Export decisions
$ unagnt policy audit export \
    --days 90 \
    --format csv \
    --output compliance/audit-q1-2026.csv
```

## See Also

- [Policy Testing Guide](./policy-testing.md)
- [Policy Simulation Guide](./policy-simulation.md)
- [Audit & Compliance Guide](./policy-audit.md)
