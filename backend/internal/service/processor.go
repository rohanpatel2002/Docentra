package service

import (
	"ai-document-assistant/internal/models"
	"ai-document-assistant/internal/repository"
	"ai-document-assistant/internal/storage"
	"ai-document-assistant/pkg/parser"
	"context"
	"fmt"
	"log"

	"github.com/pgvector/pgvector-go"
)

// Processor handles the secure workflow from raw bytes to vector search ready documents
type Processor struct {
	repo      repository.DocumentRepository
	storage   storage.Provider
	extractor *parser.Extractor
	aiSvc     AIService
}

func NewProcessor(repo repository.DocumentRepository, store storage.Provider, aiSvc AIService) *Processor {
	return &Processor{
		repo:      repo,
		storage:   store,
		extractor: parser.NewExtractor(),
		aiSvc:     aiSvc,
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
	// CLI based processing via aiservice
	chunks, err := p.aiSvc.GetEmbeddings(ctx, text)
	if err != nil {
		p.updateStatus(doc, models.StatusError)
		return fmt.Errorf("AI service processing failure: %w", err)
	}
	// Prepare chunks for batch insertion
	dbChunks := make([]models.DocumentChunk, len(chunks))
	for i, c := range chunks {
		dbChunks[i] = models.DocumentChunk{
			DocumentID: docID,
			UserID:     userID,
			Content:    c.Content,
			Embedding:  pgvector.NewVector(c.Embedding),
		}
	}
	if err := p.repo.CreateDocumentChunks(dbChunks); err != nil {
		p.updateStatus(doc, models.StatusError)
		return fmt.Errorf("failed to save chunks: %w", err)
	}
	// Mark entire document as ready for search
	doc.Status = models.StatusReady
	if err := p.repo.UpdateDocument(doc); err != nil {
		return fmt.Errorf("failed to finalize document status: %w", err)
	}

	// Log success
	log.Printf("Document %d processed via Python CLI into %d chunks and stored successfully", doc.ID, len(chunks))
	return nil
}

func (p *Processor) updateStatus(doc *models.Document, status models.DocumentStatus) {
	doc.Status = status
	_ = p.repo.UpdateDocument(doc)
}
