package context

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/NikoSokratous/unagnt/pkg/llm"
	"github.com/NikoSokratous/unagnt/pkg/memory"
)

// KnowledgeStore manages a document corpus for RAG.
type KnowledgeStore struct {
	SemanticStore     memory.SemanticStore
	EmbeddingProvider llm.EmbeddingProvider
	Documents         map[string]*Document
	ChunkSize         int
	ChunkOverlap      int
	mu                sync.RWMutex
}

// Document represents a document in the knowledge base.
type Document struct {
	ID       string
	Source   string
	Content  string
	Metadata map[string]any
	Chunks   []DocumentChunk
}

// DocumentChunk represents a chunk of a document.
type DocumentChunk struct {
	ID        string
	Content   string
	Embedding []float32
	Metadata  map[string]any
}

// NewKnowledgeStore creates a new knowledge store.
func NewKnowledgeStore(semanticStore memory.SemanticStore, embeddingProvider llm.EmbeddingProvider) *KnowledgeStore {
	return &KnowledgeStore{
		SemanticStore:     semanticStore,
		EmbeddingProvider: embeddingProvider,
		Documents:         make(map[string]*Document),
		ChunkSize:         500,
		ChunkOverlap:      50,
	}
}

// IngestDirectory loads markdown and text files from a directory.
func (k *KnowledgeStore) IngestDirectory(ctx context.Context, dir string, source string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if source == "" {
		source = dir
	}

	// Walk directory and find .md and .txt files
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".txt" {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Create document
		docID := filepath.Base(path)
		doc := &Document{
			ID:      docID,
			Source:  source,
			Content: string(content),
			Metadata: map[string]any{
				"file_path": path,
				"file_name": docID,
			},
		}

		// Split into chunks
		chunks := k.chunkDocument(doc)
		if len(chunks) == 0 {
			return nil
		}

		// Generate embeddings for chunks
		chunkTexts := make([]string, len(chunks))
		for i, chunk := range chunks {
			chunkTexts[i] = chunk.Content
		}

		embeddings, err := k.EmbeddingProvider.Embed(ctx, chunkTexts)
		if err != nil {
			return fmt.Errorf("failed to generate embeddings for %s: %w", path, err)
		}

		// Attach embeddings to chunks
		for i := range chunks {
			chunks[i].Embedding = embeddings[i]
		}
		doc.Chunks = chunks

		// Store document
		k.Documents[docID] = doc

		// Store chunks in semantic store
		for _, chunk := range chunks {
			// Add content to metadata for retrieval
			chunkMetadata := make(map[string]any)
			for k, v := range chunk.Metadata {
				chunkMetadata[k] = v
			}
			chunkMetadata["content"] = chunk.Content
			chunkMetadata["chunk_id"] = chunk.ID

			err := k.SemanticStore.Upsert(ctx, source, chunk.ID, chunk.Embedding, chunkMetadata)
			if err != nil {
				return fmt.Errorf("failed to store chunk in semantic store: %w", err)
			}
		}

		return nil
	})
}

// chunkDocument splits a document into chunks with overlap.
func (k *KnowledgeStore) chunkDocument(doc *Document) []DocumentChunk {
	// Simple chunking by tokens (approximation: 4 chars = 1 token)
	text := doc.Content
	tokenLength := len(text) / 4

	if tokenLength <= k.ChunkSize {
		// Document is small enough, single chunk
		return []DocumentChunk{
			{
				ID:      fmt.Sprintf("%s-chunk-0", doc.ID),
				Content: text,
				Metadata: map[string]any{
					"doc_id":    doc.ID,
					"source":    doc.Source,
					"chunk_idx": 0,
					"file_path": doc.Metadata["file_path"],
				},
			},
		}
	}

	// Split into chunks with overlap
	chunks := []DocumentChunk{}
	chunkSizeChars := k.ChunkSize * 4
	overlapChars := k.ChunkOverlap * 4

	// Try to split at paragraph boundaries first
	paragraphs := strings.Split(text, "\n\n")
	currentChunk := ""
	chunkIdx := 0

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// If adding this paragraph would exceed chunk size, create a chunk
		if len(currentChunk)+len(para) > chunkSizeChars && currentChunk != "" {
			chunks = append(chunks, DocumentChunk{
				ID:      fmt.Sprintf("%s-chunk-%d", doc.ID, chunkIdx),
				Content: strings.TrimSpace(currentChunk),
				Metadata: map[string]any{
					"doc_id":    doc.ID,
					"source":    doc.Source,
					"chunk_idx": chunkIdx,
					"file_path": doc.Metadata["file_path"],
				},
			})
			chunkIdx++

			// Keep overlap from previous chunk
			if len(currentChunk) > overlapChars {
				currentChunk = currentChunk[len(currentChunk)-overlapChars:]
			} else {
				currentChunk = ""
			}
		}

		currentChunk += para + "\n\n"
	}

	// Add final chunk if any content remains
	if strings.TrimSpace(currentChunk) != "" {
		chunks = append(chunks, DocumentChunk{
			ID:      fmt.Sprintf("%s-chunk-%d", doc.ID, chunkIdx),
			Content: strings.TrimSpace(currentChunk),
			Metadata: map[string]any{
				"doc_id":    doc.ID,
				"source":    doc.Source,
				"chunk_idx": chunkIdx,
				"file_path": doc.Metadata["file_path"],
			},
		})
	}

	return chunks
}

// Search finds relevant chunks for a query.
func (k *KnowledgeStore) Search(ctx context.Context, query string, topK int) ([]DocumentChunk, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	// Generate embedding for query
	embeddings, err := k.EmbeddingProvider.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings generated for query")
	}

	// Search semantic store
	// Note: We use "knowledge" as the agent name for knowledge base searches
	results, err := k.SemanticStore.Search(ctx, "knowledge", embeddings[0], topK)
	if err != nil {
		return nil, fmt.Errorf("failed to search semantic store: %w", err)
	}

	// Convert search results to document chunks
	chunks := make([]DocumentChunk, 0, len(results))
	for _, result := range results {
		chunk := DocumentChunk{
			Metadata: result.Metadata,
		}

		// Extract content from metadata if available
		if content, ok := result.Metadata["content"].(string); ok {
			chunk.Content = content
		}
		if id, ok := result.Metadata["chunk_id"].(string); ok {
			chunk.ID = id
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// Clear removes all documents from the knowledge store.
func (k *KnowledgeStore) Clear() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	k.Documents = make(map[string]*Document)
	return nil
}

// List returns all document IDs in the knowledge store.
func (k *KnowledgeStore) List() []string {
	k.mu.RLock()
	defer k.mu.RUnlock()

	ids := make([]string, 0, len(k.Documents))
	for id := range k.Documents {
		ids = append(ids, id)
	}
	return ids
}

// GetDocument retrieves a document by ID.
func (k *KnowledgeStore) GetDocument(id string) (*Document, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	doc, ok := k.Documents[id]
	return doc, ok
}
