package monitor

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var globalMonitorService *Service

// SetService sets the global monitor service
func SetService(s *Service) {
	globalMonitorService = s
}

// HandleGetRecentTransactions returns recent transactions (GET /api/admin/transactions)
func HandleGetRecentTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	transactions := globalMonitorService.GetRecentTransactions(limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transactions": transactions,
		"count":        len(transactions),
	})
}

// HandleGetTransactionByHash returns transaction by hash (GET /api/admin/transactions/:hash)
func HandleGetTransactionByHash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract hash from path (simplified - in production use proper routing)
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		http.Error(w, "Hash parameter required", http.StatusBadRequest)
		return
	}

	tx := globalMonitorService.GetTransactionByHash(hash)
	if tx == nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

// HandleGetStats returns monitoring statistics (GET /api/admin/stats)
func HandleGetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := globalMonitorService.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// HandleWebSocket provides WebSocket connection for real-time transaction updates
// (GET /api/admin/transactions/ws)
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Subscribe to transaction events
	events := globalMonitorService.GetTxEvents()

	for {
		select {
		case event := <-events:
			if event == nil {
				return
			}

			// Send transaction event to client
			txInfo := &TxInfo{
				TxHash:  event.TxHash,
				Height:  event.Height,
				Success: event.Success,
				Code:    event.Code,
				Log:     event.Log,
				Time:    time.Now(),
			}

			for _, evt := range event.Events {
				txInfo.Events = append(txInfo.Events, evt.Type)
			}

			if err := conn.WriteJSON(txInfo); err != nil {
				return
			}
		}
	}
}

