package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/osmosis-labs/osmosis/v30/server/auth"
)

// AuthService is the authentication service interface
type AuthService interface {
	RegisterWebUser(email, username, password string, ipAddress, userAgent *string) (*auth.User, *auth.Wallet, string, string, error)
	RegisterTelegramUser(telegramID int64, telegramUsername, firstName, lastName string, ipAddress, userAgent *string) (*auth.User, *auth.Wallet, string, string, error)
	LoginWebUser(email, password string, ipAddress, userAgent *string) (*auth.User, *auth.Wallet, string, string, error)
	RefreshToken(refreshToken string, ipAddress, userAgent *string) (string, string, error)
	Logout(token string) error
	LogoutAll(token string) error
	ForgotPassword(email string, ipAddress *string) (string, error)
	ResetPassword(token, newPassword string) error
	GetUserSessions(userID int64) ([]*auth.Session, error)
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
	User         *auth.User   `json:"user"`
	Wallet       *auth.Wallet `json:"wallet"`
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	Message      string       `json:"message"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	User         *auth.User   `json:"user"`
	Wallet       *auth.Wallet `json:"wallet"`
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	Message      string       `json:"message"`
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

	// Get IP address and User-Agent from request
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}
	userAgent := r.Header.Get("User-Agent")

	// Register user
	user, wallet, token, refreshToken, err := authService.RegisterWebUser(req.Email, req.Username, req.Password, &ipAddress, &userAgent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := RegisterResponse{
		User:         user,
		Wallet:       wallet,
		Token:        token,
		RefreshToken: refreshToken,
		Message:      "Registration successful",
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

	// Get IP address and User-Agent from request
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}
	userAgent := r.Header.Get("User-Agent")

	// Login user
	user, wallet, token, refreshToken, err := authService.LoginWebUser(req.Email, req.Password, &ipAddress, &userAgent)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	response := LoginResponse{
		User:         user,
		Wallet:       wallet,
		Token:        token,
		RefreshToken: refreshToken,
		Message:      "Login successful",
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

	// Get IP address and User-Agent from request
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}
	userAgent := r.Header.Get("User-Agent")

	// Register or login user
	user, wallet, token, refreshToken, err := authService.RegisterTelegramUser(
		req.ID,
		req.Username,
		req.FirstName,
		req.LastName,
		&ipAddress,
		&userAgent,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := RegisterResponse{
		User:         user,
		Wallet:       wallet,
		Token:        token,
		RefreshToken: refreshToken,
		Message:      "Telegram authentication successful",
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

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenResponse represents a refresh token response
type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	Message      string `json:"message"`
}

// handleRefresh handles token refresh
func handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if authService == nil {
		http.Error(w, "Authentication service not configured", http.StatusInternalServerError)
		return
	}

	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.RefreshToken == "" {
		http.Error(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	// Get IP address and User-Agent from request
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}
	userAgent := r.Header.Get("User-Agent")

	// Refresh token
	accessToken, refreshToken, err := authService.RefreshToken(req.RefreshToken, &ipAddress, &userAgent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	response := RefreshTokenResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		Message:      "Token refreshed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// LogoutResponse represents a logout response
type LogoutResponse struct {
	Message string `json:"message"`
}

// handleLogout handles user logout (token invalidation)
func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	// Logout (invalidate token)
	err := authService.Logout(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	response := LogoutResponse{
		Message: "Logged out successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleLogoutAll handles user logout from all devices (all tokens invalidation)
func handleLogoutAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	// Logout all (invalidate all tokens for user)
	err := authService.LogoutAll(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	response := LogoutResponse{
		Message: "Logged out from all devices successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ForgotPasswordResponse represents a forgot password response
type ForgotPasswordResponse struct {
	Message string `json:"message"`
	Token   string `json:"token,omitempty"` // Only in development, remove in production
}

// handleForgotPassword handles password reset request
func handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if authService == nil {
		http.Error(w, "Authentication service not configured", http.StatusInternalServerError)
		return
	}

	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Get IP address from request
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}

	// Generate reset token
	token, err := authService.ForgotPassword(req.Email, &ipAddress)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Always return success (don't reveal if user exists)
	response := ForgotPasswordResponse{
		Message: "If an account with that email exists, a password reset link has been sent",
	}

	// In development, return token for testing
	// TODO: Remove this in production and send email instead
	if token != "" {
		response.Token = token
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ResetPasswordResponse represents a password reset response
type ResetPasswordResponse struct {
	Message string `json:"message"`
}

// handleResetPassword handles password reset
func handleResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if authService == nil {
		http.Error(w, "Authentication service not configured", http.StatusInternalServerError)
		return
	}

	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}
	if req.NewPassword == "" {
		http.Error(w, "New password is required", http.StatusBadRequest)
		return
	}

	// Validate password strength (minimum 8 characters)
	if len(req.NewPassword) < 8 {
		http.Error(w, "Password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	// Reset password
	err := authService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := ResetPasswordResponse{
		Message: "Password has been reset successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SessionInfo represents session information for API response (without sensitive data)
type SessionInfo struct {
	ID         int64     `json:"id"`
	IPAddress  *string   `json:"ip_address,omitempty"`
	UserAgent  *string   `json:"user_agent,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	IsCurrent  bool      `json:"is_current"` // Is this the current session?
}

// SessionsResponse represents the response for sessions list
type SessionsResponse struct {
	Sessions []SessionInfo `json:"sessions"`
	Total    int           `json:"total"`
}

// handleSessions handles GET /api/auth/sessions - get all active sessions for user
func handleSessions(w http.ResponseWriter, r *http.Request) {
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

	// Validate token and get user
	user, err := authService.ValidateToken(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Get all active sessions for user
	sessions, err := authService.GetUserSessions(user.ID)
	if err != nil {
		http.Error(w, "Failed to get sessions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to API response format (hide sensitive data)
	sessionInfos := make([]SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		isCurrent := session.Token == token
		sessionInfos = append(sessionInfos, SessionInfo{
			ID:         session.ID,
			IPAddress:  session.IPAddress,
			UserAgent:  session.UserAgent,
			CreatedAt:  session.CreatedAt,
			LastUsedAt: session.LastUsedAt,
			ExpiresAt:  session.ExpiresAt,
			IsCurrent:  isCurrent,
		})
	}

	response := SessionsResponse{
		Sessions: sessionInfos,
		Total:    len(sessionInfos),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
