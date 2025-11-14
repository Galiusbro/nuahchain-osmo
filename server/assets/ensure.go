package assets

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

// EnsureAsset ensures an asset exists, creating it if necessary
func (s *Service) EnsureAsset(ctx context.Context, userID int64, req EnsureAssetRequest) (*EnsureAssetResponse, error) {
	// Get user's wallet and decrypt private key
	wallet, privKeyBytes, err := s.GetUserWallet(ctx, userID)
	if err != nil {
		return &EnsureAssetResponse{
			Status: string(transactions.StatusFailed),
			Error:  err.Error(),
		}, err
	}

	// Create ensure request for blockchain client
	ensureReq := blockchain.EnsureAssetRequest{
		Creator: wallet.Address,
		Symbol:  req.Symbol,
	}

	// Execute ensure transaction
	resp, err := s.blockchainCli.EnsureAssetWithKey(ctx, ensureReq, privKeyBytes)
	if err != nil {
		errMsg := err.Error()
		if resp != nil && resp.Error != "" {
			errMsg = resp.Error
		}
		return &EnsureAssetResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  errMsg,
		}, fmt.Errorf(errMsg)
	}

	if resp == nil || resp.TxHash == "" {
		msg := "transaction hash not returned"
		return &EnsureAssetResponse{
			Status: string(transactions.StatusFailed),
			Error:  msg,
		}, fmt.Errorf(msg)
	}

	if err := s.recordPendingTransaction(userID, transactions.OperationTypeAssetEnsure, resp.TxHash, transactions.AssetEnsureData(req.Symbol)); err != nil {
		return &EnsureAssetResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  fmt.Sprintf("failed to persist transaction: %v", err),
		}, err
	}

	return &EnsureAssetResponse{
		TxHash:  resp.TxHash,
		Status:  string(transactions.StatusPending),
		Message: "Asset ensure transaction broadcast, awaiting confirmation",
	}, nil
}
