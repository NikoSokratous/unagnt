# Policy Writing Guide

Learn how to write declarative policies to control agent behavior using CEL (Common Expression Language).

## Table of Contents

1. [Introduction](#introduction)
2. [Policy Structure](#policy-structure)
3. [CEL Syntax Primer](#cel-syntax-primer)
4. [Common Patterns](#common-patterns)
5. [Testing Policies](#testing-policies)
6. [Performance](#performance)
7. [Real-World Examples](#real-world-examples)

## Introduction

Policies let you declaratively control what agents can do. Instead of modifying code, you write rules that the runtime enforces automatically.

### When to Use Policies

- Prevent actions in production
- Require approval for high-risk operations
- Enforce compliance requirements
- Implement cost controls
- Manage environment-specific behavior

## Policy Structure

Basic YAML format:

```yaml
version: "1"
rules:
  - name: rule-identifier
    match:
      tool: tool_name           # Optional: specific tool
      environment: production   # Optional: specific env
      condition: "CEL expr"     # Optional: custom logic
      risk_score: ">= 0.8"      # Optional: risk threshold
    action: deny | require_approval | allow
    message: "Why this rule exists"
    approvers: [admin, security-team]
```

### Actions

- `deny` - Block immediately, no execution
- `require_approval` - Pause and wait for human approval
- `allow` - Explicitly permit (default if no rules match)

### Match Precedence

Rules are evaluated **top to bottom**. First matching rule wins.

## CEL Syntax Primer

CEL (Common Expression Language) is a simple expression language.

### Basic Operators

```javascript
// Comparison
risk_score >= 0.8
environment == "production"
tool != "echo"

// Logical
risk_score > 0.5 && environment == "production"
tool == "email" || tool == "slack"
!approved

// String operations
input.to.endsWith("@external.com")
input.subject.contains("urgent")
input.path.startsWith("/etc/")
```

### Available Variables

- `tool` (string) - Tool name being called
- `environment` (string) - Deployment environment (dev, staging, prod)
- `risk_score` (float) - 0.0 to 1.0 risk assessment
- `input` (map) - Tool input parameters
- `input.fieldname` (any) - Access nested input fields

### Type System

- `string` - Text
- `int` / `double` - Numbers
- `bool` - true/false
- `list` - Arrays
- `map` - Objects

## Common Patterns

### 1. Block Production Writes

```yaml
- name: no-prod-writes
  match:
    environment: production
    tool: db_write
  action: deny
  message: "Database writes are blocked in production"
```

### 2. High-Risk Approval

```yaml
- name: high-risk-gate
  match:
    risk_score: ">= 0.8"
  action: require_approval
  approvers: [security-team]
```

### 3. External Email Approval

```yaml
- name: external-email-check
  match:
    tool: send_email
    condition: "input.to.endsWith('@external.com')"
  action: require_approval
  approvers: [compliance]
  message: "External emails require compliance approval"
```

### 4. Time-Based Restrictions

```yaml
- name: after-hours-block
  match:
    condition: "hour(now()) >= 18 || hour(now()) < 8"
    tool: deploy
  action: deny
  message: "Deployments only allowed 8am-6pm"
```

Note: Time-based requires extended CEL functions (future enhancement).

### 5. Cost Control

```yaml
- name: expensive-model-approval
  match:
    condition: "model.name == 'gpt-4' && token_estimate > 10000"
  action: require_approval
  message: "High token usage detected"
```

### 6. Multi-Condition Rules

```yaml
- name: sensitive-prod-data
  match:
    condition: "environment == 'production' && (tool == 'db_read' || tool == 'file_read') && input.path.contains('/sensitive/')"
  action: require_approval
  approvers: [data-governance]
```

### 7. Allow-List Pattern

```yaml
# Deny by default, allow specific tools
- name: default-deny
  match:
    condition: "!(tool in ['echo', 'calc', 'http_request'])"
  action: deny
  message: "Only approved tools allowed"
```

## Testing Policies

### 1. Validate Syntax

```bash
unagnt policy validate --policy policy.yaml
```

### 2. Test with Scenarios

Create `test-scenarios.json`:

```json
[
  {
    "name": "prod write should be denied",
    "tool": "db_write",
    "environment": "production",
    "risk_score": 0.3,
    "input": {},
    "expect_deny": true,
    "expect_approval": false
  },
  {
    "name": "dev write should pass",
    "tool": "db_write",
    "environment": "development",
    "risk_score": 0.3,
    "input": {},
    "expect_deny": false,
    "expect_approval": false
  },
  {
    "name": "high risk needs approval",
    "tool": "deploy",
    "environment": "production",
    "risk_score": 0.9,
    "input": {},
    "expect_deny": false,
    "expect_approval": true
  }
]
```

Run tests:

```bash
unagnt policy validate --policy policy.yaml --test test-scenarios.json
```

### 3. Dry-Run with Agent

```bash
unagnt agent test --config agent.yaml
```

This shows which rules would trigger for your agent config.

## Performance

### Optimization Tips

1. **Order Rules by Frequency**
   - Put common rules first
   - Expensive CEL expressions last

2. **Simplify Conditions**
   ```yaml
   # Slow
   condition: "input.tags.exists(t, t.startsWith('sensitive-'))"
   
   # Fast
   tool: sensitive_tool
   ```

3. **Cache Policy Engine**
   - Load once, reuse for all executions
   - Policies are compiled on load

### Benchmarks

Typical policy check: **< 100 microseconds**

Complex CEL with nested loops: **~1-5 milliseconds**

## Real-World Examples

### Production Safety

```yaml
version: "1"
rules:
  # No prod writes without approval
  - name: prod-write-gate
    match:
      environment: production
      condition: "tool in ['db_write', 'file_write', 'deploy']"
    action: require_approval
    approvers: [ops-team]
  
  # No deletions ever in prod
  - name: prod-delete-block
    match:
      environment: production
      condition: "tool.contains('delete') || input.action == 'delete'"
    action: deny
    message: "Deletions not allowed in production"
```

### Compliance (GDPR/HIPAA)

```yaml
rules:
  # PII access requires audit
  - name: pii-access-log
    match:
      condition: "input.table in ['users', 'patients', 'customers']"
    action: require_approval
    approvers: [compliance-officer]
    message: "PII access requires compliance approval"
  
  # External data transfer blocked
  - name: data-transfer-block
    match:
      tool: http_request
      condition: "input.url.startsWith('http://') && !input.url.contains('.internal.')"
    action: deny
    message: "External data transfer prohibited"
```

### Cost Control

```yaml
rules:
  # Expensive models need approval
  - name: cost-gate-gpt4
    match:
      condition: "model.name.contains('gpt-4') && estimated_tokens > 50000"
    action: require_approval
    message: "High-cost operation detected"
  
  # Limit HTTP requests
  - name: rate-limit-http
    match:
      tool: http_request
      condition: "run.step_count > 20"
    action: deny
    message: "Too many HTTP requests in single run"
```

## Advanced Topics

### Dynamic Risk Scoring

Override default risk scorer:

```go
type CustomRiskScorer struct{}

func (s *CustomRiskScorer) Score(toolName string, input map[string]any) float64 {
    // Custom logic
    if toolName == "deploy" && input["environment"] == "production" {
        return 0.9
    }
    return 0.3
}
```

### Multi-Level Approvals

```yaml
- name: production-deploy
  match:
    tool: deploy
    environment: production
  action: require_approval
  approvers: [tech-lead, ops-lead]  # Requires both
```

### Conditional Approvers

```yaml
- name: large-spend
  match:
    condition: "input.amount > 10000"
  action: require_approval
  approvers: ["${input.department_head}"]  # Dynamic approver
```

Note: Dynamic approvers require runtime string interpolation (future feature).

## Debugging Policies

### 1. Check Rule Matches

```bash
unagnt policy validate --policy policy.yaml --verbose
```

### 2. Inspect Logs

```bash
unagnt logs --log-file agent.log --filter "policy_check"
```

### 3. Add Test Rules

Temporarily add a catch-all rule to see what's not matching:

```yaml
- name: debug-catch-all
  match:
    condition: "true"  # Matches everything
  action: require_approval
  message: "Debug: Would have allowed"
```

## Best Practices

1. **Start Permissive**: Begin with `allow`, add restrictions as needed
2. **Clear Messages**: Explain why rule exists for debugging
3. **Test Scenarios**: Maintain test JSON for regression testing
4. **Version Policies**: Commit policy.yaml to git
5. **Document Decisions**: Add comments explaining complex rules
6. **Review Regularly**: Audit policy logs to find gaps

## Troubleshooting

### Rule Not Matching

- Check rule order (first match wins)
- Validate CEL syntax
- Print `input` to see actual values
- Use `unagnt policy validate --test`

### Performance Issues

- Profile with `--verbose`
- Simplify CEL expressions
- Move expensive rules to bottom
- Cache static conditions

## Next Steps

- Read [API Integration Guide](api-integration.md)
- See [examples/](../../examples/) for complete applications
- Check CEL spec: https://github.com/google/cel-spec

## Questions?

Open an issue or ask in our community Discord.
