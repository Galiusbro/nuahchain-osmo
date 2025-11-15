package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Authentication errors
var (
	ErrServiceNotInitialized = errors.New("authentication service not initialized")
	ErrMissingToken          = errors.New("authorization header required")
	ErrInvalidTokenFormat    = errors.New("invalid authorization header format")
)

// Service handles authentication business logic
type Service struct {
	repo            *Repository
	walletGenerator *WalletGenerator
	jwtManager      *JWTManager
	tokenExpiry     time.Duration
	refreshExpiry   time.Duration
}

// NewService creates a new authentication service
func NewService(repo *Repository, jwtSecret string, tokenExpiry, refreshExpiry time.Duration) *Service {
	return &Service{
		repo:            repo,
		walletGenerator: NewWalletGenerator(),
		jwtManager:      NewJWTManager(jwtSecret, tokenExpiry, refreshExpiry),
		tokenExpiry:     tokenExpiry,
		refreshExpiry:   refreshExpiry,
	}
}

// RegisterWebUser registers a new user via web (email/password)
// Returns: user, wallet, accessToken, refreshToken, error
// ipAddress and userAgent are optional and used for session tracking
func (s *Service) RegisterWebUser(email, username, password string, ipAddress, userAgent *string) (*User, *Wallet, string, string, error) {
	// Check if user exists
	_, err := s.repo.GetUserByEmail(email)
	if err == nil {
		return nil, nil, "", "", errors.New("user with this email already exists")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, "", "", err
	}
	passwordHashStr := string(passwordHash)

	// Create user
	user, err := s.repo.CreateUser(&email, &username, &passwordHashStr, nil, nil)
	if err != nil {
		return nil, nil, "", "", err
	}

	// Generate wallet
	address, privKey, mnemonic, err := s.walletGenerator.GenerateWallet()
	if err != nil {
		return nil, nil, "", "", err
	}

	// Encrypt private key and mnemonic
	encryptedPrivKey, err := Encrypt(privKey)
	if err != nil {
		return nil, nil, "", "", err
	}

	encryptedMnemonic, err := Encrypt([]byte(mnemonic))
	if err != nil {
		return nil, nil, "", "", err
	}

	// Create wallet
	wallet, err := s.repo.CreateWallet(user.ID, address, encryptedPrivKey, encryptedMnemonic)
	if err != nil {
		return nil, nil, "", "", err
	}

	// Generate access token
	accessToken, err := s.jwtManager.GenerateToken(user.ID, address)
	if err != nil {
		return nil, nil, "", "", err
	}

	// Generate refresh token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, nil, "", "", err
	}

	return user, wallet, accessToken, refreshToken, nil
}

// RegisterTelegramUser registers a new user via Telegram
// Returns: user, wallet, accessToken, refreshToken, error
// ipAddress and userAgent are optional and used for session tracking
func (s *Service) RegisterTelegramUser(telegramID int64, telegramUsername, firstName, lastName string, ipAddress, userAgent *string) (*User, *Wallet, string, string, error) {
	// Check if user exists
	existingUser, err := s.repo.GetUserByTelegramID(telegramID)
	if err == nil {
		// User exists, just return wallet and new session
		wallet, err := s.repo.GetWalletByUserID(existingUser.ID)
		if err != nil {
			return nil, nil, "", "", err
		}

		accessToken, err := s.jwtManager.GenerateToken(existingUser.ID, wallet.Address)
		if err != nil {
			return nil, nil, "", "", err
		}

		refreshToken, err := s.jwtManager.GenerateRefreshToken(existingUser.ID)
		if err != nil {
			return nil, nil, "", "", err
		}

		// Save session to database
		expiresAt := time.Now().Add(s.tokenExpiry)
		refreshExpiresAt := time.Now().Add(s.refreshExpiry)
		_, err = s.repo.CreateSession(existingUser.ID, accessToken, refreshToken, expiresAt, refreshExpiresAt, ipAddress, userAgent)
		if err != nil {
			// Log error but don't fail - session tracking is optional
		}

		return existingUser, wallet, accessToken, refreshToken, nil
	}

	// Create new user
	var username *string
	if telegramUsername != "" {
		username = &telegramUsername
	}

	user, err := s.repo.CreateUser(nil, username, nil, &telegramID, &telegramUsername)
	if err != nil {
		return nil, nil, "", "", err
	}

	// Generate wallet
	address, privKey, mnemonic, err := s.walletGenerator.GenerateWallet()
	if err != nil {
		return nil, nil, "", "", err
	}

	// Encrypt private key and mnemonic
	encryptedPrivKey, err := Encrypt(privKey)
	if err != nil {
		return nil, nil, "", "", err
	}

	encryptedMnemonic, err := Encrypt([]byte(mnemonic))
	if err != nil {
		return nil, nil, "", "", err
	}

	// Create wallet
	wallet, err := s.repo.CreateWallet(user.ID, address, encryptedPrivKey, encryptedMnemonic)
	if err != nil {
		return nil, nil, "", "", err
	}

	// Generate access token
	accessToken, err := s.jwtManager.GenerateToken(user.ID, address)
	if err != nil {
		return nil, nil, "", "", err
	}

	// Generate refresh token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, nil, "", "", err
	}

	// Save session to database
	expiresAt := time.Now().Add(s.tokenExpiry)
	refreshExpiresAt := time.Now().Add(s.refreshExpiry)
	_, err = s.repo.CreateSession(user.ID, accessToken, refreshToken, expiresAt, refreshExpiresAt, ipAddress, userAgent)
	if err != nil {
		// Log error but don't fail - session tracking is optional
	}

	return user, wallet, accessToken, refreshToken, nil
}

