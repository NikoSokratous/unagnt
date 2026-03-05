# Troubleshooting Guide

## Common Issues

### Agent Not Responding

**Symptoms:**
- Agent hangs or doesn't respond
- No output after starting a run
- Timeout errors

**Possible Causes:**
1. LLM provider is down or rate-limited
2. Network connectivity issues
3. Invalid API key
4. Insufficient memory/resources

**Solutions:**

1. Check LLM provider status:
   ```bash
   curl https://status.openai.com
   ```

2. Verify API key:
   ```bash
   echo $OPENAI_API_KEY
   ```

3. Check logs:
   ```bash
   unagnt logs --agent my-agent
   ```

4. Increase timeout:
   ```yaml
   timeout: 600s  # Increase from 300s
   ```

### Context Assembly Errors

**Symptoms:**
- "Context too large" errors
- Missing expected context
- Provider fetch failures

**Solutions:**

1. Check token usage:
   ```bash
   unagnt context inspect <run-id>
   ```

2. Reduce token budgets:
   ```yaml
   assembly:
     token_budget:
       memory: 2000  # Reduce from 3000
   ```

3. Lower top-K values:
   ```yaml
   config:
     top_k: 3  # Reduce from 5
   ```

4. Disable optional providers:
   ```yaml
   - type: knowledge
     enabled: false
   ```

### Memory Issues

**Symptoms:**
- High RAM usage
- Out of memory errors
- Slow performance

**Solutions:**

1. Clear semantic memory:
   ```bash
   unagnt memory clear --semantic
   ```

2. Reduce working memory size:
   ```yaml
   memory:
     working_size: 100  # Reduce from default
   ```

3. Disable memory persistence temporarily:
   ```yaml
   memory:
     persistent: false
   ```

### Tool Execution Failures

**Symptoms:**
- Tools fail to execute
- "Tool not found" errors
- Permission denied errors

**Solutions:**

1. List available tools:
   ```bash
   unagnt tools list
   ```

2. Verify tool configuration:
   ```yaml
   tools:
     - name: search_docs
       version: "1"
   ```

3. Check tool permissions:
   ```bash
   unagnt tools check search_docs
   ```

4. Test tool directly:
   ```bash
   unagnt tools exec search_docs --input '{"query": "test"}'
   ```

### Embedding Issues

**Symptoms:**
- "Embedding generation failed"
- Semantic search returns no results
- RAG not working

**Solutions:**

1. Verify API key for OpenAI:
   ```bash
   echo $OPENAI_API_KEY
   ```

2. Test embeddings directly:
   ```bash
   unagnt context search "test query" --top-k 1
   ```

3. Switch to local embeddings:
   ```yaml
   embeddings:
     provider: local
   ```

4. Check Python dependencies (for local):
   ```bash
   pip install sentence-transformers
   python -c "from sentence_transformers import SentenceTransformer; print('OK')"
   ```

### Knowledge Base Issues

**Symptoms:**
- No documents ingested
- Search returns no results
- "Knowledge provider disabled"

**Solutions:**

1. List ingested documents:
   ```bash
   unagnt context knowledge list
   ```

2. Re-ingest documents:
   ```bash
   unagnt context ingest ./docs --source "docs"
   ```

3. Enable knowledge provider:
   ```yaml
   - type: knowledge
     enabled: true
   ```

4. Lower similarity threshold:
   ```yaml
   config:
     similarity_threshold: 0.5
   ```

## Performance Issues

### Slow Context Assembly

**Symptoms:**
- Assembly takes > 500ms
- Agent feels sluggish
- High latency

**Solutions:**

1. Enable parallel fetching:
   ```yaml
   context_assembly:
     parallel: true
   ```

2. Enable caching:
   ```yaml
   context_assembly:
     cache:
       enabled: true
       ttl: 60s
   ```

3. Profile assembly:
   ```bash
   unagnt context stats <run-id>
   ```

4. Reduce provider count:
   - Disable unused providers
   - Combine similar providers

### High API Costs

**Symptoms:**
- Unexpected high bills
- Many embedding API calls
- Excessive token usage

**Solutions:**

1. Switch to local embeddings:
   ```yaml
   embeddings:
     provider: local
   ```

2. Reduce context size:
   ```yaml
   max_context_tokens: 4000
   ```

3. Lower top-K values:
   ```yaml
   config:
     top_k: 2
   ```

4. Monitor token usage:
   ```bash
   unagnt stats --tokens
   ```

## Debugging Techniques

### Enable Debug Logging

```yaml
logging:
  level: debug
```

Or via CLI:

```bash
unagnt run --agent agent.yaml --log-level debug
```

### Inspect Context

See what context was assembled:

```bash
unagnt context inspect <run-id> --format json
```

### Explain Decisions

Understand why context was included/excluded:

```bash
unagnt context explain <run-id>
```

### Interactive Debugging

Use the debug REPL:

```bash
unagnt debug --agent agent.yaml
```

Commands:
- `context` - Show assembled context
- `context memory` - Show memory context
- `context policy` - Show policy context
- `step` - Execute next step
- `state` - Show workflow state

### Trace Execution

Enable tracing for detailed execution flow:

```yaml
observability:
  tracing:
    enabled: true
    endpoint: localhost:4318
```

View traces in Jaeger or similar tool.

## Error Messages Reference

### "API key not found"

**Cause**: Environment variable not set
**Solution**: `export OPENAI_API_KEY="sk-..."`

### "Context exceeds max tokens"

**Cause**: Assembled context is too large
**Solution**: Reduce token budgets or disable providers

### "Provider fetch timeout"

**Cause**: Provider took too long to respond
**Solution**: Increase timeout or optimize provider

### "Embedding dimension mismatch"

**Cause**: Stored embeddings have different dimensions than current model
**Solution**: Clear semantic memory and re-index

### "Tool not registered"

**Cause**: Tool not available in registry
**Solution**: Register tool or check spelling

### "Policy violation"

**Cause**: Agent attempted action blocked by policy
**Solution**: Review policy constraints or adjust permissions

## Getting Help

### Documentation

- Main docs: https://docs.Unagnt.io
- API reference: https://api.Unagnt.io/docs
- Examples: https://github.com/NikoSokratous/unagnt

### Community

- Discord: https://discord.gg/Unagnt
- GitHub Discussions: https://github.com/NikoSokratous/unagntUnagnt/discussions
- Stack Overflow: Tag `Unagnt`

### Support

- Email: support@Unagnt.io
- Enterprise support: enterprise@Unagnt.io
- Bug reports: https://github.com/NikoSokratous/unagnt/issues

## Diagnostic Commands

```bash
# Check system health
unagnt health

# Verify configuration
unagnt context validate agent.yaml

# Test connectivity
unagnt test --provider openai

# Export logs
unagnt logs export --output logs.json

# Generate diagnostic report
unagnt diagnose --output report.txt
```

## Best Practices for Troubleshooting

1. **Start simple**: Disable features one by one to isolate the issue
2. **Check logs first**: Most issues leave traces in logs
3. **Reproduce consistently**: Create minimal reproducible examples
4. **Monitor metrics**: Set up observability before issues occur
5. **Keep updated**: Use latest version with bug fixes
6. **Ask for help**: Don't hesitate to reach out to community
