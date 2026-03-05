# Policy Testing Guide

## Overview

The Unagnt policy testing framework allows you to write automated tests for your policies using YAML test files. This ensures policies behave as expected before deployment.

## Test File Format

Tests are defined in YAML files:

```yaml
name: security-policy-tests
description: Test suite for security policy
policy: security-policy
version: 1.0.0

tests:
  - name: "Test case description"
    tool: tool_name
    input:
      param1: value1
      param2: value2
    context:
      key: value
    expect:
      allowed: true
      denied: false
      reasonContains: "text"
      alert: true
      minRiskScore: 0.7
      maxRiskScore: 0.9
```

## Test Structure

### Test Suite

- `name`: Suite name (required)
- `description`: Suite description (optional)
- `policy`: Policy name to test (required)
- `version`: Policy version (optional, defaults to active)
- `tests`: Array of test cases (required)

### Test Case

- `name`: Test case name (required)
- `tool`: Tool/action being tested (required)
- `input`: Tool input parameters (required)
- `context`: Execution context (optional)
- `expect`: Expected results (required)

### Expectations

You can assert on multiple criteria:

- `allowed`: Boolean - action should be allowed
- `denied`: Boolean - action should be denied
- `reason`: Exact deny reason string
- `reasonContains`: Deny reason contains substring
- `alert`: Boolean - should trigger alert
- `minRiskScore`: Minimum risk score
- `maxRiskScore`: Maximum risk score

## Examples

### Example 1: File Access Policy Tests

```yaml
name: file-access-tests
policy: file-access-policy
version: 1.0.0

tests:
  - name: "Allow reading config files"
    tool: read_file
    input:
      path: "/config/app.yaml"
    expect:
      allowed: true
      maxRiskScore: 0.3

  - name: "Block reading sensitive files"
    tool: read_file
    input:
      path: "/data/passwords.txt"
    expect:
      denied: true
      reasonContains: "sensitive"
      minRiskScore: 0.8

  - name: "Alert on PII access"
    tool: read_file
    input:
      path: "/data/customers_pii.csv"
    expect:
      allowed: true  # Allowed but logged
      alert: true
      minRiskScore: 0.6
```

### Example 2: Network Policy Tests

```yaml
name: network-policy-tests
policy: network-policy

tests:
  - name: "Allow internal API calls"
    tool: http_request
    input:
      url: "https://internal.company.com/api/data"
      method: "GET"
    expect:
      allowed: true
      maxRiskScore: 0.2

  - name: "Block external APIs without approval"
    tool: http_request
    input:
      url: "https://external-api.com/data"
    expect:
      denied: true
      reasonContains: "external"
      minRiskScore: 0.7

  - name: "Allow external with approval"
    tool: http_request
    input:
      url: "https://external-api.com/data"
    context:
      approved_by: "admin@company.com"
      approval_id: "APR-12345"
    expect:
      allowed: true
      maxRiskScore: 0.4
```

### Example 3: Database Policy Tests

```yaml
name: database-policy-tests
policy: database-policy

tests:
  - name: "Allow SELECT queries"
    tool: sql_query
    input:
      query: "SELECT * FROM users WHERE id = ?"
      params: [123]
    expect:
      allowed: true
      maxRiskScore: 0.3

  - name: "Block DELETE without approval"
    tool: sql_query
    input:
      query: "DELETE FROM users WHERE id = ?"
      params: [123]
    expect:
      denied: true
      reasonContains: "DELETE requires approval"
      minRiskScore: 0.9

  - name: "Block DROP TABLE"
    tool: sql_query
    input:
      query: "DROP TABLE users"
    expect:
      denied: true
      reasonContains: "DROP not allowed"
      minRiskScore: 1.0
```

## Running Tests

### Via CLI

```bash
# Run tests from file
$ unagnt policy test security-policy --file tests.yaml

# Output:
security-policy-tests
Tests: 5 total, 5 passed, 0 failed, 0 skipped
Duration: 234ms

✓ Allow reading config files
✓ Block reading sensitive files
✓ Alert on PII access
✓ Block external APIs without approval
✓ Allow external with approval
```

### Via API

```go
runner := policy.NewTestRunner(simulator, store)
result, err := runner.RunTestFile(ctx, "tests.yaml")

if result.Failed > 0 {
    for _, tc := range result.TestCases {
        if tc.Status == "failed" {
            fmt.Printf("FAIL: %s - %s\n", tc.Name, tc.Message)
            fmt.Printf("  Expected: %s\n", tc.Expected)
            fmt.Printf("  Actual: %s\n", tc.Actual)
        }
    }
}
```

