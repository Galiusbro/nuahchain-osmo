package stablecoin

// BuyNDollarRequest represents the request to buy NDOLLAR with unuah
type BuyNDollarRequest struct {
	Amount string `json:"amount"` // Amount of unuah to convert to NDOLLAR (1:1)
}

// BuyNDollarResponse represents the response after buying NDOLLAR
type BuyNDollarResponse struct {
	Success       bool   `json:"success"`
	TxHash        string `json:"tx_hash"`
	NDollarAmount string `json:"ndollar_amount"` // Amount of NDOLLAR received
	NDollarDenom  string `json:"ndollar_denom"`  // Actual NDOLLAR denom (factory/.../ndollar)
	Error         string `json:"error,omitempty"`
}

// SellNDollarRequest represents the request to sell NDOLLAR
type SellNDollarRequest struct {
	Amount string `json:"amount"` // Amount of NDOLLAR to convert back to unuah (1:1)
}

// SellNDollarResponse represents the response after selling NDOLLAR
type SellNDollarResponse struct {
	Success     bool   `json:"success"`
	TxHash      string `json:"tx_hash"`
	UnuahAmount string `json:"unuah_amount"` // Amount of unuah received
	Error       string `json:"error,omitempty"`
}

