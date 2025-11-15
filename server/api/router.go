package api

import (
	"net/http"

	"github.com/osmosis-labs/osmosis/v30/server/assets"
	"github.com/osmosis-labs/osmosis/v30/server/exchange"
	"github.com/osmosis-labs/osmosis/v30/server/logger"
	"github.com/osmosis-labs/osmosis/v30/server/marketplace"
	"github.com/osmosis-labs/osmosis/v30/server/monitor"
	"github.com/osmosis-labs/osmosis/v30/server/stablecoin"
	"github.com/osmosis-labs/osmosis/v30/server/usertokens"
	"github.com/osmosis-labs/osmosis/v30/server/users"
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
	mux.HandleFunc("/api/auth/refresh", handleRefresh)
	mux.HandleFunc("/api/auth/logout", handleLogout)
	mux.HandleFunc("/api/auth/logout-all", handleLogoutAll)
	mux.HandleFunc("/api/auth/web/forgot-password", handleForgotPassword)
	mux.HandleFunc("/api/auth/web/reset-password", handleResetPassword)

	// Protected endpoints (require authentication)
	mux.HandleFunc("/api/auth/me", handleMe)
	mux.HandleFunc("/api/auth/sessions", handleSessions)

	// User profile endpoints (require authentication)
	mux.HandleFunc("/api/users/me", users.HandleGetUserProfile)
	mux.HandleFunc("/api/users/me/info", users.HandleGetUserInfoSummary)
	mux.HandleFunc("/api/users/me/tokens", users.HandleGetUserTokens)
	mux.HandleFunc("/api/users/me/upload-image", users.HandleUploadUserImage)
	mux.HandleFunc("/api/users/username", users.HandleUpdateUsername)

	// Serve uploaded images statically
	mux.HandleFunc("/uploads/images/", users.ServeUploadedImages)

	// User token endpoints (require authentication)
	mux.HandleFunc("/api/tokens/create", usertokens.HandleCreateToken)
	mux.HandleFunc("/api/tokens/buy", usertokens.HandleBuyToken)
	mux.HandleFunc("/api/tokens/sell", usertokens.HandleSellToken)

	// Marketplace endpoints (public)
	mux.HandleFunc("/api/tokens/market", marketplace.HandleGetMarketplace)
	mux.HandleFunc("/api/tokens/search", marketplace.HandleSearchTokens)
	mux.HandleFunc("/api/tokens/", marketplace.HandleGetTokenDetails) // Must be after /market and /search

	// Asset endpoints (require authentication)
	mux.HandleFunc("/api/assets/ensure", assets.HandleEnsureAsset)
	mux.HandleFunc("/api/assets/buy", assets.HandleBuyAsset)
	mux.HandleFunc("/api/assets/sell", assets.HandleSellAsset)
	mux.HandleFunc("/api/assets/margin/open", assets.HandleOpenMarginPosition)
	mux.HandleFunc("/api/assets/margin/close", assets.HandleCloseMarginPosition)

	// Stablecoin endpoints (require authentication)
	mux.HandleFunc("/api/stablecoin/buy-ndollar", stablecoin.HandleBuyNDollar)
	mux.HandleFunc("/api/stablecoin/sell-ndollar", stablecoin.HandleSellNDollar)

	// Exchange endpoints (require authentication)
	mux.HandleFunc("/api/exchange/tokens", exchange.HandleExchangeTokens)

	// Transaction status endpoint (public, works for any transaction)
	mux.HandleFunc("/api/tx/", usertokens.HandleGetTxStatus)

	// Admin monitoring endpoints
	mux.HandleFunc("/api/admin/transactions", monitor.HandleGetRecentTransactions)
	mux.HandleFunc("/api/admin/transactions/ws", monitor.HandleWebSocket)
	mux.HandleFunc("/api/admin/stats", monitor.HandleGetStats)

	// Wrap with logging middleware
	return logger.HTTPMiddleware(appLogger)(mux)
}
