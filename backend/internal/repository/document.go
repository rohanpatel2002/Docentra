package repository

import (
	"ai-document-assistant/internal/models"

	pgvector "github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// Defines the secure database interface for documents
type SearchResult struct {
	Content      string  `json:"content"`
	DocumentID   uint    `json:"document_id"`
	OriginalName string  `json:"original_name"`
	Distance     float64 `json:"distance"`
}

type DocumentRepository interface {
	CreateDocument(doc *models.Document) error
	GetDocumentByID(id uint, userID uint) (*models.Document, error)
	UpdateDocument(doc *models.Document) error
	CreateDocumentChunk(chunk *models.DocumentChunk) error
	CreateDocumentChunks(chunks []models.DocumentChunk) error
	SearchSimilarChunks(userID uint, queryVector pgvector.Vector, limit int) ([]SearchResult, error)
}
type documentRepository struct {
	db *gorm.DB
}

// Document repository instance
func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{db: db}
}

// Inserts a new Document record tracked against a specific User
func (r *documentRepository) CreateDocument(doc *models.Document) error {
	return r.db.Create(doc).Error
}

func (r *documentRepository) UpdateDocument(doc *models.Document) error {
	return r.db.Save(doc).Error
}

func (r *documentRepository) CreateDocumentChunk(chunk *models.DocumentChunk) error {
	return r.db.Create(chunk).Error
}

func (r *documentRepository) CreateDocumentChunks(chunks []models.DocumentChunk) error {
	return r.db.Create(&chunks).Error
}

func (r *documentRepository) GetDocumentByID(id uint, userID uint) (*models.Document, error) {
	var doc models.Document
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&doc).Error
	return &doc, err
}

func (r *documentRepository) SearchSimilarChunks(userID uint, queryVector pgvector.Vector, limit int) ([]SearchResult, error) {
	var results []SearchResult
	err := r.db.Table("document_chunks").
		Select("document_chunks.content, document_chunks.document_id, documents.original_name, (document_chunks.embedding <=> ?) AS distance", queryVector).
		Joins("JOIN documents ON documents.id = document_chunks.document_id").
		Where("document_chunks.user_id = ?", userID).
		Order("distance ASC").
		Limit(limit).
		Scan(&results).Error

	return results, err
}
