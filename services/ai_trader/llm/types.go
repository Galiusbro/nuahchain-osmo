package llm

import "time"

// PromptInput encapsulates the structured inputs provided to the LLM.
type PromptInput struct {
	Symbols []string  `json:"symbols"`
	NowUTC  time.Time `json:"now_utc"`
	// Latest spot prices per symbol (string for decimal safety)
	LatestPrices map[string]string `json:"latest_prices"`
	// Optional OHLCV windows keyed by timeframe (e.g., "1m","1h","1d")
	OHLCV map[string][]Candle `json:"ohlcv,omitempty"`
	// Operational limits and constraints for the decision
	Constraints Constraints `json:"constraints"`
	// Optional context summary or rationale hints
	Context string `json:"context,omitempty"`
}

// Candle represents an OHLCV bar.
type Candle struct {
	T int64  `json:"t"` // unix seconds
	O string `json:"o"`
	H string `json:"h"`
	L string `json:"l"`
	C string `json:"c"`
	V string `json:"v"` // volume (string for decimal safety)
}

// Constraints define the safety rails applied to AI decisions.
type Constraints struct {
	Market        string   `json:"market"`          // "assets" or "bondingcurve"
	MaxNotionalND string   `json:"max_notional_nd"` // NDOLLAR budget cap for a single decision
	MinNotionalND string   `json:"min_notional_nd"`
	AllowedDenoms []string `json:"allowed_denoms"`
	FeeRate       string   `json:"fee_rate"` // percent in decimal string
}

// DecisionOut is the structured result returned by the LLM.
type DecisionOut struct {
	Action       string  `json:"action"` // buy | sell | hold
	Symbol       string  `json:"symbol"`
	Amount       string  `json:"amount"` // amount in payment denom (NDOLLAR by default)
	PaymentDenom string  `json:"payment_denom"`
	Market       string  `json:"market"` // assets | bondingcurve
	Confidence   float64 `json:"confidence"`
	Rationale    string  `json:"rationale"`
}