// LoginWebUser logs in a user via web (email/password)
// Returns: user, wallet, accessToken, refreshToken, error
// ipAddress and userAgent are optional and used for session tracking
func (s *Service) LoginWebUser(email, password string, ipAddress, userAgent *string) (*User, *Wallet, string, string, error) {
	// Get user with password hash
	user, passwordHash, err := s.repo.GetUserByEmailWithPassword(email)
	if err != nil {
		return nil, nil, "", "", errors.New("invalid credentials")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return nil, nil, "", "", errors.New("invalid credentials")
	}

	// Update last login
	err = s.repo.UpdateUserLastLogin(user.ID)
	if err != nil {
		return nil, nil, "", "", err
	}

	// Get wallet
	wallet, err := s.repo.GetWalletByUserID(user.ID)
	if err != nil {
		return nil, nil, "", "", err
	}

	// Generate access token
	accessToken, err := s.jwtManager.GenerateToken(user.ID, wallet.Address)
	if err != nil {
		return nil, nil, "", "", err
	}

	// Generate refresh token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, nil, "", "", err
	}

	// Save session to database
	expiresAt := time.Now().Add(s.tokenExpiry)
	refreshExpiresAt := time.Now().Add(s.refreshExpiry)
	_, err = s.repo.CreateSession(user.ID, accessToken, refreshToken, expiresAt, refreshExpiresAt, ipAddress, userAgent)
	if err != nil {
		// Log error but don't fail login - session tracking is optional
	}

	return user, wallet, accessToken, refreshToken, nil
}

// ValidateToken validates a JWT token and returns user
func (s *Service) ValidateToken(token string) (*User, error) {
	// First validate token structure and signature
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	// Check if token is blacklisted (logout)
	tokenHash := hashToken(token)
	isBlacklisted, err := s.repo.IsTokenBlacklisted(tokenHash)
	if err != nil {
		return nil, err
	}
	if isBlacklisted {
		return nil, errors.New("token has been revoked")
	}

	user, err := s.repo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserWallet gets wallet for a user
func (s *Service) GetUserWallet(userID int64) (*Wallet, error) {
	return s.repo.GetWalletByUserID(userID)
}

// GetWalletByAddress gets wallet by address
func (s *Service) GetWalletByAddress(address string) (*Wallet, error) {
	return s.repo.GetWalletByAddress(address)
}

// RefreshToken refreshes an access token using a refresh token
// Returns: new accessToken, new refreshToken, error
// ipAddress and userAgent are optional and used for session tracking
func (s *Service) RefreshToken(refreshTokenString string, ipAddress, userAgent *string) (string, string, error) {
	// Validate refresh token
	userID, err := s.jwtManager.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	// Verify user exists
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	// Get wallet for address
	wallet, err := s.repo.GetWalletByUserID(user.ID)
	if err != nil {
		return "", "", errors.New("wallet not found")
	}

	// Generate new access token
	accessToken, err := s.jwtManager.GenerateToken(user.ID, wallet.Address)
	if err != nil {
		return "", "", err
	}

	// Generate new refresh token (rotate refresh token for security)
	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	// Save new session to database
	expiresAt := time.Now().Add(s.tokenExpiry)
	refreshExpiresAt := time.Now().Add(s.refreshExpiry)
	_, err = s.repo.CreateSession(user.ID, accessToken, newRefreshToken, expiresAt, refreshExpiresAt, ipAddress, userAgent)
	if err != nil {
		// Log error but don't fail refresh - session tracking is optional
	}

	return accessToken, newRefreshToken, nil
}

// Logout invalidates a token by adding it to the blacklist
func (s *Service) Logout(tokenString string) error {
	// Validate token to get claims (including expiry)
	claims, err := s.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return errors.New("invalid token")
	}

	// Check if already blacklisted
	tokenHash := hashToken(tokenString)
	isBlacklisted, err := s.repo.IsTokenBlacklisted(tokenHash)
	if err != nil {
		return err
	}
	if isBlacklisted {
		// Already logged out, but return success
		return nil
	}

	// Get expiry time from claims (jwt.NumericDate -> time.Time)
	var expiresAt time.Time
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	} else {
		// Fallback: use default token expiry if not set
		expiresAt = time.Now().Add(24 * time.Hour)
	}

	// Add to blacklist with token's expiry time
	err = s.repo.AddTokenToBlacklist(tokenHash, expiresAt, claims.UserID)
	if err != nil {
		return err
	}

	return nil
}

