package handlers

import (
	"ai-document-assistant/internal/models"
	"ai-document-assistant/internal/repository"
	"ai-document-assistant/pkg/utils"
	"encoding/json"
	"net/http"
	"strings"
)

type AuthHandler struct {
	userRepo repository.UserRepository
}

func NewAuthHandler(repo repository.UserRepository) *AuthHandler {
	return &AuthHandler{userRepo: repo}
}

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register user
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	// Parse the JSON payload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}
	// Sanitization and validation
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		http.Error(w, `{"error": "Email is required"}`, http.StatusBadRequest)
		return
	}
	if len(req.Password) < 8 {
		http.Error(w, `{"error": "Password must be at least 8 characters long"}`, http.StatusBadRequest)
		return
	}
	// Prevents bcrypt overflow vulnerability
	if len(req.Password) > 72 {
		http.Error(w, `{"error": "Password exceeds maximum length"}`, http.StatusBadRequest)
		return
	}
	// Check if the user already exists to prevent multiple copies or entries with the same email
	if existingUser, err := h.userRepo.GetUserByEmail(req.Email); err == nil && existingUser.ID != 0 {
		http.Error(w, `{"error": "Email is already in use"}`, http.StatusConflict)
		return
	}
	// Hash the password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, `{"error": "Internal server error during registration"}`, http.StatusInternalServerError)
		return
	}
	// New user to the database
	newUser := models.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}
	if err := h.userRepo.CreateUser(&newUser); err != nil {
		http.Error(w, `{"error": "Failed to create user account"}`, http.StatusInternalServerError)
		return
	}
	// Success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"user_id": newUser.ID,
		"email":   newUser.Email,
	})
}

// Login authentication and generates a secure JWT
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	// Parse safely
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	// Retrieves the user from the database via repository
	user, err := h.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		http.Error(w, `{"error": "Invalid email or password"}`, http.StatusUnauthorized)
		return
	}
	// Verify the submitted password against the stored secure bcrypt
	if !utils.VerifyPassword(user.PasswordHash, req.Password) {
		http.Error(w, `{"error": "Invalid email or password"}`, http.StatusUnauthorized)
		return
	}
	// Generate the JWT session
	tokenString, err := utils.GenerateJWT(user.ID)
	if err != nil {
		http.Error(w, `{"error": "Failed to generate authentication token"}`, http.StatusInternalServerError)
		return
	}
	// Send successful response with the JWT
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"token":   tokenString,
	})
}
