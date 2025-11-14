package assets

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

// OpenMarginPosition opens a leveraged position via the leverage module.
func (s *Service) OpenMarginPosition(ctx context.Context, userID int64, req OpenMarginPositionRequest) (*OpenMarginPositionResponse, error) {
	// Retrieve wallet and decrypted private key
	wallet, privKeyBytes, err := s.GetUserWallet(ctx, userID)
	if err != nil {
		return &OpenMarginPositionResponse{
			Status: string(transactions.StatusFailed),
			Error:  err.Error(),
		}, err
	}

	// Build blockchain request
	openReq := blockchain.OpenLeveragePositionRequest{
		Owner:       wallet.Address,
		Symbol:      req.Symbol,
		Side:        req.Side,
		QuoteAmount: req.QuoteAmount,
		Leverage:    req.Leverage,
	}

	resp, err := s.blockchainCli.OpenLeveragePositionWithKey(ctx, openReq, privKeyBytes)
	if resp == nil {
		resp = &blockchain.OpenLeveragePositionResponse{}
	}

	if err != nil {
		errMsg := err.Error()
		return &OpenMarginPositionResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  errMsg,
		}, fmt.Errorf(errMsg)
	}

	if resp.TxHash == "" {
		msg := "transaction hash not returned"
		return &OpenMarginPositionResponse{
			Status: string(transactions.StatusFailed),
			Error:  msg,
		}, fmt.Errorf(msg)
	}

	effectiveLeverage := resp.Leverage
	if effectiveLeverage == "" {
		effectiveLeverage = req.Leverage
	}

	side := strings.ToUpper(strings.TrimSpace(req.Side))
	if err := s.recordPendingTransaction(
		userID,
		transactions.OperationTypeAssetMarginOpen,
		resp.TxHash,
		transactions.AssetMarginOpenData(
			req.Symbol,
			side,
			req.QuoteAmount,
			effectiveLeverage,
			resp.PositionID,
			resp.BaseQuantity,
			resp.EntryPrice,
		),
	); err != nil {
		return &OpenMarginPositionResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  fmt.Sprintf("failed to persist transaction: %v", err),
		}, err
	}

	positionID := ""
	if resp.PositionID != 0 {
		positionID = strconv.FormatUint(resp.PositionID, 10)
	}

	return &OpenMarginPositionResponse{
		TxHash:       resp.TxHash,
		PositionID:   positionID,
		BaseQuantity: resp.BaseQuantity,
		EntryPrice:   resp.EntryPrice,
		Leverage:     effectiveLeverage,
		Status:       string(transactions.StatusPending),
		Message:      "Margin position opening broadcast, awaiting confirmation",
	}, nil
}

// CloseMarginPosition closes an existing leveraged position on-chain.
func (s *Service) CloseMarginPosition(ctx context.Context, userID int64, positionID uint64) (*CloseMarginPositionResponse, error) {
	// Retrieve wallet and decrypted private key
	wallet, privKeyBytes, err := s.GetUserWallet(ctx, userID)
	if err != nil {
		return &CloseMarginPositionResponse{
			Status: string(transactions.StatusFailed),
			Error:  err.Error(),
		}, err
	}

	closeReq := blockchain.CloseLeveragePositionRequest{
		Owner:      wallet.Address,
		PositionID: positionID,
	}

	resp, err := s.blockchainCli.CloseLeveragePositionWithKey(ctx, closeReq, privKeyBytes)
	if resp == nil {
		resp = &blockchain.CloseLeveragePositionResponse{}
	}

	if err != nil {
		errMsg := err.Error()
		return &CloseMarginPositionResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  errMsg,
		}, fmt.Errorf(errMsg)
	}

	if resp.TxHash == "" {
		msg := "transaction hash not returned"
		return &CloseMarginPositionResponse{
			Status: string(transactions.StatusFailed),
			Error:  msg,
		}, fmt.Errorf(msg)
	}

	if err := s.recordPendingTransaction(
		userID,
		transactions.OperationTypeAssetMarginClose,
		resp.TxHash,
		transactions.AssetMarginCloseData(positionID, resp.Pnl),
	); err != nil {
		return &CloseMarginPositionResponse{
			Status: string(transactions.StatusFailed),
			TxHash: resp.TxHash,
			Error:  fmt.Sprintf("failed to persist transaction: %v", err),
		}, err
	}

	return &CloseMarginPositionResponse{
		TxHash:  resp.TxHash,
		Pnl:     resp.Pnl,
		Status:  string(transactions.StatusPending),
		Message: "Margin position closure broadcast, awaiting confirmation",
	}, nil
}
