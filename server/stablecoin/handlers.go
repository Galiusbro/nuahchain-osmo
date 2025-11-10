package stablecoin

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/osmosis-labs/osmosis/v30/server/auth"
)

// authenticateRequest validates JWT token and returns user
func authenticateRequest(r *http.Request) (*auth.User, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, http.ErrNoCookie
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, http.ErrNoCookie
	}

	token := parts[1]
	return globalAuthService.ValidateToken(token)
}

// HandleBuyNDollar handles the HTTP request to buy NDOLLAR with unuah
func HandleBuyNDollar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate user
	user, err := authenticateRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Parse request
	var req BuyNDollarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Amount == "" {
		http.Error(w, "Amount is required", http.StatusBadRequest)
		return
	}

	// Execute operation
	resp, err := globalService.BuyNDollar(r.Context(), user.ID, req.Amount)
	if err != nil {
		// Return error response with details
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // Still 200 to allow client to see error details
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleSellNDollar handles the HTTP request to sell NDOLLAR back to unuah
func HandleSellNDollar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate user
	user, err := authenticateRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Parse request
	var req SellNDollarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Amount == "" {
		http.Error(w, "Amount is required", http.StatusBadRequest)
		return
	}

	// Execute operation
	resp, err := globalService.SellNDollar(r.Context(), user.ID, req.Amount)
	if err != nil {
		// Return error response with details
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // Still 200 to allow client to see error details
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
