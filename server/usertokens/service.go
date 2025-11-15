package usertokens

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/v30/server/auth"
	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/tokens"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
	transactionstracker "github.com/osmosis-labs/osmosis/v30/server/transactions/tracker"
)

// Service handles all user token operations
type Service struct {
	authRepo         *auth.Repository
	blockchainCli    *blockchain.Client
	transactionsRepo *transactions.Repository
	tokensRepo       *tokens.Repository
	txTracker        *transactionstracker.Tracker
}

// NewService creates a new user token service
func NewService(authRepo *auth.Repository, blockchainCli *blockchain.Client, transactionsRepo *transactions.Repository, tokensRepo *tokens.Repository, txTracker *transactionstracker.Tracker) *Service {
	return &Service{
		authRepo:         authRepo,
		blockchainCli:    blockchainCli,
		transactionsRepo: transactionsRepo,
		tokensRepo:       tokensRepo,
		txTracker:        txTracker,
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

func (s *Service) recordPendingTransaction(userID int64, operationType string, txHash string, data map[string]interface{}) error {
	if txHash == "" {
		return fmt.Errorf("empty transaction hash")
	}

	if _, err := s.transactionsRepo.CreateTransaction(transactions.CreateTransactionRequest{
		UserID:        userID,
		OperationType: operationType,
		TxHash:        txHash,
		Status:        transactions.StatusPending,
		OperationData: data,
		ErrorMessage:  nil,
	}); err != nil {
		return err
	}

	if s.txTracker != nil {
		s.txTracker.Track(txHash)
	}

	return nil
}
