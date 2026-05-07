package service

import (
	"context"
	"fmt"
	"log"

	"ai-document-assistant/internal/models"
	"ai-document-assistant/internal/repository"
	"ai-document-assistant/internal/storage"
	"ai-document-assistant/pkg/client"
	"ai-document-assistant/pkg/parser"

	"github.com/pgvector/pgvector-go"
)

// Processor handles the secure workflow from raw bytes to vector search ready documents
type Processor struct {
	repo      repository.DocumentRepository
	storage   storage.Provider
	extractor *parser.Extractor
	embedder  *client.EmbeddingClient
}

func NewProcessor(repo repository.DocumentRepository, store storage.Provider, embedder *client.EmbeddingClient) *Processor {
	return &Processor{
		repo:      repo,
		storage:   store,
		extractor: parser.NewExtractor(),
		embedder:  embedder,
	}
}

// Secure text extraction and embedding lifecycle
func (p *Processor) ProcessDocument(ctx context.Context, docID uint, userID uint) error {
	// Fetch document
	doc, err := p.repo.GetDocumentByID(docID, userID)
	if err != nil {
		return fmt.Errorf("security breach attempt or missing doc: %w", err)
	}
	// Open the file securely
	reader, err := p.storage.Get(doc.StoragePath)
	if err != nil {
		p.updateStatus(doc, models.StatusError)
		return fmt.Errorf("failed to access secure storage: %w", err)
	}
	defer reader.Close()
	// Extract purely plain
	text, err := p.extractor.Extract(reader, doc.MimeType)
	if err != nil {
		p.updateStatus(doc, models.StatusError)
		return fmt.Errorf("extraction failure: %w", err)
	}
	// Generate Ai Embeddings
	vectors, err := p.embedder.GetEmbedding(ctx, text)
	if err != nil {
		p.updateStatus(doc, models.StatusError)
		return fmt.Errorf("AI embedding failure: %w", err)
	}

	// Save embedding to DB
	float32Vectors := make([]float32, len(vectors))
	for i, v := range vectors {
		float32Vectors[i] = float32(v)
	}
	doc.Embedding = pgvector.NewVector(float32Vectors)
	doc.Status = models.StatusReady
	if err := p.repo.UpdateDocument(doc); err != nil {
		return fmt.Errorf("failed to save vectorized document: %w", err)
	}

	// Log success
	log.Printf("Document %d vectorized and stored successfully", doc.ID)
	return nil
}

func (p *Processor) updateStatus(doc *models.Document, status models.DocumentStatus) {
	doc.Status = status
	_ = p.repo.UpdateDocument(doc)
}
