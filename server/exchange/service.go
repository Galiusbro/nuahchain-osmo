package exchange

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/server/auth"
	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
	transactionstracker "github.com/osmosis-labs/osmosis/v30/server/transactions/tracker"
)

// Service handles exchange operations
type Service struct {
	blockchainCli    *blockchain.Client
	authService      *auth.Service
	transactionsRepo *transactions.Repository
	txTracker        *transactionstracker.Tracker
}

// NewService creates a new exchange service
func NewService(
	blockchainCli *blockchain.Client,
	authService *auth.Service,
	transactionsRepo *transactions.Repository,
	txTracker *transactionstracker.Tracker,
) *Service {
	return &Service{
		blockchainCli:    blockchainCli,
		authService:      authService,
		transactionsRepo: transactionsRepo,
		txTracker:        txTracker,
	}
}

// ExchangeTokens handles token exchange for unuah
func (s *Service) ExchangeTokens(ctx context.Context, userID int64, req ExchangeTokensRequest) (*ExchangeTokensResponse, error) {
	// Get user wallet
	wallet, err := s.authService.GetUserWallet(userID)
	if err != nil {
		return &ExchangeTokensResponse{
			Status:   string(transactions.StatusFailed),
			ErrorMsg: fmt.Sprintf("failed to get user wallet: %v", err),
		}, fmt.Errorf("failed to get user wallet: %w", err)
	}

	// Decrypt private key
	privKeyBytes, err := auth.Decrypt(wallet.EncryptedPrivateKey)
	if err != nil {
		return &ExchangeTokensResponse{
			Status:   string(transactions.StatusFailed),
			ErrorMsg: fmt.Sprintf("failed to decrypt private key: %v", err),
		}, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	privKeyHex := fmt.Sprintf("%x", privKeyBytes)

	// Parse amounts
	amount, ok := math.NewIntFromString(req.Amount)
	if !ok {
		return &ExchangeTokensResponse{
			Status:   string(transactions.StatusFailed),
			ErrorMsg: "invalid amount",
		}, fmt.Errorf("invalid amount: %s", req.Amount)
	}

	minOutput, ok := math.NewIntFromString(req.MinOutput)
	if !ok {
		return &ExchangeTokensResponse{
			Status:   string(transactions.StatusFailed),
			ErrorMsg: "invalid min_output",
		}, fmt.Errorf("invalid min_output: %s", req.MinOutput)
	}

	// Create blockchain request
	bReq := blockchain.ExchangeTokensRequest{
		TokenIn:    sdk.NewCoin(req.TokenDenom, amount),
		MinNuahOut: minOutput,
	}

	// Execute exchange on blockchain
	bResp, err := s.blockchainCli.ExchangeTokensWithKey(ctx, privKeyHex, bReq)
	if err != nil {
		errMsg := err.Error()
		if bResp != nil && bResp.ErrorMsg != "" {
			errMsg = bResp.ErrorMsg
		}
		return &ExchangeTokensResponse{
			Status:   string(transactions.StatusFailed),
			TxHash:   bResp.TxHash,
			ErrorMsg: errMsg,
		}, fmt.Errorf("exchange failed: %s", errMsg)
	}

	response := &ExchangeTokensResponse{
		Status:   string(transactions.StatusPending),
		TxHash:   bResp.TxHash,
		UnuahOut: "",
		ErrorMsg: "",
	}

	if bResp == nil || bResp.TxHash == "" {
		return response, nil
	}

	operationData := map[string]interface{}{
		"token_denom": req.TokenDenom,
		"amount":      req.Amount,
		"min_output":  req.MinOutput,
		"unuah_out":   "",
	}

	if bResp.NuahOut.IsPositive() {
		value := bResp.NuahOut.String()
		operationData["unuah_out"] = value
		response.UnuahOut = value
	}

	if _, createErr := s.transactionsRepo.CreateTransaction(transactions.CreateTransactionRequest{
		UserID:        userID,
		OperationType: transactions.OperationTypeExchange,
		TxHash:        bResp.TxHash,
		Status:        transactions.StatusPending,
		OperationData: operationData,
		ErrorMessage:  nil,
	}); createErr != nil {
		return &ExchangeTokensResponse{
			Status:   string(transactions.StatusFailed),
			TxHash:   bResp.TxHash,
			ErrorMsg: fmt.Sprintf("failed to persist transaction: %v", createErr),
		}, createErr
	}

	if s.txTracker != nil {
		s.txTracker.Track(bResp.TxHash)
	}

	return response, nil
}
