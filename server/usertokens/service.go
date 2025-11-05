package usertokens

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/v30/server/auth"
	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

// Service handles all user token operations
type Service struct {
	authRepo         *auth.Repository
	blockchainCli    *blockchain.Client
	transactionsRepo *transactions.Repository
}

// NewService creates a new user token service
func NewService(authRepo *auth.Repository, blockchainCli *blockchain.Client, transactionsRepo *transactions.Repository) *Service {
	return &Service{
		authRepo:         authRepo,
		blockchainCli:    blockchainCli,
		transactionsRepo: transactionsRepo,
	}
}

// GetUserWallet retrieves and decrypts the user's wallet
func (s *Service) GetUserWallet(ctx context.Context, userID int64) (*auth.Wallet, []byte, error) {
	wallet, err := s.authRepo.GetWalletByUserID(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("wallet not found: %w", err)
	}

	privKeyBytes, err := auth.Decrypt(wallet.EncryptedPrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	return wallet, privKeyBytes, nil
}

// GetTxStatus gets the status of a transaction by hash
func (s *Service) GetTxStatus(ctx context.Context, txHash string) (*blockchain.TxStatusResponse, error) {
	return s.blockchainCli.GetTxStatus(ctx, txHash)
}