// LogoutAll invalidates all tokens for the current user
func (s *Service) LogoutAll(tokenString string) error {
	// Validate token to get user ID
	claims, err := s.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return errors.New("invalid token")
	}

	// Revoke all tokens for this user
	err = s.repo.RevokeAllUserTokens(claims.UserID)
	if err != nil {
		return err
	}

	// Also add current token to blacklist (if not already)
	tokenHash := hashToken(tokenString)
	isBlacklisted, err := s.repo.IsTokenBlacklisted(tokenHash)
	if err != nil {
		return err
	}
	if !isBlacklisted {
		var expiresAt time.Time
		if claims.ExpiresAt != nil {
			expiresAt = claims.ExpiresAt.Time
		} else {
			expiresAt = time.Now().Add(24 * time.Hour)
		}
		err = s.repo.AddTokenToBlacklist(tokenHash, expiresAt, claims.UserID)
		if err != nil {
			return err
		}
	}

	return nil
}

// hashToken creates a SHA256 hash of the token for blacklist storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// ForgotPassword generates a password reset token for a user
// Returns: resetToken, error
// Note: In production, this token should be sent via email. For now, we return it directly.
func (s *Service) ForgotPassword(email string, ipAddress *string) (string, error) {
	// Get user by email
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		// Don't reveal if user exists or not (security best practice)
		// Return success even if user doesn't exist
		return "", nil
	}

	// Check if user has a password (web user, not Telegram-only)
	if user.Email == nil || *user.Email == "" {
		// User doesn't have email/password, can't reset
		return "", nil
	}

	// Invalidate all existing reset tokens for this user
	err = s.repo.InvalidateUserPasswordResetTokens(user.ID)
	if err != nil {
		return "", err
	}

	// Generate secure random token
	token, err := generateSecureToken(32)
	if err != nil {
		return "", err
	}

	// Token expires in 1 hour
	expiresAt := time.Now().Add(1 * time.Hour)

	// Save token to database
	err = s.repo.CreatePasswordResetToken(user.ID, token, expiresAt, ipAddress)
	if err != nil {
		return "", err
	}

	// TODO: Send email with reset link
	// For now, we return the token (for development/testing)
	// In production: sendEmail(user.Email, resetLink)

	return token, nil
}

// ResetPassword resets a user's password using a reset token
func (s *Service) ResetPassword(token, newPassword string) error {
	// Get reset token
	resetToken, err := s.repo.GetPasswordResetToken(token)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	// Check if token is already used
	if resetToken.Used {
		return errors.New("reset token has already been used")
	}

	// Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		return errors.New("reset token has expired")
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	passwordHashStr := string(passwordHash)

	// Update user password
	err = s.repo.UpdateUserPassword(resetToken.UserID, passwordHashStr)
	if err != nil {
		return err
	}

	// Mark token as used
	err = s.repo.MarkPasswordResetTokenAsUsed(token)
	if err != nil {
		return err
	}

	// Invalidate all other reset tokens for this user (security)
	err = s.repo.InvalidateUserPasswordResetTokens(resetToken.UserID)
	if err != nil {
		// Log but don't fail
	}

	return nil
}

// GetUserSessions gets all active sessions for a user
func (s *Service) GetUserSessions(userID int64) ([]*Session, error) {
	return s.repo.GetUserSessions(userID)
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
