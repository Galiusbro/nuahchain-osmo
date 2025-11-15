package usertokens

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/tokens"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

// Helper function to convert string to *string
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// CreateToken creates a new user token on the blockchain
func (s *Service) CreateToken(ctx context.Context, userID int64, req CreateTokenRequest) (*CreateTokenResponse, error) {
	// Get user's wallet and decrypt private key
	wallet, privKeyBytes, err := s.GetUserWallet(ctx, userID)
	if err != nil {
		return &CreateTokenResponse{
			Status: string(transactions.StatusFailed),
			Error:  err.Error(),
		}, err
	}

	// Create token request for blockchain client
	createReq := blockchain.CreateTokenRequest{
		Creator:     wallet.Address,
		Name:        req.Name,
		Symbol:      req.Symbol,
		Image:       req.Image,
		Description: req.Description,
	}

	// Create token on blockchain
	resp, err := s.blockchainCli.CreateTokenWithKey(ctx, createReq, privKeyBytes)
	if err != nil {
		errMsg := err.Error()
		if resp != nil && resp.Error != "" {
			errMsg = resp.Error
		}
		return &CreateTokenResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  errMsg,
		}, fmt.Errorf(errMsg)
	}

	if resp.TxHash == "" {
		msg := "transaction hash not returned"
		return &CreateTokenResponse{
			Status: string(transactions.StatusFailed),
			Error:  msg,
		}, fmt.Errorf(msg)
	}

	// Save token metadata to database
	if s.tokensRepo != nil {
		tokenMetadata := tokens.Token{
			Denom:          resp.Denom,
			Name:           req.Name,
			Symbol:         req.Symbol,
			Image:          stringPtr(req.Image),
			Description:    stringPtr(req.Description),
			CreatorAddress: wallet.Address,
			CreatorUserID:  &userID,
			Decimals:       6, // Default decimals for user tokens
		}
		if err := s.tokensRepo.CreateOrUpdateToken(tokenMetadata); err != nil {
			// Log error but don't fail the transaction - metadata can be saved later
			fmt.Printf("Warning: failed to save token metadata: %v\n", err)
		}
	}

	if err := s.recordPendingTransaction(userID, transactions.OperationTypeTokenCreate, resp.TxHash, transactions.TokenCreateData(resp.Denom, req.Name, req.Symbol, req.Image, req.Description)); err != nil {
		return &CreateTokenResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  fmt.Sprintf("failed to persist transaction: %v", err),
		}, err
	}

	return &CreateTokenResponse{
		Denom:   resp.Denom,
		TxHash:  resp.TxHash,
		Status:  string(transactions.StatusPending),
		Message: "Token creation broadcast, awaiting confirmation",
	}, nil
}
