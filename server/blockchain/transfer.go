package blockchain

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// SendCoinsRequest represents a request to send coins
type SendCoinsRequest struct {
	FromAddress string // Sender address
	ToAddress   string // Recipient address
	Denom       string // Denom to send (e.g., "unuah", "undollar")
	Amount      string // Amount to send (as string to handle large numbers)
}

// SendCoinsResponse represents the response from sending coins
type SendCoinsResponse struct {
	TxHash  string `json:"tx_hash,omitempty"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// SendCoinsWithKey sends coins from one address to another using a private key
func (c *Client) SendCoinsWithKey(
	ctx context.Context,
	req SendCoinsRequest,
	privKeyBytes []byte,
) (*SendCoinsResponse, error) {
	// Validate addresses
	fromAddr, err := sdk.AccAddressFromBech32(req.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid from address: %w", err)
	}

	// Validate to address (but don't need to use it)
	if _, err := sdk.AccAddressFromBech32(req.ToAddress); err != nil {
		return nil, fmt.Errorf("invalid to address: %w", err)
	}

	// Create secp256k1 private key from bytes
	if len(privKeyBytes) != 32 {
		return nil, fmt.Errorf("invalid private key length: expected 32 bytes, got %d", len(privKeyBytes))
	}
	privKey := &secp256k1.PrivKey{Key: privKeyBytes}

	// Verify the address matches the private key
	pubKey := privKey.PubKey()
	derivedAddr := sdk.AccAddress(pubKey.Address())
	if !fromAddr.Equals(derivedAddr) {
		return nil, fmt.Errorf("private key does not match from address")
	}

	// Parse amount
	amount, ok := sdkmath.NewIntFromString(req.Amount)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %s", req.Amount)
	}

	// Create send message
	msg := &banktypes.MsgSend{
		FromAddress: req.FromAddress,
		ToAddress:   req.ToAddress,
		Amount:      sdk.NewCoins(sdk.NewCoin(req.Denom, amount)),
	}

	// Validate amount
	if amount.IsZero() || amount.IsNegative() {
		return nil, fmt.Errorf("invalid amount: must be positive")
	}

	// Sign and broadcast transaction
	txHash, _, err := c.signAndBroadcastTx(ctx, req.FromAddress, privKey, pubKey, msg)
	if err != nil {
		return &SendCoinsResponse{
			Success: false,
			TxHash:  txHash,
			Error:   err.Error(),
		}, err
	}

	return &SendCoinsResponse{
		TxHash:  txHash,
		Success: true,
	}, nil
}
