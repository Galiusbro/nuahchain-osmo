package blockchain

import (
	"context"
	"encoding/hex"
	"fmt"

	abciv1beta1 "cosmossdk.io/api/cosmos/base/abci/v1beta1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stablecointypes "github.com/osmosis-labs/osmosis/v30/x/stablecoin/types"
)

// BuyNDollarRequest represents the request to buy NDOLLAR
type BuyNDollarRequest struct {
	Amount string // Amount of unuah to convert
}

// BuyNDollarResponse represents the response after buying NDOLLAR
type BuyNDollarResponse struct {
	TxHash        string
	NDollarAmount string
	NDollarDenom  string
	Success       bool
	Error         string
}

// BuyNDollarWithKey buys NDOLLAR using unuah at 1:1 ratio
func (c *Client) BuyNDollarWithKey(ctx context.Context, privKeyHex string, req BuyNDollarRequest) (*BuyNDollarResponse, error) {
	// Decode private key
	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key hex: %w", err)
	}

	privKey := &secp256k1.PrivKey{Key: privKeyBytes}
	pubKey := privKey.PubKey()
	addr := sdk.AccAddress(pubKey.Address())

	// Create the message
	msg := &stablecointypes.MsgBuyNDollar{
		Buyer:  addr.String(),
		Amount: req.Amount,
	}

	// Sign and broadcast
	txHash, txResp, err := c.signAndBroadcastTx(ctx, addr.String(), privKey, pubKey, msg)
	if err != nil {
		return &BuyNDollarResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Check for errors in the transaction
	if txResp.Code != 0 {
		return &BuyNDollarResponse{
			TxHash:  txHash,
			Success: false,
			Error:   fmt.Sprintf("transaction failed with code %d: %s", txResp.Code, txResp.RawLog),
		}, fmt.Errorf("transaction failed with code %d: %s", txResp.Code, txResp.RawLog)
	}

	// Extract response data from transaction events or logs
	ndollarAmount, ndollarDenom := c.extractBuyNDollarMetadata(txResp)

	return &BuyNDollarResponse{
		TxHash:        txHash,
		NDollarAmount: ndollarAmount,
		NDollarDenom:  ndollarDenom,
		Success:       true,
	}, nil
}

// extractBuyNDollarMetadata extracts NDOLLAR amount and denom from transaction response
func (c *Client) extractBuyNDollarMetadata(txResp *abciv1beta1.TxResponse) (string, string) {
	if txResp == nil {
		return "", ""
	}
	// Try to extract from events
	for _, event := range txResp.Events {
		if event.GetType_() == "buy_ndollar" {
			var ndollarAmount, ndollarDenom string
			for _, attr := range event.GetAttributes() {
				switch attr.GetKey() {
				case "ndollar_amount":
					ndollarAmount = attr.GetValue()
				case "ndollar_denom":
					ndollarDenom = attr.GetValue()
				}
			}
			if ndollarAmount != "" && ndollarDenom != "" {
				return ndollarAmount, ndollarDenom
			}
		}
	}

	// Fallback: use the request amount (since it's 1:1)
	return "", ""
}

// SellNDollarRequest represents the request to sell NDOLLAR
type SellNDollarRequest struct {
	Amount string // Amount of NDOLLAR to convert back to unuah
}

// SellNDollarResponse represents the response after selling NDOLLAR
type SellNDollarResponse struct {
	TxHash      string
	UnuahAmount string
	Success     bool
	Error       string
}

// SellNDollarWithKey sells NDOLLAR back to unuah at 1:1 ratio
func (c *Client) SellNDollarWithKey(ctx context.Context, privKeyHex string, req SellNDollarRequest) (*SellNDollarResponse, error) {
	// Decode private key
	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key hex: %w", err)
	}

	privKey := &secp256k1.PrivKey{Key: privKeyBytes}
	pubKey := privKey.PubKey()
	addr := sdk.AccAddress(pubKey.Address())

	// Create the message
	msg := &stablecointypes.MsgSellNDollar{
		Seller: addr.String(),
		Amount: req.Amount,
	}

	// Sign and broadcast
	txHash, txResp, err := c.signAndBroadcastTx(ctx, addr.String(), privKey, pubKey, msg)
	if err != nil {
		return &SellNDollarResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Check for errors in the transaction
	if txResp.Code != 0 {
		return &SellNDollarResponse{
			TxHash:  txHash,
			Success: false,
			Error:   fmt.Sprintf("transaction failed with code %d: %s", txResp.Code, txResp.RawLog),
		}, fmt.Errorf("transaction failed with code %d: %s", txResp.Code, txResp.RawLog)
	}

	// Extract response data from transaction events or logs
	unuahAmount := c.extractSellNDollarMetadata(txResp)

	return &SellNDollarResponse{
		TxHash:      txHash,
		UnuahAmount: unuahAmount,
		Success:     true,
	}, nil
}

// extractSellNDollarMetadata extracts unuah amount from transaction response
func (c *Client) extractSellNDollarMetadata(txResp *abciv1beta1.TxResponse) string {
	if txResp == nil {
		return ""
	}
	// Try to extract from events
	for _, event := range txResp.Events {
		if event.GetType_() == "sell_ndollar" {
			for _, attr := range event.GetAttributes() {
				if attr.GetKey() == "unuah_amount" {
					return attr.GetValue()
				}
			}
		}
	}

	// Fallback: use the request amount (since it's 1:1)
	return ""
}
