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
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Select payment denom: use provided one, or auto-select based on balance (undollar first, then unuah)
	paymentDenom := req.PaymentDenom
	if paymentDenom == "" {
		selectedDenom, err := s.blockchainCli.SelectPaymentDenom(ctx, wallet.Address, "")
		if err != nil {
			return &BuyTokenResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to select payment denom: %v", err),
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
			OperationType: transactions.OperationTypeTokenBuy,
			TxHash:        resp.TxHash,
			Status:        status,
			OperationData: transactions.TokenBuyData(req.Denom, paymentDenom, req.PaymentAmount, resp.TokensOut, resp.PricePaid),
			ErrorMessage:  errorMsg,
		})

		// Если запись в БД не удалась, логируем но не прерываем выполнение
		if createErr != nil {
			// В продакшене здесь можно добавить логирование
			_ = createErr
		}
	}

	if err != nil {
		return &BuyTokenResponse{
			Success: false,
			TxHash:  resp.TxHash,
			Error:   err.Error(),
		}, err
	}

	return &BuyTokenResponse{
		TxHash:    resp.TxHash,
		TokensOut: resp.TokensOut,
		PricePaid: resp.PricePaid,
		Success:   resp.Success,
		Error:     resp.Error,
	}, nil
}
