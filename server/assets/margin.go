package assets

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

// OpenMarginPosition opens a leveraged position via the leverage module.
func (s *Service) OpenMarginPosition(ctx context.Context, userID int64, req OpenMarginPositionRequest) (*OpenMarginPositionResponse, error) {
	// Retrieve wallet and decrypted private key
	wallet, privKeyBytes, err := s.GetUserWallet(ctx, userID)
	if err != nil {
		return &OpenMarginPositionResponse{
			Success: false,
			Error:   err.Error(),
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

	status := transactions.StatusPending
	var errorMsg *string
	if err != nil || (resp != nil && !resp.Success) {
		status = transactions.StatusFailed
		msg := err
		if msg == nil && resp != nil && resp.Error != "" {
			msg = fmt.Errorf(resp.Error)
		}
		if msg != nil {
			s := msg.Error()
			errorMsg = &s
		}
	}

	effectiveLeverage := resp.Leverage
	if effectiveLeverage == "" {
		effectiveLeverage = req.Leverage
	}

	if resp.TxHash != "" {
		side := strings.ToUpper(strings.TrimSpace(req.Side))
		tx, createErr := s.transactionsRepo.CreateTransaction(transactions.CreateTransactionRequest{
			UserID:        userID,
			OperationType: transactions.OperationTypeAssetMarginOpen,
			TxHash:        resp.TxHash,
			Status:        status,
			OperationData: transactions.AssetMarginOpenData(
				req.Symbol,
				side,
				req.QuoteAmount,
				effectiveLeverage,
				resp.PositionID,
				resp.BaseQuantity,
				resp.EntryPrice,
			),
			ErrorMessage: errorMsg,
		})
		if createErr != nil {
			// TODO: add structured logging
			_ = createErr
		} else if tx != nil && resp.Success {
			// Update status to SUCCESS after successful execution
			// Wait a bit for transaction to be included in block
			txHash := resp.TxHash
			go func() {
				time.Sleep(5 * time.Second)
				// Create new context for background goroutine
				bgCtx := context.Background()
				txStatus, err := s.blockchainCli.GetTxStatus(bgCtx, txHash)
				if err == nil && txStatus != nil && txStatus.Success {
					_ = s.transactionsRepo.UpdateTransactionByTxHash(
						txHash,
						transactions.StatusSuccess,
						nil,
						nil,
					)
				} else if txStatus != nil && !txStatus.Success {
					errorMsg := txStatus.Error
					_ = s.transactionsRepo.UpdateTransactionByTxHash(
						txHash,
						transactions.StatusFailed,
						nil,
						&errorMsg,
					)
				}
			}()
		}
	}

	if err != nil {
		return &OpenMarginPositionResponse{
			TxHash:  resp.TxHash,
			Success: false,
			Error:   err.Error(),
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
		Success:      resp.Success,
		Message:      "Margin position opening initiated",
		Error:        resp.Error,
	}, nil
}

// CloseMarginPosition closes an existing leveraged position on-chain.
func (s *Service) CloseMarginPosition(ctx context.Context, userID int64, positionID uint64) (*CloseMarginPositionResponse, error) {
	// Retrieve wallet and decrypted private key
	wallet, privKeyBytes, err := s.GetUserWallet(ctx, userID)
	if err != nil {
		return &CloseMarginPositionResponse{
			Success: false,
			Error:   err.Error(),
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

	status := transactions.StatusPending
	var errorMsg *string
	if err != nil || (resp != nil && !resp.Success) {
		status = transactions.StatusFailed
		msg := err
		if msg == nil && resp != nil && resp.Error != "" {
			msg = fmt.Errorf(resp.Error)
		}
		if msg != nil {
			s := msg.Error()
			errorMsg = &s
		}
	}

	if resp.TxHash != "" {
		tx, createErr := s.transactionsRepo.CreateTransaction(transactions.CreateTransactionRequest{
			UserID:        userID,
			OperationType: transactions.OperationTypeAssetMarginClose,
			TxHash:        resp.TxHash,
			Status:        status,
			OperationData: transactions.AssetMarginCloseData(positionID, resp.Pnl),
			ErrorMessage:  errorMsg,
		})
		if createErr != nil {
			// TODO: add structured logging
			_ = createErr
		} else if tx != nil && resp.Success {
			// Update status to SUCCESS after successful execution
			txHash := resp.TxHash
			go func() {
				time.Sleep(5 * time.Second)
				// Create new context for background goroutine
				bgCtx := context.Background()
				txStatus, err := s.blockchainCli.GetTxStatus(bgCtx, txHash)
				if err == nil && txStatus != nil && txStatus.Success {
					_ = s.transactionsRepo.UpdateTransactionByTxHash(
						txHash,
						transactions.StatusSuccess,
						nil,
						nil,
					)
				} else if txStatus != nil && !txStatus.Success {
					errorMsg := txStatus.Error
					_ = s.transactionsRepo.UpdateTransactionByTxHash(
						txHash,
						transactions.StatusFailed,
						nil,
						&errorMsg,
					)
				}
			}()
		}
	}

	if err != nil {
		return &CloseMarginPositionResponse{
			TxHash:  resp.TxHash,
			Success: false,
			Error:   err.Error(),
		}, err
	}

	return &CloseMarginPositionResponse{
		TxHash:  resp.TxHash,
		Pnl:     resp.Pnl,
		Success: resp.Success,
		Message: "Margin position closure initiated",
		Error:   resp.Error,
	}, nil
}
