package balances

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/osmosis-labs/osmosis/v30/server/auth"
)

var globalBalancesService *Service
var globalAuthService interface {
	ValidateToken(token string) (*auth.User, error)
	GetUserWallet(userID int64) (*auth.Wallet, error)
}

// SetService sets the global balances service
func SetService(s *Service) {
	globalBalancesService = s
}

// SetAuthService sets the global auth service
func SetAuthService(s interface {
	ValidateToken(token string) (*auth.User, error)
	GetUserWallet(userID int64) (*auth.Wallet, error)
}) {
	globalAuthService = s
}

// HandleGetUserBalancesFromDB handles GET /api/users/balances-db - get balances from database
func HandleGetUserBalancesFromDB(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if globalBalancesService == nil {
		http.Error(w, "Balances service not configured", http.StatusInternalServerError)
		return
	}

	// Extract token from Authorization header
	user, err := authenticateRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Get optional denom filter from query parameters
	denomFilter := r.URL.Query().Get("tokenMint")
	if denomFilter == "" {
		denomFilter = r.URL.Query().Get("denom")
	}

	// Get balances from database
	balances, err := globalBalancesService.GetUserBalancesFromDB(user.ID, denomFilter)
	if err != nil {
		http.Error(w, "Failed to get balances: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Enrich with token metadata if needed
	enrichedBalances := enrichBalancesWithMetadata(balances)

	response := map[string]interface{}{
		"balances": enrichedBalances,
		"count":    len(enrichedBalances),
		"source":   "database",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleGetUserBalancesFromBlockchain handles GET /api/users/balances - get balances from blockchain
func HandleGetUserBalancesFromBlockchain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if globalBalancesService == nil {
		http.Error(w, "Balances service not configured", http.StatusInternalServerError)
		return
	}

	// Extract token from Authorization header
	user, err := authenticateRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Get wallet address
	wallet, err := globalAuthService.GetUserWallet(user.ID)
	if err != nil {
		http.Error(w, "Failed to get wallet: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get balances from blockchain
	ctx := r.Context()
	balances, err := globalBalancesService.GetUserBalancesFromBlockchain(ctx, wallet.Address)
	if err != nil {
		http.Error(w, "Failed to get balances from blockchain: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Enrich with token metadata if needed
	enrichedBalances := enrichBalancesWithMetadata(balances)

	response := map[string]interface{}{
		"balances": enrichedBalances,
		"count":    len(enrichedBalances),
		"source":   "blockchain",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleSyncUserBalances handles POST /api/users/balances/sync - sync balances from blockchain
func HandleSyncUserBalances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if globalBalancesService == nil {
		http.Error(w, "Balances service not configured", http.StatusInternalServerError)
		return
	}

	// Extract token from Authorization header
	user, err := authenticateRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Get wallet address
	wallet, err := globalAuthService.GetUserWallet(user.ID)
	if err != nil {
		http.Error(w, "Failed to get wallet: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Sync balances
	ctx := r.Context()
	if err := globalBalancesService.SyncUserBalances(ctx, user.ID, wallet.Address); err != nil {
		http.Error(w, "Failed to sync balances: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "Balances synchronized successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleGetBalanceHistory handles GET /api/users/balances/history - get balance history
func HandleGetBalanceHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if globalBalancesService == nil {
		http.Error(w, "Balances service not configured", http.StatusInternalServerError)
		return
	}

	// Extract token from Authorization header
	user, err := authenticateRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Get optional parameters
	denom := r.URL.Query().Get("denom")
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // Default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	// Get history
	history, err := globalBalancesService.GetBalanceHistory(user.ID, denom, limit)
	if err != nil {
		http.Error(w, "Failed to get balance history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"history": history,
		"count":   len(history),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// authenticateRequest authenticates the request and returns the user
func authenticateRequest(r *http.Request) (*auth.User, error) {
	if globalAuthService == nil {
		return nil, fmt.Errorf("auth service not configured")
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	token := parts[1]
	user, err := globalAuthService.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// enrichBalancesWithMetadata enriches balances with token metadata
// This is a placeholder - in production, you'd fetch from tokens repository
func enrichBalancesWithMetadata(balances []UserBalance) []map[string]interface{} {
	result := make([]map[string]interface{}, len(balances))
	for i, balance := range balances {
		result[i] = map[string]interface{}{
			"denom":  balance.Denom,
			"amount": balance.Amount,
		}
		// TODO: Add token metadata from tokens repository
	}
	return result
}

