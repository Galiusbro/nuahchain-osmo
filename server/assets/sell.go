package assets

import (
	"context"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

// SellAsset sells an asset and receives NDOLLAR
func (s *Service) SellAsset(ctx context.Context, userID int64, req SellAssetRequest) (*SellAssetResponse, error) {
	// Get user's wallet and decrypt private key
	wallet, privKeyBytes, err := s.GetUserWallet(ctx, userID)
	if err != nil {
		return &SellAssetResponse{
			Success: false,
			Error:   err.Error(),
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
			OperationType: transactions.OperationTypeAssetSell,
			TxHash:        resp.TxHash,
			Status:        status,
			OperationData: transactions.AssetSellData(req.Symbol, req.BaseAmount, resp.PayoutNDOLLAR),
			ErrorMessage:  errorMsg,
		})

		// Если запись в БД не удалась, логируем но не прерываем выполнение
		if createErr != nil {
			// В продакшене здесь можно добавить логирование
			_ = createErr
		}
	}

	if err != nil {
		return &SellAssetResponse{
			Success: false,
			TxHash:  resp.TxHash,
			Error:   err.Error(),
		}, err
	}

	return &SellAssetResponse{
		TxHash:        resp.TxHash,
		PayoutNDOLLAR: resp.PayoutNDOLLAR,
		Success:       resp.Success,
		Error:         resp.Error,
	}, nil
}
