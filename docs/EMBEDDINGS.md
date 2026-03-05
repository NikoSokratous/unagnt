# Embeddings and Semantic Search

This guide covers how to use embeddings for semantic search and RAG (Retrieval Augmented Generation) in Unagnt.

## Table of Contents

- [Overview](#overview)
- [Embedding Providers](#embedding-providers)
- [Configuration](#configuration)
- [Semantic Search in Memory](#semantic-search-in-memory)
- [RAG with Knowledge Base](#rag-with-knowledge-base)
- [CLI Commands](#cli-commands)
- [Performance](#performance)
- [Cost Analysis](#cost-analysis)
- [Troubleshooting](#troubleshooting)

## Overview

Embeddings are vector representations of text that capture semantic meaning. Unagnt uses embeddings for:

1. **Semantic Memory Search**: Find similar past interactions based on meaning, not just keywords
2. **RAG (Retrieval Augmented Generation)**: Retrieve relevant knowledge from documents to ground agent responses

## Embedding Providers

Unagnt supports multiple embedding providers:

### OpenAI (Recommended for Production)

- **Model**: `text-embedding-3-small`
- **Dimensions**: 1536
- **Cost**: $0.02 per 1M tokens
- **Quality**: High
- **Setup**: Requires API key

**Pros:**
- High quality embeddings
- Fast API response
- No local setup required

**Cons:**
- Requires internet connection
- Costs money (very affordable)
- API key management

### Local Embeddings (Optional)

- **Model**: `all-MiniLM-L6-v2` (sentence-transformers)
- **Dimensions**: 384
- **Cost**: Free
- **Quality**: Good
- **Setup**: Requires Python and sentence-transformers

**Pros:**
- Free
- Works offline
- No API key needed

**Cons:**
- Requires Python runtime
- Slower than API
- Lower dimensionality

## Configuration

### Enable Embeddings

Add to your `agent.yaml`:

```yaml
context_assembly:
  enabled: true
  max_context_tokens: 8000
  
  # Embedding configuration
  embeddings:
    provider: openai  # or "local" or "disabled"
    model: text-embedding-3-small
    api_key_env: OPENAI_API_KEY
```

Set your API key:

```bash
export OPENAI_API_KEY="sk-..."
```

### Configure Semantic Memory

```yaml
context_assembly:
  providers:
    - type: semantic_memory
      priority: 3
      enabled: true
      config:
        top_k: 5
        similarity_threshold: 0.7
        use_embeddings: true
```

**Parameters:**
- `top_k`: Number of similar interactions to retrieve (default: 5)
- `similarity_threshold`: Minimum cosine similarity score (0-1, default: 0.7)
- `use_embeddings`: Enable semantic search (default: false)

### Configure RAG

```yaml
context_assembly:
  providers:
    - type: knowledge
      priority: 5
      enabled: true
      config:
        sources: ["./docs", "./knowledge"]
        top_k: 3
        chunk_size: 500
        chunk_overlap: 50
```

**Parameters:**
- `sources`: Directories containing documents to ingest
- `top_k`: Number of relevant chunks to retrieve (default: 3)
- `chunk_size`: Tokens per chunk (default: 500)
- `chunk_overlap`: Token overlap between chunks (default: 50)

## Semantic Search in Memory

### How It Works

1. Agent stores interactions with embeddings in semantic memory
2. On new query, generate embedding for the goal
3. Search semantic store using cosine similarity
4. Return top-K most similar past interactions

### Example

```go
// Create memory provider with embeddings
embeddingProvider := openai.NewEmbeddingClient(apiKey, "text-embedding-3-small")
memProvider := NewMemoryProvider(manager, 3)
memProvider.EmbeddingProvider = embeddingProvider
memProvider.TopK = 5
memProvider.SimilarityThreshold = 0.7

// Fetch context (includes semantic search)
fragment, err := memProvider.Fetch(ctx, input)
```

### CLI Testing

```bash
# Search for similar past interactions
unagnt context search "how do I deploy an agent?" --top-k 5
```

## RAG with Knowledge Base

### How It Works

1. **Ingestion**: Documents are split into chunks with overlap
2. **Embedding**: Each chunk is embedded using the configured provider
3. **Indexing**: Chunks are stored in a semantic vector store
4. **Retrieval**: On query, find most relevant chunks
5. **Context**: Relevant chunks are injected into agent context

### Document Ingestion

```bash
# Ingest documents from a directory
unagnt context ingest ./docs --source "documentation"

# List ingested documents
unagnt context knowledge list

# Clear knowledge base
unagnt context knowledge clear --yes
```

### Programmatic Usage

```go
// Create knowledge store
semanticStore := manager.Semantic()
embeddingProvider := openai.NewEmbeddingClient(apiKey, "text-embedding-3-small")
knowledgeStore := NewKnowledgeStore(semanticStore, embeddingProvider)

// Ingest directory
err := knowledgeStore.IngestDirectory(ctx, "./docs", "documentation")

// Search knowledge base
chunks, err := knowledgeStore.Search(ctx, "how do I configure policies?", 3)
```

### Chunking Strategy

**Default Settings:**
- Chunk size: 500 tokens (~2000 characters)
- Overlap: 50 tokens (~200 characters)

**Why Overlap?**
- Maintains context across chunk boundaries
- Prevents information loss at splits
- Improves retrieval quality

**Paragraph Preservation:**
- Chunks are split at paragraph boundaries when possible
- Ensures semantic coherence within chunks

## CLI Commands

### Context Inspection

```bash
# Inspect assembled context
unagnt context inspect <run-id>

# Explain why each piece was included
unagnt context explain <run-id>

# Show assembly statistics
unagnt context stats <run-id>
```

### Knowledge Management

```bash
# Ingest documents
unagnt context ingest ./docs --source "docs"

# List documents
unagnt context knowledge list

# Search knowledge base
unagnt context search "query" --top-k 5

# Clear knowledge base
unagnt context knowledge clear --yes
```

### Validation

```bash
# Validate configuration
unagnt context validate agent.yaml
```

## Performance

### Embedding Generation

| Provider | Speed | Quality | Cost |
|----------|-------|---------|------|
| OpenAI API | ~100ms | High | $0.02/1M tokens |
| Local (Python) | ~300ms | Good | Free |

### Search Performance

- **In-memory semantic search**: < 50ms for 10K vectors
- **Knowledge retrieval**: < 200ms (embedding + search)
- **Total context assembly**: < 500ms with all providers

### Optimization Tips

1. **Cache embeddings**: Don't regenerate for the same text
2. **Batch API calls**: Embed multiple texts in one request
3. **Tune top-K**: Fewer results = faster search
4. **Adjust chunk size**: Smaller chunks = more precise, larger = more context

## Cost Analysis

### OpenAI Embeddings

**Pricing**: $0.02 per 1M tokens

**Example costs:**
- Single query embedding (20 tokens): $0.0000004
- 1000 queries: $0.0004
- Ingest 100 documents (500K tokens): $0.01
- Monthly (10K queries + 1000 docs): ~$0.14

**Conclusion**: Very affordable for production use

### Local Embeddings

**Cost**: Free (after initial setup)

**Requirements:**
- Python 3.8+
- sentence-transformers library
- ~100MB model download
- 2-5x slower than API

## Troubleshooting

### Issue: "OPENAI_API_KEY not set"

**Solution**: Set the environment variable:

```bash
export OPENAI_API_KEY="sk-..."
```

Or update your config to use a different env var:

```yaml
embeddings:
  api_key_env: MY_OPENAI_KEY
```

### Issue: "sentence-transformers not installed"

**Solution**: Install Python dependencies:

```bash
pip install sentence-transformers
```

### Issue: "No semantic search results"

**Possible causes:**
1. `similarity_threshold` too high (try lowering to 0.5)
2. No interactions stored in semantic memory yet
3. Embeddings not enabled in config

**Debug:**

```bash
unagnt context explain <run-id>
```

### Issue: "Knowledge provider returns no results"

**Possible causes:**
1. Documents not ingested yet
2. Query doesn't match document content
3. `top_k` too low

**Debug:**

```bash
# Check if documents are ingested
unagnt context knowledge list

# Test search directly
unagnt context search "your query" --top-k 10
```

### Issue: "Context assembly is slow"

**Solutions:**
1. Enable parallel provider fetching:
   ```yaml
   context_assembly:
     parallel: true
   ```

2. Reduce `top_k` values:
   ```yaml
   config:
     top_k: 3  # instead of 10
   ```

3. Use local embeddings for high-volume scenarios

4. Implement caching for frequently accessed documents

## Best Practices

1. **Start with OpenAI**: Easiest to set up, best quality
2. **Tune similarity threshold**: Start at 0.7, adjust based on results
3. **Keep chunks focused**: 500 tokens is a good default
4. **Use descriptive sources**: Name sources clearly for citation
5. **Monitor costs**: Track API usage if using OpenAI
6. **Test retrieval quality**: Use `unagnt context search` to validate
7. **Update knowledge regularly**: Re-ingest when documents change

## Example Configuration

Complete example with all features enabled:

```yaml
name: support_agent
version: v1.0
model:
  provider: openai
  name: gpt-4

context_assembly:
  enabled: true
  max_context_tokens: 8000
  parallel: true
  
  embeddings:
    provider: openai
    model: text-embedding-3-small
    api_key_env: OPENAI_API_KEY
  
  providers:
    - type: policy
      priority: 1
      enabled: true
    
    - type: workflow_state
      priority: 2
      enabled: true
    
    - type: semantic_memory
      priority: 3
      enabled: true
      config:
        top_k: 5
        similarity_threshold: 0.7
        use_embeddings: true
    
    - type: tool_outputs
      priority: 4
      enabled: true
      config:
        max_results: 10
    
    - type: knowledge
      priority: 5
      enabled: true
      config:
        sources: ["./docs", "./knowledge"]
        top_k: 3
        chunk_size: 500
        chunk_overlap: 50
  
  assembly:
    token_budget:
      policy: 1000
      workflow_state: 500
      memory: 3000
      tool_outputs: 2000
      knowledge: 1500
```

## Next Steps

- Read [CONTEXT_ASSEMBLY.md](./CONTEXT_ASSEMBLY.md) for context engine architecture
- See [examples/knowledge-base/](../examples/knowledge-base/) for sample setup
- Check [API documentation](./API.md) for programmatic usage
