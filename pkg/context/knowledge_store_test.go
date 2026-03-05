package context

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/NikoSokratous/unagnt/pkg/llm/openai"
	"github.com/NikoSokratous/unagnt/pkg/memory"
)

func TestKnowledgeStoreIngest(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	// Create temp directory with test documents
	tmpDir := t.TempDir()

	testDoc1 := filepath.Join(tmpDir, "test1.md")
	err := os.WriteFile(testDoc1, []byte("# Test Document 1\n\nThis is a test document for knowledge base ingestion."), 0644)
	if err != nil {
		t.Fatalf("failed to write test document: %v", err)
	}

	testDoc2 := filepath.Join(tmpDir, "test2.txt")
	err = os.WriteFile(testDoc2, []byte("This is another test document with different content."), 0644)
	if err != nil {
		t.Fatalf("failed to write test document: %v", err)
	}

	// Create knowledge store
	semanticStore := memory.NewInMemorySemanticStore()
	embeddingProvider := openai.NewEmbeddingClient(apiKey, "text-embedding-3-small")
	store := NewKnowledgeStore(semanticStore, embeddingProvider)

	// Ingest directory
	err = store.IngestDirectory(context.Background(), tmpDir, "test-source")
	if err != nil {
		t.Fatalf("failed to ingest directory: %v", err)
	}

	// Verify documents were ingested
	docs := store.List()
	if len(docs) != 2 {
		t.Errorf("expected 2 documents, got %d", len(docs))
	}

	// Test search
	results, err := store.Search(context.Background(), "test document", 5)
	if err != nil {
		t.Fatalf("failed to search: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected search results, got none")
	}
}

func TestKnowledgeStoreChunking(t *testing.T) {
	semanticStore := memory.NewInMemorySemanticStore()
	embeddingProvider := openai.NewEmbeddingClient("test-key", "text-embedding-3-small")
	store := NewKnowledgeStore(semanticStore, embeddingProvider)

	// Test small document (should be single chunk)
	smallDoc := &Document{
		ID:      "small.md",
		Source:  "test",
		Content: "Short content",
		Metadata: map[string]any{
			"file_path": "small.md",
		},
	}

	chunks := store.chunkDocument(smallDoc)
	if len(chunks) != 1 {
		t.Errorf("expected 1 chunk for small document, got %d", len(chunks))
	}

	// Test large document (should be multiple chunks)
	// Use \n\n so chunkDocument splits at paragraph boundaries (ChunkSize=500 tokens ≈ 2000 chars)
	largeContent := ""
	for i := 0; i < 100; i++ {
		largeContent += "This is a paragraph with some content.\n\n"
	}

	largeDoc := &Document{
		ID:      "large.md",
		Source:  "test",
		Content: largeContent,
		Metadata: map[string]any{
			"file_path": "large.md",
		},
	}

	chunks = store.chunkDocument(largeDoc)
	if len(chunks) <= 1 {
		t.Errorf("expected multiple chunks for large document, got %d", len(chunks))
	}

	// Verify chunk metadata
	for i, chunk := range chunks {
		if chunk.ID == "" {
			t.Errorf("chunk %d has empty ID", i)
		}
		if chunk.Metadata["doc_id"] != "large.md" {
			t.Errorf("chunk %d has wrong doc_id", i)
		}
		if chunk.Metadata["chunk_idx"] != i {
			t.Errorf("chunk %d has wrong chunk_idx", i)
		}
	}
}

func TestKnowledgeStoreOperations(t *testing.T) {
	semanticStore := memory.NewInMemorySemanticStore()
	embeddingProvider := openai.NewEmbeddingClient("test-key", "text-embedding-3-small")
	store := NewKnowledgeStore(semanticStore, embeddingProvider)

	// Test list on empty store
	docs := store.List()
	if len(docs) != 0 {
		t.Errorf("expected 0 documents in empty store, got %d", len(docs))
	}

	// Add a document manually
	doc := &Document{
		ID:      "test.md",
		Source:  "test",
		Content: "test content",
		Metadata: map[string]any{
			"file_path": "test.md",
		},
	}
	store.Documents["test.md"] = doc

	// Test get document
	retrieved, ok := store.GetDocument("test.md")
	if !ok {
		t.Error("expected to retrieve document")
	}
	if retrieved.ID != "test.md" {
		t.Errorf("expected doc ID 'test.md', got '%s'", retrieved.ID)
	}

	// Test list
	docs = store.List()
	if len(docs) != 1 {
		t.Errorf("expected 1 document, got %d", len(docs))
	}

	// Test clear
	err := store.Clear()
	if err != nil {
		t.Fatalf("failed to clear store: %v", err)
	}

	docs = store.List()
	if len(docs) != 0 {
		t.Errorf("expected 0 documents after clear, got %d", len(docs))
	}
}
