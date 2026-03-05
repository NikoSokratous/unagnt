# Code Review Bot Example

Automated code reviewer that analyzes pull requests and provides feedback.

## Features

- Fetches PR diff from GitHub API
- Analyzes code quality, security, performance
- Checks for common mistakes
- Posts review comments (with approval)
- Learns from past reviews (persistent memory)

## Setup

### 1. Prerequisites

```bash
export ANTHROPIC_API_KEY=sk-ant-...
export GITHUB_TOKEN=ghp_...
```

### 2. Configure Webhook (Optional)

For automatic PR reviews:

```bash
# Add webhook to GitHub repo
# URL: https://your-unagntd.com/webhooks/code-review
# Events: Pull requests
```

### 3. Manual Run

```bash
cd examples/code-review-bot

../../bin/unagnt run \
  --config agent.yaml \
  --goal "Review PR #123 in owner/repo"
```

## Example Goals

```bash
# Review specific PR
"Review PR #456 in myorg/myrepo"

# Security audit
"Perform security audit on PR #789"

# Style check
"Check code style compliance for PR #101"
```

## Review Checklist

The bot checks for:

### Code Quality
- Naming conventions
- Function complexity
- Code duplication
- Error handling

### Security
- SQL injection risks
- XSS vulnerabilities
- Hardcoded secrets
- Unsafe dependencies

### Performance
- O(n²) algorithms
- Memory leaks
- Inefficient queries
- Large file operations

### Best Practices
- Missing tests
- Insufficient documentation
- Breaking changes without migration
- Missing error handling

## Architecture

```
GitHub Webhook
    ↓
unagntd receives event
    ↓
Agent fetches PR diff (http_request tool)
    ↓
Claude analyzes code
    ↓
Agent drafts review comments
    ↓
Policy check (approval for posting)
    ↓
Post to GitHub (if approved)
```

## Customization

### Configure Review Rules

Edit the agent system prompt to customize review criteria:

```yaml
# In agent.yaml (future feature)
system_prompt: |
  You are a code reviewer focusing on:
  - Security vulnerabilities
  - Performance issues
  - Your custom criteria here
```

### Adjust Autonomy

- **Level 1 (Cautious)**: Approve every comment
- **Level 2 (Standard)**: Auto-post non-critical comments
- **Level 3 (Autonomous)**: Fully automated reviews

## Memory Usage

The bot uses persistent memory to:
- Remember past review patterns
- Learn repository-specific conventions
- Track recurring issues

## Integration

### GitHub Actions

```yaml
name: AI Code Review
on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run review bot
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          unagnt run \
            --config examples/code-review-bot/agent.yaml \
            --goal "Review PR #${{ github.event.pull_request.number }}"
```

## Observability

View review history:

```bash
# List past reviews
unagnt logs --log-file agent.log

# Compare reviews across different model versions
unagnt diff <run-id-1> <run-id-2>
```

## Cost Estimation

Using Claude 3.5 Sonnet:
- Small PR (< 500 lines): ~$0.01-0.05
- Medium PR (500-2000 lines): ~$0.05-0.20
- Large PR (> 2000 lines): ~$0.20-0.50

## Future Enhancements

- [ ] Multi-file analysis
- [ ] AST parsing for deeper insights
- [ ] Auto-fix suggestions
- [ ] Integration with linters (eslint, pylint)
- [ ] PR summary generation

## License

MIT
