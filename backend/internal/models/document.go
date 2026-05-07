package models

import (
	"time"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type DocumentStatus string

const (
	StatusUploaded   DocumentStatus = "Uploaded"
	StatusProcessing DocumentStatus = "Processing"
	StatusReady      DocumentStatus = "Ready"
	StatusError      DocumentStatus = "Error"
)

type Document struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"not null;index" json:"user_id"`
	OriginalName string         `gorm:"not null" json:"original_name"`
	Filename     string         `gorm:"not null;uniqueIndex" json:"filename"`
	MimeType     string         `gorm:"not null" json:"mime_type"`
	Size         int64          `gorm:"not null" json:"size"`
	StoragePath  string         `gorm:"not null" json:"storage_path"`
	Status       DocumentStatus `gorm:"type:varchar(20);default:'uploaded'" json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

type DocumentChunk struct {
	ID         uint            `gorm:"primaryKey" json:"id"`
	DocumentID uint            `gorm:"not null;index" json:"document_id"`
	UserID     uint            `gorm:"not null;index" json:"user_id"`
	Content    string          `gorm:"type:text;not null" json:"content"`
	Embedding  pgvector.Vector `gorm:"type:vector(384)" json:"-"`
	CreatedAt  time.Time       `json:"created_at"`

	Document Document `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}
