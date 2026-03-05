# Research Assistant Example

Multi-step research agent that gathers information from the web, deduplicates using semantic memory, and generates comprehensive reports.

## Features

- Web scraping with intelligent URL queue management
- Semantic memory for deduplication
- Citation tracking
- Multi-source aggregation
- Markdown report generation

## Setup

```bash
cd examples/research-assistant

# Set API key
export OPENAI_API_KEY=sk-...

# Run research task
../../bin/unagnt run \
  --config agent.yaml \
  --goal "Research the latest developments in quantum computing" \
  --store research.db
```

## Example Research Goals

```bash
# Technology research
"Research the current state of WebAssembly adoption in 2024"

# Market research
"Analyze competitor pricing strategies for SaaS tools"

# Academic research
"Summarize recent papers on transformer architectures"

# News monitoring
"Track all mentions of 'climate policy' in major news outlets this week"
```

## How It Works

### Research Pipeline

1. **Query Planning**: Break goal into sub-questions
2. **Source Discovery**: Find relevant URLs
3. **Content Extraction**: Scrape and parse
4. **Deduplication**: Use semantic memory to skip duplicates
5. **Synthesis**: Combine findings
6. **Report Generation**: Create markdown summary with citations

### Semantic Memory

The agent uses vector embeddings to:
- Detect duplicate content across sources
- Find related past research
- Build knowledge graph

### Agent Workflow

```
Research Goal
    ↓
Plan Search Queries
    ↓
Execute Web Requests (http_request tool)
    ↓
Store in Semantic Memory
    ↓
Check for Similar Content (avoid duplicates)
    ↓
Synthesize Findings
    ↓
Generate Report
```

## Configuration

### Memory Settings

```yaml
memory:
  persistent: true   # Remember across runs
  semantic: true     # Enable vector search
```

### Autonomy

Set to `3` (Autonomous) for minimal interruptions.

## Output

Research results are stored in:
- **Working Memory**: Current session findings
- **Persistent Memory**: Key facts, sources
- **Semantic Memory**: Content embeddings
- **Logs**: Full research trail

### Accessing Results

```bash
# View logs
unagnt logs --log-file agent.log --run-id <run-id>

# Query semantic memory
unagnt memory query --agent-id research-assistant --query "quantum computing"
```

## Advanced Usage

### Multi-Day Research

Run multiple sessions, semantic memory accumulates:

```bash
# Day 1
unagnt run --config agent.yaml --goal "Research AI safety" --store research.db

# Day 2 - builds on previous research
unagnt run --config agent.yaml --goal "Find counter-arguments to AI safety concerns" --store research.db
```

### Custom Sources

Create a curated source list:

```json
{
  "sources": [
    "https://arxiv.org",
    "https://news.ycombinator.com",
    "https://scholar.google.com"
  ]
}
```

## Rate Limiting

Policy enforces:
- Max 100 HTTP requests per run
- Respects robots.txt
- Avoids admin/private paths

## Cost Estimation

Using GPT-4o:
- Simple research (5-10 sources): ~$0.10-0.30
- Comprehensive research (20-50 sources): ~$0.50-2.00
- Deep dive (100+ sources): ~$2.00-10.00

## Performance

- Semantic search: ~10ms per query
- Web scraping: 200-500ms per page
- Total research time: 2-10 minutes (typical)

## Future Enhancements

- [ ] PDF parsing
- [ ] Citation format (APA, MLA)
- [ ] Export to Notion/Confluence
- [ ] Scheduled monitoring
- [ ] Multi-language support

## Example Output

```markdown
# Research Report: Quantum Computing Developments

## Summary
Recent developments in quantum computing include...

## Key Findings
1. IBM announced 127-qubit processor [1]
2. Google achieved quantum advantage [2]
3. New error correction techniques [3]

## Sources
[1] https://research.ibm.com/... (accessed 2024-01-15)
[2] https://ai.google/quantum/... (accessed 2024-01-15)
[3] https://arxiv.org/... (accessed 2024-01-15)
```

## Troubleshooting

### Rate Limited

If hitting rate limits:
- Increase `timeout` in agent.yaml
- Reduce `max_steps`
- Use caching (persistent memory)

### Poor Quality Results

- Increase `temperature` for more creative synthesis
- Switch to GPT-4o for better analysis
- Provide more specific goals

## License

MIT
