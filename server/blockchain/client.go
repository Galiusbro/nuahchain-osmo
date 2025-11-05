package blockchain

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"

	abcitypes "cosmossdk.io/api/tendermint/abci"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txservice "cosmossdk.io/api/cosmos/tx/v1beta1"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	assetstypes "github.com/osmosis-labs/osmosis/v30/x/assets/types"
	bondingcurvetypes "github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
	usertokentypes "github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

// Client is a blockchain client for interacting with Cosmos SDK
type Client struct {
	nodeURL       string
	conn          *grpc.ClientConn
	msgClient     usertokentypes.MsgClient
	bondingClient bondingcurvetypes.MsgClient
	assetsClient  assetstypes.MsgClient
	txClient      txservice.ServiceClient
	authClient    authtypes.QueryClient
	bankClient    banktypes.QueryClient
	chainID       string
	encCfg        EncodingConfig
	keyring       keyring.Keyring
}

// EncodingConfig wraps the encoding configuration
type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
}

// NewClient creates a new blockchain client
func NewClient(nodeURL, chainID string) (*Client, error) {
	if nodeURL == "" {
		return nil, fmt.Errorf("node URL is required")
	}
	if chainID == "" {
		return nil, fmt.Errorf("chain ID is required")
	}

	// Create gRPC connection
	conn, err := grpc.NewClient(nodeURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to node: %w", err)
	}

	// Create clients
	msgClient := usertokentypes.NewMsgClient(conn)
	bondingClient := bondingcurvetypes.NewMsgClient(conn)
	assetsClient := assetstypes.NewMsgClient(conn)
	txClient := txservice.NewServiceClient(conn)
	authClient := authtypes.NewQueryClient(conn)
	bankClient := banktypes.NewQueryClient(conn)

	// Create encoding config
	encCfg, err := makeEncodingConfig()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create encoding config: %w", err)
	}

	// Create in-memory keyring (not used for signing, but required by client.Context)
	kb := keyring.NewInMemory(encCfg.Codec)

	return &Client{
		nodeURL:       nodeURL,
		conn:          conn,
		msgClient:     msgClient,
		bondingClient: bondingClient,
		assetsClient:  assetsClient,
		txClient:      txClient,
		authClient:    authClient,
		bankClient:    bankClient,
		chainID:       chainID,
		encCfg:        encCfg,
		keyring:       kb,
	}, nil
}

// makeEncodingConfig creates encoding config for transaction signing
func makeEncodingConfig() (EncodingConfig, error) {
	interfaceRegistry := types.NewInterfaceRegistry()

	// Register standard Cosmos SDK interfaces (this includes auth types)
	std.RegisterInterfaces(interfaceRegistry)

	// Explicitly register auth types to ensure BaseAccount is registered
	authtypes.RegisterInterfaces(interfaceRegistry)

	// Register additional module interfaces
	authztypes.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	distrtypes.RegisterInterfaces(interfaceRegistry)
	govtypesv1.RegisterInterfaces(interfaceRegistry)
	govtypesv1beta1.RegisterInterfaces(interfaceRegistry)
	stakingtypes.RegisterInterfaces(interfaceRegistry)

	// Register usertoken interfaces
	usertokentypes.RegisterInterfaces(interfaceRegistry)

	// Register bondingcurve interfaces
	bondingcurvetypes.RegisterInterfaces(interfaceRegistry)

	// Register assets interfaces
	assetstypes.RegisterInterfaces(interfaceRegistry)

	codec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := authtx.NewTxConfig(codec, authtx.DefaultSignModes)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             codec,
		TxConfig:          txConfig,
	}, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CreateTokenRequest represents a request to create a token
type CreateTokenRequest struct {
	Creator     string
	Name        string
	Symbol      string
	Image       string
	Description string
}

