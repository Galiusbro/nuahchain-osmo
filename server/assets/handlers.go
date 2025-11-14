package assets

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/osmosis-labs/osmosis/v30/server/auth"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

var (
	// assetService holds the service instance (set during initialization)
	assetService *Service
	// authService holds the auth service instance (set during initialization)
	authService *auth.Service
)

func httpStatusFromTransaction(status string) int {
	switch status {
	case string(transactions.StatusSuccess):
		return http.StatusOK
	case string(transactions.StatusFailed):
		return http.StatusInternalServerError
	default:
		return http.StatusAccepted
	}
}

// SetService sets the asset service instance
func SetService(service *Service) {
	assetService = service
}

// SetAuthService sets the auth service instance
func SetAuthService(service *auth.Service) {
	authService = service
}

// HandleEnsureAsset handles POST /api/assets/ensure
func HandleEnsureAsset(w http.ResponseWriter, r *http.Request) {
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

	// Parse request body
	var req EnsureAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}

	// Ensure asset
	resp, err := assetService.EnsureAsset(r.Context(), user.ID, req)
	if resp == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	statusCode := httpStatusFromTransaction(resp.Status)
	if err != nil && statusCode == http.StatusAccepted {
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

// HandleBuyAsset handles POST /api/assets/buy
func HandleBuyAsset(w http.ResponseWriter, r *http.Request) {
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

	// Parse request body
	var req BuyAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}
	if req.Amount == "" && req.AmountNDOLLAR == "" {
		http.Error(w, "Either amount or amount_ndollar is required", http.StatusBadRequest)
		return
	}

	// Buy asset
	resp, err := assetService.BuyAsset(r.Context(), user.ID, req)
	if resp == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	statusCode := httpStatusFromTransaction(resp.Status)
	if err != nil && statusCode == http.StatusAccepted {
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

// HandleSellAsset handles POST /api/assets/sell
func HandleSellAsset(w http.ResponseWriter, r *http.Request) {
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

	// Parse request body
	var req SellAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}
	if req.BaseAmount == "" {
		http.Error(w, "Base amount is required", http.StatusBadRequest)
		return
	}

	// Sell asset
	resp, err := assetService.SellAsset(r.Context(), user.ID, req)
	if resp == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	statusCode := httpStatusFromTransaction(resp.Status)
	if err != nil && statusCode == http.StatusAccepted {
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

// HandleOpenMarginPosition handles POST /api/assets/margin/open
func HandleOpenMarginPosition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := authenticateRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req OpenMarginPositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Symbol == "" || req.Side == "" || req.QuoteAmount == "" || req.Leverage == "" {
		http.Error(w, "symbol, side, quote_amount and leverage are required", http.StatusBadRequest)
		return
	}

	resp, err := assetService.OpenMarginPosition(r.Context(), user.ID, req)
	if resp == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	statusCode := httpStatusFromTransaction(resp.Status)
	if err != nil && statusCode == http.StatusAccepted {
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

// HandleCloseMarginPosition handles POST /api/assets/margin/close
func HandleCloseMarginPosition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := authenticateRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req CloseMarginPositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.PositionID == "" {
		http.Error(w, "position_id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(req.PositionID, 10, 64)
	if err != nil {
		http.Error(w, "position_id must be a positive integer", http.StatusBadRequest)
		return
	}

	resp, err := assetService.CloseMarginPosition(r.Context(), user.ID, id)
	if resp == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	statusCode := httpStatusFromTransaction(resp.Status)
	if err != nil && statusCode == http.StatusAccepted {
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

// authenticateRequest extracts and validates the user from the Authorization header
func authenticateRequest(r *http.Request) (*auth.User, error) {
	if assetService == nil {
		return nil, auth.ErrServiceNotInitialized
	}
	if authService == nil {
		return nil, auth.ErrServiceNotInitialized
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, auth.ErrMissingToken
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, auth.ErrInvalidTokenFormat
	}

	user, err := authService.ValidateToken(parts[1])
	if err != nil {
		return nil, err
	}

	return user, nil
}
