package exchange

// ExchangeTokensRequest represents an API request to exchange tokens for unuah
type ExchangeTokensRequest struct {
	TokenDenom string `json:"token_denom"` // e.g., "uusdc", "ueth", "ubtc"
	Amount     string `json:"amount"`      // Amount in base units (e.g., "100000000" for 100 USDC)
	MinOutput  string `json:"min_output"`  // Minimum unuah output (slippage protection)
}

// ExchangeTokensResponse represents the API response for token exchange
type ExchangeTokensResponse struct {
	Status   string `json:"status"`    // pending | success | failed
	TxHash   string `json:"tx_hash"`   // Blockchain transaction hash
	UnuahOut string `json:"unuah_out"` // Amount of unuah received (populated when known)
	ErrorMsg string `json:"error_msg,omitempty"`
}

// GetExchangeRateRequest represents a request to get current exchange rate
type GetExchangeRateRequest struct {
	TokenDenom string `json:"token_denom"`
}

// GetExchangeRateResponse represents the exchange rate information
type GetExchangeRateResponse struct {
	TokenDenom  string `json:"token_denom"`
	Rate        string `json:"rate"` // USD price per token
	LastUpdated string `json:"last_updated"`
}

// GetSupportedTokensResponse represents the list of supported tokens
type GetSupportedTokensResponse struct {
	SupportedTokens []string `json:"supported_tokens"`
}
