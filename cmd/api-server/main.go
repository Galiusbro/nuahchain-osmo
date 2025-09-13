package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/osmosis-labs/osmosis/v30/api"
)

func main() {
	// Create router
	r := mux.NewRouter()

	// Create API instance
	ndollarAPI := api.NewNDollarAPI(nil) // Pass actual app instance in production

	// Register API routes
	ndollarAPI.RegisterRoutes(r)

	// Add health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"ndollar-api"}`))
	}).Methods("GET")

	// Add CORS middleware
	r.Use(corsMiddleware)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("N$ API Server starting on port %s\n", port)
	fmt.Println("Available endpoints:")
	fmt.Println("  GET /health - Health check")
	fmt.Println("  GET /api/v1/ndollar/price - Current N$ spot price")
	fmt.Println("  GET /api/v1/ndollar/twap?period=1h - TWAP price")
	fmt.Println("  GET /api/v1/ndollar/metrics - Comprehensive metrics")
	fmt.Println("  GET /api/v1/ndollar/supply - Token supply information")
	fmt.Println("  GET /api/v1/usd/price - USD price from oracle")
	fmt.Println("  GET /api/v1/pegkeeper/status - PegKeeper status")

	// Start server
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// corsMiddleware adds CORS headers to all responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
