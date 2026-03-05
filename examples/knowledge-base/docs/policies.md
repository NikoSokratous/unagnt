# Policy Configuration Guide

## Overview

Policies in Unagnt define constraints and rules that agents must follow during execution. They provide safety, compliance, and governance controls.

## Policy Types

### 1. Tool Constraints

Restrict which tools an agent can use:

```yaml
policy:
  name: restricted_tools
  rules:
    - type: tool_constraint
      config:
        allowed: ["search_docs", "read_file"]
        denied: ["delete_file", "execute_command"]
```

### 2. Approval Requirements

Require human approval for certain actions:

```yaml
policy:
  rules:
    - type: approval_required
      config:
        conditions:
          - tool: delete_file
          - cost_threshold: 10.00
          - data_sensitivity: high
```

### 3. Rate Limits

Limit execution frequency:

```yaml
policy:
  rules:
    - type: rate_limit
      config:
        max_executions: 100
        time_window: 1h
        per_tool: true
```

### 4. Data Access

Control data access:

```yaml
policy:
  rules:
    - type: data_access
      config:
        allow_pii: false
        allow_financial: false
        allowed_scopes: ["public", "internal"]
```

## Creating Policies

### Using YAML

Create a policy file `policy.yaml`:

```yaml
name: production_policy
version: v1
description: Safety policy for production agents

rules:
  - type: tool_constraint
    priority: 1
    config:
      denied: ["rm", "delete", "drop"]
  
  - type: approval_required
    priority: 2
    config:
      conditions:
        - cost_threshold: 5.00
        - action_type: destructive
  
  - type: rate_limit
    priority: 3
    config:
      max_executions: 50
      time_window: 1h
```

Apply the policy:

```bash
unagnt policy create policy.yaml
```

### Using CLI

Create policy interactively:

```bash
unagnt policy create --name my_policy --interactive
```

### Programmatically

```go
import "github.com/Unagnt/Unagnt/pkg/policy"

engine := policy.NewEngine()

// Add tool constraint
engine.AddConstraint(policy.ToolConstraint{
    Denied: []string{"delete_file", "execute_command"},
})

// Add approval requirement
engine.AddConstraint(policy.ApprovalRequired{
    CostThreshold: 10.00,
})
```

## Policy Rules

### Tool Constraints

Control tool access:

```yaml
- type: tool_constraint
  config:
    # Whitelist approach
    allowed: ["search", "read"]
    
    # Blacklist approach
    denied: ["delete", "modify"]
    
    # Pattern matching
    denied_patterns: ["*_admin", "delete_*"]
```

### Approval Rules

Require approval for actions:

```yaml
- type: approval_required
  config:
    conditions:
      - tool: send_email
        recipients: external
      
      - cost_threshold: 5.00
      
      - data_sensitivity: high
      
      - custom: |
          return input.amount > 1000
```

### Rate Limits

Prevent excessive execution:

```yaml
- type: rate_limit
  config:
    # Global limit
    max_executions: 100
    time_window: 1h
    
    # Per-tool limit
    per_tool: true
    tool_limits:
      send_email: 10
      api_call: 50
```

### Cost Controls

Manage API costs:

```yaml
- type: cost_control
  config:
    max_daily_cost: 50.00
    max_per_run: 5.00
    alert_threshold: 40.00
```

## Attaching Policies

### To Agents

In `agent.yaml`:

```yaml
name: support_agent
policy: production_policy
```

Or specify inline:

```yaml
name: support_agent
policy:
  inline: true
  rules:
    - type: tool_constraint
      config:
        denied: ["delete_*"]
```

### To Workflows

```yaml
workflow:
  name: support_workflow
  policy: workflow_policy
  steps:
    - name: classify
      policy: classify_policy  # Step-specific policy
```

### Dynamically

```go
agent := runtime.NewAgent(config)
agent.AttachPolicy(policyEngine)
```

## Policy Enforcement

### Evaluation

Policies are evaluated before each tool execution:

