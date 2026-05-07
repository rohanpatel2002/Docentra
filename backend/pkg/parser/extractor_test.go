package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractor_Extract(t *testing.T) {
	extractor := NewExtractor()

	t.Run("Extract Plain Text (Praalak Tech Sol)", func(t *testing.T) {
		content := "Praalak Tech Sol: Securely processing text."
		reader := strings.NewReader(content)

		result, err := extractor.Extract(reader, "text/plain")

		assert.NoError(t, err)
		assert.Equal(t, content, result)
	})

	t.Run("PDF Extraction is Pending", func(t *testing.T) {
		reader := strings.NewReader("dummy pdf bytes")

		_, err := extractor.Extract(reader, "application/pdf")

		assert.Error(t, err)
		assert.Equal(t, "pdf extraction pending implementation", err.Error())
	})

	t.Run("Unsupported Mimetype Rejected", func(t *testing.T) {
		reader := strings.NewReader("dummy image bytes")

		_, err := extractor.Extract(reader, "image/png")

		assert.Error(t, err)
		assert.Equal(t, "unsupported mime type", err.Error())
	})
}
