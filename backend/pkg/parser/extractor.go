package parser

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

// Extract pulls plain text from the provided file reader safely
func (e *Extractor) Extract(reader io.Reader, mimeType string) (string, error) {
	if strings.Contains(mimeType, "text/plain") {
		return e.extractText(reader)
	}
	if strings.Contains(mimeType, "application/pdf") {
		return "", errors.New("pdf extraction pending implementation")
	}

	return "", errors.New("unsupported mime type")
}

func (e *Extractor) extractText(reader io.Reader) (string, error) {
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, reader); err != nil {
		return "", err
	}
	return buf.String(), nil
}
