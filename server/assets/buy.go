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
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Select payment denom: use provided one, or auto-select based on balance (undollar first, then unuah)
	paymentDenom := req.Denom
	if paymentDenom == "" {
		selectedDenom, err := s.blockchainCli.SelectPaymentDenom(ctx, wallet.Address, "")
		if err != nil {
			return &BuyAssetResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to select payment denom: %v", err),
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

	// Записываем транзакцию в БД
	status := transactions.StatusPending
	var errorMsg *string
	if err != nil || !resp.Success {
		status = transactions.StatusFailed
		if err != nil {
			msg := err.Error()
			errorMsg = &msg
		} else if resp.Error != "" {
			errorMsg = &resp.Error
		}
	}

	// Записываем только если есть tx_hash (транзакция была отправлена)
	if resp.TxHash != "" {
		_, createErr := s.transactionsRepo.CreateTransaction(transactions.CreateTransactionRequest{
			UserID:        userID,
			OperationType: transactions.OperationTypeAssetBuy,
			TxHash:        resp.TxHash,
			Status:        status,
			OperationData: transactions.AssetBuyData(req.Symbol, paymentDenom, paymentAmount, resp.BaseAmount),
			ErrorMessage:  errorMsg,
		})

		// Если запись в БД не удалась, логируем но не прерываем выполнение
		if createErr != nil {
			// В продакшене здесь можно добавить логирование
			_ = createErr
		}
	}

	if err != nil {
		return &BuyAssetResponse{
			Success: false,
			TxHash:  resp.TxHash,
			Error:   err.Error(),
		}, err
	}

	return &BuyAssetResponse{
		TxHash:     resp.TxHash,
		BaseAmount: resp.BaseAmount,
		Success:    resp.Success,
		Error:      resp.Error,
	}, nil
}
