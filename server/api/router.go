package api

import (
	"net/http"

	"github.com/osmosis-labs/osmosis/v30/server/logger"
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

	// Wrap with logging middleware
	return logger.HTTPMiddleware(appLogger)(mux)
}
