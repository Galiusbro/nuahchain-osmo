package blockchain

import (
	"context"
	"fmt"
	"time"

	abciv1beta1 "cosmossdk.io/api/cosmos/base/abci/v1beta1"
	txservice "cosmossdk.io/api/cosmos/tx/v1beta1"
)

// BroadcastResult represents the result of broadcasting a transaction
type BroadcastResult struct {
	TxHash  string
	TxResp  *abciv1beta1.TxResponse
	Success bool
	Code    uint32
	RawLog  string
	Error   error
}

// VerifyTxResult verifies the final result of a broadcasted transaction
// This should be called after signAndBroadcastTx to ensure the transaction actually succeeded
func (c *Client) VerifyTxResult(ctx context.Context, txHash string, maxRetries int) (*BroadcastResult, error) {
	if maxRetries <= 0 {
		maxRetries = 5
	}

	// Wait for transaction to be included in a block (increase delay for slower chains)
	txResp, err := c.WaitForTx(ctx, txHash, maxRetries, 3*time.Second)
	if err != nil {
		return &BroadcastResult{
			TxHash:  txHash,
			Success: false,
			Error:   fmt.Errorf("transaction not found after waiting: %w", err),
		}, err
	}

	// Check if transaction succeeded
	success := txResp.Code == 0
	var resultErr error
	if !success {
		resultErr = fmt.Errorf("transaction failed (code %d): %s", txResp.Code, txResp.RawLog)
	}

	return &BroadcastResult{
		TxHash:  txHash,
		TxResp:  txResp,
		Success: success,
		Code:    txResp.Code,
		RawLog:  txResp.RawLog,
		Error:   resultErr,
	}, nil
}

// WaitForTx waits for a transaction to be included in a block and returns its final status
// maxRetries: maximum number of retry attempts
// retryDelay: delay between retries
func (c *Client) WaitForTx(ctx context.Context, txHash string, maxRetries int, retryDelay time.Duration) (*abciv1beta1.TxResponse, error) {
	if maxRetries <= 0 {
		maxRetries = 5
	}
	if retryDelay <= 0 {
		retryDelay = 2 * time.Second
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(retryDelay)
		}

		req := &txservice.GetTxRequest{
			Hash: txHash,
		}

		resp, err := c.txClient.GetTx(ctx, req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.TxResponse != nil {
			return resp.TxResponse, nil
		}
	}

	return nil, fmt.Errorf("transaction not found after %d retries: %w", maxRetries, lastErr)
}

// ValidateTxResponse checks if a transaction response indicates success
// and returns a user-friendly error if not
func ValidateTxResponse(txResp *abciv1beta1.TxResponse) error {
	if txResp == nil {
		return fmt.Errorf("transaction response is nil")
	}

	if txResp.Code != 0 {
		return fmt.Errorf("transaction failed (code %d): %s", txResp.Code, txResp.RawLog)
	}

	return nil
}
