package usertokens

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/osmosis-labs/osmosis/v30/server/auth"
)

var (
	// tokenService holds the service instance (set during initialization)
	tokenService *Service
	// authService holds the auth service instance (set during initialization)
	authService *auth.Service
)

// SetService sets the token service instance
func SetService(service *Service) {
	tokenService = service
}

// SetAuthService sets the auth service instance
func SetAuthService(service *auth.Service) {
	authService = service
}

// HandleCreateToken handles POST /api/tokens/create
func HandleCreateToken(w http.ResponseWriter, r *http.Request) {
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
	var req CreateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if req.Symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}

	// Create token
	resp, err := tokenService.CreateToken(r.Context(), user.ID, req)
	if err != nil {
		apiResp := CreateTokenResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(apiResp)
		return
	}

	// Return response
	apiResp := CreateTokenResponse{
		Denom:   resp.Denom,
		TxHash:  resp.TxHash,
		Success: resp.Success,
		Message: "Token creation initiated",
		Error:   resp.Error,
	}

	w.Header().Set("Content-Type", "application/json")
	if resp.Success {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(apiResp)
}

// HandleBuyToken handles POST /api/tokens/buy
func HandleBuyToken(w http.ResponseWriter, r *http.Request) {
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
	var req BuyTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Denom == "" {
		http.Error(w, "Denom is required", http.StatusBadRequest)
		return
	}
	if req.PaymentAmount == "" {
		http.Error(w, "Payment amount is required", http.StatusBadRequest)
		return
	}

	// Buy token
	resp, err := tokenService.BuyToken(r.Context(), user.ID, req)
	if err != nil {
		apiResp := BuyTokenResponse{
			Success: false,
			TxHash:  resp.TxHash,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(apiResp)
		return
	}

	// Return response
	apiResp := BuyTokenResponse{
		TxHash:    resp.TxHash,
		TokensOut: resp.TokensOut,
		PricePaid: resp.PricePaid,
		Success:   resp.Success,
		Message:   "Token purchase initiated",
		Error:     resp.Error,
	}

	w.Header().Set("Content-Type", "application/json")
	if resp.Success {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(apiResp)
}

// HandleSellToken handles POST /api/tokens/sell
func HandleSellToken(w http.ResponseWriter, r *http.Request) {
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
	var req SellTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Denom == "" {
		http.Error(w, "Denom is required", http.StatusBadRequest)
		return
	}
	if req.TokenAmount == "" {
		http.Error(w, "Token amount is required", http.StatusBadRequest)
		return
	}

	// Sell token
	resp, err := tokenService.SellToken(r.Context(), user.ID, req)
	if err != nil {
		apiResp := SellTokenResponse{
			Success: false,
			TxHash:  resp.TxHash,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(apiResp)
		return
	}

	// Return response
	apiResp := SellTokenResponse{
		TxHash:        resp.TxHash,
		PaymentOut:    resp.PaymentOut,
		PriceReceived: resp.PriceReceived,
		Success:       resp.Success,
		Message:       "Token sale initiated",
		Error:         resp.Error,
	}

	w.Header().Set("Content-Type", "application/json")
	if resp.Success {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(apiResp)
}

// authenticateRequest extracts and validates the user from the Authorization header
func authenticateRequest(r *http.Request) (*auth.User, error) {
	if tokenService == nil {
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

// HandleGetTxStatus handles GET /api/tx/:hash
// This endpoint works for any transaction, not just token-related ones
func HandleGetTxStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract tx hash from URL path
	// Path format: /api/tx/:hash
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[3] == "" {
		http.Error(w, "Transaction hash is required", http.StatusBadRequest)
		return
	}

	txHash := parts[3]

	// Get transaction status
	resp, err := tokenService.GetTxStatus(r.Context(), txHash)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tx_hash": txHash,
			"found":   false,
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
