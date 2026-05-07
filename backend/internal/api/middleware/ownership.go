package middleware

import (
	"ai-document-assistant/internal/repository"
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type contextKey string

const DocKey contextKey = "document"

func Ownership(repo repository.DocumentRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDKey).(uint)
			if !ok {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			docIDStr := chi.URLParam(r, "docID")
			if docIDStr == "" {
				// No ID in URL, skip check (useful for global routes)
				next.ServeHTTP(w, r)
				return
			}

			docID, err := strconv.ParseUint(docIDStr, 10, 32)
			if err != nil {
				http.Error(w, "Invalid document ID", http.StatusBadRequest)
				return
			}

			doc, err := repo.GetDocumentByID(uint(docID), userID)
			if err != nil {
				http.Error(w, "Document not found or access denied", http.StatusForbidden)
				return
			}
			ctx := context.WithValue(r.Context(), DocKey, doc)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
