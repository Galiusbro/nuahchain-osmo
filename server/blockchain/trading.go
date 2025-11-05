package blockchain

import (
	"context"
	"fmt"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txservice "cosmossdk.io/api/cosmos/tx/v1beta1"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bondingcurvetypes "github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
)

// BuyFromCurveRequest represents a request to buy tokens from bonding curve
type BuyFromCurveRequest struct {
	Trader        string // Buyer address
	Denom         string // Token denom to buy
	PaymentDenom  string // Payment currency (e.g., "unuah")
	PaymentAmount string // Amount to pay
	MinTokensOut  string // Minimum tokens to receive (slippage protection)
}

// BuyFromCurveResponse represents the response from buying tokens
type BuyFromCurveResponse struct {
	TxHash    string
	TokensOut string
	PricePaid string
	Success   bool
	Error     string
}

// SellToCurveRequest represents a request to sell tokens to bonding curve
type SellToCurveRequest struct {
	Trader        string // Seller address
	Denom         string // Token denom to sell
	TokenAmount   string // Amount of tokens to sell
	PaymentDenom  string // Payment currency to receive
	MinPaymentOut string // Minimum payment to receive (slippage protection)
}

// SellToCurveResponse represents the response from selling tokens
type SellToCurveResponse struct {
	TxHash        string
	PaymentOut    string
	PriceReceived string
	Success       bool
	Error         string
}

// BuyFromCurveWithKey buys tokens from bonding curve using a private key
func (c *Client) BuyFromCurveWithKey(ctx context.Context, req BuyFromCurveRequest, privKeyBytes []byte) (*BuyFromCurveResponse, error) {
	// Create private key from bytes
	privKey := &secp256k1.PrivKey{Key: privKeyBytes}
	pubKey := privKey.PubKey()

	// Validate trader address
	if _, err := sdk.AccAddressFromBech32(req.Trader); err != nil {
		return nil, fmt.Errorf("invalid trader address: %w", err)
	}

	// Set default payment denom if not provided
	if req.PaymentDenom == "" {
		req.PaymentDenom = "unuah"
	}

	// Create buy message
	msg := &bondingcurvetypes.MsgBuyFromCurve{
		Trader:        req.Trader,
		Denom:         req.Denom,
		PaymentDenom:  req.PaymentDenom,
		PaymentAmount: req.PaymentAmount,
		MinTokensOut:  req.MinTokensOut,
	}

	// Sign and broadcast transaction
	txHash, _, err := c.signAndBroadcastTx(ctx, req.Trader, privKey, pubKey, msg)
	if err != nil {
		return &BuyFromCurveResponse{
			Success: false,
			TxHash:  txHash,
			Error:   err.Error(),
		}, err
	}

	// Extract tokens out from response
	// Note: Events extraction requires querying the transaction after it's included in a block
	// For now, we return empty and let the client query if needed
	tokensOut := ""

	return &BuyFromCurveResponse{
		TxHash:    txHash,
		TokensOut: tokensOut,
		Success:   true,
	}, nil
}

// SellToCurveWithKey sells tokens to bonding curve using a private key
func (c *Client) SellToCurveWithKey(ctx context.Context, req SellToCurveRequest, privKeyBytes []byte) (*SellToCurveResponse, error) {
	// Create private key from bytes
	privKey := &secp256k1.PrivKey{Key: privKeyBytes}
	pubKey := privKey.PubKey()

	// Validate trader address
	if _, err := sdk.AccAddressFromBech32(req.Trader); err != nil {
		return nil, fmt.Errorf("invalid trader address: %w", err)
	}

	// Set default payment denom if not provided
	if req.PaymentDenom == "" {
		req.PaymentDenom = "unuah"
	}

	// Create sell message
	msg := &bondingcurvetypes.MsgSellToCurve{
		Trader:        req.Trader,
		Denom:         req.Denom,
		TokenAmount:   req.TokenAmount,
		PaymentDenom:  req.PaymentDenom,
		MinPaymentOut: req.MinPaymentOut,
	}

	// Sign and broadcast transaction
	txHash, _, err := c.signAndBroadcastTx(ctx, req.Trader, privKey, pubKey, msg)
	if err != nil {
		return &SellToCurveResponse{
			Success: false,
			TxHash:  txHash,
			Error:   err.Error(),
		}, err
	}

	// Extract payment out from response
	// Note: Events extraction from GetTxResponse requires querying the transaction
	// For now, we return empty and let the client query if needed
	paymentOut := ""

	return &SellToCurveResponse{
		TxHash:     txHash,
		PaymentOut: paymentOut,
		Success:    true,
	}, nil
}

