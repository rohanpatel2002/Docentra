package handlers

import (
	"ai-document-assistant/internal/models"
	"ai-document-assistant/pkg/database"
	"ai-document-assistant/pkg/utils"
	"encoding/json"
	"net/http"
	"strings"
)

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register user
func Register(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	// Parse the JSON payload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}
	// Sanitization and validation
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || len(req.Password) < 8 {
		http.Error(w, `{"error": "Email is required and password must be at least 8 characters long"}`, http.StatusBadRequest)
		return
	}
	// Prevents bcrypt overflow vulnerability
	if len(req.Password) > 72 {
		http.Error(w, `{"error": "Password exceeds maximum length"}`, http.StatusBadRequest)
		return
	}
	// Check if the user already exists to prevent multiple copies or entries with the same email
	var existingUser models.User
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
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
	if err := database.DB.Create(&newUser).Error; err != nil {
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
func Login(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	// Parse safely
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	// Retrieves the user from the database
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
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
