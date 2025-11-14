package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/anypb"
)

// DecodeTxFromResponse decodes a transaction from TxResponse.Tx (Any field)
// Returns the decoded transaction with messages, signers, fee, and memo
func (c *Client) DecodeTxFromResponse(txAny *anypb.Any) (*DecodedTx, error) {
	if txAny == nil {
		return nil, fmt.Errorf("transaction Any is nil")
	}

	// Decode the transaction bytes from Any
	txBytes := txAny.Value
	if len(txBytes) == 0 {
		return nil, fmt.Errorf("transaction bytes are empty")
	}

	// Decode using TxDecoder (same as in indexer)
	decodedTx, err := c.encCfg.TxConfig.TxDecoder()(txBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction: %w", err)
	}

	// Extract messages
	messages := decodedTx.GetMsgs()
	msgsData := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		msgData := map[string]interface{}{
			"type": sdk.MsgTypeURL(msg),
		}

		// Try to extract message fields based on type
		// This is a simplified version - in production you might want to use reflection
		// or type assertions for specific message types
		if msgJSON, err := c.encCfg.Codec.MarshalJSON(msg); err == nil {
			var msgMap map[string]interface{}
			if err := json.Unmarshal(msgJSON, &msgMap); err == nil {
				msgData["value"] = msgMap
			}
		}

		msgsData = append(msgsData, msgData)
	}

	// Extract fee
	var feeAmount map[string]interface{}
	if feeTx, ok := decodedTx.(sdk.FeeTx); ok {
		fee := feeTx.GetFee()
		if len(fee) > 0 {
			feeAmount = map[string]interface{}{
				"amount": fee.String(),
				"coins":  fee,
			}
		}
	}

	// Extract signers from messages using codec
	signers := make([]string, 0)
	signerMap := make(map[string]bool) // Use map to avoid duplicates
	for _, msg := range messages {
		// Use codec to get signers (works for both v1 and legacy messages)
		signerAddrs, _, err := c.encCfg.Codec.GetMsgV1Signers(msg)
		if err == nil && len(signerAddrs) > 0 {
			for _, addr := range signerAddrs {
				addrStr := sdk.AccAddress(addr).String()
				if !signerMap[addrStr] {
					signers = append(signers, addrStr)
					signerMap[addrStr] = true
				}
			}
		}
	}

	// Extract memo
	var memo string
	if memoTx, ok := decodedTx.(sdk.TxWithMemo); ok {
		memo = memoTx.GetMemo()
	}

	return &DecodedTx{
		Messages: messages,
		MessagesData: msgsData,
		Signers: signers,
		Fee:     feeAmount,
		Memo:    memo,
	}, nil
}

// DecodeTxFromBase64 decodes a transaction from base64-encoded bytes (from WebSocket)
func (c *Client) DecodeTxFromBase64(txBase64 string) (*DecodedTx, error) {
	txBytes, err := base64.StdEncoding.DecodeString(txBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Decode using TxDecoder
	decodedTx, err := c.encCfg.TxConfig.TxDecoder()(txBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction: %w", err)
	}

	// Extract messages
	messages := decodedTx.GetMsgs()
	msgsData := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		msgData := map[string]interface{}{
			"type": sdk.MsgTypeURL(msg),
		}

		if msgJSON, err := c.encCfg.Codec.MarshalJSON(msg); err == nil {
			var msgMap map[string]interface{}
			if err := json.Unmarshal(msgJSON, &msgMap); err == nil {
				msgData["value"] = msgMap
			}
		}

		msgsData = append(msgsData, msgData)
	}

	// Extract fee
	var feeAmount map[string]interface{}
	if feeTx, ok := decodedTx.(sdk.FeeTx); ok {
		fee := feeTx.GetFee()
		if len(fee) > 0 {
			feeAmount = map[string]interface{}{
				"amount": fee.String(),
				"coins":  fee,
			}
		}
	}

	// Extract signers from messages using codec
	signers := make([]string, 0)
	signerMap := make(map[string]bool) // Use map to avoid duplicates
	for _, msg := range messages {
		// Use codec to get signers (works for both v1 and legacy messages)
		signerAddrs, _, err := c.encCfg.Codec.GetMsgV1Signers(msg)
		if err == nil && len(signerAddrs) > 0 {
			for _, addr := range signerAddrs {
				addrStr := sdk.AccAddress(addr).String()
				if !signerMap[addrStr] {
					signers = append(signers, addrStr)
					signerMap[addrStr] = true
				}
			}
		}
	}

	// Extract memo
	var memo string
	if memoTx, ok := decodedTx.(sdk.TxWithMemo); ok {
		memo = memoTx.GetMemo()
	}

	return &DecodedTx{
		Messages: messages,
		MessagesData: msgsData,
		Signers: signers,
		Fee:     feeAmount,
		Memo:    memo,
	}, nil
}

// DecodedTx represents a decoded transaction with all its components
type DecodedTx struct {
	Messages     []sdk.Msg
	MessagesData []map[string]interface{} // JSON-serializable message data
	Signers      []string
	Fee          map[string]interface{}
	Memo         string
}

