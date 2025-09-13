package api

import (
	"encoding/json"
	"net/http"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
	"github.com/osmosis-labs/osmosis/osmomath"
)

// PriceResponse represents the API response for price queries
type PriceResponse struct {
	Price      string    `json:"price"`
	Timestamp  time.Time `json:"timestamp"`
	PoolID     uint64    `json:"pool_id"`
	BaseDenom  string    `json:"base_denom"`
	QuoteDenom string    `json:"quote_denom"`
}

// USDPriceResponse represents USD price from oracle
type USDPriceResponse struct {
	Price     string    `json:"price"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

// PegKeeperStatusResponse represents pegkeeper status
type PegKeeperStatusResponse struct {
	Active       bool      `json:"active"`
	TargetPrice  string    `json:"target_price"`
	CurrentPrice string    `json:"current_price"`
	Deviation    string    `json:"deviation"`
	LastAction   string    `json:"last_action"`
	LastUpdated  time.Time `json:"last_updated"`
}

// MetricsResponse represents comprehensive N$ token metrics
type MetricsResponse struct {
	Price         string `json:"price"`
	TotalSupply   string `json:"total_supply"`
	PoolLiquidity struct {
		NUAH    string `json:"nuah"`
		NDollar string `json:"ndollar"`
	} `json:"pool_liquidity"`
	TradingVolume24h string    `json:"trading_volume_24h"`
	LastUpdated      time.Time `json:"last_updated"`
}

// NDollarAPI provides HTTP endpoints for N$ token monitoring
type NDollarAPI struct {
	app interface{} // Replace with your app interface
}

// NewNDollarAPI creates a new API instance
func NewNDollarAPI(app interface{}) *NDollarAPI {
	return &NDollarAPI{
		app: app,
	}
}

// RegisterRoutes registers all API routes
func (api *NDollarAPI) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/v1/ndollar/price", api.GetPrice).Methods("GET")
	r.HandleFunc("/api/v1/ndollar/twap", api.GetTWAP).Methods("GET")
	r.HandleFunc("/api/v1/ndollar/metrics", api.GetMetrics).Methods("GET")
	r.HandleFunc("/api/v1/ndollar/supply", api.GetSupply).Methods("GET")
	r.HandleFunc("/api/v1/usd/price", api.GetUSDPrice).Methods("GET")
	r.HandleFunc("/api/v1/pegkeeper/status", api.GetPegKeeperStatus).Methods("GET")
}

// GetPrice returns the current spot price of N$ from the AMM pool
func (api *NDollarAPI) GetPrice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Mock response - replace with actual implementation
	response := PriceResponse{
		Price:      "1.001025908322333333",
		Timestamp:  time.Now(),
		PoolID:     1,
		BaseDenom:  "factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar",
		QuoteDenom: "unuah",
	}

	json.NewEncoder(w).Encode(response)
}

// GetTWAP returns time-weighted average price over specified period
func (api *NDollarAPI) GetTWAP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Parse query parameters
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "1h" // default to 1 hour
	}

	// Mock TWAP calculation - replace with actual TWAP query
	response := PriceResponse{
		Price:      "1.000512345678901234",
		Timestamp:  time.Now(),
		PoolID:     1,
		BaseDenom:  "factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar",
		QuoteDenom: "unuah",
	}

	json.NewEncoder(w).Encode(response)
}

// GetMetrics returns comprehensive N$ token metrics
func (api *NDollarAPI) GetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Mock metrics - replace with actual queries
	response := MetricsResponse{
		Price:            "1.001025908322333333",
		TotalSupply:      "1000000",
		TradingVolume24h: "50000",
		LastUpdated:      time.Now(),
	}
	response.PoolLiquidity.NUAH = "999500"
	response.PoolLiquidity.NDollar = "1000500"

	json.NewEncoder(w).Encode(response)
}

// GetSupply returns current N$ token supply information
func (api *NDollarAPI) GetSupply(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	type SupplyResponse struct {
		TotalSupply string    `json:"total_supply"`
		Circulating string    `json:"circulating_supply"`
		Denom       string    `json:"denom"`
		LastUpdated time.Time `json:"last_updated"`
	}

	// Mock supply data - replace with actual bank module query
	response := SupplyResponse{
		TotalSupply: "1000000",
		Circulating: "1000000",
		Denom:       "factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar",
		LastUpdated: time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}

// Helper function to query TWAP (to be implemented with actual keeper)
func (api *NDollarAPI) queryTWAP(poolID uint64, baseDenom, quoteDenom string, startTime time.Time, endTime time.Time) (osmomath.Dec, error) {
	// This would use the actual TWAP keeper in a real implementation
	// return app.TwapKeeper.GetArithmeticTwap(ctx, poolID, baseDenom, quoteDenom, startTime, endTime)
	return osmomath.NewDecWithPrec(1001025908322333333, 18), nil
}

// GetUSDPrice returns current USD price from oracle
func (api *NDollarAPI) GetUSDPrice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Mock USD price - replace with actual USD oracle query
	response := USDPriceResponse{
		Price:     "1.0000",
		Timestamp: time.Now(),
		Source:    "USD Oracle",
	}

	json.NewEncoder(w).Encode(response)
}

// GetPegKeeperStatus returns current pegkeeper status
func (api *NDollarAPI) GetPegKeeperStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Mock pegkeeper status - replace with actual pegkeeper query
	response := PegKeeperStatusResponse{
		Active:       true,
		TargetPrice:  "1.0000",
		CurrentPrice: "1.0010",
		Deviation:    "0.10%",
		LastAction:   "mint",
		LastUpdated:  time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}

// Helper function to query token supply (to be implemented with actual keeper)
func (api *NDollarAPI) queryTokenSupply(denom string) (sdk.Coin, error) {
	// This would use the actual bank keeper in a real implementation
	// return app.BankKeeper.GetSupply(ctx, denom)
	return sdk.NewCoin("factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar", math.NewInt(1000000)), nil
}
