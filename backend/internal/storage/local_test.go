package storage

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalStorage_Save_And_Get(t *testing.T) {
	// Temporary secure directory
	tempBaseDir := t.TempDir()
	localStorage, err := NewLocalStorage(tempBaseDir)
	require.NoError(t, err, "Failed to initialize local storage")
	// Mock user and text file
	userID := uint(1)
	secureFilename := "praalaktech.txt"
	fileBytes := []byte("This is a highly confidential document regarding project Ai Document Assistant")
	t.Run("Save Successfully", func(t *testing.T) {
		relativePath, err := localStorage.Save(userID, secureFilename, bytes.NewReader(fileBytes))
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join("user_1", secureFilename), relativePath)
		writtenBytes, err := os.ReadFile(filepath.Join(tempBaseDir, relativePath))
		require.NoError(t, err)
		assert.Equal(t, fileBytes, writtenBytes)
	})
	t.Run("Get Successfully", func(t *testing.T) {
		relativePath := filepath.Join("user_1", secureFilename)
		retrievedStream, err := localStorage.Get(relativePath)
		assert.NoError(t, err)
		defer retrievedStream.Close()
		retrievedBytes, err := io.ReadAll(retrievedStream)
		assert.NoError(t, err)
		assert.Equal(t, fileBytes, retrievedBytes)
	})
	t.Run("Deny Directory Traversal Saves", func(t *testing.T) {
		maliciousFile := "../../../etc/passwd"
		_, err := localStorage.Save(userID, maliciousFile, bytes.NewReader([]byte("hacked")))
		assert.ErrorIs(t, err, ErrInvalidFilename)
		maliciousFile2 := "\\windows\\system32\\config"
		_, err2 := localStorage.Save(userID, maliciousFile2, bytes.NewReader([]byte("hacked")))
		assert.ErrorIs(t, err2, ErrInvalidFilename)
	})
	// Test File doesnt exist
	t.Run("Get Missing File", func(t *testing.T) {
		_, err := localStorage.Get("user_1/ghostfile.pdf")
		assert.ErrorIs(t, err, ErrFileNotFound)
	})
	// Test Delete
	t.Run("Delete Successfully", func(t *testing.T) {
		relativePath := filepath.Join("user_1", secureFilename)
		err := localStorage.Delete(relativePath)
		assert.NoError(t, err)
		_, statErr := os.Stat(filepath.Join(tempBaseDir, relativePath))
		assert.True(t, os.IsNotExist(statErr), "File should be completely erased from disk")
	})
}
