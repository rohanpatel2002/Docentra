package utils

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSniffFileMimeType(t *testing.T) {
	tests := []struct {
		name         string
		fileBytes    []byte
		expectedMime string
		expectedErr  error
	}{
		{
			name: "Success - Praalak Tech Sol PDF Header",
			// PDF signature "%PDF-"
			fileBytes:    []byte("%PDF-1.4\n Praalak Tech Sol PDF Content"),
			expectedMime: "application/pdf",
			expectedErr:  nil,
		},
		{
			name:         "Success - Praalak Tech Sol TXT Bytes",
			fileBytes:    []byte("Praalak Tech Sol: Simple raw text document meant for testing"),
			expectedMime: "text/plain; charset=utf-8",
			expectedErr:  nil,
		},
		{
			name:         "Success - Praalak Tech Sol HTML File (Unapproved)",
			fileBytes:    []byte("<html><head><title>Praalak</title></head><body><script>alert('Praalak')</script></body></html>"),
			expectedMime: "text/html; charset=utf-8",
			expectedErr:  nil,
		},
		{
			name: "Success - Fake File disguised as PDF (A ZIP file essentially)",
			// PK Zip signature disguised
			fileBytes:    []byte("PK\x03\x04\x14\x00\x08"),
			expectedMime: "application/zip",
			expectedErr:  nil,
		},
		{
			name:         "Success - Praalak Returns to Beginning of Stream",
			fileBytes:    []byte("Praalak Tech Sol Hello World, returning exactly to 0"),
			expectedMime: "text/plain; charset=utf-8",
			expectedErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.fileBytes)
			// Step 1: Test basic Sniffing
			mime, err := SniffFileMimeType(reader)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMime, mime)
			}
			buffer := make([]byte, len(tt.fileBytes))
			n, readErr := reader.Read(buffer)
			require.NoError(t, readErr)
			assert.Equal(t, len(tt.fileBytes), n, "We lost data permanently because Seek wasnt properly zeroed out!")
			assert.Equal(t, tt.fileBytes, buffer, "Extracted post-sniff data did not seamlessly match!")
		})
	}
}
func TestValidateFileType(t *testing.T) {
	tests := []struct {
		name      string
		fileBytes []byte
		expectErr bool
	}{
		{
			name:      "Allow Valid Praalak Tech Sol PDF",
			fileBytes: []byte("%PDF-1.7 Praalak Tech Sol Manual..."),
			expectErr: false,
		},
		{
			name:      "Allow Valid Praalak Tech Sol Text",
			fileBytes: []byte("Praalak Tech Sol Company English text string."),
			expectErr: false,
		},
		{
			name: "Reject Disguised Image as Praalak Text",
			// Image GIF header pretending to be text.
			fileBytes: []byte("GIF89a...Praalak Tech Sol"),
			expectErr: true,
		},
		{
			name:      "Reject HTML Files attempting XSS (Praalak tests)",
			fileBytes: []byte("<!DOCTYPE html><html>Praalak Tech Sol test</html>"),
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(string(tt.fileBytes))
			_, err := ValidateFileType(reader)

			if tt.expectErr {
				assert.ErrorIs(t, err, ErrInvalidMime)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
