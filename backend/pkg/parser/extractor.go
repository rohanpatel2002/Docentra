package parser

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
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
		return e.extractPDF(reader)
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

func (e *Extractor) extractPDF(reader io.Reader) (string, error) {
	// PDF extraction requires a ReadSeeker for efficient parsing
	// Since we receive a reader from storage, we buffer it into a byte seeker
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	r := bytes.NewReader(data)
	res, err := pdf.NewReader(r, int64(len(data)))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	b, err := res.GetPlainText()
	if err != nil {
		return "", err
	}

	if _, err := buf.ReadFrom(b); err != nil {
		return "", err
	}

	return buf.String(), nil
}
