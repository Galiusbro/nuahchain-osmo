package users

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/osmosis-labs/osmosis/v30/server/auth"
	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/tokens"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

// AuthServiceInterface defines the interface for authentication service methods we need
type AuthServiceInterface interface {
	ValidateToken(token string) (*auth.User, error)
	GetUserWallet(userID int64) (*auth.Wallet, error)
	GetUserSessions(userID int64) ([]*auth.Session, error)
}

// Service provides user profile management functionality
type Service struct {
	authService      AuthServiceInterface
	authRepo         *auth.Repository
	blockchainCli    *blockchain.Client
	transactionsRepo *transactions.Repository
	tokensRepo       *tokens.Repository
	imageValidator   *ImageValidator
	uploadPath       string // Path to store uploaded images
}

// NewService creates a new user service
func NewService(authService AuthServiceInterface, authRepo *auth.Repository, blockchainCli *blockchain.Client, transactionsRepo *transactions.Repository, tokensRepo *tokens.Repository, uploadPath string) *Service {
	if uploadPath == "" {
		uploadPath = "./uploads/images" // Default path
	}

	return &Service{
		authService:      authService,
		authRepo:         authRepo,
		blockchainCli:    blockchainCli,
		transactionsRepo: transactionsRepo,
		tokensRepo:       tokensRepo,
		imageValidator:   NewImageValidator(),
		uploadPath:       uploadPath,
	}
}

// GetUserProfile gets full user profile information
func (s *Service) GetUserProfile(userID int64) (*UserProfile, error) {
	// Get user from repository
	user, err := s.authRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get wallet
	wallet, err := s.authService.GetUserWallet(userID)
	if err != nil {
		// Wallet might not exist, continue without it
		wallet = nil
	}

	var walletInfo *WalletInfo
	if wallet != nil {
		// Get balances from blockchain
		var balances []Balance
		if s.blockchainCli != nil {
			ctx := context.Background()
			coins, err := s.blockchainCli.GetAllBalances(ctx, wallet.Address)
			if err == nil {
				balances = make([]Balance, 0, len(coins))
				for _, coin := range coins {
					// Only include non-zero balances
					if !coin.Amount.IsZero() {
						balances = append(balances, Balance{
							Denom:  coin.Denom,
							Amount: coin.Amount.String(),
						})
					}
				}
			}
		}

		walletInfo = &WalletInfo{
			Address:   wallet.Address,
			CreatedAt: wallet.CreatedAt,
			Balances:  balances,
		}
	}

	// Get transaction statistics
	var txStats *transactions.UserTransactionStats
	if s.transactionsRepo != nil {
		stats, err := s.transactionsRepo.GetUserTransactionStats(userID)
		if err == nil {
			txStats = stats
		}
	}

	// Calculate unique assets owned (from balances)
	assetsOwned := 0
	if walletInfo != nil && len(walletInfo.Balances) > 0 {
		assetDenoms := make(map[string]bool)
		for _, balance := range walletInfo.Balances {
			// Count only asset/* and factory/* denoms (exclude base denoms like unuah, undollar)
			if strings.HasPrefix(balance.Denom, "asset/") || strings.HasPrefix(balance.Denom, "factory/") {
				assetDenoms[balance.Denom] = true
			}
		}
		assetsOwned = len(assetDenoms)
	}

	// Get tokens count from statistics if available
	tokensCount := 0
	if txStats != nil {
		tokensCount = txStats.TokensCreated
	}

	// Get active sessions count
	activeSessions := 0
	sessions, err := s.authService.GetUserSessions(userID)
	if err == nil {
		activeSessions = len(sessions)
	}

	return &UserProfile{
		User:           user,
		Wallet:         walletInfo,
		TokensCount:    tokensCount,
		ActiveSessions: activeSessions,
		Stats:          txStats,
		AssetsOwned:    assetsOwned,
	}, nil
}

// UpdateUserImageURL updates user's image_url in database
func (s *Service) UpdateUserImageURL(userID int64, imageURL string) error {
	return s.authRepo.UpdateUserImageURL(userID, imageURL)
}

