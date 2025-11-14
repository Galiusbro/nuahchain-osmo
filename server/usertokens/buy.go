package usertokens

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

// BuyToken buys tokens from bonding curve
func (s *Service) BuyToken(ctx context.Context, userID int64, req BuyTokenRequest) (*BuyTokenResponse, error) {
	// Get user's wallet and decrypt private key
	wallet, privKeyBytes, err := s.GetUserWallet(ctx, userID)
	if err != nil {
		return &BuyTokenResponse{
			Status: string(transactions.StatusFailed),
			Error:  err.Error(),
		}, err
	}

	// Select payment denom: use provided one, or auto-select based on balance (undollar first, then unuah)
	paymentDenom := req.PaymentDenom
	if paymentDenom == "" {
		selectedDenom, err := s.blockchainCli.SelectPaymentDenom(ctx, wallet.Address, "")
		if err != nil {
			return &BuyTokenResponse{
				Status: string(transactions.StatusFailed),
				Error:  fmt.Sprintf("failed to select payment denom: %v", err),
			}, err
		}
		paymentDenom = selectedDenom
	}

	// Create buy request for blockchain client
	buyReq := blockchain.BuyFromCurveRequest{
		Trader:        wallet.Address,
		Denom:         req.Denom,
		PaymentDenom:  paymentDenom,
		PaymentAmount: req.PaymentAmount,
		MinTokensOut:  req.MinTokensOut,
	}

	// Execute buy transaction
	resp, err := s.blockchainCli.BuyFromCurveWithKey(ctx, buyReq, privKeyBytes)
	if err != nil {
		errMsg := err.Error()
		if resp != nil && resp.Error != "" {
			errMsg = resp.Error
		}
		return &BuyTokenResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  errMsg,
		}, fmt.Errorf(errMsg)
	}

	if resp == nil || resp.TxHash == "" {
		msg := "transaction hash not returned"
		return &BuyTokenResponse{
			Status: string(transactions.StatusFailed),
			Error:  msg,
		}, fmt.Errorf(msg)
	}

	if err := s.recordPendingTransaction(
		userID,
		transactions.OperationTypeTokenBuy,
		resp.TxHash,
		transactions.TokenBuyData(req.Denom, paymentDenom, req.PaymentAmount, resp.TokensOut, resp.PricePaid),
	); err != nil {
		return &BuyTokenResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  fmt.Sprintf("failed to persist transaction: %v", err),
		}, err
	}

	return &BuyTokenResponse{
		TxHash:    resp.TxHash,
		TokensOut: resp.TokensOut,
		PricePaid: resp.PricePaid,
		Status:    string(transactions.StatusPending),
		Message:   "Token purchase broadcast, awaiting confirmation",
	}, nil
}
