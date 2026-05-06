package utils

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestGenerateJWT(t *testing.T) {
	// Mock environment variable
	os.Setenv("JWT_SECRET", "praalak_test_key_999")
	defer os.Unsetenv("JWT_SECRET")
	userID := uint(42)
	// Basic Generation
	tokenString, err := GenerateJWT(userID)
	assert.NoError(t, err, "GenerateJWT should not return an error")

	// Format validation
	parts := strings.Split(tokenString, ".")
	assert.Len(t, parts, 3, "Expected standard JWT format with 3 parts separated by dots")
}
func TestValidateJWT(t *testing.T) {
	os.Setenv("JWT_SECRET", "praalak_test_key_999")
	defer os.Unsetenv("JWT_SECRET")
	userID := uint(42)
	validToken, err := GenerateJWT(userID)
	assert.NoError(t, err, "Failed to prepare valid token")
	// Expired token manually for testing
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		"iat":     time.Now().Unix(),
	})
	expiredTokenString, _ := expiredToken.SignedString([]byte("praalak_test_key_999"))
	// Token signed with the wrong secret
	wrongSecretToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	wrongSecretTokenString, _ := wrongSecretToken.SignedString([]byte("wrong_key"))
	// Table Driven Testing
	tests := []struct {
		name        string
		tokenString string
		expectedID  uint
		expectError bool
	}{
		{"Valid Token", validToken, 42, false},
		{"Expired Token", expiredTokenString, 0, true},
		{"Invalid Secret", wrongSecretTokenString, 0, true},
		{"Malformed String", "fake.token.string", 0, true},
		{"Empty String", "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := ValidateJWT(tt.tokenString)
			if tt.expectError {
				assert.Error(t, err, "ValidateJWT() expected error for %s", tt.name)
			} else {
				assert.NoError(t, err, "ValidateJWT() unexpected error for %s", tt.name)
				assert.Equal(t, tt.expectedID, id, "ValidateJWT() returned incorrect ID")
			}
		})
	}
}
