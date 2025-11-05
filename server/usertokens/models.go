package usertokens

// Request and response models for user token operations

// CreateTokenRequest represents the request to create a new token
type CreateTokenRequest struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Image       string `json:"image,omitempty"`
	Description string `json:"description,omitempty"`
}

// CreateTokenResponse represents the response from token creation
type CreateTokenResponse struct {
	Denom   string `json:"denom"`
	TxHash  string `json:"tx_hash"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// BuyTokenRequest represents the request to buy tokens from bonding curve
type BuyTokenRequest struct {
	Denom         string `json:"denom"`                    // Token denom to buy (e.g., factory/creator/symbol)
	PaymentAmount string `json:"payment_amount"`           // Amount to pay (in payment_denom)
	PaymentDenom  string `json:"payment_denom,omitempty"`  // Payment currency (defaults to unuah)
	MinTokensOut  string `json:"min_tokens_out,omitempty"` // Minimum tokens to receive (slippage protection)
}

// BuyTokenResponse represents the response from token purchase
type BuyTokenResponse struct {
	TxHash    string `json:"tx_hash"`
	TokensOut string `json:"tokens_out,omitempty"` // Actual tokens received
	PricePaid string `json:"price_paid,omitempty"` // Price paid per token
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
}

// SellTokenRequest represents the request to sell tokens to bonding curve
type SellTokenRequest struct {
	Denom         string `json:"denom"`                     // Token denom to sell
	TokenAmount   string `json:"token_amount"`              // Amount of tokens to sell
	PaymentDenom  string `json:"payment_denom,omitempty"`   // Payment currency to receive (defaults to unuah)
	MinPaymentOut string `json:"min_payment_out,omitempty"` // Minimum payment to receive (slippage protection)
}

// SellTokenResponse represents the response from token sale
type SellTokenResponse struct {
	TxHash        string `json:"tx_hash"`
	PaymentOut    string `json:"payment_out,omitempty"`    // Actual payment received
	PriceReceived string `json:"price_received,omitempty"` // Price received per token
	Success       bool   `json:"success"`
	Message       string `json:"message,omitempty"`
	Error         string `json:"error,omitempty"`
}

// TokenInfo represents information about a token
type TokenInfo struct {
	Denom        string `json:"denom"`
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
	Image        string `json:"image,omitempty"`
	Description  string `json:"description,omitempty"`
	Creator      string `json:"creator"`
	TotalSupply  string `json:"total_supply"`
	CurrentPrice string `json:"current_price,omitempty"`
}
