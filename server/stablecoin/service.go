package stablecoin

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v30/server/auth"
	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
	transactionstracker "github.com/osmosis-labs/osmosis/v30/server/transactions/tracker"
)

// Service handles stablecoin operations
type Service struct {
	authService      *auth.Service
	blockchainCli    *blockchain.Client
	transactionsRepo *transactions.Repository
	txTracker        *transactionstracker.Tracker
}

// NewService creates a new stablecoin service
func NewService(authService *auth.Service, blockchainCli *blockchain.Client, transactionsRepo *transactions.Repository, txTracker *transactionstracker.Tracker) *Service {
	return &Service{
		authService:      authService,
		blockchainCli:    blockchainCli,
		transactionsRepo: transactionsRepo,
		txTracker:        txTracker,
	}
}

var globalService *Service
var globalAuthService *auth.Service

// SetService sets the global service instance
func SetService(s *Service) {
	globalService = s
}

// SetAuthService sets the global auth service instance
func SetAuthService(s *auth.Service) {
	globalAuthService = s
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
