package quotes

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

var globalQuotesService *Service

// SetService sets the global quotes service
func SetService(s *Service) {
	globalQuotesService = s
}

// HandleGetTradeQuote handles GET /api/quote/trade - get trade quote for bonding curve
func HandleGetTradeQuote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if globalQuotesService == nil {
		http.Error(w, "Quotes service not configured", http.StatusInternalServerError)
		return
	}

	// Parse query parameters
	denom := r.URL.Query().Get("denom")
	operation := strings.ToLower(r.URL.Query().Get("operation"))
	amount := r.URL.Query().Get("amount")
	paymentDenom := r.URL.Query().Get("payment_denom")

	// Validate required parameters
	if denom == "" {
		http.Error(w, "denom parameter is required", http.StatusBadRequest)
		return
	}
	if operation == "" {
		http.Error(w, "operation parameter is required (buy or sell)", http.StatusBadRequest)
		return
	}
	if operation != "buy" && operation != "sell" {
		http.Error(w, "operation must be 'buy' or 'sell'", http.StatusBadRequest)
		return
	}
	if amount == "" {
		http.Error(w, "amount parameter is required", http.StatusBadRequest)
		return
	}

	// URL decode denom (in case it contains special characters)
	denom, err := url.QueryUnescape(denom)
	if err != nil {
		denom = strings.ReplaceAll(denom, "%2F", "/")
	}

	// Create request
	req := TradeQuoteRequest{
		Denom:        denom,
		Operation:    operation,
		Amount:       amount,
		PaymentDenom: paymentDenom,
	}

	// Get quote
	quote, err := globalQuotesService.GetTradeQuote(r.Context(), req)
	if err != nil {
		// Return JSON error response instead of plain text
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Failed to get trade quote",
			"message": err.Error(),
			"denom":   req.Denom,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(quote)
}

// HandleGetSwapQuote handles GET /api/quote/swap - get swap quote for exchange
func HandleGetSwapQuote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if globalQuotesService == nil {
		http.Error(w, "Quotes service not configured", http.StatusInternalServerError)
		return
	}

	// Parse query parameters
	tokenIn := r.URL.Query().Get("token_in")
	amountIn := r.URL.Query().Get("amount_in")

	// Validate required parameters
	if tokenIn == "" {
		http.Error(w, "token_in parameter is required", http.StatusBadRequest)
		return
	}
	if amountIn == "" {
		http.Error(w, "amount_in parameter is required", http.StatusBadRequest)
		return
	}

	// Create request
	req := SwapQuoteRequest{
		TokenIn:  tokenIn,
		AmountIn: amountIn,
	}

	// Get quote
	quote, err := globalQuotesService.GetSwapQuote(r.Context(), req)
	if err != nil {
		// Return JSON error response instead of plain text
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":    "Failed to get swap quote",
			"message":  err.Error(),
			"token_in": req.TokenIn,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(quote)
}

// HandleGetSupportedTokens handles GET /api/quote/supported-tokens - get list of supported tokens
func HandleGetSupportedTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if globalQuotesService == nil {
		http.Error(w, "Quotes service not configured", http.StatusInternalServerError)
		return
	}

	// Get supported tokens
	tokens, err := globalQuotesService.GetSupportedTokens(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Failed to get supported tokens",
			"message": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokens)
}
