package users

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var globalUserService *Service

// SetService sets the global user service
func SetService(s *Service) {
	globalUserService = s
}

// HandleGetUserProfile handles GET /api/users/me - get full user profile
func HandleGetUserProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if globalUserService == nil {
		http.Error(w, "User service not configured", http.StatusInternalServerError)
		return
	}

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	token := parts[1]

	// Validate token and get user
	user, err := globalUserService.authService.ValidateToken(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Get full user profile
	profile, err := globalUserService.GetUserProfile(user.ID)
	if err != nil {
		http.Error(w, "Failed to get user profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}

// HandleUploadUserImage handles POST /api/users/me/upload-image - upload user profile image
func HandleUploadUserImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if globalUserService == nil {
		http.Error(w, "User service not configured", http.StatusInternalServerError)
		return
	}

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	token := parts[1]

	// Validate token and get user
	user, err := globalUserService.authService.ValidateToken(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Parse multipart form (max 10MB)
	err = r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image file is required: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Upload and save image
	imageURL, err := globalUserService.SaveUserImage(user.ID, file, header)
	if err != nil {
		http.Error(w, "Failed to upload image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message":   "Image uploaded successfully",
		"image_url": imageURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ServeUploadedImages serves uploaded images statically
// This should be registered as a file server route
func ServeUploadedImages(w http.ResponseWriter, r *http.Request) {
	if globalUserService == nil {
		http.Error(w, "User service not configured", http.StatusInternalServerError)
		return
	}

	// Extract filename from path
	filename := filepath.Base(r.URL.Path)
	filePath := filepath.Join(globalUserService.uploadPath, filename)

	// Check if file exists
	file, err := os.Open(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	// Get file info for Content-Type
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Failed to get file info", http.StatusInternalServerError)
		return
	}

	// Set appropriate headers
	ext := filepath.Ext(filename)
	contentType := "application/octet-stream"
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Copy file to response
	io.Copy(w, file)
}

// HandleGetUserInfoSummary handles GET /api/users/me/info - get brief user profile summary
func HandleGetUserInfoSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if globalUserService == nil {
		http.Error(w, "User service not configured", http.StatusInternalServerError)
		return
	}

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	token := parts[1]

	// Validate token and get user
	user, err := globalUserService.authService.ValidateToken(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Get user info summary
	summary, err := globalUserService.GetUserInfoSummary(user.ID)
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(summary)
}

// HandleUpdateUsername handles PATCH /api/users/username - update user's username
func HandleUpdateUsername(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if globalUserService == nil {
		http.Error(w, "User service not configured", http.StatusInternalServerError)
		return
	}

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	token := parts[1]

	// Validate token and get user
	user, err := globalUserService.authService.ValidateToken(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update username
	if err := globalUserService.UpdateUsername(user.ID, req.Username); err != nil {
		http.Error(w, "Failed to update username: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get updated user
	updatedUser, err := globalUserService.authRepo.GetUserByID(user.ID)
	if err != nil {
		http.Error(w, "Failed to get updated user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message":  "Username updated successfully",
		"username": updatedUser.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleGetUserTokens handles GET /api/users/me/tokens - get list of user's tokens
func HandleGetUserTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if globalUserService == nil {
		http.Error(w, "User service not configured", http.StatusInternalServerError)
		return
	}

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	token := parts[1]

	// Validate token and get user
	user, err := globalUserService.authService.ValidateToken(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Get user tokens
	tokens, err := globalUserService.GetUserTokens(user.ID)
	if err != nil {
		http.Error(w, "Failed to get user tokens: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"tokens": tokens,
		"count":  len(tokens),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

