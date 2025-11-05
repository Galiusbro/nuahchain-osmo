package usertokens

import (
	"context"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

// CreateToken creates a new user token on the blockchain
func (s *Service) CreateToken(ctx context.Context, userID int64, req CreateTokenRequest) (*CreateTokenResponse, error) {
	// Get user's wallet and decrypt private key
	wallet, privKeyBytes, err := s.GetUserWallet(ctx, userID)
	if err != nil {
		return &CreateTokenResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Create token request for blockchain client
	createReq := blockchain.CreateTokenRequest{
		Creator:     wallet.Address,
		Name:        req.Name,
		Symbol:      req.Symbol,
		Image:       req.Image,
		Description: req.Description,
	}

	// Create token on blockchain
	resp, err := s.blockchainCli.CreateTokenWithKey(ctx, createReq, privKeyBytes)

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
			OperationType: transactions.OperationTypeTokenCreate,
			TxHash:        resp.TxHash,
			Status:        status,
			OperationData: transactions.TokenCreateData(resp.Denom, req.Name, req.Symbol, req.Image, req.Description),
			ErrorMessage:  errorMsg,
		})

		// Если запись в БД не удалась, логируем но не прерываем выполнение
		if createErr != nil {
			// В продакшене здесь можно добавить логирование
			_ = createErr
		}
	}

	if err != nil {
		return &CreateTokenResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	return &CreateTokenResponse{
		Denom:   resp.Denom,
		TxHash:  resp.TxHash,
		Success: resp.Success,
		Error:   resp.Error,
	}, nil
}
