package usertokens

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

// SellToken sells tokens to bonding curve
func (s *Service) SellToken(ctx context.Context, userID int64, req SellTokenRequest) (*SellTokenResponse, error) {
	// Get user's wallet and decrypt private key
	wallet, privKeyBytes, err := s.GetUserWallet(ctx, userID)
	if err != nil {
		return &SellTokenResponse{
			Status: string(transactions.StatusFailed),
			Error:  err.Error(),
		}, err
	}

	// Select payment denom: use provided one, or auto-select based on balance (undollar first, then unuah)
	paymentDenom := req.PaymentDenom
	if paymentDenom == "" {
		selectedDenom, err := s.blockchainCli.SelectPaymentDenom(ctx, wallet.Address, "")
		if err != nil {
			return &SellTokenResponse{
				Status: string(transactions.StatusFailed),
				Error:  fmt.Sprintf("failed to select payment denom: %v", err),
			}, err
		}
		paymentDenom = selectedDenom
	}

	// Create sell request for blockchain client
	sellReq := blockchain.SellToCurveRequest{
		Trader:        wallet.Address,
		Denom:         req.Denom,
		TokenAmount:   req.TokenAmount,
		PaymentDenom:  paymentDenom,
		MinPaymentOut: req.MinPaymentOut,
	}

	// Execute sell transaction
	resp, err := s.blockchainCli.SellToCurveWithKey(ctx, sellReq, privKeyBytes)
	if err != nil {
		errMsg := err.Error()
		if resp != nil && resp.Error != "" {
			errMsg = resp.Error
		}
		return &SellTokenResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  errMsg,
		}, fmt.Errorf(errMsg)
	}

	if resp == nil || resp.TxHash == "" {
		msg := "transaction hash not returned"
		return &SellTokenResponse{
			Status: string(transactions.StatusFailed),
			Error:  msg,
		}, fmt.Errorf(msg)
	}

	if err := s.recordPendingTransaction(
		userID,
		transactions.OperationTypeTokenSell,
		resp.TxHash,
		transactions.TokenSellData(req.Denom, req.TokenAmount, paymentDenom, resp.PaymentOut, resp.PriceReceived),
	); err != nil {
		return &SellTokenResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  fmt.Sprintf("failed to persist transaction: %v", err),
		}, err
	}

	return &SellTokenResponse{
		TxHash:        resp.TxHash,
		PaymentOut:    resp.PaymentOut,
		PriceReceived: resp.PriceReceived,
		Status:        string(transactions.StatusPending),
		Message:       "Token sale broadcast, awaiting confirmation",
	}, nil
}
