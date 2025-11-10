package stablecoin

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/v30/server/auth"
	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
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
		return &SellNDollarResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	return &SellNDollarResponse{
		Success:     resp.Success,
		TxHash:      resp.TxHash,
		UnuahAmount: resp.UnuahAmount,
		Error:       resp.Error,
	}, nil
}

