package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/osmosis-labs/osmosis/v30/server/auth"
)

// AuthService is the authentication service interface
type AuthService interface {
	RegisterWebUser(email, username, password string) (*auth.User, *auth.Wallet, string, error)
	RegisterTelegramUser(telegramID int64, telegramUsername, firstName, lastName string) (*auth.User, *auth.Wallet, string, error)
	LoginWebUser(email, password string) (*auth.User, *auth.Wallet, string, error)
	ValidateToken(token string) (*auth.User, error)
	GetUserWallet(userID int64) (*auth.Wallet, error)
}

var authService AuthService

// SetAuthService sets the authentication service
func SetAuthService(service AuthService) {
	authService = service
}

// RegisterRequest represents a web registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterResponse represents a registration response
type RegisterResponse struct {
	User    *auth.User   `json:"user"`
	Wallet  *auth.Wallet `json:"wallet"`
	Token   string       `json:"token"`
	Message string       `json:"message"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	User    *auth.User   `json:"user"`
	Wallet  *auth.Wallet `json:"wallet"`
	Token   string       `json:"token"`
	Message string       `json:"message"`
}

// TelegramAuthRequest represents a Telegram authentication request
type TelegramAuthRequest struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

// handleRegister handles user registration via web
func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if authService == nil {
		http.Error(w, "Authentication service not configured", http.StatusInternalServerError)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Register user
	user, wallet, token, err := authService.RegisterWebUser(req.Email, req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := RegisterResponse{
		User:    user,
		Wallet:  wallet,
		Token:   token,
		Message: "Registration successful",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleLogin handles user login via web
func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if authService == nil {
		http.Error(w, "Authentication service not configured", http.StatusInternalServerError)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Login user
	user, wallet, token, err := authService.LoginWebUser(req.Email, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	response := LoginResponse{
		User:    user,
		Wallet:  wallet,
		Token:   token,
		Message: "Login successful",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleTelegramAuth handles Telegram authentication
func handleTelegramAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if authService == nil {
		http.Error(w, "Authentication service not configured", http.StatusInternalServerError)
		return
	}

	var req TelegramAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Verify Telegram hash (requires BOT_TOKEN)
	// For now, we'll skip verification in development

	// Register or login user
	user, wallet, token, err := authService.RegisterTelegramUser(
		req.ID,
		req.Username,
		req.FirstName,
		req.LastName,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := RegisterResponse{
		User:    user,
		Wallet:  wallet,
		Token:   token,
		Message: "Telegram authentication successful",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleMe returns current user information (requires authentication)
func handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if authService == nil {
		http.Error(w, "Authentication service not configured", http.StatusInternalServerError)
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
	user, err := authService.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Get wallet for user
	wallet, err := authService.GetUserWallet(user.ID)

	response := map[string]interface{}{
		"user": user,
	}

	if err == nil {
		response["wallet"] = map[string]interface{}{
			"id":      wallet.ID,
			"address": wallet.Address,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
