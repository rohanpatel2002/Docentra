package handlers

import (
	"ai-document-assistant/internal/api/middleware"
	"ai-document-assistant/internal/models"
	"ai-document-assistant/internal/repository"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	pgvector "github.com/pgvector/pgvector-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Simulates database actions for documents
type MockDocumentRepository struct {
	mock.Mock
}

func (m *MockDocumentRepository) CreateDocument(doc *models.Document) error {
	args := m.Called(doc)
	if args.Error(0) == nil {
		doc.ID = 1 // simulate auto increment
	}
	return args.Error(0)
}
func (m *MockDocumentRepository) GetDocumentByID(id uint, userID uint) (*models.Document, error) {
	args := m.Called(id, userID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Document), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDocumentRepository) UpdateDocument(doc *models.Document) error {
	args := m.Called(doc)
	return args.Error(0)
}

func (m *MockDocumentRepository) CreateDocumentChunk(chunk *models.DocumentChunk) error {
	args := m.Called(chunk)
	return args.Error(0)
}

func (m *MockDocumentRepository) CreateDocumentChunks(chunks []models.DocumentChunk) error {
	args := m.Called(chunks)
	return args.Error(0)
}

func (m *MockDocumentRepository) SearchSimilarChunks(userID uint, queryVector pgvector.Vector, limit int) ([]repository.SearchResult, error) {
	args := m.Called(userID, queryVector, limit)
	if args.Get(0) != nil {
		return args.Get(0).([]repository.SearchResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDocumentRepository) DeleteDocument(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func (m *MockDocumentRepository) GetDocumentsByUserID(userID uint) ([]models.Document, error) {
	args := m.Called(userID)
	if args.Get(0) != nil {
		return args.Get(0).([]models.Document), args.Error(1)
	}
	return nil, args.Error(1)
}

// Simulates the physical hard drive storage disk
type MockStorageProvider struct {
	mock.Mock
}

func (m *MockStorageProvider) Save(userID uint, filename string, content io.Reader) (string, error) {
	args := m.Called(userID, filename, content)
	return args.String(0), args.Error(1)
}

func (m *MockStorageProvider) Get(relativePath string) (io.ReadCloser, error) {
	args := m.Called(relativePath)
	if args.Get(0) != nil {
		return args.Get(0).(io.ReadCloser), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStorageProvider) Delete(relativePath string) error {
	args := m.Called(relativePath)
	return args.Error(0)
}
func createMultipartRequest(t *testing.T, fieldname, filename, fileContent string, userID uint) *http.Request {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	if filename != "" {
		part, err := writer.CreateFormFile(fieldname, filename)
		require.NoError(t, err)
		part.Write([]byte(fileContent))
	}
	writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/documents", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// Inject the mocked JWT authentication identity so our handler succeeds
	if userID > 0 {
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
	}
	return req
}
func TestDocumentHandler_UploadDocument(t *testing.T) {
	t.Run("Success - Valid Text File Upload (Praalak Tech Sol)", func(t *testing.T) {
		mockRepo := new(MockDocumentRepository)
		mockStorage := new(MockStorageProvider)
		handler := NewDocumentHandler(mockRepo, mockStorage, nil)
		req := createMultipartRequest(t, "document", "praalak_proposal.txt", "Praalak Tech Sol document content.", 42)
		mockStorage.On("Save", uint(42), mock.AnythingOfType("string"), mock.Anything).Return("user_42/uuid.txt", nil)
		mockRepo.On("CreateDocument", mock.AnythingOfType("*models.Document")).Return(nil)
		w := httptest.NewRecorder()
		handler.UploadDocument(w, req)
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusCreated, res.StatusCode)
		var respBody map[string]interface{}
		json.NewDecoder(res.Body).Decode(&respBody)
		assert.Equal(t, "Document uploaded and stored securely", respBody["message"])
		mockStorage.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
	t.Run("Failure - Unauthenticated User", func(t *testing.T) {
		handler := NewDocumentHandler(new(MockDocumentRepository), new(MockStorageProvider), nil)
		req := createMultipartRequest(t, "document", "missing_auth.txt", "content", 0)
		w := httptest.NewRecorder()
		handler.UploadDocument(w, req)
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})
	t.Run("Failure - Dangerous Disguised MIME Type", func(t *testing.T) {
		handler := NewDocumentHandler(new(MockDocumentRepository), new(MockStorageProvider), nil)
		// Malicious payload mimicking a PNG file
		maliciousContent := "\x89PNG\r\n\x1a\n\x00\x00"
		req := createMultipartRequest(t, "document", "evil.txt", maliciousContent, 42)
		w := httptest.NewRecorder()
		handler.UploadDocument(w, req)
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusUnsupportedMediaType, res.StatusCode)
	})
	t.Run("Failure - DB Failure Handles Storage Rollback", func(t *testing.T) {
		mockRepo := new(MockDocumentRepository)
		mockStorage := new(MockStorageProvider)
		handler := NewDocumentHandler(mockRepo, mockStorage, nil)
		req := createMultipartRequest(t, "document", "db_crash.pdf", "%PDF-1.4 Data", 42)
		// Storage writes successfully
		mockStorage.On("Save", uint(42), mock.AnythingOfType("string"), mock.Anything).Return("user_42/uuid.pdf", nil)
		// But DB immediately crashes magically!
		mockRepo.On("CreateDocument", mock.AnythingOfType("*models.Document")).Return(errors.New("db down"))
		// We expect the handler to actively hunt down the file and delete it permanently as a security rollback
		mockStorage.On("Delete", "user_42/uuid.pdf").Return(nil)
		w := httptest.NewRecorder()
		handler.UploadDocument(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
		mockStorage.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}