// signAndBroadcastTx is a helper method to sign and broadcast transactions
func (c *Client) signAndBroadcastTx(ctx context.Context, address string, privKey *secp256k1.PrivKey, pubKey cryptotypes.PubKey, msg sdk.Msg) (string, *txservice.GetTxResponse, error) {
	// Get account info
	accountResp, err := c.authClient.Account(ctx, &authtypes.QueryAccountRequest{
		Address: address,
	})
	if err != nil {
		return "", nil, fmt.Errorf("failed to get account: %w", err)
	}

	var account authtypes.AccountI
	if err := c.encCfg.InterfaceRegistry.UnpackAny(accountResp.Account, &account); err != nil {
		return "", nil, fmt.Errorf("failed to unpack account: %w", err)
	}

	// Build transaction
	txBuilder := c.encCfg.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msg); err != nil {
		return "", nil, fmt.Errorf("failed to set messages: %w", err)
	}

	// Set gas and fees (trading requires less gas than token creation)
	txBuilder.SetGasLimit(500000)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin("unuah", sdkmath.NewInt(2000))))

	// Create signer data
	signerData := authsigning.SignerData{
		ChainID:       c.chainID,
		AccountNumber: account.GetAccountNumber(),
		Sequence:      account.GetSequence(),
	}

	// Create signature
	sig := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: signerData.Sequence,
	}

	if err := txBuilder.SetSignatures(sig); err != nil {
		return "", nil, fmt.Errorf("failed to set signatures: %w", err)
	}

	// Get sign bytes
	protoSignMode, err := authsigning.APISignModeToInternal(signingv1beta1.SignMode_SIGN_MODE_DIRECT)
	if err != nil {
		return "", nil, fmt.Errorf("failed to convert sign mode: %w", err)
	}

	signBytes, err := authsigning.GetSignBytesAdapter(
		ctx,
		c.encCfg.TxConfig.SignModeHandler(),
		protoSignMode,
		signerData,
		txBuilder.GetTx(),
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get sign bytes: %w", err)
	}

	// Sign the transaction
	sigBytes, err := privKey.Sign(signBytes)
	if err != nil {
		return "", nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Update signature
	sig = signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: sigBytes,
		},
		Sequence: signerData.Sequence,
	}

	if err := txBuilder.SetSignatures(sig); err != nil {
		return "", nil, fmt.Errorf("failed to update signatures: %w", err)
	}

	// Encode transaction
	txBytes, err := c.encCfg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return "", nil, fmt.Errorf("failed to encode transaction: %w", err)
	}

	// Compute transaction hash
	txHash := computeTxHash(txBytes)

	// Broadcast transaction
	broadcastReq := &txservice.BroadcastTxRequest{
		Mode:    txservice.BroadcastMode_BROADCAST_MODE_SYNC,
		TxBytes: txBytes,
	}

	broadcastResp, err := c.txClient.BroadcastTx(ctx, broadcastReq)
	if err != nil {
		return txHash, nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	txResponse := broadcastResp.TxResponse
	if txResponse == nil {
		return txHash, nil, fmt.Errorf("empty transaction response")
	}

	// Check if transaction succeeded
	if txResponse.Code != 0 {
		errorMsg := ""
		if txResponse.RawLog != "" {
			errorMsg = txResponse.RawLog
		} else if txResponse.Info != "" {
			errorMsg = txResponse.Info
		} else {
			errorMsg = "transaction failed"
		}
		return txHash, nil, fmt.Errorf("transaction failed with code %d: %s", txResponse.Code, errorMsg)
	}

	// Return successful response (we don't query the tx again, just return the sync response)
	return txHash, nil, nil
}
