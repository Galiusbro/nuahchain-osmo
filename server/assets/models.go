package assets

// Request and response models for asset operations

// EnsureAssetRequest represents the request to ensure an asset exists
type EnsureAssetRequest struct {
	Symbol string `json:"symbol"` // Asset symbol (e.g., "GOLD", "BTC")
}

// EnsureAssetResponse represents the response from ensuring an asset
type EnsureAssetResponse struct {
	TxHash  string `json:"tx_hash"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// BuyAssetRequest represents the request to buy an asset
type BuyAssetRequest struct {
	Symbol        string `json:"symbol"`                   // Asset symbol (e.g., "GOLD")
	Denom         string `json:"denom,omitempty"`          // Payment denom (e.g., "NDOLLAR", "unuah", or factory denom)
	Amount        string `json:"amount,omitempty"`         // Amount in the specified denom
	AmountNDOLLAR string `json:"amount_ndollar,omitempty"` // Deprecated: amount in NDOLLAR (for backward compatibility)
}

// BuyAssetResponse represents the response from buying an asset
type BuyAssetResponse struct {
	TxHash     string `json:"tx_hash"`
	BaseAmount string `json:"base_amount,omitempty"` // Amount of asset received
	Success    bool   `json:"success"`
	Message    string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
}

// SellAssetRequest represents the request to sell an asset
type SellAssetRequest struct {
	Symbol     string `json:"symbol"`      // Asset symbol
	BaseAmount string `json:"base_amount"` // Amount of asset to sell
}

// SellAssetResponse represents the response from selling an asset
type SellAssetResponse struct {
	TxHash        string `json:"tx_hash"`
	PayoutNDOLLAR string `json:"payout_ndollar,omitempty"` // Amount of NDOLLAR received
	Success       bool   `json:"success"`
	Message       string `json:"message,omitempty"`
	Error         string `json:"error,omitempty"`
}

// OpenMarginPositionRequest represents the request to open a leveraged position
type OpenMarginPositionRequest struct {
	Symbol      string `json:"symbol"`       // Asset symbol
	Side        string `json:"side"`         // long / short
	QuoteAmount string `json:"quote_amount"` // Margin amount in micro NDOLLAR
	Leverage    string `json:"leverage"`     // Leverage multiplier (decimal string)
}

// OpenMarginPositionResponse represents the response from opening a leveraged position
type OpenMarginPositionResponse struct {
	TxHash       string `json:"tx_hash"`
	PositionID   string `json:"position_id,omitempty"`
	BaseQuantity string `json:"base_quantity,omitempty"`
	EntryPrice   string `json:"entry_price,omitempty"`
	Leverage     string `json:"leverage,omitempty"`
	Success      bool   `json:"success"`
	Message      string `json:"message,omitempty"`
	Error        string `json:"error,omitempty"`
}

// CloseMarginPositionRequest represents the request to close a leveraged position
type CloseMarginPositionRequest struct {
	PositionID string `json:"position_id"`
}

// CloseMarginPositionResponse represents the response from closing a leveraged position
type CloseMarginPositionResponse struct {
	TxHash  string `json:"tx_hash"`
	Pnl     string `json:"pnl,omitempty"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// AssetInfo represents information about an asset
type AssetInfo struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price,omitempty"` // Current price from oracle
	Denom  string `json:"denom,omitempty"` // Asset denom (e.g., "asset/GOLD")
}
