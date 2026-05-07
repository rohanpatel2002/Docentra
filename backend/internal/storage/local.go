package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	// ErrInvalidFilename is returned when a filename contains path traversal characters like ".."
	ErrInvalidFilename = errors.New("invalid or insecure filename detected")
	// ErrFileNotFound is returned when the requested file does not exist on disk
	ErrFileNotFound = errors.New("file not found")
)

// Provider abstracts file storage away from the business logic
type Provider interface {
	Save(userID uint, filename string, content io.Reader) (string, error)
	Get(relativePath string) (io.ReadCloser, error)
	Delete(relativePath string) error
}

// LocalStorage implements the Provider interface using the local server filesystem.
type LocalStorage struct {
	baseDir string
}

// NewLocalStorage initializes a secure local storage manager with the given base directory.
func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	if err := os.MkdirAll(baseDir, 0750); err != nil {
		return nil, fmt.Errorf("Failed to create base storage directory: %w", err)
	}
	return &LocalStorage{baseDir: baseDir}, nil
}

// Save streams an incoming file securely to the disk, grouping by the user's ID.
func (s *LocalStorage) Save(userID uint, filename string, content io.Reader) (string, error) {
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return "", ErrInvalidFilename
	}
	userSubDir := fmt.Sprintf("user_%d", userID)
	userDirPath := filepath.Join(s.baseDir, userSubDir)
	if err := os.MkdirAll(userDirPath, 0750); err != nil {
		return "", fmt.Errorf("Failed to create user storage directory: %w", err)
	}
	fullPath := filepath.Join(userDirPath, filename)
	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if os.IsExist(err) {
			return "", fmt.Errorf("file already exists")
		}
		return "", fmt.Errorf("failed to create file descriptor: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, content); err != nil {
		return "", fmt.Errorf("failed to write file content to disk: %w", err)
	}
	return filepath.Join(userSubDir, filename), nil
}

// Get securely retrieves a stored file for reading.
func (s *LocalStorage) Get(relativePath string) (io.ReadCloser, error) {
	cleanPath := filepath.Clean(relativePath)
	if strings.Contains(cleanPath, "..") || strings.HasPrefix(cleanPath, "/") {
		return nil, ErrInvalidFilename
	}
	fullPath := filepath.Join(s.baseDir, cleanPath)
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("failed to retrieve file: %w", err)
	}
	return file, nil
}

// Delete securely wipes a file from the server's disk.
func (s *LocalStorage) Delete(relativePath string) error {
	cleanPath := filepath.Clean(relativePath)
	if strings.Contains(cleanPath, "..") || strings.HasPrefix(cleanPath, "/") {
		return ErrInvalidFilename
	}
	fullPath := filepath.Join(s.baseDir, cleanPath)
	err := os.Remove(fullPath)
	if err != nil && os.IsNotExist(err) {
		return ErrFileNotFound
	}
	return err
}
