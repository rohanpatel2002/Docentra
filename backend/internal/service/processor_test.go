package service

import (
	"ai-document-assistant/internal/models"
	"ai-document-assistant/internal/repository"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	pgvector "github.com/pgvector/pgvector-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Handles DB
type MockDocRepo struct {
	mock.Mock
}

func (m *MockDocRepo) CreateDocument(doc *models.Document) error {
	args := m.Called(doc)
	return args.Error(0)
}
func (m *MockDocRepo) GetDocumentByID(id uint, userID uint) (*models.Document, error) {
	args := m.Called(id, userID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Document), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDocRepo) UpdateDocument(doc *models.Document) error {
	args := m.Called(doc)
	return args.Error(0)
}

func (m *MockDocRepo) CreateDocumentChunk(chunk *models.DocumentChunk) error {
	args := m.Called(chunk)
	return args.Error(0)
}

func (m *MockDocRepo) CreateDocumentChunks(chunks []models.DocumentChunk) error {
	args := m.Called(chunks)
	return args.Error(0)
}

func (m *MockDocRepo) SearchSimilarChunks(userID uint, queryVector pgvector.Vector, limit int) ([]repository.SearchResult, error) {
	args := m.Called(userID, queryVector, limit)
	if args.Get(0) != nil {
		return args.Get(0).([]repository.SearchResult), args.Error(1)
	}
	return nil, args.Error(1)
}

// Handles file access mocking
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Save(userID uint, filename string, content io.Reader) (string, error) {
	args := m.Called(userID, filename, content)
	return args.String(0), args.Error(1)
}
func (m *MockStorage) Get(path string) (io.ReadCloser, error) {
	args := m.Called(path)
	if args.Get(0) != nil {
		return args.Get(0).(io.ReadCloser), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockStorage) Delete(path string) error {
	return m.Called(path).Error(0)
}

// MockAIService is used for testing
type MockAIService struct {
	mock.Mock
}

func (m *MockAIService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	args := m.Called(ctx, text)
	if args.Get(0) != nil {
		return args.Get(0).([]float32), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAIService) GetEmbeddings(ctx context.Context, text string) ([]ChunkResult, error) {
	args := m.Called(ctx, text)
	if args.Get(0) != nil {
		return args.Get(0).([]ChunkResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestProcessor_ProcessDocument(t *testing.T) {
	t.Run("Success - Full Pipeline orchestrates correctly", func(t *testing.T) {
		repo := new(MockDocRepo)
		storage := new(MockStorage)
		aiSvc := new(MockAIService)

		proc := NewProcessor(repo, storage, aiSvc)
		doc := &models.Document{
			ID:          1,
			UserID:      42,
			MimeType:    "text/plain",
			StoragePath: "user_42/file.txt",
		}

		chunks := []ChunkResult{
			{Content: "Praalak tech sol encrypted content", Embedding: []float32{0.1, 0.2, 0.3}},
		}

		repo.On("GetDocumentByID", uint(1), uint(42)).Return(doc, nil)
		aiSvc.On("GetEmbeddings", mock.Anything, "Praalak tech sol encrypted content").Return(chunks, nil)
		repo.On("CreateDocumentChunks", mock.Anything).Return(nil)
		repo.On("UpdateDocument", mock.Anything).Return(nil)
		storage.On("Get", "user_42/file.txt").Return(io.NopCloser(strings.NewReader("Praalak tech sol encrypted content")), nil)

		err := proc.ProcessDocument(context.Background(), 1, 42)
		assert.NoError(t, err)
		assert.Equal(t, models.StatusReady, doc.Status)
		repo.AssertExpectations(t)
		storage.AssertExpectations(t)
		aiSvc.AssertExpectations(t)
	})

	t.Run("Failure - Secure Tenant Isolation Check", func(t *testing.T) {
		repo := new(MockDocRepo)
		storage := new(MockStorage)
		aiSvc := new(MockAIService)
		proc := NewProcessor(repo, storage, aiSvc)

		repo.On("GetDocumentByID", uint(1), uint(99)).Return(nil, errors.New("access denied"))

		err := proc.ProcessDocument(context.Background(), 1, 99)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "security breach attempt")
	})
}
