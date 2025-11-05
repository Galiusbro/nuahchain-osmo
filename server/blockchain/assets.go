package blockchain

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	assetstypes "github.com/osmosis-labs/osmosis/v30/x/assets/types"
)

// EnsureAssetRequest represents a request to ensure an asset exists
type EnsureAssetRequest struct {
	Creator string // Creator address
	Symbol  string // Asset symbol (e.g., "GOLD", "BTC")
}

// EnsureAssetResponse represents the response from ensuring an asset
type EnsureAssetResponse struct {
	TxHash  string
	Success bool
	Error   string
}

// BuyAssetRequest represents a request to buy an asset
type BuyAssetRequest struct {
	Buyer         string // Buyer address
	Symbol        string // Asset symbol
	Denom         string // Payment denom (e.g., "NDOLLAR", "unuah", or factory denom)
	Amount        string // Amount in the specified denom
	AmountNDOLLAR string // Deprecated: amount in NDOLLAR (for backward compatibility)
}

// BuyAssetResponse represents the response from buying an asset
type BuyAssetResponse struct {
	TxHash     string
	BaseAmount string // Amount of asset received
	Success    bool
	Error      string
}

// SellAssetRequest represents a request to sell an asset
type SellAssetRequest struct {
	Seller     string // Seller address
	Symbol     string // Asset symbol
	BaseAmount string // Amount of asset to sell
}

// SellAssetResponse represents the response from selling an asset
type SellAssetResponse struct {
	TxHash        string
	PayoutNDOLLAR string // Amount of NDOLLAR received
	Success       bool
	Error         string
}

// EnsureAssetWithKey ensures an asset exists by signing and broadcasting a transaction
func (c *Client) EnsureAssetWithKey(
	ctx context.Context,
	req EnsureAssetRequest,
	privKeyBytes []byte,
) (*EnsureAssetResponse, error) {
	// Validate creator address
	creatorAddr, err := sdk.AccAddressFromBech32(req.Creator)
	if err != nil {
		return nil, fmt.Errorf("invalid creator address: %w", err)
	}

	// Create secp256k1 private key from bytes
	if len(privKeyBytes) != 32 {
		return nil, fmt.Errorf("invalid private key length: expected 32 bytes, got %d", len(privKeyBytes))
	}
	privKey := &secp256k1.PrivKey{Key: privKeyBytes}

	// Verify the address matches the private key
	pubKey := privKey.PubKey()
	derivedAddr := sdk.AccAddress(pubKey.Address())
	if !creatorAddr.Equals(derivedAddr) {
		return nil, fmt.Errorf("private key does not match creator address")
	}

	// Create message
	msg := &assetstypes.MsgEnsureAsset{
		Creator: req.Creator,
		Symbol:  req.Symbol,
	}

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid message: %w", err)
	}

	// Sign and broadcast transaction
	txHash, _, err := c.signAndBroadcastTx(ctx, req.Creator, privKey, pubKey, msg)
	if err != nil {
		return &EnsureAssetResponse{
			Success: false,
			TxHash:  txHash,
			Error:   err.Error(),
		}, err
	}

	return &EnsureAssetResponse{
		TxHash:  txHash,
		Success: true,
	}, nil
}

// BuyAssetWithKey buys an asset by signing and broadcasting a transaction
func (c *Client) BuyAssetWithKey(
	ctx context.Context,
	req BuyAssetRequest,
	privKeyBytes []byte,
) (*BuyAssetResponse, error) {
	// Validate buyer address
	buyerAddr, err := sdk.AccAddressFromBech32(req.Buyer)
	if err != nil {
		return nil, fmt.Errorf("invalid buyer address: %w", err)
	}

	// Create secp256k1 private key from bytes
	if len(privKeyBytes) != 32 {
		return nil, fmt.Errorf("invalid private key length: expected 32 bytes, got %d", len(privKeyBytes))
	}
	privKey := &secp256k1.PrivKey{Key: privKeyBytes}

	// Verify the address matches the private key
	pubKey := privKey.PubKey()
	derivedAddr := sdk.AccAddress(pubKey.Address())
	if !buyerAddr.Equals(derivedAddr) {
		return nil, fmt.Errorf("private key does not match buyer address")
	}

	// Create message - support both old and new format
	msg := &assetstypes.MsgBuyAsset{
		Buyer:          req.Buyer,
		Symbol:         req.Symbol,
		Amount_NDOLLAR: req.AmountNDOLLAR, // Deprecated but still supported
		Denom:          req.Denom,
		Amount:         req.Amount,
	}

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid message: %w", err)
	}

	// Sign and broadcast transaction
	txHash, _, err := c.signAndBroadcastTx(ctx, req.Buyer, privKey, pubKey, msg)
	if err != nil {
		return &BuyAssetResponse{
			Success: false,
			TxHash:  txHash,
			Error:   err.Error(),
		}, err
	}

	// Extract base amount from events (will be empty for SYNC mode, can be queried later)
	baseAmount := ""

	return &BuyAssetResponse{
		TxHash:     txHash,
		BaseAmount: baseAmount,
		Success:    true,
	}, nil
}

// SellAssetWithKey sells an asset by signing and broadcasting a transaction
func (c *Client) SellAssetWithKey(
	ctx context.Context,
	req SellAssetRequest,
	privKeyBytes []byte,
) (*SellAssetResponse, error) {
	// Validate seller address
	sellerAddr, err := sdk.AccAddressFromBech32(req.Seller)
	if err != nil {
		return nil, fmt.Errorf("invalid seller address: %w", err)
	}

	// Create secp256k1 private key from bytes
	if len(privKeyBytes) != 32 {
		return nil, fmt.Errorf("invalid private key length: expected 32 bytes, got %d", len(privKeyBytes))
	}
	privKey := &secp256k1.PrivKey{Key: privKeyBytes}

	// Verify the address matches the private key
	pubKey := privKey.PubKey()
	derivedAddr := sdk.AccAddress(pubKey.Address())
	if !sellerAddr.Equals(derivedAddr) {
		return nil, fmt.Errorf("private key does not match seller address")
	}

	// Create message
	msg := &assetstypes.MsgSellAsset{
		Seller:     req.Seller,
		Symbol:     req.Symbol,
		BaseAmount: req.BaseAmount,
	}

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid message: %w", err)
	}

	// Sign and broadcast transaction
	txHash, _, err := c.signAndBroadcastTx(ctx, req.Seller, privKey, pubKey, msg)
	if err != nil {
		return &SellAssetResponse{
			Success: false,
			TxHash:  txHash,
			Error:   err.Error(),
		}, err
	}

	// Extract payout from events (will be empty for SYNC mode, can be queried later)
	payoutNDOLLAR := ""

	return &SellAssetResponse{
		TxHash:        txHash,
		PayoutNDOLLAR: payoutNDOLLAR,
		Success:       true,
	}, nil
}
