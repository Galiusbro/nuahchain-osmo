package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// DBHealthResponse represents the database health check response
type DBHealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
}

// DBHealthChecker is an interface for database health checks
type DBHealthChecker interface {
	HealthCheck(ctx context.Context) error
}

var dbHealthChecker DBHealthChecker

// SetDBHealthChecker sets the database health checker instance
func SetDBHealthChecker(checker DBHealthChecker) {
	dbHealthChecker = checker
}

// handleDBHealth handles the database health check endpoint
func handleDBHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := DBHealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
	}

	statusCode := http.StatusOK

	if dbHealthChecker != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := dbHealthChecker.HealthCheck(ctx); err != nil {
			response.Status = "error"
			response.Error = err.Error()
			statusCode = http.StatusServiceUnavailable
		}
	} else {
		response.Status = "error"
		response.Error = "database health checker not configured"
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
