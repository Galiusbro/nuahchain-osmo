package blockchain

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	abciv1beta1 "cosmossdk.io/api/cosmos/base/abci/v1beta1"
	txservice "cosmossdk.io/api/cosmos/tx/v1beta1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	leveragetypes "github.com/osmosis-labs/osmosis/v30/x/leverage/types"
)

// OpenLeveragePositionRequest represents a request to open a leveraged (margin) position.
type OpenLeveragePositionRequest struct {
	Owner       string // Account address of the trader
	Symbol      string // Asset symbol
	Side        string // "long" or "short"
	QuoteAmount string // Margin amount in micro NDOLLAR
	Leverage    string // Leverage multiplier (decimal string, e.g. "3")
}

// OpenLeveragePositionResponse represents the response from opening a leveraged position.
type OpenLeveragePositionResponse struct {
	TxHash       string
	PositionID   uint64
	BaseQuantity string
	EntryPrice   string
	Leverage     string
	Success      bool
	Error        string
}

// CloseLeveragePositionRequest represents a request to close an existing leveraged position.
type CloseLeveragePositionRequest struct {
	Owner      string
	PositionID uint64
}

// CloseLeveragePositionResponse represents the response from closing a leveraged position.
type CloseLeveragePositionResponse struct {
	TxHash  string
	Pnl     string
	Success bool
	Error   string
}

// OpenLeveragePositionWithKey opens a leveraged position by signing and broadcasting a transaction with the provided private key.
func (c *Client) OpenLeveragePositionWithKey(
	ctx context.Context,
	req OpenLeveragePositionRequest,
	privKeyBytes []byte,
) (*OpenLeveragePositionResponse, error) {
	ownerAddr, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return &OpenLeveragePositionResponse{Success: false, Error: fmt.Sprintf("invalid owner address: %v", err)}, fmt.Errorf("invalid owner address: %w", err)
	}

	if len(privKeyBytes) != 32 {
		err := fmt.Errorf("invalid private key length: expected 32 bytes, got %d", len(privKeyBytes))
		return &OpenLeveragePositionResponse{Success: false, Error: err.Error()}, err
	}

	privKey := &secp256k1.PrivKey{Key: privKeyBytes}
	pubKey := privKey.PubKey()
	derivedAddr := sdk.AccAddress(pubKey.Address())
	if !ownerAddr.Equals(derivedAddr) {
		err := fmt.Errorf("private key does not match owner address")
		return &OpenLeveragePositionResponse{Success: false, Error: err.Error()}, err
	}

	side, err := parseLeverageSide(req.Side)
	if err != nil {
		return &OpenLeveragePositionResponse{Success: false, Error: err.Error()}, err
	}

	msg := leveragetypes.NewMsgOpenPosition(req.Owner, req.Symbol, side, req.QuoteAmount, req.Leverage)
	if err := msg.ValidateBasic(); err != nil {
		return &OpenLeveragePositionResponse{Success: false, Error: err.Error()}, fmt.Errorf("invalid message: %w", err)
	}

	txHash, txResp, err := c.signAndBroadcastTx(ctx, req.Owner, privKey, pubKey, msg)
	if err != nil {
		return &OpenLeveragePositionResponse{TxHash: txHash, Success: false, Error: err.Error()}, err
	}

	// Try to extract from sync response first
	// Note: In BROADCAST_MODE_SYNC, Data may be empty initially, but we try anyway
	fmt.Printf("DEBUG OpenLeveragePositionWithKey: txResp.Data length: %d\n", len(txResp.Data))
	positionID, baseQty, entryPrice, leverage := c.extractOpenPositionMetadata(txResp)
	fmt.Printf("DEBUG OpenLeveragePositionWithKey: After sync extract, positionID: %d\n", positionID)

	// If Data is empty or positionID not found, wait for transaction to be included and query it
	if (txResp.Data == "" || positionID == 0) && txHash != "" {
		fmt.Printf("DEBUG OpenLeveragePositionWithKey: Starting retry loop for txHash: %s\n", txHash)
		// Wait a bit for transaction to be included in block
		time.Sleep(3 * time.Second)

		// Retry up to 3 times with increasing delays
		for attempt := 1; attempt <= 3; attempt++ {
			if attempt > 1 {
				delay := time.Duration(attempt*2) * time.Second
				fmt.Printf("DEBUG OpenLeveragePositionWithKey: Attempt %d, waiting %v\n", attempt, delay)
				time.Sleep(delay)
			} else {
				fmt.Printf("DEBUG OpenLeveragePositionWithKey: Attempt %d\n", attempt)
			}

			// Query the transaction via gRPC to get full response with Data field
			getTxResp, err := c.txClient.GetTx(ctx, &txservice.GetTxRequest{Hash: txHash})
			if err != nil {
				fmt.Printf("DEBUG OpenLeveragePositionWithKey: GetTx error on attempt %d: %v\n", attempt, err)
				continue
			}
			if getTxResp == nil || getTxResp.TxResponse == nil {
				fmt.Printf("DEBUG OpenLeveragePositionWithKey: Empty response on attempt %d\n", attempt)
				continue
			}
			if getTxResp.TxResponse.Data == "" {
				fmt.Printf("DEBUG OpenLeveragePositionWithKey: Data still empty on attempt %d\n", attempt)
				continue
			}
			fmt.Printf("DEBUG OpenLeveragePositionWithKey: Got Data on attempt %d, length: %d\n", attempt, len(getTxResp.TxResponse.Data))
			// Re-extract from queried transaction
			positionID, baseQty, entryPrice, leverage = c.extractOpenPositionMetadata(getTxResp.TxResponse)
			fmt.Printf("DEBUG OpenLeveragePositionWithKey: After retry extract, positionID: %d\n", positionID)
			if positionID != 0 {
				fmt.Printf("DEBUG OpenLeveragePositionWithKey: Successfully extracted positionID: %d\n", positionID)
				break
			}
		}
	}

	if leverage == "" {
		leverage = req.Leverage
	}

	return &OpenLeveragePositionResponse{
		TxHash:       txHash,
		PositionID:   positionID,
		BaseQuantity: baseQty,
		EntryPrice:   entryPrice,
		Leverage:     leverage,
		Success:      true,
	}, nil
}

