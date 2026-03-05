# End-to-End Example: CLI Assistant

Get from zero to a working AI agent in under 2 minutes.

## 1. Install

```bash
go install github.com/NikoSokratous/unagnt/cmd/unagnt@latest
```

## 2. Set API Key

```bash
export OPENAI_API_KEY=sk-...
```

## 3. Agent Configuration

The example uses a minimal YAML config. See [examples/cli-assistant/agent.yaml](../../examples/cli-assistant/agent.yaml):

```yaml
name: cli-assistant
version: "1.0"
description: AI assistant for terminal command help

model:
  provider: openai
  name: gpt-4o-mini
  temperature: 0.3

autonomy_level: 1  # Cautious - requires approval for risky commands

tools:
  - name: echo
    version: "1"
  - name: http_request
    version: "1"

policy: ./policy.yaml
max_steps: 5
timeout: 30s
```

## 4. Run

```bash
# From the repo root after cloning
unagnt run --config examples/cli-assistant/agent.yaml --goal "List all files in current directory"
```

Or from the example directory:

```bash
cd examples/cli-assistant
unagnt run --config agent.yaml --goal "List all files in current directory"
```

## 5. Example Output

You should see output similar to:

```
[Step 1] Planning...
[Step 2] Executing tool: echo
[Step 3] Final response

The files in the current directory are:
- README.md
- agent.yaml
- policy.yaml
...
```

## 6. Approximate Cost and Duration

- **Model**: gpt-4o-mini
- **Typical cost**: ~$0.001–0.002 per run for a simple query
- **Duration**: ~5–15 seconds depending on network and model load

## Next Steps

- Try other goals: `unagnt run --config agent.yaml --goal "Show disk usage for /home"`
- Add your own tools and policies
- See [QUICKSTART.md](../../QUICKSTART.md) for server mode and workflows
