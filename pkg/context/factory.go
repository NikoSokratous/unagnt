package context

import (
	"context"
	"fmt"
	"os"

	"github.com/NikoSokratous/unagnt/internal/config"
	"github.com/NikoSokratous/unagnt/pkg/llm"
	"github.com/NikoSokratous/unagnt/pkg/llm/local"
	"github.com/NikoSokratous/unagnt/pkg/llm/openai"
	"github.com/NikoSokratous/unagnt/pkg/memory"
	"github.com/NikoSokratous/unagnt/pkg/policy"
)

// NewContextEngineFromConfig creates a fully configured context engine.
func NewContextEngineFromConfig(ctx context.Context, cfg *config.ContextAssemblyConfig, memoryMgr *memory.Manager, policyEngine *policy.Engine) (*ContextEngine, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("context assembly is disabled in configuration")
	}

	// 1. Create embedding provider based on config
	var embeddingProvider llm.EmbeddingProvider
	if cfg.Embeddings.Provider != "disabled" && cfg.Embeddings.Provider != "" {
		switch cfg.Embeddings.Provider {
		case "openai":
			apiKey := os.Getenv(cfg.Embeddings.APIKeyEnv)
			if apiKey == "" {
				return nil, fmt.Errorf("OpenAI API key not found in environment variable %s", cfg.Embeddings.APIKeyEnv)
			}
			embeddingProvider = openai.NewEmbeddingClient(apiKey, cfg.Embeddings.Model)
		case "local":
			embeddingProvider = local.NewEmbeddingClient(cfg.Embeddings.Model)
		default:
			return nil, fmt.Errorf("unknown embedding provider: %s", cfg.Embeddings.Provider)
		}
	}

	// 2. Create assembler
	assembler := NewDefaultAssembler()
	if cfg.Assembly.TokenBudget != nil {
		// Token budget will be passed in AssemblyConfig during Assemble
	}

	// 3. Create engine
	engineConfig := EngineConfig{
		MaxTokens: cfg.MaxContextTokens,
		Parallel:  cfg.Parallel,
	}
	engine := NewEngine(engineConfig, assembler)

	// 4. Create providers based on config
	for _, providerConfig := range cfg.Providers {
		if !providerConfig.Enabled {
			continue
		}

		var provider ContextProvider
		switch providerConfig.Type {
		case "memory", "semantic_memory":
			if memoryMgr == nil {
				continue
			}
			memProvider := NewMemoryProvider(memoryMgr, providerConfig.Priority)

			// Configure memory provider
			if topK, ok := providerConfig.Config["top_k"].(int); ok {
				memProvider.TopK = topK
			}
			if threshold, ok := providerConfig.Config["similarity_threshold"].(float64); ok {
				memProvider.SimilarityThreshold = threshold
			}
			if includeWorking, ok := providerConfig.Config["include_working"].(bool); ok {
				memProvider.IncludeWorking = includeWorking
			}
			if includePersistent, ok := providerConfig.Config["include_persistent"].(bool); ok {
				memProvider.IncludePersistent = includePersistent
			}

			// Enable semantic search if embeddings are configured
			if useEmbeddings, ok := providerConfig.Config["use_embeddings"].(bool); ok && useEmbeddings && embeddingProvider != nil {
				memProvider.EmbeddingProvider = embeddingProvider
			}

			provider = memProvider

		case "policy":
			if policyEngine == nil {
				continue
			}
			provider = NewPolicyProvider(policyEngine, providerConfig.Priority)

		case "workflow":
			provider = NewWorkflowProvider(providerConfig.Priority)

		case "tool_output":
			maxResults := 5
			if mr, ok := providerConfig.Config["max_results"].(int); ok {
				maxResults = mr
			}
			provider = NewToolOutputProvider(providerConfig.Priority, maxResults)

		case "knowledge":
			if embeddingProvider == nil {
				continue // Knowledge provider requires embeddings
			}

			// Create knowledge store
			semanticStore := memoryMgr.Semantic()
			if semanticStore == nil {
				continue
			}

			knowledgeStore := NewKnowledgeStore(semanticStore, embeddingProvider)

			// Configure chunk size and overlap
			if chunkSize, ok := providerConfig.Config["chunk_size"].(int); ok {
				knowledgeStore.ChunkSize = chunkSize
			}
			if chunkOverlap, ok := providerConfig.Config["chunk_overlap"].(int); ok {
				knowledgeStore.ChunkOverlap = chunkOverlap
			}

			// Ingest directories if configured
			if sources, ok := providerConfig.Config["sources"].([]interface{}); ok {
				for _, s := range sources {
					if sourceDir, ok := s.(string); ok {
						if _, err := os.Stat(sourceDir); err == nil {
							err := knowledgeStore.IngestDirectory(ctx, sourceDir, sourceDir)
							if err != nil {
								// Log error but continue
								fmt.Fprintf(os.Stderr, "Warning: failed to ingest %s: %v\n", sourceDir, err)
							}
						}
					}
				}
			}

			// Create knowledge provider
			var sourcesStr []string
			if sources, ok := providerConfig.Config["sources"].([]interface{}); ok {
				for _, s := range sources {
					if sourceDir, ok := s.(string); ok {
						sourcesStr = append(sourcesStr, sourceDir)
					}
				}
			}

			knowledgeProvider := NewKnowledgeProvider(providerConfig.Priority, sourcesStr)
			knowledgeProvider.KnowledgeStore = knowledgeStore
			knowledgeProvider.SetEnabled(true)

			if topK, ok := providerConfig.Config["top_k"].(int); ok {
				knowledgeProvider.TopK = topK
			}

			provider = knowledgeProvider

		default:
			return nil, fmt.Errorf("unknown provider type: %s", providerConfig.Type)
		}

		if provider != nil {
			engine.AddProvider(provider)
		}
	}

	return engine, nil
}
