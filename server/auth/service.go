package auth

import (
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
}

// NewService creates a new authentication service
func NewService(repo *Repository, jwtSecret string, tokenExpiry, refreshExpiry time.Duration) *Service {
	return &Service{
		repo:            repo,
		walletGenerator: NewWalletGenerator(),
		jwtManager:      NewJWTManager(jwtSecret, tokenExpiry, refreshExpiry),
	}
}

// RegisterWebUser registers a new user via web (email/password)
func (s *Service) RegisterWebUser(email, username, password string) (*User, *Wallet, string, error) {
	// Check if user exists
	_, err := s.repo.GetUserByEmail(email)
	if err == nil {
		return nil, nil, "", errors.New("user with this email already exists")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, "", err
	}
	passwordHashStr := string(passwordHash)

	// Create user
	user, err := s.repo.CreateUser(&email, &username, &passwordHashStr, nil, nil)
	if err != nil {
		return nil, nil, "", err
	}

	// Generate wallet
	address, privKey, mnemonic, err := s.walletGenerator.GenerateWallet()
	if err != nil {
		return nil, nil, "", err
	}

	// Encrypt private key and mnemonic
	encryptedPrivKey, err := Encrypt(privKey)
	if err != nil {
		return nil, nil, "", err
	}

	encryptedMnemonic, err := Encrypt([]byte(mnemonic))
	if err != nil {
		return nil, nil, "", err
	}

	// Create wallet
	wallet, err := s.repo.CreateWallet(user.ID, address, encryptedPrivKey, encryptedMnemonic)
	if err != nil {
		return nil, nil, "", err
	}

	// Generate session token
	token, err := s.jwtManager.GenerateToken(user.ID, address)
	if err != nil {
		return nil, nil, "", err
	}

	return user, wallet, token, nil
}

// RegisterTelegramUser registers a new user via Telegram
func (s *Service) RegisterTelegramUser(telegramID int64, telegramUsername, firstName, lastName string) (*User, *Wallet, string, error) {
	// Check if user exists
	existingUser, err := s.repo.GetUserByTelegramID(telegramID)
	if err == nil {
		// User exists, just return wallet and new session
		wallet, err := s.repo.GetWalletByUserID(existingUser.ID)
		if err != nil {
			return nil, nil, "", err
		}

		token, err := s.jwtManager.GenerateToken(existingUser.ID, wallet.Address)
		if err != nil {
			return nil, nil, "", err
		}

		return existingUser, wallet, token, nil
	}

	// Create new user
	var username *string
	if telegramUsername != "" {
		username = &telegramUsername
	}

	user, err := s.repo.CreateUser(nil, username, nil, &telegramID, &telegramUsername)
	if err != nil {
		return nil, nil, "", err
	}

	// Generate wallet
	address, privKey, mnemonic, err := s.walletGenerator.GenerateWallet()
	if err != nil {
		return nil, nil, "", err
	}

	// Encrypt private key and mnemonic
	encryptedPrivKey, err := Encrypt(privKey)
	if err != nil {
		return nil, nil, "", err
	}

	encryptedMnemonic, err := Encrypt([]byte(mnemonic))
	if err != nil {
		return nil, nil, "", err
	}

	// Create wallet
	wallet, err := s.repo.CreateWallet(user.ID, address, encryptedPrivKey, encryptedMnemonic)
	if err != nil {
		return nil, nil, "", err
	}

	// Generate session token
	token, err := s.jwtManager.GenerateToken(user.ID, address)
	if err != nil {
		return nil, nil, "", err
	}

	return user, wallet, token, nil
}

// LoginWebUser logs in a user via web (email/password)
func (s *Service) LoginWebUser(email, password string) (*User, *Wallet, string, error) {
	// Get user with password hash
	user, passwordHash, err := s.repo.GetUserByEmailWithPassword(email)
	if err != nil {
		return nil, nil, "", errors.New("invalid credentials")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return nil, nil, "", errors.New("invalid credentials")
	}

	// Update last login
	err = s.repo.UpdateUserLastLogin(user.ID)
	if err != nil {
		return nil, nil, "", err
	}

	// Get wallet
	wallet, err := s.repo.GetWalletByUserID(user.ID)
	if err != nil {
		return nil, nil, "", err
	}

	// Generate session token
	token, err := s.jwtManager.GenerateToken(user.ID, wallet.Address)
	if err != nil {
		return nil, nil, "", err
	}

	return user, wallet, token, nil
}

// ValidateToken validates a JWT token and returns user
func (s *Service) ValidateToken(token string) (*User, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, err
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