```
Agent decides to use tool
      ↓
Policy Engine evaluates constraints
      ↓
If allowed: Execute tool
If denied: Return error
If approval needed: Request approval
```

### Approval Workflows

When approval is required:

1. Agent action is paused
2. Approval request sent to configured channel (email, webhook, UI)
3. Approver reviews request
4. Approval granted or denied
5. Agent continues or aborts

Configure approval channels:

```yaml
policy:
  approval:
    channels:
      - type: webhook
        url: https://approval-service.com/webhook
      
      - type: email
        to: approvers@company.com
      
      - type: slack
        channel: "#agent-approvals"
```

## Testing Policies

### Dry Run

Test policy without executing:

```bash
unagnt policy test policy.yaml --scenario scenario.yaml
```

### Validation

Validate policy syntax:

```bash
unagnt policy validate policy.yaml
```

### Simulation

Simulate agent behavior with policy:

```bash
unagnt simulate --agent agent.yaml --policy policy.yaml
```

## Policy Inheritance

Policies can inherit from parent policies:

```yaml
name: strict_policy
extends: base_policy
rules:
  - type: tool_constraint
    config:
      denied: ["*_admin"]
```

This adds rules to those in `base_policy`.

## Conditional Policies

Apply policies based on conditions:

```yaml
policy:
  rules:
    - type: conditional
      condition: "env == 'production'"
      then:
        - type: approval_required
          config:
            cost_threshold: 1.00
      else:
        - type: approval_required
          config:
            cost_threshold: 10.00
```

## Monitoring Policy Violations

### Logging

All violations are logged:

```bash
unagnt policy violations --agent my-agent
```

### Metrics

Track violation metrics:

```bash
unagnt metrics --type policy_violations
```

### Alerts

Configure alerts:

```yaml
policy:
  alerts:
    on_violation:
      - type: email
        to: security@company.com
      - type: slack
        channel: "#security-alerts"
```

## Best Practices

1. **Start permissive**: Begin with loose policies and tighten over time
2. **Test thoroughly**: Use dry-run mode before deploying
3. **Document rules**: Add descriptions to policy rules
4. **Version policies**: Use version control for policy files
5. **Monitor violations**: Set up alerts and review regularly
6. **Layer policies**: Use multiple policies for different concerns
7. **Provide feedback**: Give clear error messages when policies block actions

## Example Policies

### Development Policy

```yaml
name: dev_policy
description: Permissive policy for development
rules:
  - type: tool_constraint
    config:
      denied: ["delete_prod_*"]
  
  - type: rate_limit
    config:
      max_executions: 1000
      time_window: 1h
```

### Production Policy

```yaml
name: prod_policy
description: Strict policy for production
rules:
  - type: tool_constraint
    config:
      denied: ["*_admin", "delete_*", "execute_*"]
  
  - type: approval_required
    config:
      conditions:
        - cost_threshold: 1.00
        - data_sensitivity: medium
  
  - type: rate_limit
    config:
      max_executions: 100
      time_window: 1h
  
  - type: cost_control
    config:
      max_daily_cost: 10.00
```

### Compliance Policy

```yaml
name: compliance_policy
description: Policy for regulated environments
rules:
  - type: data_access
    config:
      allow_pii: false
      allow_financial: false
      audit_all_access: true
  
  - type: approval_required
    config:
      conditions:
        - data_type: pii
        - data_type: financial
        - action_type: export
  
  - type: audit
    config:
      log_all_actions: true
      retention_days: 365
```

## Troubleshooting

### Policy not enforcing

Check policy is attached:
```bash
unagnt agent describe my-agent --show-policy
```

### Unexpected denials

Explain policy evaluation:
```bash
unagnt policy explain --agent my-agent --action delete_file
```

### Approval not working

Check approval channel configuration and connectivity.

## Next Steps

- Review example policies in `examples/policies/`
- Read about policy engine architecture in docs
- Set up approval workflows
- Configure monitoring and alerts
