package assets

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

// BuyAsset buys an asset using payment denom
func (s *Service) BuyAsset(ctx context.Context, userID int64, req BuyAssetRequest) (*BuyAssetResponse, error) {
	// Get user's wallet and decrypt private key
	wallet, privKeyBytes, err := s.GetUserWallet(ctx, userID)
	if err != nil {
		return &BuyAssetResponse{
			Status: string(transactions.StatusFailed),
			Error:  err.Error(),
		}, err
	}

	// Select payment denom: use provided one, or auto-select based on balance (undollar first, then unuah)
	paymentDenom := req.Denom
	if paymentDenom == "" {
		selectedDenom, err := s.blockchainCli.SelectPaymentDenom(ctx, wallet.Address, "")
		if err != nil {
			return &BuyAssetResponse{
				Status: string(transactions.StatusFailed),
				Error:  fmt.Sprintf("failed to select payment denom: %v", err),
			}, err
		}
		paymentDenom = selectedDenom
	}

	// Determine payment amount - prefer new format (amount) over deprecated (amount_ndollar)
	paymentAmount := req.Amount
	if paymentAmount == "" && req.AmountNDOLLAR != "" {
		// If using deprecated format, use NDOLLAR denom
		paymentAmount = req.AmountNDOLLAR
		if paymentDenom == "" {
			paymentDenom = "NDOLLAR"
		}
	}

	// Create buy request for blockchain client
	buyReq := blockchain.BuyAssetRequest{
		Buyer:         wallet.Address,
		Symbol:        req.Symbol,
		Denom:         paymentDenom,
		Amount:        paymentAmount,
		AmountNDOLLAR: req.AmountNDOLLAR, // Keep for backward compatibility
	}

	// Execute buy transaction
	resp, err := s.blockchainCli.BuyAssetWithKey(ctx, buyReq, privKeyBytes)
	if err != nil {
		errMsg := err.Error()
		if resp != nil && resp.Error != "" {
			errMsg = resp.Error
		}
		return &BuyAssetResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  errMsg,
		}, fmt.Errorf(errMsg)
	}

	if resp == nil || resp.TxHash == "" {
		msg := "transaction hash not returned"
		return &BuyAssetResponse{
			Status: string(transactions.StatusFailed),
			Error:  msg,
		}, fmt.Errorf(msg)
	}

	if err := s.recordPendingTransaction(userID, transactions.OperationTypeAssetBuy, resp.TxHash, transactions.AssetBuyData(req.Symbol, paymentDenom, paymentAmount, resp.BaseAmount)); err != nil {
		return &BuyAssetResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  fmt.Sprintf("failed to persist transaction: %v", err),
		}, err
	}

	return &BuyAssetResponse{
		TxHash:     resp.TxHash,
		BaseAmount: resp.BaseAmount,
		Status:     string(transactions.StatusPending),
		Message:    "Asset purchase broadcast, awaiting confirmation",
	}, nil
}
