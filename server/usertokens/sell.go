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
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Select payment denom: use provided one, or auto-select based on balance (undollar first, then unuah)
	paymentDenom := req.PaymentDenom
	if paymentDenom == "" {
		selectedDenom, err := s.blockchainCli.SelectPaymentDenom(ctx, wallet.Address, "")
		if err != nil {
			return &SellTokenResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to select payment denom: %v", err),
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
			OperationType: transactions.OperationTypeTokenSell,
			TxHash:        resp.TxHash,
			Status:        status,
			OperationData: transactions.TokenSellData(req.Denom, req.TokenAmount, paymentDenom, resp.PaymentOut, resp.PriceReceived),
			ErrorMessage:  errorMsg,
		})

		// Если запись в БД не удалась, логируем но не прерываем выполнение
		if createErr != nil {
			// В продакшене здесь можно добавить логирование
			_ = createErr
		}
	}

	if err != nil {
		return &SellTokenResponse{
			Success: false,
			TxHash:  resp.TxHash,
			Error:   err.Error(),
		}, err
	}

	return &SellTokenResponse{
		TxHash:        resp.TxHash,
		PaymentOut:    resp.PaymentOut,
		PriceReceived: resp.PriceReceived,
		Success:       resp.Success,
		Error:         resp.Error,
	}, nil
}
