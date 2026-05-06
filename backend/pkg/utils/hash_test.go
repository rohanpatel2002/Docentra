package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	password := "Pr@alaktechs0l123"
	// Standard hashing
	hash1, err := HashPassword(password)
	assert.NoError(t, err, "HashPassword should not return an error")
	assert.NotEmpty(t, hash1, "Expected a non empty hash string")
	// Salting behavior
	hash2, err := HashPassword(password)
	assert.NoError(t, err, "HashPassword on second attempt should not fail")
	assert.NotEqual(t, hash1, hash2, "Expected different hashes for the same password due to random salting")
	// bcrypt limits
	longPassword := strings.Repeat("A", 73)
	_, err = HashPassword(longPassword)
	assert.Error(t, err, "Expected an error when hashing a password longer than 72 bytes")
}
func TestVerifyPassword(t *testing.T) {
	password := "RohanPatel!@#"
	hash, err := HashPassword(password)
	assert.NoError(t, err, "Failed to hash preparation password")
	// Table driven tests
	tests := []struct {
		name             string
		providedPassword string
		expectedResult   bool
	}{
		{"Correct password", password, true},
		{"Incorrect password", "Lakhansamani123", false},
		{"Empty password", "", false},
		{"Similar password", "RohanPatel!@# ", false}, // Extra space at the end
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyPassword(hash, tt.providedPassword)
			assert.Equal(t, tt.expectedResult, result, "VerifyPassword result mismatch")
		})
	}
	// Malformed or invalid hash
	t.Run("Malformed Hash", func(t *testing.T) {
		result := VerifyPassword("not_a_real_bcrypt_hash", password)
		assert.False(t, result, "VerifyPassword() should return false for a malformed hash")
	})
}
