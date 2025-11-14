package blockchain

import (
	"context"
	"encoding/hex"
	"fmt"

	abciv1beta1 "cosmossdk.io/api/cosmos/base/abci/v1beta1"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	exchangetypes "github.com/osmosis-labs/osmosis/v30/x/exchange/types"
)

// ExchangeTokensRequest represents a request to exchange tokens for unuah
type ExchangeTokensRequest struct {
	TokenIn    sdk.Coin
	MinNuahOut math.Int
}

// ExchangeTokensResponse represents the response from exchanging tokens
type ExchangeTokensResponse struct {
	TxHash   string
	NuahOut  math.Int
	Success  bool
	ErrorMsg string
}

// ExchangeTokensWithKey exchanges supported tokens (ETH, BTC, USDC, etc.) for unuah using a private key
func (c *Client) ExchangeTokensWithKey(
	ctx context.Context,
	privKeyHex string,
	req ExchangeTokensRequest,
) (*ExchangeTokensResponse, error) {
	// Decode private key
	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key hex: %w", err)
	}

	privKey := &secp256k1.PrivKey{Key: privKeyBytes}
	pubKey := privKey.PubKey()
	addr := sdk.AccAddress(pubKey.Address())

	// Create exchange message
	msg := exchangetypes.NewMsgExchangeTokens(
		addr.String(),
		req.TokenIn,
		req.MinNuahOut,
	)

	// Sign and broadcast transaction
	txHash, txResp, err := c.signAndBroadcastTx(ctx, addr.String(), privKey, pubKey, msg)
	if err != nil {
		return &ExchangeTokensResponse{
			TxHash:   txHash,
			Success:  false,
			ErrorMsg: err.Error(),
		}, err
	}

	// Extract nuah output from transaction response if available
	nuahOut, err := extractExchangeOutput(txResp)
	if err != nil {
		fmt.Printf("Warning: couldn't extract nuah output: %v\n", err)
		nuahOut = math.ZeroInt()
	}

	return &ExchangeTokensResponse{
		TxHash:   txHash,
		NuahOut:  nuahOut,
		Success:  true,
		ErrorMsg: "",
	}, nil
}

// extractExchangeOutput extracts the nuah output amount from the transaction response
func extractExchangeOutput(txResp *abciv1beta1.TxResponse) (math.Int, error) {
	if txResp == nil || txResp.Data == "" {
		return math.ZeroInt(), fmt.Errorf("empty transaction response data")
	}

	// Decode transaction response data
	var txMsgData sdk.TxMsgData
	if err := txMsgData.Unmarshal([]byte(txResp.Data)); err != nil {
		return math.ZeroInt(), fmt.Errorf("failed to unmarshal tx data: %w", err)
	}

	// Find exchange response in message responses
	for _, msgData := range txMsgData.MsgResponses {
		var exchangeResp exchangetypes.MsgExchangeTokensResponse
		if err := exchangeResp.Unmarshal(msgData.Value); err != nil {
			continue // Not an exchange response, skip
		}
		return exchangeResp.NuahOut, nil
	}

	return math.ZeroInt(), fmt.Errorf("exchange response not found in transaction data")
}
