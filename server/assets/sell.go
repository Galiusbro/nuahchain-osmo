package assets

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

// SellAsset sells an asset and receives NDOLLAR
func (s *Service) SellAsset(ctx context.Context, userID int64, req SellAssetRequest) (*SellAssetResponse, error) {
	// Get user's wallet and decrypt private key
	wallet, privKeyBytes, err := s.GetUserWallet(ctx, userID)
	if err != nil {
		return &SellAssetResponse{
			Status: string(transactions.StatusFailed),
			Error:  err.Error(),
		}, err
	}

	// Create sell request for blockchain client
	sellReq := blockchain.SellAssetRequest{
		Seller:     wallet.Address,
		Symbol:     req.Symbol,
		BaseAmount: req.BaseAmount,
	}

	// Execute sell transaction
	resp, err := s.blockchainCli.SellAssetWithKey(ctx, sellReq, privKeyBytes)
	if err != nil {
		errMsg := err.Error()
		if resp != nil && resp.Error != "" {
			errMsg = resp.Error
		}
		return &SellAssetResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  errMsg,
		}, fmt.Errorf(errMsg)
	}

	if resp == nil || resp.TxHash == "" {
		msg := "transaction hash not returned"
		return &SellAssetResponse{
			Status: string(transactions.StatusFailed),
			Error:  msg,
		}, fmt.Errorf(msg)
	}

	if err := s.recordPendingTransaction(userID, transactions.OperationTypeAssetSell, resp.TxHash, transactions.AssetSellData(req.Symbol, req.BaseAmount, resp.PayoutNDOLLAR)); err != nil {
		return &SellAssetResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  fmt.Sprintf("failed to persist transaction: %v", err),
		}, err
	}

	return &SellAssetResponse{
		TxHash:        resp.TxHash,
		PayoutNDOLLAR: resp.PayoutNDOLLAR,
		Status:        string(transactions.StatusPending),
		Message:       "Asset sale broadcast, awaiting confirmation",
	}, nil
}