// UploadUserImage uploads and saves a user's profile image
func (s *Service) UploadUserImage(userID int64, file multipart.File, header *multipart.FileHeader) (string, error) {
	// Validate image
	if err := s.imageValidator.ValidateImage(file, header); err != nil {
		return "", fmt.Errorf("image validation failed: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg" // Default extension
	}

	filename := fmt.Sprintf("user_%d_%d%s", userID, time.Now().Unix(), ext)
	filePath := filepath.Join(s.uploadPath, filename)

	// Ensure upload directory exists
	if err := os.MkdirAll(s.uploadPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, file)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Return relative URL path (will be served via static file server)
	imageURL := fmt.Sprintf("/uploads/images/%s", filename)

	return imageURL, nil
}

// SaveUserImage uploads and saves a user's profile image, updating the database
func (s *Service) SaveUserImage(userID int64, file multipart.File, header *multipart.FileHeader) (string, error) {
	// Upload and save image
	imageURL, err := s.UploadUserImage(userID, file, header)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	// Update user's image_url in database
	err = s.UpdateUserImageURL(userID, imageURL)
	if err != nil {
		// Try to delete uploaded file if DB update fails
		os.Remove(filepath.Join(s.uploadPath, filepath.Base(imageURL)))
		return "", fmt.Errorf("failed to update user image URL: %w", err)
	}

	return imageURL, nil
}

// GetUserInfoSummary gets a brief summary of user profile (lightweight version)
func (s *Service) GetUserInfoSummary(userID int64) (*UserInfoSummary, error) {
	// Get user from repository
	user, err := s.authRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get wallet address only
	var walletAddress string
	wallet, err := s.authService.GetUserWallet(userID)
	if err == nil && wallet != nil {
		walletAddress = wallet.Address
	}

	// Get basic stats (only counts, no detailed breakdown)
	var totalTransactions, tokensCreated int
	if s.transactionsRepo != nil {
		stats, err := s.transactionsRepo.GetUserTransactionStats(userID)
		if err == nil && stats != nil {
			totalTransactions = stats.TotalTransactions
			tokensCreated = stats.TokensCreated
		}
	}

	return &UserInfoSummary{
		ID:                user.ID,
		Username:          user.Username,
		Email:             user.Email,
		ImageURL:          user.ImageURL,
		WalletAddress:     walletAddress,
		TotalTransactions: totalTransactions,
		TokensCreated:     tokensCreated,
		CreatedAt:         user.CreatedAt,
	}, nil
}

// UpdateUsername updates user's username
func (s *Service) UpdateUsername(userID int64, username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Validate username format (basic validation)
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters")
	}
	if len(username) > 100 {
		return fmt.Errorf("username must be less than 100 characters")
	}

	return s.authRepo.UpdateUsername(userID, username)
}

// GetUserTokens gets list of tokens owned by the user
func (s *Service) GetUserTokens(userID int64) ([]UserToken, error) {
	// Get wallet
	wallet, err := s.authService.GetUserWallet(userID)
	if err != nil {
		return []UserToken{}, nil // Return empty list if no wallet
	}

	// Get balances from blockchain
	var userTokens []UserToken
	var denoms []string
	if s.blockchainCli != nil {
		ctx := context.Background()
		coins, err := s.blockchainCli.GetAllBalances(ctx, wallet.Address)
		if err != nil {
			return []UserToken{}, fmt.Errorf("failed to get balances: %w", err)
		}

		// Filter only token denoms (factory/* and asset/*) and collect denoms
		for _, coin := range coins {
			// Skip zero balances
			if coin.Amount.IsZero() {
				continue
			}

			// Only include factory/* and asset/* denoms (exclude base denoms like unuah, undollar)
			if strings.HasPrefix(coin.Denom, "factory/") || strings.HasPrefix(coin.Denom, "asset/") {
				denoms = append(denoms, coin.Denom)
				userTokens = append(userTokens, UserToken{
					Denom:  coin.Denom,
					Amount: coin.Amount.String(),
				})
			}
		}
	}

	// Get token metadata from database if available
	if s.tokensRepo != nil && len(denoms) > 0 {
		tokenMetadataMap, err := s.tokensRepo.GetTokensByDenoms(denoms)
		if err == nil {
			// Enrich user tokens with metadata
			for i := range userTokens {
				if metadata, found := tokenMetadataMap[userTokens[i].Denom]; found {
					userTokens[i].Name = metadata.Name
					userTokens[i].Symbol = metadata.Symbol
					if metadata.Image != nil {
						userTokens[i].Image = *metadata.Image
					}
					userTokens[i].Decimals = metadata.Decimals
				}
			}
		}
	}

	return userTokens, nil
}

// UserToken represents a token owned by the user
type UserToken struct {
	Denom     string `json:"denom"`
	Amount    string `json:"amount"`
	Name      string `json:"name,omitempty"`
	Symbol    string `json:"symbol,omitempty"`
	Image     string `json:"image,omitempty"`
	Decimals  int    `json:"decimals,omitempty"`
}

// UserProfile represents extended user profile information
type UserProfile struct {
	User            *auth.User                      `json:"user"`
	Wallet          *WalletInfo                     `json:"wallet,omitempty"`
	TokensCount     int                             `json:"tokens_count,omitempty"`
	ActiveSessions  int                             `json:"active_sessions,omitempty"`
	Stats           *transactions.UserTransactionStats `json:"stats,omitempty"`
	AssetsOwned     int                             `json:"assets_owned,omitempty"`
}

// WalletInfo represents wallet information for profile
type WalletInfo struct {
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	Balances  []Balance `json:"balances,omitempty"`
}

// Balance represents a token balance
type Balance struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// UserInfoSummary represents a brief summary of user profile (lightweight)
type UserInfoSummary struct {
	ID                int64      `json:"id"`
	Username          *string    `json:"username,omitempty"`
	Email             *string    `json:"email,omitempty"`
	ImageURL          *string    `json:"image_url,omitempty"`
	WalletAddress     string     `json:"wallet_address,omitempty"`
	TotalTransactions int        `json:"total_transactions,omitempty"`
	TokensCreated     int        `json:"tokens_created,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