## Test Output

### Successful Test

```
✓ Block reading sensitive files
```

### Failed Test

```
✗ Block reading sensitive files - allowed expectation mismatch
    Expected: allowed=false
    Actual:   allowed=true
```

### Detailed Results

```go
type TestResult struct {
    TestFile   string
    TotalTests int
    Passed     int
    Failed     int
    Skipped    int
    Duration   string
    TestCases  []TestCaseResult
}
```

## Best Practices

### 1. Test Coverage

Ensure comprehensive coverage:

```yaml
tests:
  # Happy path
  - name: "Normal operation allowed"
    ...
    expect:
      allowed: true

  # Edge cases
  - name: "Boundary condition"
    ...

  # Error cases
  - name: "Invalid input denied"
    ...
    expect:
      denied: true

  # Security cases
  - name: "High risk operation blocked"
    ...
    expect:
      denied: true
      minRiskScore: 0.8
```

### 2. Descriptive Names

Use clear, descriptive test names:

**Good:**
```yaml
- name: "Block file deletion in production directory"
- name: "Allow read access to public files"
```

**Bad:**
```yaml
- name: "Test 1"
- name: "Delete test"
```

### 3. Test Organization

Group related tests:

```yaml
# File: file-access-tests.yaml
tests:
  - name: "Read operations - config files"
  - name: "Read operations - data files"
  - name: "Write operations - temp directory"
  - name: "Write operations - protected directory"
```

### 4. Context Testing

Test with different contexts:

```yaml
tests:
  - name: "Operation without approval"
    tool: delete_file
    input:
      path: "/data/important.db"
    expect:
      denied: true

  - name: "Operation with approval"
    tool: delete_file
    input:
      path: "/data/important.db"
    context:
      approved_by: "admin@company.com"
    expect:
      allowed: true
```

### 5. Risk Score Validation

Validate risk scoring:

```yaml
tests:
  - name: "Low risk operation"
    ...
    expect:
      maxRiskScore: 0.3

  - name: "High risk operation"
    ...
    expect:
      minRiskScore: 0.8
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Policy Tests

on: [push, pull_request]

jobs:
  test-policies:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Setup Go
        uses: actions/setup-go@v2
        
      - name: Build unagnt
        run: go build -o unagnt ./cmd/unagnt
        
      - name: Test policies
        run: |
          ./unagnt policy test security-policy --file tests/security_policy_test.yaml
          ./unagnt policy test data-access --file tests/data_access_test.yaml
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running policy tests..."

for test_file in tests/*_test.yaml; do
    if ! unagnt policy test $(basename $test_file _test.yaml) --file $test_file; then
        echo "Policy tests failed!"
        exit 1
    fi
done

echo "All policy tests passed ✓"
```

## Regression Testing

Test policy changes don't break existing behavior:

```bash
# Before changing policy
$ unagnt policy test security-policy --file tests.yaml > results-before.txt

# After changing policy
$ unagnt policy test security-policy --file tests.yaml > results-after.txt

# Compare
$ diff results-before.txt results-after.txt
```

## Advanced Features

### Parameterized Tests

Use YAML anchors for DRY tests:

```yaml
.common-context: &common_context
  context:
    user_id: "user-123"
    timestamp: "2026-02-26T10:00:00Z"

tests:
  - name: "Test 1"
    <<: *common_context
    tool: read_file
    input:
      path: "/file1.txt"
    expect:
      allowed: true

  - name: "Test 2"
    <<: *common_context
    tool: read_file
    input:
      path: "/file2.txt"
    expect:
      allowed: true
```

### Test Fixtures

Create reusable test data:

```yaml
setup:
  test_user: "test@example.com"
  test_path: "/tmp/test"

tests:
  - name: "Test with fixtures"
    tool: read_file
    input:
      path: "${test_path}/file.txt"
      user: "${test_user}"
    expect:
      allowed: true
```

## Troubleshooting

### Test Fails Unexpectedly

1. Check policy version:
```bash
$ unagnt policy versions security-policy
```

2. Verify policy content:
```bash
$ unagnt policy get security-policy 1.0.0
```

3. Run in verbose mode (future feature)

### Assertion Errors

Common issues:

- **Risk score mismatch**: Adjust `minRiskScore`/`maxRiskScore` ranges
- **Reason not matching**: Use `reasonContains` for partial matches
- **Context not applied**: Verify context structure matches policy expectations

## See Also

- [Policy Versioning Guide](./policy-versioning.md)
- [Policy Simulation Guide](./policy-simulation.md)
- [Writing Policy Rules](./policy-rules.md)
