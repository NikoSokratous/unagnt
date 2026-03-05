# Knowledge Base Example

This example demonstrates how to use Unagnt's RAG (Retrieval Augmented Generation) capabilities with a knowledge base.

## Overview

The support agent in this example has access to:
- Semantic memory for past interactions
- A knowledge base ingested from markdown documents
- Context assembly that automatically retrieves relevant information

## Setup

### 1. Set OpenAI API Key

```bash
export OPENAI_API_KEY="sk-..."
```

Or use local embeddings (requires Python):

```bash
pip install sentence-transformers
```

Then update `agent.yaml`:

```yaml
embeddings:
  provider: local
  model: all-MiniLM-L6-v2
```

### 2. Ingest Documents

The `docs/` directory contains sample documentation. Ingest it:

```bash
unagnt context ingest ./docs --source "documentation"
```

This will:
- Scan all `.md` and `.txt` files
- Split them into ~500 token chunks
- Generate embeddings for each chunk
- Store in the semantic index

### 3. Verify Ingestion

```bash
unagnt context knowledge list
```

### 4. Test Search

```bash
unagnt context search "how do I configure policies?" --top-k 5
```

## Running the Agent

```bash
unagnt run --agent agent.yaml --goal "Help me understand how to deploy an agent"
```

The agent will:
1. Generate an embedding for your goal
2. Search the knowledge base for relevant docs
3. Retrieve the top 3 most relevant chunks
4. Include them in the context for the LLM
5. Generate a response grounded in your documentation

## What's in the Knowledge Base?

The `docs/` directory contains:
- **api-guide.md**: API usage and authentication
- **deployment.md**: Deployment strategies and best practices
- **troubleshooting.md**: Common issues and solutions
- **policies.md**: Policy configuration guide

## How RAG Works

1. **Ingestion** (one-time):
   ```
   Document → Chunks → Embeddings → Vector Store
   ```

2. **Retrieval** (per query):
   ```
   Query → Embedding → Semantic Search → Top-K Chunks → LLM Context
   ```

3. **Generation**:
   ```
   Context + Query → LLM → Grounded Response
   ```

## Configuration Options

### Chunk Size

Larger chunks = more context, fewer chunks:

```yaml
config:
  chunk_size: 800  # tokens
  chunk_overlap: 100
```

### Retrieval

More results = more context, higher token usage:

```yaml
config:
  top_k: 5
```

### Token Budget

Allocate more tokens to knowledge if documents are verbose:

```yaml
assembly:
  token_budget:
    knowledge: 2500
```

## Testing Different Queries

Try these queries to see RAG in action:

```bash
# Should retrieve from api-guide.md
unagnt run --agent agent.yaml --goal "How do I authenticate with the API?"

# Should retrieve from deployment.md
unagnt run --agent agent.yaml --goal "What are the deployment best practices?"

# Should retrieve from troubleshooting.md
unagnt run --agent agent.yaml --goal "My agent is failing, how do I debug?"
```

## Inspecting Context

See what was retrieved:

```bash
unagnt context inspect <run-id>
```

Look for the "knowledge" fragment to see which chunks were included.

## Adding Your Own Documents

1. Add `.md` or `.txt` files to `docs/`
2. Re-ingest:
   ```bash
   unagnt context ingest ./docs --source "documentation"
   ```
3. Test retrieval:
   ```bash
   unagnt context search "your query"
   ```

## Cost Considerations

With OpenAI embeddings:
- Ingesting 100 docs (~500K tokens): $0.01
- Each query (~20 tokens): $0.0000004
- Very affordable for production use

With local embeddings:
- Free, but 2-5x slower
- Good for development and testing

## Next Steps

- Read [EMBEDDINGS.md](../../docs/EMBEDDINGS.md) for detailed embedding guide
- Read [CONTEXT_ASSEMBLY.md](../../docs/CONTEXT_ASSEMBLY.md) for context engine architecture
- Experiment with different chunking strategies
- Try combining semantic memory + knowledge base
- Add more documents to expand knowledge coverage

## Troubleshooting

### "No results found"

- Check documents were ingested: `unagnt context knowledge list`
- Lower similarity threshold in config
- Increase `top_k`
- Try different query phrasing

### "Embeddings API error"

- Verify `OPENAI_API_KEY` is set
- Check internet connection
- Or switch to local embeddings

### "Context too large"

- Reduce `top_k` in knowledge config
- Decrease `chunk_size`
- Lower `token_budget` for knowledge
