package balances

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
)

// Service provides balance-related operations
type Service struct {
	repo          *Repository
	blockchainCli *blockchain.Client
	indexer       *Indexer
}

// NewService creates a new balances service
func NewService(repo *Repository, blockchainCli *blockchain.Client, indexer *Indexer) *Service {
	return &Service{
		repo:          repo,
		blockchainCli: blockchainCli,
		indexer:       indexer,
	}
}

// GetUserBalancesFromDB gets user balances from database
func (s *Service) GetUserBalancesFromDB(userID int64, denomFilter string) ([]UserBalance, error) {
	return s.repo.GetUserBalances(userID, denomFilter)
}

// GetUserBalancesFromBlockchain gets user balances directly from blockchain
func (s *Service) GetUserBalancesFromBlockchain(ctx context.Context, address string) ([]UserBalance, error) {
	if s.blockchainCli == nil {
		return nil, fmt.Errorf("blockchain client not configured")
	}

	// Get balances from blockchain
	coins, err := s.blockchainCli.GetAllBalances(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get balances from blockchain: %w", err)
	}

	// Convert to UserBalance format
	var balances []UserBalance
	for _, coin := range coins {
		// Skip zero balances
		if coin.Amount.IsZero() {
			continue
		}

		balances = append(balances, UserBalance{
			Address: address,
			Denom:   coin.Denom,
			Amount:  coin.Amount.String(),
		})
	}

	return balances, nil
}

// SyncUserBalances synchronizes user balances from blockchain to database
func (s *Service) SyncUserBalances(ctx context.Context, userID int64, address string) error {
	if s.blockchainCli == nil {
		return fmt.Errorf("blockchain client not configured")
	}

	// Get balances from blockchain
	blockchainBalances, err := s.GetUserBalancesFromBlockchain(ctx, address)
	if err != nil {
		return err
	}

	// Get balances from database
	dbBalances, err := s.repo.GetUserBalances(userID, "")
	if err != nil {
		return err
	}

	// Create map of DB balances for quick lookup
	dbBalanceMap := make(map[string]*UserBalance)
	for i := range dbBalances {
		dbBalanceMap[dbBalances[i].Denom] = &dbBalances[i]
	}

	// Update or insert balances
	for _, bcBalance := range blockchainBalances {
		dbBalance, exists := dbBalanceMap[bcBalance.Denom]

		// Check if update is needed
		needsUpdate := !exists || dbBalance.Amount != bcBalance.Amount

		if needsUpdate {
			req := UpdateBalanceRequest{
				UserID:    userID,
				Address:   address,
				Denom:     bcBalance.Denom,
				Amount:    bcBalance.Amount,
				TxHash:    "sync", // Special marker for sync
				Height:    0,      // Unknown for sync
				EventType: "sync",
			}

			if dbBalance != nil {
				req.AmountBefore = &dbBalance.Amount
			}

			if err := s.repo.UpsertBalance(req); err != nil {
				return fmt.Errorf("failed to sync balance for %s: %w", bcBalance.Denom, err)
			}
		}
	}

	return nil
}

// GetBalanceHistory gets balance history for a user
func (s *Service) GetBalanceHistory(userID int64, denom string, limit int) ([]BalanceHistory, error) {
	return s.repo.GetBalanceHistory(userID, denom, limit)
}
