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

// AssetInfo represents information about an asset
type AssetInfo struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price,omitempty"` // Current price from oracle
	Denom  string `json:"denom,omitempty"` // Asset denom (e.g., "asset/GOLD")
}
