# CLI Assistant Example

AI-powered terminal assistant using Agent Runtime.

## Features

- Natural language to shell command translation
- Safe command execution with approval gates
- Command explanation and documentation
- History analysis

## Setup

```bash
cd examples/cli-assistant

# Set API key
export OPENAI_API_KEY=sk-...

# Run the assistant
../../bin/unagnt run --config agent.yaml --goal "List all files in current directory"
```

## Example Goals

```bash
# File operations
unagnt run --config agent.yaml --goal "Find all .go files modified in the last week"

# System info
unagnt run --config agent.yaml --goal "Show disk usage for /home"

# Network
unagnt run --config agent.yaml --goal "Check if port 8080 is open"

# Git operations
unagnt run --config agent.yaml --goal "Show git commits from last 24 hours"
```

## Safety Features

The policy.yaml enforces:
- **Blocked**: Destructive commands (`rm -rf`, `dd`, `mkfs`)
- **Approval Required**: Sudo commands, network commands
- **Auto-Approved**: Read-only commands

## Architecture

```
User Question
    ↓
Agent Planning (GPT-4o-mini)
    ↓
Tool Selection
    ↓
Policy Check (approval gate)
    ↓
Command Execution (if approved)
    ↓
Result to User
```

## Policy Configuration

See `policy.yaml` for rules. Customize for your security requirements.

## Extending

Add custom tools for:
- File system operations
- Git commands
- Docker management
- System monitoring

## Security Notes

- Never run untrusted agents with unrestricted autonomy
- Review all sudo commands before approval
- Logs are stored in `agent.log` for audit

## License

MIT
