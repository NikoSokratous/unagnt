package context

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/NikoSokratous/unagnt/pkg/llm/openai"
	"github.com/NikoSokratous/unagnt/pkg/memory"
	"github.com/NikoSokratous/unagnt/pkg/runtime"
)

func TestSemanticSearchIntegration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	// Create memory manager
	manager := memory.NewManager("test-agent", nil, memory.NewInMemorySemanticStore())

	// Store some test interactions in semantic memory
	semanticStore := manager.Semantic()
	if semanticStore == nil {
		t.Fatal("semantic store not available")
	}

	// Create embedding provider
	embeddingProvider := openai.NewEmbeddingClient(apiKey, "text-embedding-3-small")

	// Generate embeddings for test interactions
	testInteractions := []string{
		"How do I deploy an agent to production?",
		"What are the memory configuration options?",
		"How do I configure policies for my agent?",
	}

	embeddings, err := embeddingProvider.Embed(context.Background(), testInteractions)
	if err != nil {
		t.Fatalf("failed to generate embeddings: %v", err)
	}

	// Store interactions with embeddings
	for i, interaction := range testInteractions {
		err := semanticStore.Upsert(context.Background(), "test-agent", fmt.Sprintf("interaction-%d", i), embeddings[i], map[string]any{
			"text": interaction,
		})
		if err != nil {
			t.Fatalf("failed to store interaction %d: %v", i, err)
		}
	}

	// Create memory provider with embeddings
	memProvider := NewMemoryProvider(manager, 3)
	memProvider.EmbeddingProvider = embeddingProvider
	memProvider.TopK = 3
	memProvider.SimilarityThreshold = 0.5

	// Test semantic search
	input := ContextInput{
		StepInput: runtime.StepInput{
			AgentName: "test-agent",
			Goal:      "How do I set up agent deployment?",
		},
	}

	fragment, err := memProvider.Fetch(context.Background(), input)
	if err != nil {
		t.Fatalf("failed to fetch context: %v", err)
	}

	if fragment == nil {
		t.Fatal("expected context fragment, got nil")
	}

	// Verify fragment contains semantic results
	if fragment.Content == "" {
		t.Error("expected non-empty content")
	}

	t.Logf("Semantic search results:\n%s", fragment.Content)
}

func TestMemoryProviderWithoutEmbeddings(t *testing.T) {
	// Create memory manager
	manager := memory.NewManager("test-agent", nil, memory.NewInMemorySemanticStore())

	// Add some working memory
	working := manager.Working()
	working.Set("key1", "value1")
	working.Set("key2", "value2")

	// Create memory provider without embeddings
	memProvider := NewMemoryProvider(manager, 3)
	memProvider.EmbeddingProvider = nil

	input := ContextInput{
		StepInput: runtime.StepInput{
			AgentName: "test-agent",
			Goal:      "test goal",
		},
	}

	fragment, err := memProvider.Fetch(context.Background(), input)
	if err != nil {
		t.Fatalf("failed to fetch context: %v", err)
	}

	if fragment == nil {
		t.Fatal("expected context fragment, got nil")
	}

	// Verify fragment contains working memory but not semantic results
	if fragment.Content == "" {
		t.Error("expected non-empty content")
	}

	// Should contain working memory
	if !containsString(fragment.Content, "key1") {
		t.Error("expected working memory in content")
	}

	// Should not contain semantic memory (since no embedding provider)
	if containsString(fragment.Content, "Similar past interactions") {
		t.Error("should not have semantic results without embedding provider")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
