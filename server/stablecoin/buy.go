package stablecoin

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/v30/server/auth"
	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

// BuyNDollar converts unuah to NDOLLAR at 1:1 ratio
func (s *Service) BuyNDollar(ctx context.Context, userID int64, amount string) (*BuyNDollarResponse, error) {
	// Get user wallet
	wallet, err := s.authService.GetUserWallet(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user wallet: %w", err)
	}

	// Decrypt private key
	privKeyBytes, err := auth.Decrypt(wallet.EncryptedPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	privKeyHex := fmt.Sprintf("%x", privKeyBytes)

	// Call blockchain
	req := blockchain.BuyNDollarRequest{
		Amount: amount,
	}

	resp, err := s.blockchainCli.BuyNDollarWithKey(ctx, privKeyHex, req)
	if err != nil {
		errMsg := err.Error()
		if resp != nil && resp.Error != "" {
			errMsg = resp.Error
		}
		return &BuyNDollarResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  errMsg,
		}, fmt.Errorf(errMsg)
	}

	if resp == nil || resp.TxHash == "" {
		msg := "transaction hash not returned"
		return &BuyNDollarResponse{
			Status: string(transactions.StatusFailed),
			Error:  msg,
		}, fmt.Errorf(msg)
	}

	if err := s.recordPendingTransaction(userID, transactions.OperationTypeStablecoinBuy, resp.TxHash, map[string]interface{}{
		"amount":         amount,
		"ndollar_amount": resp.NDollarAmount,
		"ndollar_denom":  resp.NDollarDenom,
	}); err != nil {
		return &BuyNDollarResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  fmt.Sprintf("failed to persist transaction: %v", err),
		}, err
	}

	return &BuyNDollarResponse{
		Status:        string(transactions.StatusPending),
		TxHash:        resp.TxHash,
		NDollarAmount: resp.NDollarAmount,
		NDollarDenom:  resp.NDollarDenom,
	}, nil
}
