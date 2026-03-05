package context

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/llm"
	"github.com/NikoSokratous/unagnt/pkg/memory"
)

// MemoryProvider retrieves context from memory systems.
type MemoryProvider struct {
	Manager             *memory.Manager
	priority            int
	TopK                int
	SimilarityThreshold float64
	IncludeWorking      bool
	IncludePersistent   bool
	EmbeddingProvider   llm.EmbeddingProvider
}

// NewMemoryProvider creates a new memory provider.
func NewMemoryProvider(manager *memory.Manager, priority int) *MemoryProvider {
	return &MemoryProvider{
		Manager:             manager,
		priority:            priority,
		TopK:                5,
		SimilarityThreshold: 0.7,
		IncludeWorking:      true,
		IncludePersistent:   true,
	}
}

// Name returns the provider name.
func (p *MemoryProvider) Name() string {
	return "memory"
}

// Priority returns the provider priority.
func (p *MemoryProvider) Priority() int {
	return p.priority
}

// Fetch retrieves memory context.
func (p *MemoryProvider) Fetch(ctx context.Context, input ContextInput) (*ContextFragment, error) {
	if p.Manager == nil {
		return nil, nil
	}

	parts := []string{}

	// 1. Working memory (current session context)
	if p.IncludeWorking {
		working := p.Manager.Working()
		if working != nil {
			workingData := working.All()
			if len(workingData) > 0 {
				workingStr := p.formatWorkingMemory(workingData)
				if workingStr != "" {
					parts = append(parts, "Current session context:\n"+workingStr)
				}
			}
		}
	}

	// 2. Persistent memory (facts and preferences)
	if p.IncludePersistent {
		persistent := p.Manager.Persistent()
		if persistent != nil {
			// Note: PersistentMemory interface doesn't have List method
			// This is a placeholder for future enhancement
			// For now, we skip persistent memory retrieval
		}
	}

	// 3. Semantic memory (similar past interactions)
	semantic := p.Manager.Semantic()
	if semantic != nil && input.StepInput.Goal != "" && p.EmbeddingProvider != nil {
		// Generate embedding for the goal
		embeddings, err := p.EmbeddingProvider.Embed(ctx, []string{input.StepInput.Goal})
		if err == nil && len(embeddings) > 0 {
			// Search semantic store
			results, err := semantic.Search(ctx, input.StepInput.AgentName, embeddings[0], p.TopK)
			if err == nil && len(results) > 0 {
				semanticStr := p.formatSemanticResults(results, p.SimilarityThreshold)
				if semanticStr != "" {
					parts = append(parts, "Similar past interactions:\n"+semanticStr)
				}
			}
		}
	}

	if len(parts) == 0 {
		return nil, nil
	}

	content := strings.Join(parts, "\n\n")
	tokenCount := len(content) / 4 // Simple estimate

	return &ContextFragment{
		ProviderName: p.Name(),
		Type:         FragmentTypeMemory,
		Content:      content,
		Priority:     p.priority,
		TokenCount:   tokenCount,
		Metadata:     map[string]any{"source": "memory"},
		Timestamp:    time.Now(),
	}, nil
}

// formatWorkingMemory formats working memory data.
func (p *MemoryProvider) formatWorkingMemory(data map[string]any) string {
	if len(data) == 0 {
		return ""
	}

	parts := []string{}
	for key, value := range data {
		parts = append(parts, fmt.Sprintf("  %s: %v", key, value))
	}

	return strings.Join(parts, "\n")
}

// formatSemanticResults formats semantic search results.
func (p *MemoryProvider) formatSemanticResults(results []memory.SearchResult, threshold float64) string {
	filtered := []memory.SearchResult{}
	for _, r := range results {
		if r.Score >= threshold {
			filtered = append(filtered, r)
		}
	}

	if len(filtered) == 0 {
		return ""
	}

	parts := []string{}
	for i, r := range filtered {
		content := fmt.Sprintf("  %d. [Score: %.2f]", i+1, r.Score)
		if text, ok := r.Metadata["text"].(string); ok {
			content += " " + text
		}
		parts = append(parts, content)
	}
	return strings.Join(parts, "\n")
}
