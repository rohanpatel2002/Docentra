package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-document-assistant/internal/models"
	"ai-document-assistant/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(user *models.User) error {
	args := m.Called(user)
	if args.Error(0) == nil {
		user.ID = 1
	}
	return args.Error(0)
}
func (m *MockUserRepository) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) GetUserByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		payload        map[string]interface{}
		setupMock      func(*MockUserRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success - Valid Registration",
			payload: map[string]interface{}{
				"email":    "test@example.com",
				"password": "praalaktech123",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("GetUserByEmail", "test@example.com").Return((*models.User)(nil), errors.New("record not found"))
				m.On("CreateUser", mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "User registered successfully",
		},
		{
			name: "Failure - Existing Email",
			payload: map[string]interface{}{
				"email":    "existing@example.com",
				"password": "praalaktech123",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("GetUserByEmail", "existing@example.com").Return(&models.User{ID: 1, Email: "existing@example.com"}, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   "Email is already in use",
		},
		{
			name: "Failure - Weak Password",
			payload: map[string]interface{}{
				"email":    "test@example.com",
				"password": "short", // Less than 8 chars
			},
			setupMock: func(m *MockUserRepository) {
				// Should not reach DB calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Password must be at least 8 characters long",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)
			authHandler := NewAuthHandler(mockRepo)
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			authHandler.Register(w, req)
			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.expectedStatus, res.StatusCode)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
			mockRepo.AssertExpectations(t)
		})
	}
}
func TestAuthHandler_Login(t *testing.T) {
	validPassword := "securepassword123"
	hashedPassword, err := utils.HashPassword(validPassword)
	assert.NoError(t, err)
	validUser := &models.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
	}
	tests := []struct {
		name           string
		payload        map[string]interface{}
		setupMock      func(*MockUserRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success - Valid Credentials",
			payload: map[string]interface{}{
				"email":    "test@example.com",
				"password": validPassword,
			},
			setupMock: func(m *MockUserRepository) {
				// Lookup successful
				m.On("GetUserByEmail", "test@example.com").Return(validUser, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Login successful",
		},
		{
			name: "Failure - Non existent User",
			payload: map[string]interface{}{
				"email":    "unknown@example.com",
				"password": validPassword,
			},
			setupMock: func(m *MockUserRepository) {
				m.On("GetUserByEmail", "unknown@example.com").Return((*models.User)(nil), errors.New("record not found"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid email or password", // Prevents guessing if accounts exist
		},
		{
			name: "Failure - Wrong Password",
			payload: map[string]interface{}{
				"email":    "test@example.com",
				"password": "wrongpassword!",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("GetUserByEmail", "test@example.com").Return(validUser, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid email or password",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)
			authHandler := NewAuthHandler(mockRepo)
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			authHandler.Login(w, req)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.expectedStatus, res.StatusCode)
			assert.Contains(t, w.Body.String(), tt.expectedBody)

			mockRepo.AssertExpectations(t)
		})
	}
}
