package stablecoin

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/v30/server/auth"
	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
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
		return &BuyNDollarResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	return &BuyNDollarResponse{
		Success:       resp.Success,
		TxHash:        resp.TxHash,
		NDollarAmount: resp.NDollarAmount,
		NDollarDenom:  resp.NDollarDenom,
		Error:         resp.Error,
	}, nil
}
