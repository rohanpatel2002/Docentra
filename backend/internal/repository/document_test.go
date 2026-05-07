package repository

import (
	"ai-document-assistant/internal/models"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentRepository_CreateDocument(t *testing.T) {
	// Defined in user_test
	db, mock, mockDB := setupMockDB(t)
	defer mockDB.Close()
	repo := NewDocumentRepository(db)
	now := time.Now().Truncate(time.Microsecond)
	doc := &models.Document{
		UserID:       42,
		OriginalName: "praalak_tech_sol_doc.txt",
		Filename:     "uuid-1234.txt",
		MimeType:     "text/plain",
		Size:         1024,
		StoragePath:  "user_1/uuid-1234.txt",
		Status:       models.StatusUploaded,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	mock.ExpectBegin()
	// GORM Insert statement for Document
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "documents"`)).
		WithArgs(
			doc.UserID,
			doc.OriginalName,
			doc.Filename,
			doc.MimeType,
			doc.Size,
			doc.StoragePath,
			doc.Status,
			doc.CreatedAt,
			doc.UpdatedAt,
			sqlmock.AnyArg(), // deleted_at
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.CreateDocument(doc)

	assert.NoError(t, err)
	assert.Equal(t, uint(1), doc.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}
func TestDocumentRepository_GetDocumentByID(t *testing.T) {
	db, mock, mockDB := setupMockDB(t)
	defer mockDB.Close()
	repo := NewDocumentRepository(db)
	docID := uint(1)
	userID := uint(42)
	now := time.Now().Truncate(time.Microsecond)
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "original_name", "filename", "mime_type", "size", "storage_path", "status", "created_at", "updated_at", "deleted_at",
	}).AddRow(
		docID, userID, "praalak_tech_sol_doc.txt", "uuid-1234.txt", "text/plain", 1024, "user_42/uuid-1234.txt", models.StatusUploaded, now, now, nil,
	)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "documents" WHERE (id = $1 AND user_id = $2) AND "documents"."deleted_at" IS NULL ORDER BY "documents"."id" LIMIT $3`)).
		WithArgs(docID, userID, 1).
		WillReturnRows(rows)
	resDoc, err := repo.GetDocumentByID(docID, userID)
	assert.NoError(t, err)
	assert.NotNil(t, resDoc)
	assert.Equal(t, docID, resDoc.ID)
	assert.Equal(t, userID, resDoc.UserID)
	assert.Equal(t, "praalak_tech_sol_doc.txt", resDoc.OriginalName)
	require.NoError(t, mock.ExpectationsWereMet())
}