// CreateTokenResponse represents the response from creating a token
type CreateTokenResponse struct {
	Denom   string `json:"denom"`
	TxHash  string `json:"tx_hash,omitempty"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// CreateTokenWithKey creates a token by signing and broadcasting a transaction
func (c *Client) CreateTokenWithKey(
	ctx context.Context,
	req CreateTokenRequest,
	privKeyBytes []byte,
) (*CreateTokenResponse, error) {
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
	msg := &usertokentypes.MsgCreateToken{
		Creator:     req.Creator,
		Name:        req.Name,
		Symbol:      req.Symbol,
		Image:       req.Image,
		Description: req.Description,
	}

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid message: %w", err)
	}

	// Get account info to get account number and sequence
	accountResp, err := c.authClient.Account(ctx, &authtypes.QueryAccountRequest{
		Address: req.Creator,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	var account authtypes.AccountI
	if err := c.encCfg.InterfaceRegistry.UnpackAny(accountResp.Account, &account); err != nil {
		return nil, fmt.Errorf("failed to unpack account: %w", err)
	}

	// Build transaction
	txBuilder := c.encCfg.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msg); err != nil {
		return nil, fmt.Errorf("failed to set messages: %w", err)
	}

	// Set gas and fees
	// Token creation requires significant gas for minting distributions (~1.1M gas)
	txBuilder.SetGasLimit(1500000)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin("unuah", sdkmath.NewInt(5000))))

	// Create signer data
	signerData := authsigning.SignerData{
		ChainID:       c.chainID,
		AccountNumber: account.GetAccountNumber(),
		Sequence:      account.GetSequence(),
	}

	// Create signature with empty signature first (needed for SIGN_MODE_DIRECT)
	sig := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: signerData.Sequence,
	}

	if err := txBuilder.SetSignatures(sig); err != nil {
		return nil, fmt.Errorf("failed to set signatures: %w", err)
	}

	// Get sign bytes - convert to internal sign mode
	protoSignMode, err := authsigning.APISignModeToInternal(signingv1beta1.SignMode_SIGN_MODE_DIRECT)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sign mode: %w", err)
	}

	signBytes, err := authsigning.GetSignBytesAdapter(
		ctx,
		c.encCfg.TxConfig.SignModeHandler(),
		protoSignMode,
		signerData,
		txBuilder.GetTx(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get sign bytes: %w", err)
	}

	// Sign the transaction
	sigBytes, err := privKey.Sign(signBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Update signature with actual signature
	sig = signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: sigBytes,
		},
		Sequence: signerData.Sequence,
	}

	if err := txBuilder.SetSignatures(sig); err != nil {
		return nil, fmt.Errorf("failed to update signatures: %w", err)
	}

	// Encode transaction
	txBytes, err := c.encCfg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, fmt.Errorf("failed to encode transaction: %w", err)
	}

	// Broadcast transaction
	// Use BROADCAST_MODE_SYNC to get CheckTx response immediately
	// SYNC mode returns events in the response, which we can use to extract denom
	broadcastReq := &txservice.BroadcastTxRequest{
		Mode:    txservice.BroadcastMode_BROADCAST_MODE_SYNC,
		TxBytes: txBytes,
	}

	broadcastResp, err := c.txClient.BroadcastTx(ctx, broadcastReq)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	txResponse := broadcastResp.TxResponse
	if txResponse == nil {
		return nil, fmt.Errorf("empty transaction response")
	}

	// Determine transaction hash, preferring the node's response and
	// falling back to recomputing from the raw tx bytes.
	txHash := sanitizeTxHash(txResponse.Txhash)
	if txHash == "" && len(txBytes) > 0 {
		txHash = computeTxHash(txBytes)
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
		return &CreateTokenResponse{
			Success: false,
			TxHash:  txHash,
			Error:   errorMsg,
		}, fmt.Errorf("transaction failed with code %d: %s", txResponse.Code, errorMsg)
	}

	// Extract denom from transaction events
	denom := extractDenomFromEvents(txResponse.Events)

	// If denom not found in SYNC response events, wait a bit and query the transaction
	// SYNC mode may not include all events, so we query the tx after it's included in a block
	if denom == "" && txHash != "" {
		// Wait a moment for transaction to be included in a block
		time.Sleep(3 * time.Second)

		// Try to query the transaction to get full events
		denom = c.queryDenomFromTx(ctx, txHash)
	}

	// If still not found, try to extract from RawLog
	if denom == "" && txResponse.RawLog != "" {
		denom = extractDenomFromLogs(txResponse.RawLog)
	}

	// Fallback: compute denom from formula factory/{creator}/{symbol_lowercase}
	// This is the standard format used by tokenfactory and usertoken modules
	if denom == "" {
		symbolLower := strings.ToLower(req.Symbol)
		// Sanitize symbol (remove invalid characters, similar to keeper logic)
		symbolLower = strings.TrimSpace(symbolLower)
		if len(symbolLower) > 44 {
			symbolLower = symbolLower[:44]
		}
		denom = fmt.Sprintf("factory/%s/%s", req.Creator, symbolLower)
	}

	return &CreateTokenResponse{
		Denom:   denom,
		TxHash:  txHash,
		Success: true,
	}, nil
}

// extractDenomFromEvents extracts the denom from transaction events
func extractDenomFromEvents(events []*abcitypes.Event) string {
	for _, event := range events {
		eventType := event.GetType_()

		// Check for create_token event (emitted by usertoken module)
		if eventType == "create_token" {
			for _, attr := range event.GetAttributes() {
				key := attr.GetKey()
				if key == "denom" || key == "denom.denom" {
					return attr.GetValue()
				}
			}
		}

		// Check message events for action type
		if eventType == "message" {
			var hasCreateTokenAction bool
			for _, attr := range event.GetAttributes() {
				if attr.GetKey() == "action" {
					value := attr.GetValue()
					if value == "/osmosis.usertoken.v1beta1.MsgCreateToken" ||
						value == "osmosis.usertoken.v1beta1.MsgCreateToken" ||
						strings.Contains(value, "MsgCreateToken") {
						hasCreateTokenAction = true
						break
					}
				}
			}
			// If we found create token action, look for denom in attributes
			if hasCreateTokenAction {
				for _, attr := range event.GetAttributes() {
					if attr.GetKey() == "denom" {
						return attr.GetValue()
					}
				}
			}
		}

		// Check for token factory events
		if strings.Contains(eventType, "factory") || strings.Contains(eventType, "denom") {
			for _, attr := range event.GetAttributes() {
				key := attr.GetKey()
				if key == "denom" || key == "new_token_denom" || key == "factory_denom" {
					return attr.GetValue()
				}
			}
		}
	}

	// If denom not found in events, return empty
	return ""
}

// queryDenomFromTx queries a transaction by hash and extracts denom from events
func (c *Client) queryDenomFromTx(ctx context.Context, txHash string) string {
	// txHash is already in hex string format (lowercase)
	// GetTx expects hex string, so we can use it directly
	getTxReq := &txservice.GetTxRequest{
		Hash: txHash,
	}

	// Try a few times in case tx is not yet in block
	for i := 0; i < 5; i++ {
		getTxResp, err := c.txClient.GetTx(ctx, getTxReq)
		if err == nil && getTxResp != nil && getTxResp.TxResponse != nil {
			// Extract denom from events
			denom := extractDenomFromEvents(getTxResp.TxResponse.Events)
			if denom != "" {
				return denom
			}
		}
		// Wait longer between attempts
		if i < 4 {
			time.Sleep(2 * time.Second)
		}
	}

	return ""
}

// extractDenomFromLogs extracts denom from transaction logs (JSON format)
func extractDenomFromLogs(logStr string) string {
	// Logs are JSON array, try to parse and find denom in events
	// This is a fallback if events parsing fails
	// Format: [{"events":[{"type":"create_token","attributes":[{"key":"denom","value":"..."}]}]}]
	if strings.Contains(logStr, "denom") {
		// Try to extract factory/... pattern
		re := regexp.MustCompile(`factory/[a-z0-9]+/[a-z0-9_]+`)
		if matches := re.FindString(logStr); matches != "" {
			return matches
		}
	}
	return ""
}

// Helper function to create client context (for compatibility if needed)
func (c *Client) ClientContext() client.Context {
	return client.Context{}.
		WithCodec(c.encCfg.Codec).
		WithInterfaceRegistry(c.encCfg.InterfaceRegistry).
		WithTxConfig(c.encCfg.TxConfig).
		WithChainID(c.chainID).
		WithKeyring(c.keyring).
		WithGRPCClient(c.conn).
		WithBroadcastMode("sync")
}

// computeTxHash returns the canonical lower-case hex hash for the provided tx bytes.
func computeTxHash(txBytes []byte) string {
	if len(txBytes) == 0 {
		return ""
	}
	hashBytes := tmhash.Sum(txBytes)
	return strings.ToLower(hex.EncodeToString(hashBytes))
}

// sanitizeTxHash validates the hash returned by the node and normalises it to lower-case.
func sanitizeTxHash(hash string) string {
	hash = strings.TrimSpace(hash)
	if len(hash) != 64 {
		return ""
	}
	hash = strings.ToLower(hash)
	for i := 0; i < len(hash); i++ {
		ch := hash[i]
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			return ""
		}
	}
	return hash
}

// GetBalance queries the balance for a specific denom for an address
func (c *Client) GetBalance(ctx context.Context, address string, denom string) (sdkmath.Int, error) {
	req := &banktypes.QueryBalanceRequest{
		Address: address,
		Denom:   denom,
	}

	resp, err := c.bankClient.Balance(ctx, req)
	if err != nil {
		return sdkmath.ZeroInt(), err
	}

	if resp.Balance == nil {
		return sdkmath.ZeroInt(), nil
	}

	return resp.Balance.Amount, nil
}

// SelectPaymentDenom selects the appropriate payment denom based on user's balance
// Priority: undollar first, then unuah
func (c *Client) SelectPaymentDenom(ctx context.Context, address string, preferredDenom string) (string, error) {
	// If preferred denom is provided, use it if balance is sufficient
	if preferredDenom != "" {
		balance, err := c.GetBalance(ctx, address, preferredDenom)
		if err == nil && balance.IsPositive() {
			return preferredDenom, nil
		}
	}

	// Try undollar first
	undollarBalance, err := c.GetBalance(ctx, address, "undollar")
	if err == nil && undollarBalance.IsPositive() {
		return "undollar", nil
	}

	// Fallback to unuah
	unuahBalance, err := c.GetBalance(ctx, address, "unuah")
	if err == nil && unuahBalance.IsPositive() {
		return "unuah", nil
	}

	// If neither has balance, return undollar as default (transaction will fail with insufficient funds)
	return "undollar", nil
}

// GetTxStatus queries the status of a transaction by hash using REST API for full response
func (c *Client) GetTxStatus(ctx context.Context, txHash string) (*TxStatusResponse, error) {
	// Sanitize and validate hash
	txHash = sanitizeTxHash(txHash)
	if txHash == "" {
		return nil, fmt.Errorf("invalid transaction hash")
	}

	// Use REST API to get full transaction response with all logs
	// REST API provides more complete information than gRPC
	// Convert gRPC URL (localhost:9090) to RPC URL (localhost:26657)
	host := c.nodeURL
	// Replace gRPC port with RPC port
	if strings.Contains(host, ":9090") {
		host = strings.Replace(host, ":9090", ":26657", 1)
	} else if !strings.Contains(host, ":") {
		// If no port, assume RPC port
		host = host + ":26657"
	} else if !strings.Contains(host, ":26657") {
		// If different port, try to replace or append
		parts := strings.Split(host, ":")
		if len(parts) == 2 {
			host = parts[0] + ":26657"
		}
	}
	// Ensure localhost is 127.0.0.1 for HTTP requests
	host = strings.Replace(host, "localhost", "127.0.0.1", 1)
	restURL := fmt.Sprintf("http://%s/tx?hash=0x%s", host, txHash)

	// Make HTTP request to REST API
	req, err := http.NewRequestWithContext(ctx, "GET", restURL, nil)
	if err != nil {
		return &TxStatusResponse{
			TxHash:  txHash,
			Found:   false,
			Success: false,
			Error:   fmt.Sprintf("failed to create request: %v", err),
		}, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &TxStatusResponse{
			TxHash:  txHash,
			Found:   false,
			Success: false,
			Error:   fmt.Sprintf("failed to query transaction: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	var restResponse struct {
		Result struct {
			Hash     string `json:"hash"`
			Height   string `json:"height"`
			TxResult struct {
				Code      int    `json:"code"`
				Codespace string `json:"codespace"`
				Log       string `json:"log"`
				Info      string `json:"info"`
				GasWanted string `json:"gas_wanted"`
				GasUsed   string `json:"gas_used"`
				Events    []struct {
					Type       string `json:"type"`
					Attributes []struct {
						Key   string `json:"key"`
						Value string `json:"value"`
					} `json:"attributes"`
				} `json:"events"`
			} `json:"tx_result"`
		} `json:"result"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&restResponse); err != nil {
		return &TxStatusResponse{
			TxHash:  txHash,
			Found:   false,
			Success: false,
			Error:   fmt.Sprintf("failed to parse response: %v", err),
		}, nil
	}

	// Check if transaction was found
	if restResponse.Error.Code != 0 {
		return &TxStatusResponse{
			TxHash:  txHash,
			Found:   false,
			Success: false,
			Error:   restResponse.Error.Data,
		}, nil
	}

	if restResponse.Result.Hash == "" {
		return &TxStatusResponse{
			TxHash:  txHash,
			Found:   false,
			Success: false,
			Error:   "transaction not found",
		}, nil
	}

	txResult := restResponse.Result.TxResult

	// Parse height
	height := int64(0)
	if restResponse.Result.Height != "" {
		if h, err := strconv.ParseInt(restResponse.Result.Height, 10, 64); err == nil {
			height = h
		}
	}

	// Parse gas
	gasWanted := int64(0)
	gasUsed := int64(0)
	if txResult.GasWanted != "" {
		if gw, err := strconv.ParseInt(txResult.GasWanted, 10, 64); err == nil {
			gasWanted = gw
		}
	}
	if txResult.GasUsed != "" {
		if gu, err := strconv.ParseInt(txResult.GasUsed, 10, 64); err == nil {
			gasUsed = gu
		}
	}

	// Determine success: code 0 AND no error in log
	success := txResult.Code == 0
	// Check log for error messages even if code is 0
	if success && txResult.Log != "" {
		// Parse log to check for errors
		logLower := strings.ToLower(txResult.Log)
		if strings.Contains(logLower, "error") || strings.Contains(logLower, "failed") ||
			strings.Contains(logLower, "insufficient") || strings.Contains(logLower, "invalid") {
			// Even with code 0, there might be errors in log
			success = false
		}
	}

	// Extract error message
	errorMsg := ""
	if !success {
		if txResult.Log != "" {
			errorMsg = txResult.Log
		} else if txResult.Info != "" {
			errorMsg = txResult.Info
		} else if txResult.Code != 0 {
			errorMsg = fmt.Sprintf("transaction failed with code %d", txResult.Code)
		}
	}

	return &TxStatusResponse{
		TxHash:    txHash,
		Found:     true,
		Success:   success,
		Code:      txResult.Code,
		Codespace: txResult.Codespace,
		Height:    height,
		GasUsed:   gasUsed,
		GasWanted: gasWanted,
		Log:       txResult.Log,
		Error:     errorMsg,
		Events:    convertEvents(txResult.Events),
	}, nil
}

// convertEvents converts REST API events to a simpler format
func convertEvents(events []struct {
	Type       string `json:"type"`
	Attributes []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"attributes"`
}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(events))
	for _, event := range events {
		eventMap := map[string]interface{}{
			"type": event.Type,
		}
		attrs := make(map[string]string)
		for _, attr := range event.Attributes {
			attrs[attr.Key] = attr.Value
		}
		eventMap["attributes"] = attrs
		result = append(result, eventMap)
	}
	return result
}

// TxStatusResponse represents the status of a transaction
type TxStatusResponse struct {
	TxHash    string                   `json:"tx_hash"`
	Found     bool                     `json:"found"`
	Success   bool                     `json:"success"`
	Code      int                      `json:"code,omitempty"`
	Codespace string                   `json:"codespace,omitempty"`
	Height    int64                    `json:"height,omitempty"`
	GasUsed   int64                    `json:"gas_used,omitempty"`
	GasWanted int64                    `json:"gas_wanted,omitempty"`
	Log       string                   `json:"log,omitempty"`
	Error     string                   `json:"error,omitempty"`
	Events    []map[string]interface{} `json:"events,omitempty"`
}
