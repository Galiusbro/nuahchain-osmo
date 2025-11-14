package exchange

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/osmosis-labs/osmosis/v30/server/auth"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

var globalExchangeService *Service
var globalAuthService *auth.Service

// SetService sets the global exchange service instance
func SetService(s *Service) {
	globalExchangeService = s
}

// SetAuthService sets the global auth service instance
func SetAuthService(s *auth.Service) {
	globalAuthService = s
}

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

// authenticateRequest validates the JWT token and returns the user
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

// HandleExchangeTokens handles the exchange tokens endpoint
func HandleExchangeTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate user
	user, err := authenticateRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req ExchangeTokensRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Execute exchange
	resp, err := globalExchangeService.ExchangeTokens(r.Context(), user.ID, req)
	if resp == nil {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, "unknown error", http.StatusInternalServerError)
		}
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
