package assets

import (
	"context"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

// EnsureAsset ensures an asset exists, creating it if necessary
func (s *Service) EnsureAsset(ctx context.Context, userID int64, req EnsureAssetRequest) (*EnsureAssetResponse, error) {
	// Get user's wallet and decrypt private key
	wallet, privKeyBytes, err := s.GetUserWallet(ctx, userID)
	if err != nil {
		return &EnsureAssetResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Create ensure request for blockchain client
	ensureReq := blockchain.EnsureAssetRequest{
		Creator: wallet.Address,
		Symbol:  req.Symbol,
	}

	// Execute ensure transaction
	resp, err := s.blockchainCli.EnsureAssetWithKey(ctx, ensureReq, privKeyBytes)

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
			OperationType: transactions.OperationTypeAssetEnsure,
			TxHash:        resp.TxHash,
			Status:        status,
			OperationData: transactions.AssetEnsureData(req.Symbol),
			ErrorMessage:  errorMsg,
		})

		// Если запись в БД не удалась, логируем но не прерываем выполнение
		if createErr != nil {
			// В продакшене здесь можно добавить логирование
			_ = createErr
		}
	}

	if err != nil {
		return &EnsureAssetResponse{
			Success: false,
			TxHash:  resp.TxHash,
			Error:   err.Error(),
		}, err
	}

	return &EnsureAssetResponse{
		TxHash:  resp.TxHash,
		Success: resp.Success,
		Error:   resp.Error,
	}, nil
}