// CloseLeveragePositionWithKey closes an existing leveraged position using the provided private key.
func (c *Client) CloseLeveragePositionWithKey(
	ctx context.Context,
	req CloseLeveragePositionRequest,
	privKeyBytes []byte,
) (*CloseLeveragePositionResponse, error) {
	ownerAddr, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return &CloseLeveragePositionResponse{Success: false, Error: fmt.Sprintf("invalid owner address: %v", err)}, fmt.Errorf("invalid owner address: %w", err)
	}

	if len(privKeyBytes) != 32 {
		err := fmt.Errorf("invalid private key length: expected 32 bytes, got %d", len(privKeyBytes))
		return &CloseLeveragePositionResponse{Success: false, Error: err.Error()}, err
	}

	privKey := &secp256k1.PrivKey{Key: privKeyBytes}
	pubKey := privKey.PubKey()
	derivedAddr := sdk.AccAddress(pubKey.Address())
	if !ownerAddr.Equals(derivedAddr) {
		err := fmt.Errorf("private key does not match owner address")
		return &CloseLeveragePositionResponse{Success: false, Error: err.Error()}, err
	}

	msg := leveragetypes.NewMsgClosePosition(req.Owner, req.PositionID)
	if err := msg.ValidateBasic(); err != nil {
		return &CloseLeveragePositionResponse{Success: false, Error: err.Error()}, fmt.Errorf("invalid message: %w", err)
	}

	txHash, txResp, err := c.signAndBroadcastTx(ctx, req.Owner, privKey, pubKey, msg)
	if err != nil {
		return &CloseLeveragePositionResponse{TxHash: txHash, Success: false, Error: err.Error()}, err
	}

	pnl := c.extractClosePositionMetadata(txResp)

	return &CloseLeveragePositionResponse{
		TxHash:  txHash,
		Pnl:     pnl,
		Success: true,
	}, nil
}

