package utils

import (
	"errors"
	"io"
	"net/http"
	"strings"
)

var (
	// Returned when an upload exceeds the globally allowed maximum.
	ErrFileTooLarge = errors.New("file size exceeds the maximum allowed limit")
	// Returned when the detected file signature doesn't match our allowed list.
	ErrInvalidMime = errors.New("file type is not permitted for upload")
	// Returned if we could not read the file bytes to determine its type.
	ErrParseMime = errors.New("could not read file to detect secure MIME type")
)

const (
	// 10 MB to prevent DDoS via excessive processing limits
	MaxUploadSize = 10 << 20
)

// Support pure text files and PDFs to prevent malicious payloads like disguised exes or html files
var AllowedMimeTypes = map[string]bool{
	"application/pdf":                true,
	"text/plain; charset=utf-8":      true,
	"text/plain; charset=iso-8859-1": true,
	"text/plain":                     true,
}

// SniffFileMimeType reads exactly the first 512 bytes of a file to securely determine its real MIME type
// using the algorithm described in the HTML5 specification (http://mimesniff.spec.whatwg.org/).
// This entirely bypasses potentially spoofed file extensions.
func SniffFileMimeType(file io.ReadSeeker) (string, error) {
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	// It's perfectly fine if the initial file is smaller than 512 bytes (io.EOF).
	if err != nil && err != io.EOF {
		return "", ErrParseMime
	}
	if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
		return "", ErrParseMime
	}
	mimeType := http.DetectContentType(buffer[:n])
	return mimeType, nil
}

// Performs both the byte sniffing and the strict comparison against allowed formats
func ValidateFileType(file io.ReadSeeker) (string, error) {
	mimeType, err := SniffFileMimeType(file)
	if err != nil {
		return "", err
	}
	if !AllowedMimeTypes[mimeType] && !strings.HasPrefix(mimeType, "text/plain") {
		return mimeType, ErrInvalidMime
	}
	return mimeType, nil
}
