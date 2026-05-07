package handlers

import (
	"ai-document-assistant/internal/api/middleware"
	"ai-document-assistant/internal/models"
	"ai-document-assistant/internal/repository"
	"ai-document-assistant/internal/storage"
	"ai-document-assistant/pkg/parser"
	"ai-document-assistant/pkg/utils"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// Document uploads, utilizing abstracted storage and database layers
type DocumentHandler struct {
	docRepo repository.DocumentRepository
	storage storage.Provider
}

// Initializes the handler with injected dependencies
func NewDocumentHandler(repo repository.DocumentRepository, store storage.Provider) *DocumentHandler {
	return &DocumentHandler{
		docRepo: repo,
		storage: store,
	}
}

// Handles multipart file uploads.
func (h *DocumentHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, utils.MaxUploadSize)
	if err := r.ParseMultipartForm(utils.MaxUploadSize); err != nil {
		http.Error(w, `{"error": "File size exceeds the allowed limit or invalid request"}`, http.StatusRequestEntityTooLarge)
		return
	}
	// Extract securely verified user ID from JWT context
	userIDValue := r.Context().Value(middleware.UserIDKey)
	if userIDValue == nil {
		http.Error(w, `{"error": "Unauthorized access"}`, http.StatusUnauthorized)
		return
	}
	userID := userIDValue.(uint)
	// Fetch the file from the multipart stream
	file, header, err := r.FormFile("document")
	if err != nil {
		http.Error(w, `{"error": "Missing 'document' field in upload"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()
	// Validate File Type by Byte Sniffing (Bypassing spoofed extensions)
	mimeType, err := utils.ValidateFileType(file)
	if err != nil {
		http.Error(w, `{"error": "Invalid or dangerous file type detected"}`, http.StatusUnsupportedMediaType)
		return
	}
	// Generate a Cryptographically Secure Random Filename
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".txt" // Fallback based on text detection
	}
	secureFilename := uuid.New().String() + ext
	// Store the file via our secure abstraction layer
	storagePath, err := h.storage.Save(userID, secureFilename, file)
	if err != nil {
		http.Error(w, `{"error": "Failed to securely save file to storage"}`, http.StatusInternalServerError)
		return
	}
	// Track the documented state cleanly in our database model
	doc := models.Document{
		UserID:       userID,
		OriginalName: header.Filename,
		Filename:     secureFilename,
		MimeType:     mimeType,
		Size:         header.Size,
		StoragePath:  storagePath,
		Status:       models.StatusReady, // Updating to Ready directly for txt context
	}

	extractor := parser.NewExtractor()
	// Re-open saved file to extract text without eating RAM
	fileReader, err := h.storage.Get(storagePath)
	if err == nil {
		defer fileReader.Close()
		_, _ = extractor.Extract(fileReader, mimeType) // We can save extracted text to DB or VDB later
	}

	if err := h.docRepo.CreateDocument(&doc); err != nil {
		// Prevent ghost storage leak
		_ = h.storage.Delete(storagePath)
		http.Error(w, `{"error": "Failed to record document metadata"}`, http.StatusInternalServerError)
		return
	}
	// Send sanitized response JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Document uploaded and stored securely",
		"document_id": doc.ID,
		"filename":    doc.OriginalName,
		"status":      doc.Status,
	})
}
