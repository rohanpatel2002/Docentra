package repository

import (
	"ai-document-assistant/internal/models"

	"gorm.io/gorm"
)

// Defines the secure database interface for documents
type DocumentRepository interface {
	CreateDocument(doc *models.Document) error
	GetDocumentByID(id uint, userID uint) (*models.Document, error)
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
func (r *documentRepository) GetDocumentByID(id uint, userID uint) (*models.Document, error) {
	var doc models.Document
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&doc).Error
	return &doc, err
}
