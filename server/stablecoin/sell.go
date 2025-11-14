package stablecoin

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/v30/server/auth"
	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

// SellNDollar converts NDOLLAR back to unuah at 1:1 ratio
func (s *Service) SellNDollar(ctx context.Context, userID int64, amount string) (*SellNDollarResponse, error) {
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
	req := blockchain.SellNDollarRequest{
		Amount: amount,
	}

	resp, err := s.blockchainCli.SellNDollarWithKey(ctx, privKeyHex, req)
	if err != nil {
		errMsg := err.Error()
		if resp != nil && resp.Error != "" {
			errMsg = resp.Error
		}
		return &SellNDollarResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  errMsg,
		}, fmt.Errorf(errMsg)
	}

	if resp == nil || resp.TxHash == "" {
		msg := "transaction hash not returned"
		return &SellNDollarResponse{
			Status: string(transactions.StatusFailed),
			Error:  msg,
		}, fmt.Errorf(msg)
	}

	if err := s.recordPendingTransaction(userID, transactions.OperationTypeStablecoinSell, resp.TxHash, map[string]interface{}{
		"amount":       amount,
		"unuah_amount": resp.UnuahAmount,
	}); err != nil {
		return &SellNDollarResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  fmt.Sprintf("failed to persist transaction: %v", err),
		}, err
	}

	return &SellNDollarResponse{
		Status:      string(transactions.StatusPending),
		TxHash:      resp.TxHash,
		UnuahAmount: resp.UnuahAmount,
	}, nil
}
