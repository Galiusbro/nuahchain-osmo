package marketplace

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var globalMarketplaceService *Service

// SetService sets the global marketplace service
func SetService(s *Service) {
	globalMarketplaceService = s
}

// HandleGetMarketplace handles GET /api/tokens/market - list all tokens
func HandleGetMarketplace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if globalMarketplaceService == nil {
		http.Error(w, "Marketplace service not configured", http.StatusInternalServerError)
		return
	}

	// Parse query parameters
	limit := uint64(100) // Default
	offset := uint64(0)

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.ParseUint(limitStr, 10, 64); err == nil {
			limit = parsed
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.ParseUint(offsetStr, 10, 64); err == nil {
			offset = parsed
		}
	}

	// Get tokens
	tokens, err := globalMarketplaceService.GetMarketplaceTokens(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "Failed to get marketplace tokens: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"tokens": tokens,
		"count":  len(tokens),
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleSearchTokens handles GET /api/tokens/search - search tokens by name/symbol
func HandleSearchTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if globalMarketplaceService == nil {
		http.Error(w, "Marketplace service not configured", http.StatusInternalServerError)
		return
	}

	// Get query parameter
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	// Parse pagination
	limit := uint64(50) // Default for search
	offset := uint64(0)

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.ParseUint(limitStr, 10, 64); err == nil {
			limit = parsed
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.ParseUint(offsetStr, 10, 64); err == nil {
			offset = parsed
		}
	}

	// Search tokens
	tokens, err := globalMarketplaceService.SearchTokens(r.Context(), query, limit, offset)
	if err != nil {
		http.Error(w, "Failed to search tokens: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"tokens": tokens,
		"count":  len(tokens),
		"query":  query,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleGetTokenDetails handles GET /api/tokens/{denom}/details - get token details
func HandleGetTokenDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if globalMarketplaceService == nil {
		http.Error(w, "Marketplace service not configured", http.StatusInternalServerError)
		return
	}

	// Extract denom from URL path
	// Expected format: /api/tokens/{denom}/details
	// Denom can contain slashes (e.g., factory/creator/symbol)
	path := r.URL.Path

	// Remove /api/tokens/ prefix
	prefix := "/api/tokens/"
	if !strings.HasPrefix(path, prefix) {
		http.Error(w, "Invalid URL format. Expected: /api/tokens/{denom}/details", http.StatusBadRequest)
		return
	}

	// Remove /details suffix
	suffix := "/details"
	if !strings.HasSuffix(path, suffix) {
		http.Error(w, "Invalid URL format. Expected: /api/tokens/{denom}/details", http.StatusBadRequest)
		return
	}

	// Extract denom between prefix and suffix
	// path = /api/tokens/factory/creator/symbol/details
	// After TrimPrefix: factory/creator/symbol/details
	// After TrimSuffix: factory/creator/symbol
	denom := strings.TrimPrefix(path, prefix)
	denom = strings.TrimSuffix(denom, suffix)

	// URL decode denom properly
	decoded, err := url.QueryUnescape(denom)
	if err == nil {
		denom = decoded
	} else {
		// Fallback: replace %2F with /
		denom = strings.ReplaceAll(denom, "%2F", "/")
	}

	if denom == "" {
		http.Error(w, "Token denom is required", http.StatusBadRequest)
		return
	}

	// Get token details
	details, err := globalMarketplaceService.GetTokenDetails(r.Context(), denom)
	if err != nil {
		http.Error(w, "Failed to get token details: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if details == nil {
		http.Error(w, "Token not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(details)
}

