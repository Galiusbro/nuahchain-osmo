package api

import (
	"net/http"

	"github.com/osmosis-labs/osmosis/v30/server/assets"
	"github.com/osmosis-labs/osmosis/v30/server/logger"
	"github.com/osmosis-labs/osmosis/v30/server/usertokens"
)

// NewRouter creates and returns a new HTTP router with all routes configured
func NewRouter(appLogger *logger.Logger) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoints
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/health/db", handleDBHealth)

	// Authentication endpoints (public)
	mux.HandleFunc("/api/auth/register", handleRegister)
	mux.HandleFunc("/api/auth/login", handleLogin)
	mux.HandleFunc("/api/auth/telegram", handleTelegramAuth)

	// Protected endpoints (require authentication)
	mux.HandleFunc("/api/auth/me", handleMe)

	// User token endpoints (require authentication)
	mux.HandleFunc("/api/tokens/create", usertokens.HandleCreateToken)
	mux.HandleFunc("/api/tokens/buy", usertokens.HandleBuyToken)
	mux.HandleFunc("/api/tokens/sell", usertokens.HandleSellToken)

	// Asset endpoints (require authentication)
	mux.HandleFunc("/api/assets/ensure", assets.HandleEnsureAsset)
	mux.HandleFunc("/api/assets/buy", assets.HandleBuyAsset)
	mux.HandleFunc("/api/assets/sell", assets.HandleSellAsset)

	// Transaction status endpoint (public, works for any transaction)
	mux.HandleFunc("/api/tx/", usertokens.HandleGetTxStatus)

	// Wrap with logging middleware
	return logger.HTTPMiddleware(appLogger)(mux)
}