// parseLeverageSide normalises a textual side into the leverage module enum.
func parseLeverageSide(side string) (leveragetypes.Side, error) {
	switch strings.ToUpper(strings.TrimSpace(side)) {
	case "LONG", "SIDE_LONG":
		return leveragetypes.Side_SIDE_LONG, nil
	case "SHORT", "SIDE_SHORT":
		return leveragetypes.Side_SIDE_SHORT, nil
	default:
		return leveragetypes.Side_SIDE_UNSPECIFIED, fmt.Errorf("invalid leverage side: %s", side)
	}
}

// extractOpenPositionMetadata attempts to unpack MsgOpenPositionResponse from the tx response to retrieve position metadata.
func (c *Client) extractOpenPositionMetadata(txResp *abciv1beta1.TxResponse) (uint64, string, string, string) {
	if txResp == nil || txResp.Data == "" {
		return 0, "", "", ""
	}

	dataBytes, err := decodeTxData(txResp.Data)
	if err != nil {
		return 0, "", "", ""
	}

	txMsgData := &sdk.TxMsgData{}
	if err := c.encCfg.Codec.Unmarshal(dataBytes, txMsgData); err != nil {
		fmt.Printf("DEBUG extractOpenPositionMetadata: Failed to unmarshal TxMsgData: %v, data length: %d\n", err, len(dataBytes))
		return 0, "", "", ""
	}

	const typeURL = "/osmosis.leverage.v1.MsgOpenPositionResponse"
	fmt.Printf("DEBUG extractOpenPositionMetadata: Found %d message responses\n", len(txMsgData.MsgResponses))

	for i, any := range txMsgData.MsgResponses {
		fmt.Printf("DEBUG extractOpenPositionMetadata[%d]: TypeURL=%s, Value length=%d\n", i, any.TypeUrl, len(any.Value))
		if any.TypeUrl != typeURL {
			continue
		}
		// Unmarshal directly from protobuf bytes instead of using UnpackAny
		// UnpackAny requires registration, but we can unmarshal protobuf directly
		var msg leveragetypes.MsgOpenPositionResponse
		if err := c.encCfg.Codec.Unmarshal(any.Value, &msg); err != nil {
			fmt.Printf("DEBUG extractOpenPositionMetadata: Failed to unmarshal protobuf: %v\n", err)
			continue
		}
		if msg.Position == nil {
			fmt.Printf("DEBUG extractOpenPositionMetadata: Position is nil in response\n")
			return 0, "", "", ""
		}
		fmt.Printf("DEBUG extractOpenPositionMetadata: Successfully extracted position_id: %d\n", msg.Position.Id)
		return msg.Position.Id, msg.Position.BaseQty, msg.Position.EntryPrice, msg.Position.Leverage
	}

	fmt.Printf("DEBUG extractOpenPositionMetadata: No matching type URL found\n")

	return 0, "", "", ""
}

// extractClosePositionMetadata unpacks MsgClosePositionResponse to obtain realised PnL.
func (c *Client) extractClosePositionMetadata(txResp *abciv1beta1.TxResponse) string {
	if txResp == nil || txResp.Data == "" {
		return ""
	}

	dataBytes, err := decodeTxData(txResp.Data)
	if err != nil {
		return ""
	}

	txMsgData := &sdk.TxMsgData{}
	if err := c.encCfg.Codec.Unmarshal(dataBytes, txMsgData); err != nil {
		return ""
	}

	const typeURL = "/osmosis.leverage.v1.MsgClosePositionResponse"
	for _, any := range txMsgData.MsgResponses {
		if any.TypeUrl != typeURL {
			continue
		}
		// Unmarshal directly from protobuf bytes instead of using UnpackAny
		var msg leveragetypes.MsgClosePositionResponse
		if err := c.encCfg.Codec.Unmarshal(any.Value, &msg); err != nil {
			continue
		}
		return msg.Pnl
	}

	return ""
}

func decodeTxData(data string) ([]byte, error) {
	trimmed := strings.TrimPrefix(strings.TrimSpace(data), "0x")
	if trimmed == "" {
		return nil, fmt.Errorf("empty data")
	}
	return hex.DecodeString(trimmed)
}
